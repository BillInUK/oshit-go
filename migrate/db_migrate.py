#!/usr/bin/env python3
"""
meme_db → oshit_db 数据库迁移脚本

旧库 (meme_db): 列名大多为 CamelCase（少数表例外），表名 snake_case
新库 (oshit_db): 表名和列名均为 snake_case

约束:
  - 旧库多出的列不导入（只迁移新库有的列）
  - 旧库没有但新库有的列，使用新库 DEFAULT 或跳过
  - 大表预演只导 1000 条
  - 新工程独有表 / 手动导入表跳过

使用方式:
    python3 db_migrate.py                              # 预演迁移
    python3 db_migrate.py --dry-run                    # 仅打印，不执行
    python3 db_migrate.py --tables t_native_account_info  # 只迁移指定表
    python3 db_migrate.py --rebuild-ddl                # 重建 DDL 后迁移
"""

import argparse
import logging
import os
from dataclasses import dataclass, field
from typing import Optional

import json
from decimal import Decimal
from datetime import datetime, date

import psycopg2
import psycopg2.extras
from psycopg2.extensions import adapt, register_adapter

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] %(message)s",
    datefmt="%H:%M:%S",
)
log = logging.getLogger("migrate")


# ─── psycopg2: 让 dict/list 自动序列化为 JSON 字符串 ─────────────
def _adapt_dict(d):
    return adapt(json.dumps(d))

def _adapt_list_as_json(lst):
    """仅当 list 内含 dict 时走 JSON，否则走默认 adapt（支持 Postgres ARRAY）"""
    if lst and isinstance(lst[0], dict):
        return adapt(json.dumps(lst))
    return adapt(lst)

register_adapter(dict, _adapt_dict)

# psycopg2 读取自定义类型（如 ulid）时可能返回非标准 Python 对象
# 预处理每行数据，将未知类型转为字符串以确保可插入
_SAFE_TYPES = (str, int, float, bool, type(None), datetime, date, Decimal, bytes)

def _safe_row(row):
    """将行中的非标准类型转为字符串，确保 psycopg2 能序列化"""
    result = []
    for val in row:
        if isinstance(val, _SAFE_TYPES):
            result.append(val)
        elif isinstance(val, dict):
            result.append(json.dumps(val))
        elif isinstance(val, (list, tuple)):
            # 数组：递归处理元素
            result.append([str(v) if not isinstance(v, _SAFE_TYPES) else v for v in val])
        else:
            # ULID 等自定义类型 → 转为字符串
            result.append(str(val))
    return tuple(result)


# ─── 数据库连接配置 ──────────────────────────────────────────────

SOURCE_DB = dict(
    host="172.31.3.251",
    port=5432,
    user="postgres",
    password="xxxxx",
    dbname="meme_db",
)

TARGET_DB = dict(
    host="172.31.3.251",
    port=5432,
    user="postgres",
    password="xxxxx",
    dbname="oshit_db",
)

# ─── 表迁移定义 ───────────────────────────────────────────────────

@dataclass
class TableMigration:
    new_table: str                              # 新表名 (oshit_db)
    old_table: str                              # 旧表名 (meme_db)
    column_map: dict                            # {new_col: old_col}，old_col 为旧库实际列名
    limit: Optional[int] = None                 # 行数限制
    where: Optional[str] = None                 # WHERE 条件（用旧库列名）
    defaults: dict = field(default_factory=dict) # 新表需要填充的默认值 {new_col: value}
    on_conflict_skip: bool = False              # 唯一键冲突时跳过（ON CONFLICT DO NOTHING）
    note: str = ""


# ─── 时间戳简写 ──────────────────────────────────────────────────
# 大多数旧表用 CreateTime/UpdateTime
CT = {"created_at": "CreateTime", "updated_at": "UpdateTime"}
CT_ID = {"record_id": "RecordId", **CT}
# 少数旧表用 CreatedAt/UpdatedAt
CA = {"created_at": "CreatedAt", "updated_at": "UpdatedAt"}
CA_ID = {"record_id": "RecordId", **CA}

# ─── TTL 过滤条件（用旧库列名，迁移时丢弃超过 TTL 的数据）────────
TTL_2D_CT = """"CreateTime" >= NOW() - INTERVAL '2 days'"""
TTL_7D_CT = """"CreateTime" >= NOW() - INTERVAL '7 days'"""
TTL_1M_CT = """"CreateTime" >= NOW() - INTERVAL '1 month'"""
TTL_2D_CA = """"CreatedAt" >= NOW() - INTERVAL '2 days'"""


# ═══════════════════════════════════════════════════════════════════
# 所有表迁移映射（基于旧库实际列名）
# ═══════════════════════════════════════════════════════════════════

