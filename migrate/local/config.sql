-- 基础配置
delete from t_system_config;
insert into t_system_config(env)values(1);

-- aws配置
delete from t_aws_config;
insert into t_aws_config
(access_key_id,secret_access_key,region)
values
    ('ENC~2mNxs8K/N+vv2+U1/deebykeBI2bK7mzOQeunQYBmm3c5Wd5H+CAXC3b5DzWmIOL9LUMV5HF2peiJzWyo//7gw==','ENC~vLPX1hybMcS8uHKpk/4OqFF+NEfTimIiP9ITODaSLg81Yujg/Oc2hJULLdyUb5EQp1tTML3nmNkstSM8xZ/HFwWMWJhCUuRsHAwt4legNv4=','ap-southeast-1');

-- rpc 配置
INSERT INTO public.t_rpc_endpoint
    (scope, provider, endpoint, api_key, wss_endpoint, wss_api_key, weight)
VALUES
    ('env', 'helius', 'https://devnet.helius-rpc.com', 'ENC~tzszBl6r9HXBDBxbzVzIWK08Hme5hlY9hgYu6PLid7onKzLs3wL1GS1XhFkvS2wcnkEyK7tb/oByRfnxHXwLsr0AFlrB2PnsmCguRpscSuE=', 'wss://devnet.helius-rpc.com', 'ENC~tzszBl6r9HXBDBxbzVzIWK08Hme5hlY9hgYu6PLid7onKzLs3wL1GS1XhFkvS2wcnkEyK7tb/oByRfnxHXwLsr0AFlrB2PnsmCguRpscSuE=', 1),
    ('mainnet', 'helius', 'https://mainnet.helius-rpc.com', 'ENC~tzszBl6r9HXBDBxbzVzIWK08Hme5hlY9hgYu6PLid7onKzLs3wL1GS1XhFkvS2wcnkEyK7tb/oByRfnxHXwLsr0AFlrB2PnsmCguRpscSuE=', 'wss://mainnet.helius-rpc.com', 'ENC~tzszBl6r9HXBDBxbzVzIWK08Hme5hlY9hgYu6PLid7onKzLs3wL1GS1XhFkvS2wcnkEyK7tb/oByRfnxHXwLsr0AFlrB2PnsmCguRpscSuE=', 1);

delete from t_chain_config;
insert into t_chain_config
(chain_name,decimals,symbol)
values
    ('solana',9,'SOL');

delete from t_token_config;
insert into t_token_config
(token_name,token_symbol,decimals,mint)
values
    ('OShit','OShit',3,'wtnrTujJqBRUknLRhQQcUSwzAzx8LvcxKXEuBwvFnJM');

delete from t_fee_tolerance;
insert into t_fee_tolerance(max_less_rate)values(0.05);

-- 业务配置
delete from t_service_info;

insert into t_service_info
    (service,sub_service,address,webhook,mq_group,mq_topic,hook_type,tx_source,confirm,multi_sign,created_at,updated_at)
values
    ('reward','take token','GmKsGRytiVoeMZGmBVCWPcUzJGHVqcvzhP5K9cstdr3E','','reward','ServiceTransaction',0,0,true,true,now(),now());

insert into t_service_info
    (service,sub_service,address,webhook,mq_group,mq_topic,hook_type,tx_source,confirm,multi_sign,created_at,updated_at)
values
    ('reward','give token','AjhUm6o9eV2xV9G2ZPH3pSb8DhTAjb27MDrTkKMrDVVZ','','reward','ServiceTransaction',0,0,true,true,now(),now());

insert into t_service_info
    (service,sub_service,address,webhook,mq_group,mq_topic,hook_type,tx_source,confirm,multi_sign,created_at,updated_at)
values
    ('reward','lottery','GmKsGRytiVoeMZGmBVCWPcUzJGHVqcvzhP5K9cstdr3E','','reward','ServiceTransaction',0,0,true,true,now(),now());

insert into t_service_info
    (service,sub_service,address,webhook,mq_group,mq_topic,hook_type,tx_source,confirm,multi_sign,created_at,updated_at)
