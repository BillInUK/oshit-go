# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

### Build
```bash
# Build all packages
go build ./...

# Build a specific service
cd app/base/api && go build .
cd app/reward/api && go build .
```

### Run
```bash
# Working directory MUST be the service directory — config reads from ./etc/*.yaml
cd app/base/api && go run base.go      # HTTP :1100, Dubbo :20880
cd app/reward/api && go run reward.go  # HTTP :1200
```

### Test
Tests are **integration tests** that call live HTTP APIs. Both services, PostgreSQL, Redis, Kafka, and Nacos must be running first.

```bash
# Run all tests in a package
cd app/reward/api && go test ./test/... -v

# Run a single test
cd app/reward/api && go test ./test/... -v -run TestTakeToken
cd app/reward/api && go test ./test/... -v -run TestGiveToken
cd app/base/api  && go test ./test/... -v -run TestLogin
```

### Generate DAL (GORM gen)
Do **not** manually edit `*.gen.go` files under `common/pkg/dal/`. Re-run the generator if DB schema changes.

## Architecture

### Service Responsibilities

**base-api** owns all infrastructure concerns:
- Signs transactions with service private keys (`tx.Signatures[1]`) and broadcasts to Solana
- Runs `TxScanTask` (every 3s) to detect on-chain confirmations and emit Kafka messages
- Runs `TxExpireTask` (every 5s) as a failsafe for unconfirmed txns > 5 min old
- Exposes fee/price/invite data via HTTP and Dubbo Triple to reward-api

**reward-api** owns all business logic:
- Validates and decodes user-submitted Solana transactions before forwarding to base
- Consumes Kafka `ServiceTransaction` topic to write business records after on-chain confirmation
- Two completed patterns: `TakeToken` (platform → user) and `GiveToken` (user → user + platform reward)

### Critical Data Flow

```
User submits tx → reward validates → base signs + broadcasts → TxScanTask detects → Kafka → reward writes records
```

`txId` is always `tx.Signatures[0]` (the user's signature), never the service signature at index 1.

`t_service_tx` is written by **both** base (in `SendTransaction`) and reward (in `recordTakeToken`/`recordGiveToken`). The reward-side insert must handle unique key conflicts.

### Shared Packages (`common/`)

- `common/pkg/dal/model/` + `common/pkg/dal/query/` — GORM auto-generated; never edit directly
- `common/pkg/entity/` — Kafka message structs (`KafkaTxMsg`, `NewScannedTx`, `NewExpiredTx`) and decoded transaction structs shared between services
- `common/pkg/pb/base/` — Protobuf generated code for Dubbo Triple RPC
- `common/utils/` — Cross-service utilities: `amount.go` (raw ↔ UI conversion using `big.Float`), `jsypt_util.go` (private key decryption), `solana_util.go`
- `app/utils/` — App-layer utilities: `tx.go` (transaction pre-check, decode, service-layer decode), `rpc.go` (distributed rate limiting)

### Context Pattern

Both services use a `CoreContext` (infra: DB, Redis, RPC, Kafka, configs) embedded in a `ServiceContext` (business-specific: business configs, task manager, Dubbo client). Logic files receive `*svc.ServiceContext` directly.

### Adding a New Business (e.g., Lottery)

Follow the GiveToken pattern:
1. `logic/{business}/logic.go` — HTTP handler logic (GetConfig, GetTxInfo, CommitTx)
2. `logic/{business}/decode.go` — Transaction parsing and rule validation
3. `logic/{business}/kafka.go` — `HandleScannedTx` / `HandleExpiredTx` implementations
4. Register Kafka handlers in `setup.go` via `TaskMgr.RegisterScannedTxHandler`
5. Add routes in `handler/routes.go`

### Key Gotchas

- **json-iterator** is configured with `MarshalFloatWith6Digits: false` — do not use `encoding/json` for HTTP responses or float precision will be lost.
- **Kafka consumer** auto-commits offsets (`kafka-go` `ReadMessage`); failed message processing is not retried. Business logic must be idempotent. Fund flow writes use `OnConflict` upsert for safety.
- **Distributed locks**: periodic tasks use `WithTries(1)` (skip if locked); per-tx locks have no try limit (guarantees exactly-once processing).
- **Private keys** are jasypt-encrypted (`PBEWithHMACSHA512AndAES_256`) in DB. Decryption password is in `svc/context.go`. base uses `t_service_key` (indexed by service+subService); reward uses `t_reward_key_config` (indexed by service).
- **`t_fee_statistics`** is capped at 10,000 rows — overflow writes overwrite oldest rows rather than inserting new ones.
- **GiveToken `/commit-tx`** route is currently commented out in `handler/routes.go`.
