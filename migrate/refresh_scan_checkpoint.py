#!/usr/bin/env python3
"""
refresh_scan_checkpoint.py - Refresh t_tx_scan_info checkpoints in config SQL files.

Reads the config SQL file, extracts RPC endpoint and scan info records,
queries Solana RPC for the latest transaction signature per pda_account,
and rewrites the t_tx_scan_info INSERT statements with fresh until_tx_id and slot.

Usage:
    python3 refresh_scan_checkpoint.py --rpc-url "https://devnet.helius-rpc.com/?api-key=xxx"
    python3 refresh_scan_checkpoint.py --rpc-url "..." --mainnet-rpc-url "..." --config testnet/config.sql
    python3 refresh_scan_checkpoint.py --rpc-url "..." --dry-run
"""

import argparse
import json
import os
import re
import sys
import urllib.request

# ─── RPC URL construction (matches Go common/utils/rpc_pool.go) ──────────

def build_rpc_url(provider: str, endpoint: str, api_key: str) -> str:
    if not api_key:
        return endpoint
    if provider == "quicknode":
        return f"{endpoint}/{api_key}"
    elif provider == "helius":
        return f"{endpoint}/?api-key={api_key}"
    return endpoint


# ─── SQL parsing ─────────────────────────────────────────────────────────

def parse_rpc_endpoints(sql: str):
    """Extract RPC endpoints by scope, return dict {scope: url}."""
    pattern = re.compile(
        r"\(\s*'(env|mainnet)'\s*,\s*'(\w+)'\s*,\s*'([^']+)'\s*,\s*'([^']*)'\s*,",
        re.IGNORECASE,
    )
    endpoints = {}
    for m in pattern.finditer(sql):
        scope, provider, endpoint, api_key = m.group(1), m.group(2), m.group(3), m.group(4)
        url = build_rpc_url(provider, endpoint, api_key)
        endpoints[scope] = url
        print(f"  RPC [{scope}]: {provider} -> {url}")
    return endpoints


def parse_scan_info_records(sql: str):
    """Extract all t_tx_scan_info INSERT records."""
    records = []
    # Match INSERT INTO ... VALUES (...) blocks for t_tx_scan_info
    pattern = re.compile(
        r"INSERT\s+INTO\s+public\.t_tx_scan_info\s*"
        r"\([^)]+\)\s*VALUES\s*\(\s*"
        r"'([^']+)'\s*,\s*"   # service
        r"'([^']+)'\s*,\s*"   # sub_service
        r"'([^']+)'\s*,\s*"   # native_account
        r"'([^']+)'\s*,\s*"   # pda_account
        r"'([^']*)'\s*,\s*"   # until_tx_id
        r"'([^']*)'\s*,\s*"   # before_tx_id
        r"'([^']+)'\s*,",     # slot
        re.IGNORECASE | re.DOTALL,
    )
    for m in pattern.finditer(sql):
        records.append({
            "service": m.group(1),
            "sub_service": m.group(2),
            "native_account": m.group(3),
            "pda_account": m.group(4),
            "until_tx_id": m.group(5),
            "before_tx_id": m.group(6),
            "slot": m.group(7),
        })
    return records


# ─── Solana RPC ──────────────────────────────────────────────────────────

def get_latest_signature(rpc_url: str, pda_account: str):
    """Call getSignaturesForAddress and return (signature, slot) of the latest tx."""
    payload = json.dumps({
        "jsonrpc": "2.0",
        "id": 1,
        "method": "getSignaturesForAddress",
        "params": [
            pda_account,
            {"commitment": "finalized", "limit": 1},
        ],
    }).encode()

    req = urllib.request.Request(
        rpc_url,
        data=payload,
        headers={"Content-Type": "application/json"},
    )
    try:
        with urllib.request.urlopen(req, timeout=30) as resp:
            data = json.loads(resp.read())
    except Exception as e:
        print(f"    RPC error for {pda_account}: {e}")
        return None

    results = data.get("result", [])
    if not results:
        print(f"    No transactions found for {pda_account}")
        return None

    latest = results[0]
    return latest["signature"], latest["slot"]


# ─── SQL rewriting ───────────────────────────────────────────────────────

def rebuild_scan_info_sql(records: list[dict]) -> str:
    """Generate the replacement SQL block for t_tx_scan_info."""
    lines = ["delete from t_tx_scan_info;\n"]
    for r in records:
        lines.append(
            f"INSERT INTO public.t_tx_scan_info\n"
            f"    (service, sub_service, native_account, pda_account, until_tx_id, before_tx_id, slot, created_at, updated_at)\n"
            f"VALUES\n"
            f"    ('{r['service']}', '{r['sub_service']}', '{r['native_account']}', "
            f"'{r['pda_account']}', '{r['until_tx_id']}', '', '{r['slot']}', NOW(),NOW());\n"
        )
    return "\n".join(lines)