values
    ('reward','reward code','584AMuM1HkV4wRMMVPZuZy9g9mZbSAcTF7QiBrHJaZFE','','reward','ServiceTransaction',0,0,true,true,now(),now());

insert into t_service_info
    (service,sub_service,address,webhook,mq_group,mq_topic,hook_type,tx_source,confirm,multi_sign,created_at,updated_at)
values
    ('reward','campaign quote','2NVji8RvQAFhg4YJKuxqhdMjWMLJmWbKm5MBvSJmTUHL','','reward','ServiceTransaction',0,0,true,true,now(),now());

insert into t_service_info
    (service,sub_service,address,webhook,mq_group,mq_topic,hook_type,tx_source,confirm,multi_sign,created_at,updated_at)
values
    ('pos','pos reward','C2E7K1fDUzpihX77xMNnhYNidRHkWnLMTRejWRvfjkDH','','pos','ServiceTransaction',0,0,true,true,now(),now());

insert into t_service_info
    (service,sub_service,address,webhook,mq_group,mq_topic,hook_type,tx_source,confirm,multi_sign,created_at,updated_at)
values
    ('pos','stake token','CyLTEgvmqVF9dPJkT6bMgccfXL7G26EXRAM9FEuP5ki6','','Stake','ServiceTransaction',0,0,true,true,now(),now());

insert into t_service_info
    (service,sub_service,address,webhook,mq_group,mq_topic,hook_type,tx_source,confirm,multi_sign,created_at,updated_at)
values
    ('pos','stake reward','H5WmBY45gxP8rj7gecLXsv6yNHqHFXNH4Acmp2U9E2Tb','','Stake','ServiceTransaction',0,0,true,true,now(),now());

insert into t_service_info
    (service,sub_service,address,webhook,mq_group,mq_topic,hook_type,tx_source,confirm,multi_sign,created_at,updated_at)
values
    ('pos','stake leader reward','2yRkofKW7xKRbN79MHKGX8HFyuHZEtTJDhTwAQjJHnMX','','Stake','ServiceTransaction',0,0,true,true,now(),now());

insert into t_service_info
    (service,sub_service,address,webhook,mq_group,mq_topic,hook_type,tx_source,confirm,multi_sign,created_at,updated_at)
values
    ('pos','market buy token','HtNfUbDaBamCBPWCFiESkXpewvwVLkwrSWRjjV8FNT7i','','Stake','ServiceTransaction',1,0,true,true,now(),now());

insert into t_service_info
    (service,sub_service,address,webhook,mq_group,mq_topic,hook_type,tx_source,confirm,multi_sign,created_at,updated_at)
values
    ('gs-relay','contribution','AUq3iXbJjBjED6ZN2JqJDdZ3mknd7cXv9HQqNdYuEZ6q','','gs-relay-activity','ServiceTransaction',1,0,true,true,now(),now());

insert into t_service_info
    (service,sub_service,address,webhook,mq_group,mq_topic,hook_type,tx_source,confirm,multi_sign,created_at,updated_at)
values
    ('gs-relay','payout','AUq3iXbJjBjED6ZN2JqJDdZ3mknd7cXv9HQqNdYuEZ6q','','gs-relay-activity','ServiceTransaction',1,0,true,true,now(),now());

insert into t_service_info
    (service,sub_service,address,webhook,mq_group,mq_topic,hook_type,tx_source,confirm,multi_sign,created_at,updated_at)
values
    ('gs-relay','refund','AUq3iXbJjBjED6ZN2JqJDdZ3mknd7cXv9HQqNdYuEZ6q','','gs-relay-activity','ServiceTransaction',1,0,true,true,now(),now());


-- 业务私钥
delete from t_service_key;

insert into t_service_key
    (service,sub_service,encrypted_key,created_at,updated_at)
values
    ('reward','take token','0sCivDuWlbgtW/tc1i3KPVlkChE/2iu+AvTn0UZ5TBIoYEmfRdyRKcXCZe+pQfWcJsflnE53zDqUyGNPF9ReU54gChS4WosO0WhOP/jc1u+MIwix37rElwUa8BXS64pKwXoFbMSMd5t51VL9uM5O8rrpXWN9NUCeo4aIJDDg9FE=',NOW(),NOW());