MIGRATIONS: list[TableMigration] = [

    # ─── base_structure.sql ───────────────────────────────────────

    TableMigration(
        new_table="t_system_config",
        old_table="t_system_config",
        # 旧列: Env, CreateTime, UpdateTime
        column_map={"env": "Env", **CT},
    ),

    # 配置表，跳过，后续会手动迁移
    # TableMigration(
    #     new_table="t_aws_config",
    #     old_table="t_aws_config",
    #     # 旧列: AccessKeyId, SecretAccessKey, Region (无时间戳列)
    #     column_map={
    #         "access_key_id": "AccessKeyId",
    #         "secret_access_key": "SecretAccessKey",
    #         "region": "Region",
    #     },
    # ),

    # 配置表，跳过，后续会手动迁移
    # TableMigration(
    #     new_table="t_chain_config",
    #     old_table="t_chain_config",
    #     # 旧列: Chain, RpcUrl, WssUrl, Decimal, Symbol, CreateTime, UpdateTime
    #     column_map={
    #         "chain_name": "Chain","decimals": "Decimal", "symbol": "Symbol",
    #         **CT,
    #     },
    # ),

    # 配置表，跳过，后续会手动迁移
    # TableMigration(
    #     new_table="t_token_config",
    #     old_table="t_sol_token_config",
    #     # 旧列: Brand, TokenSymbol, Decimal, TokenMintAccount, + 更多旧库独有列
    #     column_map={
    #         "token_name": "Brand", "token_symbol": "TokenSymbol",
    #         "decimals": "Decimal", "mint": "TokenMintAccount",
    #         **CT,
    #     },
    # ),

    # 配置表，跳过，后续会手动迁移
    # TableMigration(
    #     new_table="t_fee_tolerance",
    #     old_table="t_fee_tolerance",
    #     # 旧列: Brand, TokenSymbol, MaxLessRate (无时间戳)
    #     column_map={"max_less_rate": "MaxLessRate"},
    # ),

    TableMigration(
        new_table="t_fee_statistics",
        old_table="t_sol_fee_statistics",
        # 旧列: RecordId, Slot, TransactionIndex, BlockHash, TransactionId, ComputeUnitPrice, ComputeUnitLimit, UnitsConsumed, Fee, CreateTime, UpdateTime
        column_map={
            "slot": "Slot", "tx_index": "TransactionIndex",
            "block_hash": "BlockHash", "tx_id": "TransactionId",
            "price": "ComputeUnitPrice", "unit_limit": "ComputeUnitLimit",
            "units_consumed": "UnitsConsumed", "fee": "Fee",
            **CT_ID,
        },
        where=TTL_2D_CT,
        limit=1000,
        note="TTL 2天，分区表",
    ),

    TableMigration(
        new_table="t_qn_fee",
        old_table="t_sol_qn_fee",
        # 旧列: Id, Slot, LowAvg, MediumAvg, HighAvg, CreateTime, UpdateTime
        column_map={
            "id": "Id", "slot": "Slot",
            "low_avg": "LowAvg", "medium_avg": "MediumAvg", "high_avg": "HighAvg",
            **CT,
        },
        where=TTL_2D_CT,
        note="TTL 2天，分区表",
    ),

    TableMigration(
        new_table="t_native_account_info",
        old_table="t_sol_native_account_info",
        # 旧列: RecordId, Brand, TokenSymbol, NativeAccount, TokenAccount, InviteCode, CreateTime, UpdateTime
        column_map={
            "native_account": "NativeAccount", "token_account": "TokenAccount",
            "invite_code": "InviteCode",
            **CT_ID,
        },
    ),

    TableMigration(
        new_table="t_invite_relation",
        old_table="t_sol_determine_invite_record",
        # 旧表 t_invite_relation 是空表，实际数据在 t_sol_determine_invite_record
        # 旧列: RecordId, Brand, TokenSymbol, InviterTokenAccount, InviterNativeAccount,
        #        InviteeTokenAccount, InviteeNativeAccount, TransferTxId, InviteChannel,
        #        CreateTime, UpdateTime, Level
        # 新表有 UNIQUE(invitee) 和 UNIQUE(tx_id)，冲突时跳过
        column_map={
            "inviter": "InviterTokenAccount", "invitee": "InviteeNativeAccount",
            "channel": "InviteChannel", "inviter_level": "Level",
            "tx_id": "TransferTxId",
            **CT_ID,
        },
        on_conflict_skip=True,
        note="唯一键(invitee, tx_id)冲突时跳过",
    ),

    TableMigration(
        new_table="t_hacker_account",
        old_table="t_reward_hacker",
        # 旧列: RecordId, NativeAccount, CreateTime, UpdateTime
        column_map={"native_account": "NativeAccount", **CT_ID},
    ),

    TableMigration(
        new_table="t_exclude_account",
        old_table="t_reward_exclude",
        # 旧列: Chain, Account, Brand, Remark, CreateTime, UpdateTime
        # 注意: 旧库用 Account 而非 NativeAccount
        column_map={
            "native_account": "Account", "remark": "Remark",
            **CT,
        },
    ),

    # ─── reward_structure.sql ─────────────────────────────────────

    # 配置表，跳过，后续会手动迁移
    # TableMigration(
    #     new_table="t_level_dist",
    #     old_table="t_sol_transfer_reward_distribution",
    #     # 旧列: Brand, TokenSymbol, Level
    #     column_map={"dist_level": "Level"},
    # ),

    # 配置表，跳过，后续会手动迁移
    # TableMigration(
    #     new_table="t_level_ratio",
    #     old_table="t_sol_transfer_reward_claim",
    #     # 旧列: Brand, TokenSymbol, Level, ClaimRatio
    #     column_map={"dist_level": "Level", "ratio": "ClaimRatio"},
    # ),

    # 配置表，跳过，后续会手动迁移
    # TableMigration(
    #     new_table="t_discount_rate",
    #     old_table="t_reward_discount_rate",
    #     # 旧列: RecordId, Rate, CreateTime (无 UpdateTime)
    #     # 注意: 新表只有 record_id, rate 两列，无时间戳
    #     column_map={"record_id": "RecordId", "rate": "Rate"},
    # ),

    # 配置表，跳过，后续会手动配置
    # TableMigration(
    #     new_table="t_take_token_config",
    #     old_table="t_sol_official_give_token_reward_rule",
    #     # 旧列: RecordId, InviteCode, Brand, TokenSymbol, Decimal, TokenMintAccount,
    #     #        RewardTokenAccount, RewardNativeAccount, DexNativeAccount, Amount,
    #     #        InviteAmount, DexFeeRate, MaxDexFee, Interval, Default, RewardInviter,
    #     #        DetermineInvite, CreateTime, UpdateTime
    #     column_map={
    #         "invite_code": "InviteCode", "reward_account": "RewardNativeAccount",
    #         "cost_account": "DexNativeAccount", "amount": "Amount",
    #         "invite_amount": "InviteAmount", "cost_fee_rate": "DexFeeRate",
    #         "max_cost_fee": "MaxDexFee", "is_default": "Default",
    #         "reward_inviter": "RewardInviter", "invited": "DetermineInvite",
    #         **CT_ID,
    #     },
    #     where='"Default" = true',
    #     note="只导出 Default=true 的默认规则",
    # ),

    TableMigration(
        new_table="t_take_token_record",
        old_table="t_sol_official_give_token_record",
        # 旧列: RecordId, Brand, TokenSymbol, TokenMintAccount, RewardTokenAccount,
        #        RewardNativeAccount, ReceiptTokenAccount, ReceiptNativeAccount,
        #        DexNativeAccount, RewardTxId, Amount, DexFee, UseInviteCode, InviteCode,
        #        CreateTime, UpdateTime, State, RefBlockHash, LastValidBlockHeight, DetermineInvite
        column_map={
            "reward_account": "RewardNativeAccount", "receipt_account": "ReceiptNativeAccount",
            "cost_account": "DexNativeAccount", "tx_id": "RewardTxId",
            "amount": "Amount", "cost_fee": "DexFee",
            "use_invite_code": "UseInviteCode", "invite_code": "InviteCode",
            "tx_state": "State", "invited": "DetermineInvite",
            **CT_ID,
        },
        where=TTL_1M_CT,
        limit=1000,
        note="TTL 1月，分区表",
    ),

    TableMigration(
        new_table="t_daily_claim_stats",
        old_table="t_daily_claim_stats",
        # 旧列: RecordId, NativeAccount, TakeShitDate, TakeShitCount, NeedLottery,
        #        LastTakeTime, TotalLottery, TotalTake, CreateTime, UpdateTime, LotteryCount
        column_map={
            "native_account": "NativeAccount", "take_date": "TakeShitDate",
            "take_count": "TakeShitCount", "need_lottery": "NeedLottery",
            "last_take_time": "LastTakeTime",
            "total_lottery": "TotalLottery", "total_take": "TotalTake",
            "lottery_count": "LotteryCount",
            **CT_ID,
        },
        where=TTL_1M_CT,
        note="TTL 1月，按 take_date 分区",
    ),

    TableMigration(
        new_table="t_lottery_reward",
        old_table="t_reward_lottery",
        # 旧列: RecordId, NativeAccount, RewardAmount, RewardType, State, Pending, Day, CreateTime, UpdateTime
        column_map={
            "native_account": "NativeAccount", "reward_amount": "RewardAmount",
            "reward_type": "RewardType", "reward_state": "State",
            "pending": "Pending", "reward_day": "Day",
            **CT_ID,
        },
        where=TTL_1M_CT,
        limit=1000,
        note="TTL 1月，按 reward_day 分区",
    ),

    TableMigration(
        new_table="t_lottery_claim",
        old_table="t_reward_lottery_claim_record",
        # 旧列: RecordId, RewardIds, TxId, RefBlockHash, LastValidBlockHeight, State, CreateTime, UpdateTime
        column_map={
            "reward_ids": "RewardIds", "tx_id": "TxId", "tx_state": "State",
            **CT_ID,
        },
        where=TTL_1M_CT,
        limit=1000,
        note="TTL 1月，分区表",
    ),

    # 配置表，跳过，后续会手动迁移
    # TableMigration(
    #     new_table="t_give_token_config",
    #     old_table="t_sol_transfer_token_reward_rule",
    #     # 旧列: Brand, TokenSymbol, Decimal, TokenMintAccount, RewardTokenAccount,
    #     #        RewardNativeAccount, AccountExistSlot, AccountExistBufferSlot,
    #     #        RewardRate, MaxRewardPerTx, CreateTime, UpdateTime
    #     # 注意: 旧库无 DexNativeAccount(cost_account) 和 RewardValidAddressRate(valid_rate)
    #     # cost_account 和 valid_rate 是 NOT NULL，需要提供默认值
    #     column_map={
    #         "reward_account": "RewardNativeAccount",
    #         "reward_rate": "RewardRate",
    #         "max_valid_reward": "MaxRewardPerTx",
    #         **CT,
    #     },
    #     defaults={"cost_account": "6MeXfYMhXpQSz3fqHtEa72V1XgKG7WGsECDy9jEv9e2K", "valid_rate": 300},
    #     note="cost_account 测试环境=6MeXfY..，生产=4DZ3ry..；valid_rate 固定 300",
    # ),

    TableMigration(
        new_table="t_give_token_record",
        old_table="t_sol_transfer_checked_record",
        # 旧列: RecordId, TokenMintAccount, FromTokenAccount, FromNativeAccount,
        #        ReceiptTokenAccount, ReceiptNativeAccount, OwnerNativeAccount,
        #        TransferTxId, InstructionIndex, TransferAmount, CreateTime, UpdateTime,
        #        State, RefBlockHash, LastValidBlockHeight
        column_map={
            "from_account": "FromNativeAccount", "receipt_account": "ReceiptNativeAccount",
            "tx_id": "TransferTxId", "amount": "TransferAmount", "tx_state": "State",
            **CT_ID,
        },
        where=TTL_1M_CT,
        limit=1000,
        note="TTL 1月，分区表",
    ),

    # 配置表，跳过，后续会手动迁移
    # TableMigration(
    #     new_table="t_reward_code_config",
    #     old_table="t_reward_code_rule",
    #     # 旧列: RecordId, Brand, TokenSymbol, Decimal, TokenMintAccount,
    #     #        RewardTokenAccount, RewardNativeAccount, DexNativeAccount,
    #     #        DexFeeRate, MaxDexFee, CreateTime, UpdateTime
    #     column_map={
    #         "reward_account": "RewardNativeAccount", "cost_account": "DexNativeAccount",
    #         **CT_ID,
    #     },
    # ),

    TableMigration(
        new_table="t_reward_code",
        old_table="t_reward_code",
        # 旧列: RecordId, RewardCode, RewardAmount, TxId, RefBlockHash,
        #        LastValidBlockHeight, State, CreateTime, UpdateTime, ExpireTime, NativeAccount
        # 注意: 旧库 State→reward_state, 无 TxState 列(tx_state), ExpireTime→expired_at
        column_map={
            "reward_code": "RewardCode", "reward_amount": "RewardAmount",
            "reward_state": "State", "native_account": "NativeAccount",
            "tx_id": "TxId", "expired_at": "ExpireTime",
            **CT_ID,
        },
        where=TTL_1M_CT,
        note="TTL 1月，分区表",
    ),

    # 配置表，跳过，后续会手动迁移
    # TableMigration(
    #     new_table="t_reward_code_fee",
    #     old_table="t_reward_code_fee",
    #     # 旧列: Amount, CostFeeRate, CreateTime, UpdateTime
    #     column_map={"amount": "Amount", "fee_rate": "CostFeeRate", **CT},
    # ),

    # 配置表，跳过，后续会手动迁移
    # TableMigration(
    #     new_table="t_campaign_quote_config",
    #     old_table="t_sol_campaign_exchange_score_rule",
    #     # 旧列: RecordId, Decimal, TokenMintAccount, RewardTokenAccount,
    #     #        RewardNativeAccount, Rate, DexNativeAccount, CostRate, CreateTime, UpdateTime
    #     column_map={
    #         "reward_account": "RewardNativeAccount", "cost_account": "DexNativeAccount",
    #         "quote_rate": "Rate", "cost_rate": "CostRate",
    #         **CT_ID,
    #     },
    # ),


    TableMigration(
        new_table="t_campaign_quote_limit",
        old_table="t_global_daily_exchange_limit",
        # 旧列: id, daily_limit, quota_date, created_at, updated_at, session (已经是 snake_case)
        column_map={
            "id": "id", "daily_limit": "daily_limit", "quota_date": "quota_date",
            "session": "session",
            "created_at": "created_at", "updated_at": "updated_at",
        },
    ),

    TableMigration(
        new_table="t_user_daily_quota",
        old_table="t_user_daily_exchange_quota",
        # 旧列: id, user_id, quota_date, max_quota, frozen_quota, available_quota, created_at, updated_at (snake_case)
        column_map={
            "id": "id", "user_id": "user_id", "quota_date": "quota_date",
            "max_quota": "max_quota", "frozen_quota": "frozen_quota",
            "available_quota": "available_quota",
            "created_at": "created_at", "updated_at": "updated_at",
        },
    ),

    TableMigration(
        new_table="t_campaign_quote_record",
        old_table="t_sol_exchange_campaign_score_to_token_record",
        # 旧列: RecordId, TokenMintAccount, RewardTokenAccount, RewardNativeAccount,
        #        ReceiptTokenAccount, ReceiptNativeAccount, Provider, UserId,
        #        ExchangeTxId, ScoreFlowId, ScoreTxId, Amount, Score, State,
        #        RefBlockHash, LastValidBlockHeight, CreateTime, UpdateTime
        # 注意: 旧库无 Session, UserQuotaDate 列
        column_map={
            "reward_account": "RewardNativeAccount", "receipt_account": "ReceiptNativeAccount",
            "provider": "Provider", "user_id": "UserId",
            "tx_id": "ExchangeTxId", "score_flow_id": "ScoreFlowId",
            "score_tx_id": "ScoreTxId", "amount": "Amount",
            "score": "Score", "quote_state": "State",
            **CT_ID,
        },
        defaults={"session": 0, "user_quota_date": "2024-01-01"},
        where=TTL_1M_CT,
        limit=1000,
        note="TTL 1月，分区表；旧库无 session/user_quota_date，填默认值",
    ),

    # ─── pos_structure.sql ────────────────────────────────────────

    # 配置表，跳过，后续会手动迁移
    # TableMigration(
    #     new_table="t_pos_reward_config",
    #     old_table="t_sol_pos_reward_rule",
    #     # 旧列: RecordId, Brand, TokenSymbol, Decimal, TokenMintAccount, RewardTokenAccount,
    #     #        RewardNativeAccount, DexNativeAccount, DexFeeRate, MaxDexFee,
    #     #        QuoteTokenAmount, CreateTime, UpdateTime
    #     column_map={
    #         "reward_account": "RewardNativeAccount", "cost_account": "DexNativeAccount",
    #         "cost_fee_rate": "DexFeeRate", "max_cost_fee": "MaxDexFee",
    #         "quote_token_amount": "QuoteTokenAmount",
    #         **CT_ID,
    #     },
    # ),

    # 配置表，跳过，后续会手动迁移
    # TableMigration(
    #     new_table="t_pos_star_level_rule",
    #     old_table="t_sol_pos_star_level_rule",
    #     # 旧列: RecordId, Amount, GroupAmount, StarLevel, Rate, CreateTime, UpdateTime
    #     column_map={
    #         "amount": "Amount", "group_amount": "GroupAmount",
    #         "star_level": "StarLevel", "rate": "Rate",
    #         **CT_ID,
    #     },
    # ),


    TableMigration(
        new_table="t_pos_star_whitelist",
        old_table="t_sol_pos_star_level_config",
        # 旧列: RecordId, NativeAccount, StarLevel, Rate, CreateTime, UpdateTime
        column_map={
            "native_account": "NativeAccount", "star_level": "StarLevel", "rate": "Rate",
            **CT_ID,
        },
    ),

    # 配置表，跳过，后续会手动迁移
    # TableMigration(
    #     new_table="t_pos_mission_config",
    #     old_table="t_sol_pos_mission_config",
    #     # 旧列: RecordId, RewardType, Starred, Rate, CreateTime, UpdateTime
    #     column_map={
    #         "reward_type": "RewardType", "starred": "Starred", "rate": "Rate",
    #         **CT_ID,
    #     },
    # ),

    TableMigration(
        new_table="t_pos_snap_shot",
        old_table="t_sol_pos_snap_shot",
        # 旧列: RecordId, NativeAccount, Amount, StarLevel, Rate, RangeBase, Day, CreateTime, UpdateTime
        column_map={
            "native_account": "NativeAccount", "amount": "Amount",
            "star_level": "StarLevel", "rate": "Rate",
            "range_base": "RangeBase", "snap_day": "Day",
            **CT_ID,
        },
        where=TTL_1M_CT,
        limit=1000,
        note="TTL 1月，按 snap_day 分区",
    ),

    TableMigration(
        new_table="t_pos_reward",
        old_table="t_sol_pos_reward",
        # 旧列: RecordId, GroupId, NativeAccount, StarLevel, Base, Rate, RewardAmount,
        #        RewardType, State, Starred, Pending, Day, CreateTime, UpdateTime
        column_map={
            "group_id": "GroupId", "native_account": "NativeAccount",
            "star_level": "StarLevel", "base": "Base", "rate": "Rate",
            "reward_amount": "RewardAmount", "reward_type": "RewardType",
            "reward_state": "State", "starred": "Starred",
            "pending": "Pending", "snap_day": "Day",
            **CT_ID,
        },
        where=TTL_1M_CT,
        limit=1000,
        note="TTL 1月，按 snap_day 分区",
    ),

    TableMigration(
        new_table="t_pos_reward_claim",
        old_table="t_sol_pos_reward_claim_record",
        # 旧列: RecordId, RewardIds, TxId, RefBlockHash, LastValidBlockHeight, State, CreateTime, UpdateTime
        column_map={
            "reward_ids": "RewardIds", "tx_id": "TxId", "tx_state": "State",
            **CT_ID,
        },
        where=TTL_1M_CT,
        limit=1000,
        note="TTL 1月，分区表",
    ),

    # ─── stake_structure.sql ──────────────────────────────────────

    # 配置表，跳过，后续会手动迁移
    # TableMigration(
    #     new_table="t_stake_amm_config",
    #     old_table="t_stake_amm_config",
    #     # 旧列: QuoteToken, PublicKey (无时间戳)
    #     column_map={"quote_token": "QuoteToken", "public_key": "PublicKey"},
    # ),

    # 配置表，跳过，后续会手动迁移
    # TableMigration(
    #     new_table="t_stake_token_pool",
    #     old_table="t_stake_token_pool",
    #     # 旧列: Source, FromTokenAccount (无时间戳)
    #     column_map={"source_account": "Source", "from_token_account": "FromTokenAccount"},
    # ),

    # 配置表，跳过，后续会手动迁移
    # TableMigration(
    #     new_table="t_stake_fix_rate_config",
    #     old_table="t_sol_stake_fix_interest_config",
    #     # 旧列: RecordId, MinAmount, StakeType, FixRate, IndividualRate, CreateTime, UpdateTime
    #     column_map={
    #         "min_amount": "MinAmount", "stake_type": "StakeType",
    #         "fix_rate": "FixRate", "individual_rate": "IndividualRate",
    #         **CT_ID,
    #     },
    # ),

    # 配置表，跳过，后续会手动迁移
    # TableMigration(
    #     new_table="t_stake_invite_dist",
    #     old_table="t_sol_stake_invite_dist",
    #     # 旧列: Level
    #     column_map={"dist_level": "Level"},
    # ),

    # 配置表，跳过，后续会手动迁移
    # TableMigration(
    #     new_table="t_stake_invite_rate",
    #     old_table="t_sol_stake_invite_rate",
    #     # 旧列: Level, Rate
    #     column_map={"dist_level": "Level", "rate": "Rate"},
    # ),

    # 配置表，跳过，后续会手动迁移
    # TableMigration(
    #     new_table="t_stake_star_level_rule",
    #     old_table="t_sol_stake_star_level_rule",
    #     # 旧列: RecordId, Amount, GroupAmount, StarLevel, Rate, CreateTime, UpdateTime
    #     column_map={
    #         "amount": "Amount", "group_amount": "GroupAmount",
    #         "star_level": "StarLevel", "rate": "Rate",
    #         **CT_ID,
    #     },
    # ),

    TableMigration(
        new_table="t_stake_star_whitelist",
        old_table="t_sol_stake_star_level_config",
        # 旧列: RecordId, NativeAccount, StarLevel, Rate, CreateTime, UpdateTime
        column_map={
            "native_account": "NativeAccount", "star_level": "StarLevel", "rate": "Rate",
            **CT_ID,
        },
    ),

    # 配置表，跳过，后续会手动迁移
    # TableMigration(
    #     new_table="t_stake_reward_config",
    #     old_table="t_sol_stake_reward_rule",
    #     # 旧列: RecordId, Decimal, TokenMintAccount, StakeAdmin, ProgramId,
    #     #        FaucetTokenAccount, FaucetNativeAccount, RewardTokenAccount,
    #     #        RewardNativeAccount, DexNativeAccount, QuoteTokenAmount, DexFeeRate,
    #     #        CreateTime, UpdateTime
    #     column_map={
    #         "program_id": "ProgramId", "reward_account": "RewardNativeAccount",
    #         "cost_account": "DexNativeAccount", "quote_token_amount": "QuoteTokenAmount",
    #         "cost_fee_rate": "DexFeeRate",
    #         **CT_ID,
    #     },
    # ),

    TableMigration(
        new_table="t_stake_reward",
        old_table="t_sol_stake_reward",
        # 旧列: RecordId, GroupId, NativeAccount, StarLevel, Base, Rate, RewardAmount,
        #        RewardType, State, Starred, Pending, Day, CreateTime, UpdateTime, TxId
        # 注意: 旧库无 stake_type，默认 0
        column_map={
            "group_id": "GroupId", "native_account": "NativeAccount",
            "star_level": "StarLevel", "base": "Base", "rate": "Rate",
            "reward_amount": "RewardAmount", "reward_type": "RewardType",
            "reward_state": "State", "starred": "Starred",
            "tx_id": "TxId", "pending": "Pending", "snap_day": "Day",
            **CT_ID,
        },
        defaults={"stake_type": 0},
        where="""("CreateTime" >= NOW() - INTERVAL '1 month') OR ("RewardType" = 0 AND "State" = 0)""",
        limit=1000,
        note="TTL 1月，DELETE+VACUUM清理；保留未领取的固定利息(type=0,state=0)",
    ),

    TableMigration(
        new_table="t_stake_reward_claim",
        old_table="t_sol_stake_reward_claim_record",
        # 旧列: RecordId, RewardIds, TxId, RefBlockHash, LastValidBlockHeight, State, CreateTime, UpdateTime
        column_map={
            "reward_ids": "RewardIds", "tx_id": "TxId", "tx_state": "State",
            **CT_ID,
        },
        where=TTL_1M_CT,
        limit=1000,
        note="TTL 1月，DELETE+VACUUM清理",
    ),

    TableMigration(
        new_table="t_stake_snap_shot",
        old_table="t_sol_stake_snap_shot",
        # 旧列: RecordId, NativeAccount, Amount, StakeType, Day, CreateTime, UpdateTime
        column_map={
            "native_account": "NativeAccount", "amount": "Amount",
            "stake_type": "StakeType", "snap_day": "Day",
            **CT_ID,
        },
        where=TTL_1M_CT,
        limit=1000,
        note="TTL 1月，按 snap_day 分区",
    ),

    TableMigration(
        new_table="t_stake_record",
        old_table="t_stake_record",
        # 旧列: RecordId, Staker, StakeAmount, State, StakeTxHash, UsedTxIds, LockedTxIds, CreatedAt, UpdatedAt
        column_map={
            "staker": "Staker", "stake_amount": "StakeAmount",
            "status": "Status", "stake_tx_hash": "StakeTxHash",
            "used_tx_ids": "UsedTxIds", "locked_tx_ids": "LockedTxIds",
            **CA_ID,
        },
        limit=1000,
    ),

    TableMigration(
        new_table="t_stake_buy_token",
        old_table="t_stake_buy_token",
        # 旧列: RecordId, TxId, Slot, Source, Destination, Amount, Locked, LockedBy,
        #        LockedAt, StakedAmount, RemainingAmount, CreatedAt
        # 注意: 旧库无 Expired, ExpiredAt, UpdatedAt
        column_map={
            "tx_id": "TxId", "slot": "Slot",
            "from_account": "Source", "to_account": "Destination",
            "amount": "Amount", "locked": "Locked",
            "locked_by": "LockedBy", "locked_at": "LockedAt",
            "staked_amount": "StakedAmount", "remaining_amount": "RemainingAmount",
            "created_at": "CreatedAt",
        },
        where=TTL_2D_CA,
        limit=1000,
        note="TTL 2天，分区表",
    ),

    # 配置表，跳过，后续会手动迁移
    # TableMigration(
    #     new_table="t_stake_total_leader",
    #     old_table="t_stake_total_area_leader",
    #     # 旧列: RecordId, NativeAccount, Share, CreateTime, UpdateTime
    #     column_map={
    #         "native_account": "NativeAccount", "stake_share": "Share",
    #         **CT_ID,
    #     },
    # ),

    # 配置表，跳过，后续会手动迁移
    # TableMigration(
    #     new_table="t_stake_leader",
    #     old_table="t_stake_area_leader",
    #     # 旧列: RecordId, NativeAccount, Level, Share, Leader, CreateTime, UpdateTime
    #     column_map={
    #         "native_account": "NativeAccount", "leader_level": "Level",
    #         "up_leader": "Leader",
    #         **CT_ID,
    #     },
    # ),

    TableMigration(
        new_table="t_stake_leader_reward",
        old_table="t_stake_area_leader_reward",
        # 旧列: RecordId, NativeAccount, Staker, RewardType, BaseAmount, Rate,
        #        RewardAmount, State, Pending, TxId, CreateTime, UpdateTime
        column_map={
            "native_account": "NativeAccount", "staker": "Staker",
            "reward_type": "RewardType", "base_amount": "BaseAmount",
            "stake_share": "Rate", "reward_amount": "RewardAmount",
            "reward_state": "State", "pending": "Pending", "tx_id": "TxId",
            **CT_ID,
        },
        limit=1000,
    ),

    TableMigration(
        new_table="t_stake_leader_reward_claim",
        old_table="t_stake_area_leader_reward_claim_record",
        # 旧列: RecordId, NativeAccount, RewardIds, TxId, RefBlockHash, LastValidBlockHeight, State, CreateTime, UpdateTime
        column_map={
            "native_account": "NativeAccount", "reward_ids": "RewardIds",
            "tx_id": "TxId", "tx_state": "State",
            **CT_ID,
        },
        limit=1000,
    ),

    # ─── 直接平移（旧库已是 snake_case）───────────────────────────

    # TableMigration(
    #     new_table="t_stake_team_reward_deduction",
    #     old_table="t_stake_team_reward_deduction",
    #     # 旧列: native_account, total, deducted, remaining, created_at, updated_at (已 snake_case)
    #     column_map={
    #         "native_account": "native_account", "total": "total",
    #         "deducted": "deducted", "remaining": "remaining",
    #         "created_at": "created_at", "updated_at": "updated_at",
    #     },
    #     note="新旧一致，直接平移",
    # ),
    #
    # TableMigration(
    #     new_table="t_stake_team_reward_deduction_log",
    #     old_table="t_stake_team_reward_deduction_log",
    #     # 旧列: record_id, native_account, snap_day, reward_type, original, deduction, created_at, updated_at (已 snake_case)
    #     column_map={
    #         "record_id": "record_id", "native_account": "native_account",
    #         "snap_day": "snap_day", "reward_type": "reward_type",
    #         "original": "original", "deduction": "deduction",
    #         "created_at": "created_at", "updated_at": "updated_at",
    #     },
    #     note="新旧一致，直接平移",
    # ),

    # ─── 大表: t_fund_flow ────────────────────────────────────────

    # TableMigration(
    #     new_table="t_fund_flow",
    #     old_table="t_sol_fund_flow",
    #     # 旧列 (snake_case): record_id, brand, token_symbol, is_token, from_native_account,
    #     #    to_native_account, tx_id, direction, service_type, flow_type, decimals, amount,
    #     #    create_time, update_time
    #     # 注意: 旧库列名是 snake_case，from_native_account→from_account, to_native_account→to_account
    #     column_map={
    #         "record_id": "record_id",
    #         "is_token": "is_token", "from_account": "from_native_account",
    #         "to_account": "to_native_account", "tx_id": "tx_id",
    #         "direction": "direction", "service_type": "service_type",
    #         "flow_type": "flow_type", "decimals": "decimals", "amount": "amount",
    #         "created_at": "create_time", "updated_at": "update_time",
    #     },
    #     limit=1000,
    # ),
]