def replace_scan_info_block(sql: str, new_block: str) -> str:
    """Replace the t_tx_scan_info section in the SQL file."""
    # Find the range: from "delete from t_tx_scan_info;" to the last INSERT INTO t_tx_scan_info ... ;
    start_pattern = re.compile(r"^delete\s+from\s+t_tx_scan_info\s*;", re.IGNORECASE | re.MULTILINE)
    start_match = start_pattern.search(sql)
    if not start_match:
        print("  ERROR: Could not find 'delete from t_tx_scan_info' in SQL file")
        return sql

    # Find the last INSERT INTO t_tx_scan_info ... NOW(),NOW()); ending
    # Use .*? with DOTALL to handle NOW() nested parens and quoted column names
    end_pattern = re.compile(
        r"INSERT\s+INTO\s+public\.t_tx_scan_info\s*\(.*?\)\s*VALUES\s*\(.*?NOW\(\)\s*\)\s*;",
        re.IGNORECASE | re.DOTALL,
    )
    last_end = start_match.start()
    for m in end_pattern.finditer(sql, start_match.start()):
        last_end = m.end()

    if last_end <= start_match.start():
        print("  ERROR: Could not find t_tx_scan_info INSERT statements")
        return sql

    return sql[:start_match.start()] + new_block + sql[last_end:]


# ─── Main ────────────────────────────────────────────────────────────────

def main():
    parser = argparse.ArgumentParser(description="Refresh t_tx_scan_info checkpoints with latest on-chain signatures")
    parser.add_argument("--config", default="testnet/config.sql", help="Config SQL file path (default: testnet/config.sql)")
    parser.add_argument("--dry-run", action="store_true", help="Preview changes without writing to file")
    parser.add_argument("--rpc-url", required=True, help="Env RPC URL (e.g. https://devnet.helius-rpc.com/?api-key=xxx)")
    parser.add_argument("--mainnet-rpc-url", help="Mainnet RPC URL (required if config has market buy token scans)")
    args = parser.parse_args()

    config_path = os.path.join(os.path.dirname(__file__), args.config)
    if not os.path.exists(config_path):
        print(f"ERROR: Config file not found: {config_path}")
        sys.exit(1)

    print(f"Reading {args.config} ...")
    with open(config_path, "r") as f:
        sql = f.read()

    # 1. RPC endpoints from command line args
    endpoints = {"env": args.rpc_url}
    print(f"  RPC [env]: {args.rpc_url}")
    if args.mainnet_rpc_url:
        endpoints["mainnet"] = args.mainnet_rpc_url
        print(f"  RPC [mainnet]: {args.mainnet_rpc_url}")

    # 2. Parse existing scan info records
    records = parse_scan_info_records(sql)
    if not records:
        print("ERROR: No t_tx_scan_info records found in config")
        sys.exit(1)
    print(f"  Found {len(records)} scan info records")

    # sub_service that requires mainnet RPC (e.g. DEX purchase tracking)
    mainnet_sub_services = {"market buy token"}

    # 3. Query latest signature for each unique (pda_account, rpc_url) pair
    pda_cache = {}
    for r in records:
        pda = r["pda_account"]
        use_mainnet = r["sub_service"] in mainnet_sub_services
        rpc_url = endpoints.get("mainnet") if use_mainnet else endpoints["env"]

        if not rpc_url:
            print(f"  SKIP {pda} ({r['service']}/{r['sub_service']}): no {'mainnet' if use_mainnet else 'env'} RPC configured")
            continue

        cache_key = (pda, rpc_url)
        if cache_key not in pda_cache:
            label = f"{r['service']}/{r['sub_service']}"
            if use_mainnet:
                label += " [mainnet]"
            print(f"  Querying {pda} ({label}) ...")
            result = get_latest_signature(rpc_url, pda)
            if result:
                sig, slot = result
                pda_cache[cache_key] = (sig, slot)
                print(f"    -> sig={sig[:20]}...  slot={slot}")
            else:
                print(f"    -> SKIP (keeping old values)")

    # 4. Update records with new signatures
    updated = 0
    for r in records:
        pda = r["pda_account"]
        use_mainnet = r["sub_service"] in mainnet_sub_services
        rpc_url = endpoints.get("mainnet") if use_mainnet else endpoints["env"]
        cache_key = (pda, rpc_url)
        if cache_key in pda_cache:
            sig, slot = pda_cache[cache_key]
            r["until_tx_id"] = sig
            r["slot"] = str(slot)
            updated += 1

    print(f"\n  Updated {updated}/{len(records)} records")

    # 5. Rebuild and replace SQL block
    new_block = rebuild_scan_info_sql(records)

    if args.dry_run:
        print("\n--- DRY RUN: New t_tx_scan_info block ---")
        print(new_block)
        return

    new_sql = replace_scan_info_block(sql, new_block)
    with open(config_path, "w") as f:
        f.write(new_sql)
    print(f"\n  Written to {args.config}")


if __name__ == "__main__":
    main()