insert into t_service_key
    (service,sub_service,encrypted_key,created_at,updated_at)
values
    ('reward','give token','0g78tKE/TLzl/tOm16/96IZELZH9q64LawBT4tZY4XTboPvNLkZwJMv3Ieo2h3Z3IKqp/Hl2Cw2nTufxVGGWbOmmEihCzd3TuCeUKjBYHq5++/UIMBLlUZfKN6it9ZOYyASE2Bn1K5SY4FJ2CmnNTsGnIriAGlCOeZztEW7vPR8=',NOW(),NOW());

insert into t_service_key
    (service,sub_service,encrypted_key,created_at,updated_at)
values
    ('reward','lottery','W4OQKGKWq3d6fNKWjXP5ShxBIEug3ZQbBzmcSHZtmCohOAnhZ3a7ClNQ7v+Ab02lFk8p2aZcapV840R78iFxpwR5lJL/xm5ZArSQccG6x1nSWPMEw6otmiMpG8bHtWZBhIk9yF3aqcMEbW/Qh1BK1bHGVrMx7g46fzVPRDfH9Mg=',NOW(),NOW());

insert into t_service_key
    (service,sub_service,encrypted_key,created_at,updated_at)
values
    ('reward','reward code','UG5VJEVHBKZBJ1JF3c1/by9mrI2Jc0PTFmiIMOgGPhMud/885x65TeCYnjdkaX1N70N+Rh5CrBV0T7VSRrZ3lT1L01EXJ3B9MlmcBld2xOnDF1jL+t4Vdm/kylSHArhJ7oucJokKFYk0Trh3J00tM1r1aQQSSltTaASMycgshvc=',NOW(),NOW());

insert into t_service_key
    (service,sub_service,encrypted_key,created_at,updated_at)
values
    ('reward','campaign quote','3KU/wRvdyP2b+6TXaelumobcPXeGqR9mnZZKnvLHRPkh5jWRTjczKVcH0oiTrKKba5mqAMVUd0pAhipXkHZ8jL624dqIybR3AELyXTWCDqCsG1fGjWsmHRHPOwAsU0EtDsMuBZWmXlPEcSPVlzr7BBPREphvB3ohI/8c7Vk6VmQ=',NOW(),NOW());

insert into t_service_key
    (service,sub_service,encrypted_key,created_at,updated_at)
values
    ('pos','pos reward','qDzkq4Y5h4FQ2DN0RGm3Cq4pbtrxX9cfxrboJUBWLEpgWdc6Lskz01LTnEWCxki0ywpXLEsnDjWnPfc9kzZr1S4sv2/n+EkYoNj32qFV9kxsMjpHTMj0ivr3ukbSXscfL/neXKIrS9lsedgX3WGtZvc50fUgIAInYX3GCcE3HeI=',NOW(),NOW());

insert into t_service_key
    (service,sub_service,encrypted_key,created_at,updated_at)
values
    ('pos','stake token','MOoGsyzGtBhmzH0ntTdj/zAEnvLOw0UDKMvR4mc8Lmy7RdTH2nPR8Bnjhv2LzrEnZyosuoiwVRVJFtchkcv4HB4RUIhCZDcwd86a75qbr3yYSo2x+y0GXs76vuRpvnuiYTshOxBKc+84ItSSBlw+08Gq+kwj0qZ74aRmA/ElHQ0=',NOW(),NOW());

insert into t_service_key
    (service,sub_service,encrypted_key,created_at,updated_at)
values
    ('pos','stake reward','d6XA694aFMnzgh1yqPxpwZ0ip8UPalfmYcSmq5emsNuVRUkamhQf8nw2ALbI8gewLX2Y4ubR9tl11e6UWXymw5sdhG9fDtQYqL1TvCT25Od5avUqd+4rCm+bk9Wm5hZQttRVAAcnf9KTMvAaIUdC7R+C8maQKs90Rl/g49VinLw=',NOW(),NOW());

insert into t_service_key
    (service,sub_service,encrypted_key,created_at,updated_at)