# ─── 多表合并: t_service_tx ← t_reward_tx + t_pos_tx ─────────────

@dataclass
class MergeTableMigration:
    new_table: str
    sources: list  # [(old_table, service_name, sub_service_default, column_map)]
    limit_per_source: Optional[int] = None
    where: Optional[str] = None


MERGE_MIGRATIONS: list[MergeTableMigration] = [
    MergeTableMigration(
        new_table="t_service_tx",
        sources=[
            (
                "t_reward_tx",
                "Reward",
                "TakeToken",
                # 旧列: RecordId, TxId, TxType, CreateTime, State
                # 注意: 旧库无 RetryCount/NextRetryTime/MaxRetries/UpdateTime
                # TxType 可能对应 sub_service，暂时用默认值
                {
                    "record_id": "RecordId", "tx_id": "TxId", "tx_state": "State",
                    "created_at": "CreateTime",
                },
            ),
            (
                "t_pos_tx",
                "Pos",
                "PosReward",
                {
                    "record_id": "RecordId", "tx_id": "TxId", "tx_state": "State",
                    "created_at": "CreateTime",
                },
            ),
        ],
        limit_per_source=500,
        where=""" "State" = 0 AND "CreateTime" >= NOW() - INTERVAL '7 days'""",
    ),
]


