-- RPC 配置重构迁移脚本
-- 1. 新增 t_rpc_endpoint 表（统一管理所有 RPC 节点）
-- 2. 精简 t_chain_config（去掉 rpc_url, wss_url）
-- 3. 删除 t_user_wallet_rpc_config、t_mainnet_rpc_config

-- Step 1: 创建新表
DROP TABLE IF EXISTS public.t_rpc_endpoint;
CREATE TABLE public.t_rpc_endpoint
(
    record_id    public.ulid DEFAULT public.gen_ulid() NOT NULL,
    scope        varchar(32)   NOT NULL,   -- 'env' = 跟随环境(devnet/mainnet), 'mainnet' = 固定主网
    provider     varchar(32)   NOT NULL,   -- 'quicknode', 'helius', 'alchemy', 'custom'
    endpoint     varchar(512)  NOT NULL,   -- 不含 api key 的 base URL
    api_key      varchar(256)  DEFAULT '',
    wss_endpoint varchar(512)  DEFAULT '',
    wss_api_key  varchar(256)  DEFAULT '',
    weight       integer       DEFAULT 1 NOT NULL,  -- 轮询权重, 0 表示禁用
    created_at   timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at   timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX uq_rpc_endpoint_scope_provider_endpoint ON public.t_rpc_endpoint (scope, provider, endpoint);

-- Step 2: 精简 t_chain_config（去掉 rpc_url, wss_url）
ALTER TABLE public.t_chain_config DROP COLUMN IF EXISTS rpc_url;
ALTER TABLE public.t_chain_config DROP COLUMN IF EXISTS wss_url;

-- Step 3: 删除旧表
DROP TABLE IF EXISTS public.t_user_wallet_rpc_config;
DROP TABLE IF EXISTS public.t_mainnet_rpc_config;

-- Step 4: 插入示例数据（测试网）
-- INSERT INTO public.t_rpc_endpoint (scope, provider, endpoint, api_key, wss_endpoint, wss_api_key, weight)
-- VALUES
--   ('env', 'quicknode', 'https://solitary-solitary-brook.solana-devnet.quiknode.pro', '59ff9976f07ec18f5fceb2766ebecbb9b2247bc8', 'wss://solitary-solitary-brook.solana-devnet.quiknode.pro', '59ff9976f07ec18f5fceb2766ebecbb9b2247bc8', 1),
--   ('mainnet', 'helius', 'https://mainnet.helius-rpc.com', 'a4309444-6229-433a-a89f-3fbe85f5f043', '', '', 1);

INSERT INTO public.t_rpc_endpoint (scope, provider, endpoint, api_key, wss_endpoint, wss_api_key, weight)
VALUES
  ('env', 'quicknode', 'https://solitary-solitary-brook.solana-devnet.quiknode.pro', '59ff9976f07ec18f5fceb2766ebecbb9b2247bc8', 'wss://solitary-solitary-brook.solana-devnet.quiknode.pro', '59ff9976f07ec18f5fceb2766ebecbb9b2247bc8', 1),
  ('mainnet', 'helius', 'https://mainnet.helius-rpc.com', 'a4309444-6229-433a-a89f-3fbe85f5f043', '', '', 1);