values
    ('pos','stake leader reward','rPA75VSKyaUOzkdslSyBh93ZTDNQ7inqe8TuzmuN11QP4OM70mIh0+UTt6Ijb0xcf9YCxUn1eQBuw3lCKnu4o4nHyKaycRMkrFeDeKs9QjgTIbKJDWbCkMEGcPBLJve5TM0BeMBPAcR43kaRytiYIQVIl9UPc8UlS6u/sB2FN6M=',NOW(),NOW());

insert into t_service_key
    (service,sub_service,encrypted_key,created_at,updated_at)
values
    ('gs-relay','contribution','sqDwITATtSmuUrVo1J6FB3iRPlbr4hUhUR0CAs8rphOnCAUYWTDtdDtT1cFSW7zpclN9wrTDfKBm2kZybR+x0QSYZrXimUJNp4JJBXcHIqVAOH8v+o2t/wbjNO1wd5z59x5wHIoJ1iBD5wDISMWxS5iponamObzrvnAnqtND+q8=',NOW(),NOW());

insert into t_service_key
    (service,sub_service,encrypted_key,created_at,updated_at)
values
    ('gs-relay','payout','sqDwITATtSmuUrVo1J6FB3iRPlbr4hUhUR0CAs8rphOnCAUYWTDtdDtT1cFSW7zpclN9wrTDfKBm2kZybR+x0QSYZrXimUJNp4JJBXcHIqVAOH8v+o2t/wbjNO1wd5z59x5wHIoJ1iBD5wDISMWxS5iponamObzrvnAnqtND+q8=',NOW(),NOW());

insert into t_service_key
    (service,sub_service,encrypted_key,created_at,updated_at)
values
    ('gs-relay','refund','sqDwITATtSmuUrVo1J6FB3iRPlbr4hUhUR0CAs8rphOnCAUYWTDtdDtT1cFSW7zpclN9wrTDfKBm2kZybR+x0QSYZrXimUJNp4JJBXcHIqVAOH8v+o2t/wbjNO1wd5z59x5wHIoJ1iBD5wDISMWxS5iponamObzrvnAnqtND+q8=',NOW(),NOW());


-- 业务交易扫描表
delete from t_tx_scan_info;

INSERT INTO public.t_tx_scan_info
    (service, sub_service, native_account, pda_account, until_tx_id, before_tx_id, slot, created_at, updated_at)
VALUES
    ('reward', 'take token', 'GmKsGRytiVoeMZGmBVCWPcUzJGHVqcvzhP5K9cstdr3E', 'EHEu46gQMTFw1ieiok5XVYLV9MrKUFyk6sRLDjUpEQAd', '4LwXXoEouKiQBKoVjmcWeTmgJmsnYjnyTf7YZ8Mv5X1FgQ7ZHQTsLKUJxYa2Ly9ixiBWKvegie8AJZZfNWXGNkQK', '', '468683431', NOW(),NOW());

INSERT INTO public.t_tx_scan_info
    (service, sub_service, native_account, pda_account, until_tx_id, before_tx_id, slot, created_at, updated_at)
VALUES
    ('reward', 'give token', 'AjhUm6o9eV2xV9G2ZPH3pSb8DhTAjb27MDrTkKMrDVVZ', '6zimN4MMo5CJW7VwhD3nAdXt6EpVoc5V7SCnZiS1kLPd', '4Ly2tdZ1mnFnAK76kk3gyEek6mKL24rB5zWnZCthPCf5db971seK6URCm9XfhWzbtWQm1d7fcXfxEkwt3ETrfvPQ', '', '468683609', NOW(),NOW());

INSERT INTO public.t_tx_scan_info
    (service, sub_service, native_account, pda_account, until_tx_id, before_tx_id, slot, created_at, updated_at)
VALUES
    ('reward', 'lottery', 'GmKsGRytiVoeMZGmBVCWPcUzJGHVqcvzhP5K9cstdr3E', 'EHEu46gQMTFw1ieiok5XVYLV9MrKUFyk6sRLDjUpEQAd', '4LwXXoEouKiQBKoVjmcWeTmgJmsnYjnyTf7YZ8Mv5X1FgQ7ZHQTsLKUJxYa2Ly9ixiBWKvegie8AJZZfNWXGNkQK', '', '468683431', NOW(),NOW());