# ─── 直接平移表（旧库 snake_case，新库自动建表后原样复制）─────────

DIRECT_COPY_TABLES: list[str] = [
    "user_basic",
    "user_extend_info",
    "user_score",
    "score_ledger",
    "score_operation_log",
    "social_media_user_info",
    "user_invitation_enroll",
    "user_posts_task",
    "user_posts_task_record",
    "user_sharing_task",
    "user_sharing_task_event_record",
    "user_activity_message",
    "user_reward_record",
    "user_share_reward_record",
    "user_favorite_product",
    "user_game_item",
    "user_game_item_op_record",
    "login_log",
    "invitation_relation",
    "mini_game_item",
    "game_score",
    "game_score_operation_log",
    "audit_comment_template_multi_lang",
]


# ─── 跳过的表 ─────────────────────────────────────────────────────

SKIP_TABLES = {
    "t_mainnet_rpc_config":          "新工程新增表",
    "t_service_info":                "新工程新增表",
    "t_service_key":                 "需手动导入（合并 t_reward_key_config + t_pos_reward_key_config）",
    "t_tx_scan_info":                "需手动导入（合并 t_reward_scan_info + t_pos_scan_info）",
    "t_stake_leader_reward_config":  "新工程新增表",
    "t_reward_key_config":           "新工程内部表",
}


