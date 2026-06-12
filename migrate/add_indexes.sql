-- =====================================================
-- 索引补充迁移脚本
--
-- 根据业务代码中的查询模式分析，以下表缺少必要索引。
-- 已有索引的表不在此列。
--
-- 执行方式：psql -U meme_server -d oshit_db -f add_indexes.sql
-- 所有语句使用 IF NOT EXISTS / CONCURRENTLY，可安全重复执行。
-- =====================================================

-- ===================== t_native_account_info =====================
-- auth.go: WHERE native_account = ?
-- auth.go: WHERE invite_code = ?
-- tx.go:   WHERE token_account = ?
CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS idx_native_account_info_native_account
    ON public.t_native_account_info (native_account);

CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS idx_native_account_info_invite_code
    ON public.t_native_account_info (invite_code);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_native_account_info_token_account
    ON public.t_native_account_info (token_account);

-- ===================== t_invite_relation =====================
-- 已有: uq_invite_relation_invitee (invitee)
-- invite.go: WHERE inviter = ?
-- stake_logic.go: 递归 CTE 中 WHERE invitee = ? (已有) + JOIN on invitee = ip.inviter
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_invite_relation_inviter
    ON public.t_invite_relation (inviter);

-- ===================== t_service_tx =====================
-- scan.go / expire.go / tx.go: WHERE tx_id = ?
-- scan.go: Find all (全表扫描, tx_state 过滤)
CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS idx_service_tx_tx_id
    ON public.t_service_tx (tx_id);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_service_tx_tx_state
    ON public.t_service_tx (tx_state) WHERE tx_state = 0;

-- ===================== t_hacker_account =====================
-- 业务代码中通过 native_account 查询黑名单
CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS idx_hacker_account_native_account
    ON public.t_hacker_account (native_account);

-- ===================== t_exclude_account =====================
-- 业务代码中通过 native_account 查询排除名单
CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS idx_exclude_account_native_account
    ON public.t_exclude_account (native_account);

-- ===================== t_take_token_record =====================
-- take/logic.go: WHERE tx_id = ?
-- take/logic.go: WHERE receipt_account = ? AND use_invite_code = ? AND tx_state = ?
-- take/kafka.go: WHERE tx_id = ? (UPDATE)
CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS idx_take_token_record_tx_id
    ON public.t_take_token_record (tx_id);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_take_token_record_receipt_invite
    ON public.t_take_token_record (receipt_account, use_invite_code, tx_state);

-- ===================== t_give_token_record =====================
-- give/logic.go: WHERE tx_id = ?
-- give/logic.go: WHERE from_account = ? ORDER BY created_at DESC
-- give/logic.go: WHERE from_account = ? AND created_at >= ? AND created_at < ?
-- give/logic.go: WHERE receipt_account = ? AND tx_state >= ?
-- give/kafka.go: WHERE tx_id = ? (UPDATE)
CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS idx_give_token_record_tx_id
    ON public.t_give_token_record (tx_id);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_give_token_record_from_account
    ON public.t_give_token_record (from_account, created_at DESC);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_give_token_record_receipt_account
    ON public.t_give_token_record (receipt_account, tx_state);

-- ===================== t_reward_code =====================
-- rewardcode/logic.go: WHERE reward_code = ?
-- rewardcode/kafka.go: WHERE tx_id = ?
CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS idx_reward_code_reward_code
    ON public.t_reward_code (reward_code);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_reward_code_tx_id
    ON public.t_reward_code (tx_id) WHERE tx_id IS NOT NULL;

-- ===================== t_lottery_reward =====================
-- lottery/logic.go: 按 native_account + reward_day + reward_state 查询待领取奖励
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_lottery_reward_account_day_state
    ON public.t_lottery_reward (native_account, reward_day, reward_state);

-- ===================== t_lottery_claim =====================
-- lottery/kafka.go: WHERE tx_id = ?
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_lottery_claim_tx_id
    ON public.t_lottery_claim (tx_id) WHERE tx_id IS NOT NULL;

-- ===================== t_stake_record =====================
-- stake_logic.go: 通过 staker 查询质押记录
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_stake_record_staker
    ON public.t_stake_record (staker);

-- ===================== t_stake_leader =====================
-- stake_logic.go: WHERE native_account = ? (递归 CTE 中高频查询)
CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS idx_stake_leader_native_account
    ON public.t_stake_leader (native_account);

-- ===================== t_stake_reward =====================
-- 已有: (native_account, reward_state, pending), (native_account, snap_day, reward_type, starred)
-- stake/reward_kafka.go: WHERE tx_id = ? (高频 UPDATE)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_stake_reward_tx_id
    ON public.t_stake_reward (tx_id) WHERE tx_id IS NOT NULL;

-- ===================== t_stake_snap_shot =====================
-- 已有: (native_account, stake_type, snap_day)
-- stake/snapshot_logic.go: WHERE snap_day = ? (快照计算时全天扫描)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_stake_snap_shot_snap_day
    ON public.t_stake_snap_shot (snap_day);

-- ===================== t_stake_buy_token =====================
-- 已有: (tx_id) UNIQUE, (locked), (locked_at), (locked_by), (slot)
-- stake_logic.go: WHERE from_account = ? AND expired = false AND remaining_amount > 0
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_stake_buy_token_from_account_active
    ON public.t_stake_buy_token (from_account) WHERE expired = false AND remaining_amount > 0;