INSERT INTO public.t_tx_scan_info
    (service, sub_service, native_account, pda_account, until_tx_id, before_tx_id, slot, created_at, updated_at)
VALUES
    ('reward', 'reward code', '584AMuM1HkV4wRMMVPZuZy9g9mZbSAcTF7QiBrHJaZFE', '7AqjQqMB6WGUd6Tc84b81GDRVz65pcCiEePcZZtMBjSm', '59jM2yhLQHxy7hmJU2hctR5JctcA4GqVLgS5owzt1XGYUKqFWhqJ8urHVo7e3qcWsGdFxRphNk8TPHQTvdHTx2d5', '', '463871508', NOW(),NOW());

INSERT INTO public.t_tx_scan_info
    (service, sub_service, native_account, pda_account, until_tx_id, before_tx_id, slot, created_at, updated_at)
VALUES
    ('reward', 'campaign quote', 'Gacw8xMnWtdSefjTqFhutA6yEdhymDThN7p95vRrThms', '2NVji8RvQAFhg4YJKuxqhdMjWMLJmWbKm5MBvSJmTUHL', '4Vteqp2SFBsq4EJF3dKvGJqQxL3CubrtnyE3pA6DCWaRZPoUJ8wekR7rDTWA4wHwiAKvECLD5BHxRJf49CU4MGuv', '', '451413105', NOW(),NOW());

INSERT INTO public.t_tx_scan_info
    (service, sub_service, native_account, pda_account, until_tx_id, before_tx_id, slot, created_at, updated_at)
VALUES
    ('pos', 'pos reward', 'C2E7K1fDUzpihX77xMNnhYNidRHkWnLMTRejWRvfjkDH', '2q66HpBbncGSy3iTMoJhhFZFeSsc6JofwV8xoRdGxtdK', '2EjeyS6EBekiFNNpt76rMRRj7BQZnfTj9wwGZGK4VSEo6uCUz8wV83F26RSZK32uvNovrefoxLwL2tX5h6cozjdF', '', '468179450', NOW(),NOW());

INSERT INTO public.t_tx_scan_info
    (service, sub_service, native_account, pda_account, until_tx_id, before_tx_id, slot, created_at, updated_at)
VALUES
    ('pos', 'market buy token', 'HtNfUbDaBamCBPWCFiESkXpewvwVLkwrSWRjjV8FNT7i', 'HtNfUbDaBamCBPWCFiESkXpewvwVLkwrSWRjjV8FNT7i', '3Ujgy4P55C29AqL3QooUYj3RQsdQipktFKpnKo8t6pVfWHMcmUYuXutP1XACdexJkJHtKTBHur14e2ojA9UnEGSJ', '', '428168081', NOW(),NOW());

INSERT INTO public.t_tx_scan_info
    (service, sub_service, native_account, pda_account, until_tx_id, before_tx_id, slot, created_at, updated_at)
VALUES
    ('pos', 'stake token', 'CyLTEgvmqVF9dPJkT6bMgccfXL7G26EXRAM9FEuP5ki6', 'CyLTEgvmqVF9dPJkT6bMgccfXL7G26EXRAM9FEuP5ki6', '4CATtfwvXKKdwj4exQsg6EobJFhHye8TnbkSQJUa1mPG9zm9YZj66ghKYbH5Tn2M6N8BaCA5kQ6J8VWMrHjzYMAB', '', '468179889', NOW(),NOW());

INSERT INTO public.t_tx_scan_info
    (service, sub_service, native_account, pda_account, until_tx_id, before_tx_id, slot, created_at, updated_at)
VALUES
    ('pos', 'stake reward', 'H5WmBY45gxP8rj7gecLXsv6yNHqHFXNH4Acmp2U9E2Tb', 'ERYPoieDaHoh9jz1Gbmi9whHLd1QyWPKANvtZTnECufw', '3erA9CwvNeoWn8J63Q4PVdbgs7vaYkFX71wMdgyt7AW1dwWCeKeVqWp5YgKtNiKYbwH82btRfXz6boiUBUsJ9KYg', '', '468683779', NOW(),NOW());

INSERT INTO public.t_tx_scan_info
    (service, sub_service, native_account, pda_account, until_tx_id, before_tx_id, slot, created_at, updated_at)