# ─── 核心迁移逻辑 ─────────────────────────────────────────────────

class Migrator:
    def __init__(self, dry_run: bool = False, tables: list[str] = None):
        self.dry_run = dry_run
        self.filter_tables = set(tables) if tables else None
        self.src_conn = None
        self.dst_conn = None
        self.stats = {"success": 0, "skipped": 0, "failed": 0, "rows_total": 0}
        self.skipped_tables: list[tuple[str, str]] = []   # [(table, reason)]
        self.failed_tables: list[tuple[str, str]] = []    # [(table, reason)]

    def connect(self):
        log.info("连接源数据库 (meme_db) ...")
        self.src_conn = psycopg2.connect(**SOURCE_DB)
        self.src_conn.set_session(readonly=True)
        log.info("连接目标数据库 (oshit_db) ...")
        self.dst_conn = psycopg2.connect(**TARGET_DB)
        log.info("数据库连接成功")

    def close(self):
        if self.src_conn:
            self.src_conn.close()
        if self.dst_conn:
            self.dst_conn.close()

    def migrate_table(self, m: TableMigration):
        if self.filter_tables and m.new_table not in self.filter_tables:
            return

        log.info(f"═══ {m.old_table} → {m.new_table} ═══")
        if m.note:
            log.info(f"  备注: {m.note}")

        # 构造 SELECT（用双引号包裹列名以支持 CamelCase 和 snake_case）
        old_cols = list(m.column_map.values())
        new_cols = list(m.column_map.keys())
        select_cols = ", ".join(f'"{c}"' for c in old_cols)
        query = f'SELECT {select_cols} FROM "{m.old_table}"'
        if m.where:
            query += f" WHERE {m.where}"
        if m.limit:
            query += f" LIMIT {m.limit}"

        log.info(f"  SQL: {query[:200]}{'...' if len(query) > 200 else ''}")

        if self.dry_run:
            log.info(f"  [DRY RUN] 跳过")
            self.stats["skipped"] += 1
            self.skipped_tables.append((m.new_table, "dry-run"))
            return

        # 读取源数据
        src_cur = self.src_conn.cursor()
        try:
            src_cur.execute(query)
        except Exception as e:
            log.error(f"  ✗ 查询失败: {e}")
            self.src_conn.rollback()
            self.stats["failed"] += 1
            self.failed_tables.append((m.new_table, f"查询失败: {e}"))
            return
        rows = src_cur.fetchall()
        src_cur.close()

        if not rows:
            log.info(f"  源表无数据，跳过")
            self.stats["skipped"] += 1
            self.skipped_tables.append((m.new_table, "源表无数据"))
            return

        # 组装 INSERT 列（包含默认值列）
        all_new_cols = list(new_cols)
        for def_col in m.defaults:
            if def_col not in all_new_cols:
                all_new_cols.append(def_col)

        placeholders = ", ".join(["%s"] * len(all_new_cols))
        insert_cols = ", ".join(all_new_cols)
        insert_sql = f"INSERT INTO {m.new_table} ({insert_cols}) VALUES ({placeholders})"
        if m.on_conflict_skip:
            insert_sql += " ON CONFLICT DO NOTHING"

        # 清空目标表并写入
        dst_cur = self.dst_conn.cursor()
        try:
            dst_cur.execute(f"DELETE FROM {m.new_table}")
            deleted = dst_cur.rowcount
            if deleted > 0:
                log.info(f"  清空目标表，删除 {deleted} 行")

            count = 0
            first_err = None
            for row in rows:
                values = list(_safe_row(row))
                for def_col in m.defaults:
                    if def_col not in new_cols:
                        values.append(m.defaults[def_col])
                try:
                    dst_cur.execute(insert_sql, values)
                    count += 1
                except Exception as e:
                    first_err = str(e).strip().split("\n")[0]
                    log.warning(f"  插入失败: {first_err}")
                    self.dst_conn.rollback()
                    dst_cur = self.dst_conn.cursor()
                    dst_cur.execute(f"DELETE FROM {m.new_table}")
                    break

            self.dst_conn.commit()
            if first_err:
                log.error(f"  ✗ 插入失败，已导入 {count} 行")
                if count == 0:
                    self.stats["failed"] += 1
                    self.failed_tables.append((m.new_table, first_err))
                else:
                    self.stats["success"] += 1
            else:
                log.info(f"  ✓ 导入 {count} 行")
                self.stats["success"] += 1
            self.stats["rows_total"] += count
        except Exception as e:
            log.error(f"  ✗ 写入失败: {e}")
            self.dst_conn.rollback()
            self.stats["failed"] += 1
            self.failed_tables.append((m.new_table, f"写入失败: {e}"))

    def migrate_merge_table(self, m: MergeTableMigration):
        if self.filter_tables and m.new_table not in self.filter_tables:
            return

        log.info(f"═══ 合并迁移 → {m.new_table} ═══")

        if self.dry_run:
            for old_table, svc, sub_svc, _ in m.sources:
                log.info(f"  [DRY RUN] {old_table} → service={svc}, sub_service={sub_svc}")
            self.stats["skipped"] += 1
            return

        dst_cur = self.dst_conn.cursor()
        dst_cur.execute(f"DELETE FROM {m.new_table}")
        deleted = dst_cur.rowcount
        if deleted > 0:
            log.info(f"  清空目标表，删除 {deleted} 行")

        total_count = 0
        for old_table, service_name, sub_service_default, column_map in m.sources:
            old_cols = list(column_map.values())
            new_cols = list(column_map.keys())
            select_cols = ", ".join(f'"{c}"' for c in old_cols)
            query = f'SELECT {select_cols} FROM "{old_table}"'
            if m.where:
                query += f" WHERE {m.where}"
            if m.limit_per_source:
                query += f" LIMIT {m.limit_per_source}"

            log.info(f"  源: {old_table} → service={service_name}")
            src_cur = self.src_conn.cursor()
            try:
                src_cur.execute(query)
            except Exception as e:
                log.warning(f"  查询 {old_table} 失败: {e}")
                self.src_conn.rollback()
                continue
            rows = src_cur.fetchall()
            src_cur.close()

            if not rows:
                log.info(f"  {old_table} 无数据")
                continue

            all_new_cols = new_cols + ["service", "sub_service"]
            placeholders = ", ".join(["%s"] * len(all_new_cols))
            insert_cols = ", ".join(all_new_cols)
            insert_sql = f"INSERT INTO {m.new_table} ({insert_cols}) VALUES ({placeholders})"

            count = 0
            for row in rows:
                values = list(row) + [service_name, sub_service_default]
                try:
                    dst_cur.execute(insert_sql, values)
                    count += 1
                except Exception as e:
                    log.warning(f"  插入失败: {e}")
                    self.dst_conn.rollback()
                    dst_cur = self.dst_conn.cursor()
                    continue

            log.info(f"  从 {old_table} 导入 {count} 行")
            total_count += count

        self.dst_conn.commit()
        log.info(f"  ✓ 合并共 {total_count} 行")
        self.stats["success"] += 1
        self.stats["rows_total"] += total_count

    def _generate_create_table_ddl(self, table_name: str) -> str:
        """从源库 information_schema 读取表结构，生成 CREATE TABLE DDL"""
        cur = self.src_conn.cursor()

        # 获取列信息
        cur.execute("""
            SELECT column_name, data_type, character_maximum_length,
                   numeric_precision, numeric_scale, is_nullable, column_default,
                   udt_name
            FROM information_schema.columns
            WHERE table_schema = 'public' AND table_name = %s
            ORDER BY ordinal_position
        """, (table_name,))
        columns = cur.fetchall()
        if not columns:
            return ""

        # 获取主键列
        cur.execute("""
            SELECT kcu.column_name
            FROM information_schema.table_constraints tc
            JOIN information_schema.key_column_usage kcu
                ON tc.constraint_name = kcu.constraint_name
                AND tc.table_schema = kcu.table_schema
            WHERE tc.table_schema = 'public'
                AND tc.table_name = %s
                AND tc.constraint_type = 'PRIMARY KEY'
            ORDER BY kcu.ordinal_position
        """, (table_name,))
        pk_cols = [r[0] for r in cur.fetchall()]
        cur.close()

        col_defs = []
        for col_name, data_type, max_len, num_prec, num_scale, nullable, default, udt_name in columns:
            # 将 SERIAL/BIGSERIAL sequence default 转换为 SERIAL 类型
            is_serial = default and "nextval(" in str(default)
            if is_serial and udt_name == "int8":
                type_str = "BIGSERIAL"
            elif is_serial and udt_name == "int4":
                type_str = "SERIAL"
            elif data_type == "character varying" and max_len:
                type_str = f"VARCHAR({max_len})"
            elif data_type == "numeric" and num_prec and num_scale is not None:
                type_str = f"NUMERIC({num_prec},{num_scale})"
            elif data_type == "ARRAY":
                type_str = udt_name.lstrip("_") + "[]"
            elif udt_name == "int4":
                type_str = "INTEGER"
            elif udt_name == "int8":
                type_str = "BIGINT"
            elif udt_name == "int2":
                type_str = "SMALLINT"
            elif udt_name == "bool":
                type_str = "BOOLEAN"
            elif udt_name == "float8":
                type_str = "DOUBLE PRECISION"
            elif udt_name == "float4":
                type_str = "REAL"
            elif udt_name == "timestamptz":
                type_str = "TIMESTAMP WITH TIME ZONE"
            elif udt_name == "timestamp":
                type_str = "TIMESTAMP WITHOUT TIME ZONE"
            else:
                type_str = data_type.upper()

            parts = [f'    "{col_name}" {type_str}']
            if nullable == "NO" and not is_serial:
                parts.append("NOT NULL")
            if default and not is_serial:
                parts.append(f"DEFAULT {default}")

            col_defs.append(" ".join(parts))

        if pk_cols:
            pk_str = ", ".join(f'"{c}"' for c in pk_cols)
            col_defs.append(f"    PRIMARY KEY ({pk_str})")

        ddl = f'DROP TABLE IF EXISTS "{table_name}";\n'
        ddl += f'CREATE TABLE "{table_name}" (\n'
        ddl += ",\n".join(col_defs)
        ddl += "\n);"
        return ddl

    def migrate_direct_copy(self, table_name: str):
        """直接平移: 从源库自动建表 + 全量复制数据"""
        if self.filter_tables and table_name not in self.filter_tables:
            return

        log.info(f"═══ 直接平移: {table_name} ═══")

        # 1. 生成 DDL
        ddl = self._generate_create_table_ddl(table_name)
        if not ddl:
            log.error(f"  ✗ 源库中找不到表 {table_name}")
            self.stats["failed"] += 1
            self.failed_tables.append((table_name, "源库中找不到表"))
            return

        if self.dry_run:
            log.info(f"  [DRY RUN] DDL:\n{ddl[:300]}...")
            self.stats["skipped"] += 1
            self.skipped_tables.append((table_name, "dry-run"))
            return

        # 2. 在目标库建表（用独立连接避免事务冲突）
        try:
            ddl_conn = psycopg2.connect(**TARGET_DB)
            ddl_conn.autocommit = True
            ddl_cur = ddl_conn.cursor()
            ddl_cur.execute(ddl)
            ddl_cur.close()
            ddl_conn.close()
            log.info(f"  目标表已创建/重建")
        except Exception as e:
            log.error(f"  ✗ 建表失败: {e}")
            self.stats["failed"] += 1
            self.failed_tables.append((table_name, f"建表失败: {e}"))
            return

        # 3. 获取列名
        src_cur = self.src_conn.cursor()
        src_cur.execute("""
            SELECT column_name FROM information_schema.columns
            WHERE table_schema = 'public' AND table_name = %s
            ORDER BY ordinal_position
        """, (table_name,))
        col_names = [r[0] for r in src_cur.fetchall()]

        # 排除 SERIAL 自增列的 default（插入时显式提供值）
        select_cols = ", ".join(f'"{c}"' for c in col_names)
        query = f'SELECT {select_cols} FROM "{table_name}"'

        # 4. 读取数据
        try:
            src_cur.execute(query)
        except Exception as e:
            log.error(f"  ✗ 查询失败: {e}")
            self.src_conn.rollback()
            self.stats["failed"] += 1
            self.failed_tables.append((table_name, f"查询失败: {e}"))
            return
        rows = src_cur.fetchall()
        src_cur.close()

        if not rows:
            log.info(f"  源表无数据")
            self.stats["success"] += 1
            return

        # 5. 插入数据
        insert_cols = ", ".join(f'"{c}"' for c in col_names)
        placeholders = ", ".join(["%s"] * len(col_names))
        insert_sql = f'INSERT INTO "{table_name}" ({insert_cols}) VALUES ({placeholders})'

        dst_cur = self.dst_conn.cursor()
        count = 0
        first_err = None
        for row in rows:
            try:
                dst_cur.execute(insert_sql, _safe_row(row))
                count += 1
            except Exception as e:
                if not first_err:
                    first_err = str(e).strip().split("\n")[0]
                    log.warning(f"  插入失败: {first_err}")
                self.dst_conn.rollback()
                dst_cur = self.dst_conn.cursor()
                break

        self.dst_conn.commit()

        # 6. 重置 SERIAL 序列
        for col in col_names:
            try:
                dst_cur.execute(f"""
                    SELECT pg_get_serial_sequence('"{table_name}"', '{col}')
                """)
                seq = dst_cur.fetchone()[0]
                if seq:
                    dst_cur.execute(f"""
                        SELECT setval('{seq}', COALESCE((SELECT MAX("{col}") FROM "{table_name}"), 1))
                    """)
                    self.dst_conn.commit()
            except Exception:
                self.dst_conn.rollback()

        if first_err:
            log.warning(f"  ✗ 插入失败，已导入 {count} 行")
            if count == 0:
                self.stats["failed"] += 1
                self.failed_tables.append((table_name, first_err))
            else:
                self.stats["success"] += 1
        else:
            log.info(f"  ✓ 导入 {count} 行")
            self.stats["success"] += 1
        self.stats["rows_total"] += count

    def run(self):
        try:
            self.connect()

            log.info("")
            log.info("═══ 跳过的表 ═══")
            for table, reason in SKIP_TABLES.items():
                log.info(f"  {table}: {reason}")
            log.info("")

            # 字段映射迁移
            for m in MIGRATIONS:
                self.migrate_table(m)

            # 多表合并迁移
            for m in MERGE_MIGRATIONS:
                self.migrate_merge_table(m)

            # 直接平移（自动建表 + 全量复制）
            for table_name in DIRECT_COPY_TABLES:
                self.migrate_direct_copy(table_name)

            # 打印汇总
            log.info("")
            log.info("═══ 迁移完成 ═══")
            log.info(f"  成功: {self.stats['success']}")
            log.info(f"  跳过: {self.stats['skipped']}")
            log.info(f"  失败: {self.stats['failed']}")
            log.info(f"  总行数: {self.stats['rows_total']}")

            if self.skipped_tables:
                log.info("")
                log.info("── 跳过的表 ──")
                for table, reason in self.skipped_tables:
                    log.info(f"  {table}: {reason}")

            if self.failed_tables:
                log.info("")
                log.error("── 失败的表 ──")
                for table, reason in self.failed_tables:
                    log.error(f"  {table}: {reason}")
        finally:
            self.close()