VALUES
    ('pos', 'stake leader reward', '2yRkofKW7xKRbN79MHKGX8HFyuHZEtTJDhTwAQjJHnMX', 'HqAp7uFZAckGA9jKFCZXg7rq1GX9CB2kwpGBKoq5geg6', '4EjeyG9GvWGhFVYjJDnXHA7Pije91HoFSfpUoMAg9dW68Xf1jDxCvdB5n3m74rAQM7cXZ1GxL9oad5j3FpjT6urp', '', '467320074', NOW(),NOW());

INSERT INTO public.t_tx_scan_info
    (service, sub_service, native_account, pda_account, until_tx_id, before_tx_id, slot, created_at, updated_at)
VALUES
    ('gs-relay', 'contribution', 'AUq3iXbJjBjED6ZN2JqJDdZ3mknd7cXv9HQqNdYuEZ6q', '7BV3LUAhoq4JpsLYejzx5dCdPpesB2ubAevxMNEoSbWE', '5xJDwFhzR1d9h2BLncVKUUCKdwwYdY38HNKRMiXFPPYRMMkq83Y4AMZJQvCqRnDXJkQSSanb7wEzNWritHCNLvkh', '', '467262544', NOW(),NOW());

INSERT INTO public.t_tx_scan_info
    (service, sub_service, native_account, pda_account, until_tx_id, before_tx_id, slot, created_at, updated_at)
VALUES
    ('gs-relay', 'payout', 'AUq3iXbJjBjED6ZN2JqJDdZ3mknd7cXv9HQqNdYuEZ6q', '7BV3LUAhoq4JpsLYejzx5dCdPpesB2ubAevxMNEoSbWE', '5xJDwFhzR1d9h2BLncVKUUCKdwwYdY38HNKRMiXFPPYRMMkq83Y4AMZJQvCqRnDXJkQSSanb7wEzNWritHCNLvkh', '', '467262544', NOW(),NOW());

INSERT INTO public.t_tx_scan_info
    (service, sub_service, native_account, pda_account, until_tx_id, before_tx_id, slot, created_at, updated_at)
VALUES
    ('gs-relay', 'refund', 'AUq3iXbJjBjED6ZN2JqJDdZ3mknd7cXv9HQqNdYuEZ6q', '7BV3LUAhoq4JpsLYejzx5dCdPpesB2ubAevxMNEoSbWE', '5xJDwFhzR1d9h2BLncVKUUCKdwwYdY38HNKRMiXFPPYRMMkq83Y4AMZJQvCqRnDXJkQSSanb7wEzNWritHCNLvkh', '', '467262544', NOW(),NOW());



-- reward服务初始化数据
delete from t_level_dist;
insert into t_level_dist(dist_level)values(2);

delete from t_level_ratio;
insert into t_level_ratio(dist_level,ratio)values(1,10);
insert into t_level_ratio(dist_level,ratio)values(2,5);

delete from t_discount_rate;
insert into t_discount_rate(rate)values(1.25);

-- take token 配置
delete from t_take_token_config;
insert into public.t_take_token_config
    (invite_code,reward_account,cost_account,amount,invite_amount,cost_fee_rate,invited_rate,max_cost_fee,is_default,reward_inviter,invited,created_at,updated_at)
values
    (NULL,'GmKsGRytiVoeMZGmBVCWPcUzJGHVqcvzhP5K9cstdr3E', '6MeXfYMhXpQSz3fqHtEa72V1XgKG7WGsECDy9jEv9e2K', 500000, 1500000, 200,20,75000,true,true,true,NOW(),NOW());

-- give token 配置
delete from public.t_give_token_config;
insert into public.t_give_token_config
    (reward_account,cost_account,max_reward,max_valid_reward,reward_rate,valid_rate,created_at,updated_at)
values
    ('AjhUm6o9eV2xV9G2ZPH3pSb8DhTAjb27MDrTkKMrDVVZ','6MeXfYMhXpQSz3fqHtEa72V1XgKG7WGsECDy9jEv9e2K',500000,1500000,200,300,NOW(),NOW());

-- lottery 配置
DELETE FROM t_lottery_config;
INSERT INTO t_lottery_config
    (reward_account,cost_account,cost_amount,cost_fee_rate,created_at,updated_at)
VALUES
    ('GmKsGRytiVoeMZGmBVCWPcUzJGHVqcvzhP5K9cstdr3E', '6MeXfYMhXpQSz3fqHtEa72V1XgKG7WGsECDy9jEv9e2K',500000,110,NOW(),NOW());

-- reward code
delete from t_reward_code_config;
INSERT INTO public.t_reward_code_config
    (reward_account,cost_account)
VALUES
    ('584AMuM1HkV4wRMMVPZuZy9g9mZbSAcTF7QiBrHJaZFE','6MeXfYMhXpQSz3fqHtEa72V1XgKG7WGsECDy9jEv9e2K');

delete from t_reward_code_fee;
INSERT INTO public.t_reward_code_fee(amount,fee_rate)VALUES(5000000,8);
INSERT INTO public.t_reward_code_fee(amount,fee_rate)VALUES(1000000,20);
INSERT INTO public.t_reward_code_fee(amount,fee_rate)VALUES(500000,30);

-- campaign 配置
delete from t_campaign_quote_config;
INSERT INTO t_campaign_quote_config
    (reward_account,cost_account,quote_rate,cost_rate)
VALUES
    ('2NVji8RvQAFhg4YJKuxqhdMjWMLJmWbKm5MBvSJmTUHL','6MeXfYMhXpQSz3fqHtEa72V1XgKG7WGsECDy9jEv9e2K',500,17);

-- pos奖励配置
delete from t_pos_reward_config;
INSERT INTO t_pos_reward_config
    (reward_account,cost_account,cost_fee_rate,max_cost_fee,quote_token_amount,created_at,updated_at)
VALUES
    ('C2E7K1fDUzpihX77xMNnhYNidRHkWnLMTRejWRvfjkDH','6MeXfYMhXpQSz3fqHtEa72V1XgKG7WGsECDy9jEv9e2K',200,40000,500000,NOW(),NOW());

-- pos星级配置
delete from t_pos_star_level_rule;
INSERT INTO t_pos_star_level_rule(amount,group_amount,star_level,rate,created_at,updated_at)VALUES(1000000,10000000,1,10,NOW(),NOW());
INSERT INTO t_pos_star_level_rule(amount,group_amount,star_level,rate,created_at,updated_at)VALUES(2000000,20000000,2,20,NOW(),NOW());
INSERT INTO t_pos_star_level_rule(amount,group_amount,star_level,rate,created_at,updated_at)VALUES(3000000,30000000,3,30,NOW(),NOW());
INSERT INTO t_pos_star_level_rule(amount,group_amount,star_level,rate,created_at,updated_at)VALUES(4000000,40000000,4,40,NOW(),NOW());
INSERT INTO t_pos_star_level_rule(amount,group_amount,star_level,rate,created_at,updated_at)VALUES(5000000,50000000,5,50,NOW(),NOW());

-- pos任务配置
delete from t_pos_mission_config;
INSERT INTO t_pos_mission_config (reward_type,starred,rate,created_at,updated_at) VALUES (1, false, 10.00, NOW(),NOW());
INSERT INTO t_pos_mission_config (reward_type,starred,rate,created_at,updated_at) VALUES (2, false, 10.00, NOW(),NOW());
INSERT INTO t_pos_mission_config (reward_type,starred,rate,created_at,updated_at) VALUES (3, false, 10.00, NOW(),NOW());
INSERT INTO t_pos_mission_config (reward_type,starred,rate,created_at,updated_at) VALUES (1, true,  10.00, NOW(),NOW());
INSERT INTO t_pos_mission_config (reward_type,starred,rate,created_at,updated_at) VALUES (2, true,  10.00, NOW(),NOW());
INSERT INTO t_pos_mission_config (reward_type,starred,rate,created_at,updated_at) VALUES (3, true,  10.00, NOW(),NOW());
INSERT INTO t_pos_mission_config (reward_type,starred,rate,created_at,updated_at) VALUES (0, false, 15.00, NOW(),NOW());
INSERT INTO t_pos_mission_config (reward_type,starred,rate,created_at,updated_at) VALUES (0, true,  15.00, NOW(),NOW());