def rebuild_local_ddl():
    ddl_dir = os.path.join(os.path.dirname(__file__), "..", "repositories", "structure")
    ddl_dir = os.path.abspath(ddl_dir)
    ddl_files = sorted(
        os.path.join(ddl_dir, f) for f in os.listdir(ddl_dir) if f.endswith(".sql")
    )
    if not ddl_files:
        log.error(f"在 {ddl_dir} 下未找到 SQL 文件")
        return

    log.info(f"重建本地 DDL，共 {len(ddl_files)} 个文件")
    conn = psycopg2.connect(**TARGET_DB)
    conn.autocommit = True
    cur = conn.cursor()
    for path in ddl_files:
        fname = os.path.basename(path)
        log.info(f"  执行: {fname}")
        with open(path, "r") as f:
            sql = f.read()
        try:
            cur.execute(sql)
            log.info(f"  ✓ {fname}")
        except Exception as e:
            log.error(f"  ✗ {fname}: {e}")
    cur.close()
    conn.close()
    log.info("DDL 重建完成")


def main():
    parser = argparse.ArgumentParser(description="meme_db → oshit_db 数据库迁移")
    parser.add_argument("--dry-run", action="store_true", help="仅打印，不执行写入")
    parser.add_argument("--tables", nargs="+", help="只迁移指定新表名")
    parser.add_argument("--rebuild-ddl", action="store_true", help="重建本地 DDL 后再迁移")
    parser.add_argument("--only-rebuild-ddl", action="store_true", help="只重建 DDL")
    parser.add_argument("--source-host", help="覆盖源库 host")
    parser.add_argument("--source-password", help="覆盖源库密码")
    args = parser.parse_args()

    if args.source_host:
        SOURCE_DB["host"] = args.source_host
    if args.source_password:
        SOURCE_DB["password"] = args.source_password

    if args.only_rebuild_ddl:
        rebuild_local_ddl()
        return

    if args.rebuild_ddl:
        rebuild_local_ddl()

    migrator = Migrator(dry_run=args.dry_run, tables=args.tables)
    migrator.run()


if __name__ == "__main__":
    main()