-- 质押AMM配置
delete from t_stake_amm_config;
insert into public.t_stake_amm_config(quote_token,public_key)values('solana','46uzvWDstrwNtEpBSFrcVPx4ZaTMDpjarQYWpq82Z58p');

-- 质押池配置
delete from t_stake_token_pool;
insert into public.t_stake_token_pool(source_account,from_token_account)values('Raydium','Gh6MjRrJFBU9HcYMYKBbaD8fX1dv3CjGtDDhtVThD9v3');

-- 质押每日固定利息
delete from t_stake_fix_rate_config;
insert into public.t_stake_fix_rate_config(min_amount,stake_type,rate_tier,fix_rate,individual_rate,created_at,updated_at)values(100000,0,1,70,70,now(),now());
insert into public.t_stake_fix_rate_config(min_amount,stake_type,rate_tier,fix_rate,individual_rate,created_at,updated_at)values(100000,0,2,63,63,now(),now());
insert into public.t_stake_fix_rate_config(min_amount,stake_type,rate_tier,fix_rate,individual_rate,created_at,updated_at)values(100000,1,1,100,100,now(),now());
insert into public.t_stake_fix_rate_config(min_amount,stake_type,rate_tier,fix_rate,individual_rate,created_at,updated_at)values(100000,1,2,90,90,now(),now());

-- 邀请奖励级别
delete from t_stake_invite_dist;
insert into public.t_stake_invite_dist(dist_level)values(2);

-- 邀请奖励每个级别的奖励费率
delete from t_stake_invite_rate;
insert into public.t_stake_invite_rate(dist_level,rate)values(1,10);
insert into public.t_stake_invite_rate(dist_level,rate)values(2,5);

-- 质押星级配置
delete from t_stake_star_level_rule;
insert into public.t_stake_star_level_rule(amount,group_amount,star_level,rate,created_at,updated_at)values(100000000,0,1,8,now(),now());
insert into public.t_stake_star_level_rule(amount,group_amount,star_level,rate,created_at,updated_at)values(200000000,0,2,16,now(),now());
insert into public.t_stake_star_level_rule(amount,group_amount,star_level,rate,created_at,updated_at)values(300000000,0,3,24,now(),now());
insert into public.t_stake_star_level_rule(amount,group_amount,star_level,rate,created_at,updated_at)values(400000000,0,4,32,now(),now());
insert into public.t_stake_star_level_rule(amount,group_amount,star_level,rate,created_at,updated_at)values(500000000,0,5,40,now(),now());
insert into public.t_stake_star_level_rule(amount,group_amount,star_level,rate,created_at,updated_at)values(600000000,0,6,48,now(),now());

-- 质押奖励发放配置表
delete from t_stake_reward_config;
insert into public.t_stake_reward_config
(program_id,reward_account,cost_account,quote_token_amount,cost_fee_rate,created_at,updated_at)
values
    ('CyLTEgvmqVF9dPJkT6bMgccfXL7G26EXRAM9FEuP5ki6','H5WmBY45gxP8rj7gecLXsv6yNHqHFXNH4Acmp2U9E2Tb','6MeXfYMhXpQSz3fqHtEa72V1XgKG7WGsECDy9jEv9e2K',500000,110,now(),now());

-- 插入区域经理配置
delete from t_stake_leader_reward_config;
insert into public.t_stake_leader_reward_config(reward_account)values('2yRkofKW7xKRbN79MHKGX8HFyuHZEtTJDhTwAQjJHnMX');

-- 插入总区域经理配置
delete from t_stake_total_leader;
INSERT INTO t_stake_total_leader (native_account, stake_share, created_at, updated_at) VALUES ('G6xxsFzFHPhLCUg4Qq8aun2EcQ3hTcLvwb6pUWMsVBKa', 7, '2026-04-05 17:34:19.128958', '2026-04-05 17:34:19.128958');
INSERT INTO t_stake_total_leader (native_account, stake_share, created_at, updated_at) VALUES ('CH6nEGuiYF5kenkavr4KMLEvY6DP7cKgQKrh9t2UiHX9', 3, '2026-04-05 17:34:19.130259', '2026-04-05 17:34:19.130259');

