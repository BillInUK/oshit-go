-- 基础配置
delete from t_system_config;
insert into t_system_config(env)values(1);

delete from t_aws_config;
insert into t_aws_config
(access_key_id,secret_access_key,region)
values
    ('AKIAUBM64K3O3GFSHN7W','7eGqmuBxHO8x82s8gU6pjQRVtU4bkCgE5XC8qTvI','ap-southeast-1');

delete from t_chain_config;
insert into t_chain_config
(chain_name,rpc_url,wss_url,decimals,symbol)
values
    ('solana','https://solitary-solitary-brook.solana-devnet.quiknode.pro/59ff9976f07ec18f5fceb2766ebecbb9b2247bc8/','wss://solitary-solitary-brook.solana-devnet.quiknode.pro/59ff9976f07ec18f5fceb2766ebecbb9b2247bc8/',9,'SOL');

delete from t_token_config;
insert into t_token_config
(token_name,token_symbol,decimals,mint)
values
    ('OShit','OShit',3,'wtnrTujJqBRUknLRhQQcUSwzAzx8LvcxKXEuBwvFnJM');

delete from t_fee_tolerance;
insert into t_fee_tolerance(max_less_rate)values(0.05);

delete from t_user_wallet_rpc_config;
insert into t_user_wallet_rpc_config
(chain_name,rpc_url,wss_url)
values
    ('solana','https://solitary-solitary-brook.solana-devnet.quiknode.pro/59ff9976f07ec18f5fceb2766ebecbb9b2247bc8/','wss://solitary-solitary-brook.solana-devnet.quiknode.pro/59ff9976f07ec18f5fceb2766ebecbb9b2247bc8/');

delete from t_mainnet_rpc_config;
insert into t_mainnet_rpc_config
(chain_name,rpc_url,wss_url)
values
    ('solana','https://mainnet.helius-rpc.com/?api-key=a4309444-6229-433a-a89f-3fbe85f5f043','wss://mainnet.helius-rpc.com/?api-key=a4309444-6229-433a-a89f-3fbe85f5f043');

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

-- 业务私钥
delete from t_service_key;

insert into t_service_key
(service,sub_service,encrypted_key,created_at,updated_at)
values
    ('reward','take token','Ue0THMXv3m604QOnN6H/eJ+WTcBUHnKf1fpsYyaQE6PkVvVtVpNQCjy/9wRhNPVJhPdkUPJ/AsLyQFQGPahwnGIPzzXN9oCE1vDM5uoJbSFZrPdZ2+kiyKvLDMvQcBWi3Qf65tq0llORRT9OygvntL9DJe+2H6qlfkHPdNatWus=',NOW(),NOW());

insert into t_service_key
(service,sub_service,encrypted_key,created_at,updated_at)
values
    ('reward','give token','CNshWghLiZv3hEFznuZo24A9aVmh8tEe0Ivxsf7uRxz+BILC1OkD+izKs6N10gYtVA6RHh9WOX0oqm5HLDsEkuD8CYbra9UTdDxpv9PewRBS4yRGdIOJpyPY4jMNtKGlBPaB5Xi/xgyOVJXabXJhDhPsxisGRY205/g803XTK/w=',NOW(),NOW());

insert into t_service_key
(service,sub_service,encrypted_key,created_at,updated_at)
values
    ('reward','lottery','Ue0THMXv3m604QOnN6H/eJ+WTcBUHnKf1fpsYyaQE6PkVvVtVpNQCjy/9wRhNPVJhPdkUPJ/AsLyQFQGPahwnGIPzzXN9oCE1vDM5uoJbSFZrPdZ2+kiyKvLDMvQcBWi3Qf65tq0llORRT9OygvntL9DJe+2H6qlfkHPdNatWus=',NOW(),NOW());

insert into t_service_key
(service,sub_service,encrypted_key,created_at,updated_at)
values
    ('reward','reward code','bQgVW5UNj48lhiuc/BLQxjAo1iawe0hGt1jzYG5+azHmKal2w/EVX8XQWTr0Ehsbp06nprl1POsC90hsJOpkDoC5FOEjMT5bGojiEcl7c+gTRKj3AeO2iig+RLaz34U6VC/OdwEV4tEz8NXgG4GCfeO9/HgsTDeLZ+h1WlOeqjw=',NOW(),NOW());

insert into t_service_key
(service,sub_service,encrypted_key,created_at,updated_at)
values
    ('reward','campaign quote','B85fDrCcJmQWHAxSZ+3G0cAUCkTPp+b8R5OH/CtR6OLcoOboHk1k0+DvjM+BIn1rmY2bM19WdM0N4XOZSS5o1Rb8AgrT+ggCcRmGYYUyAHSIj7fnLCMkR62MV3KAVne5ziIoyb+hXQ6jlfDyq+ScO56Vykst5sYkSG7/v9AMoLQ=',NOW(),NOW());

insert into t_service_key
(service,sub_service,encrypted_key,created_at,updated_at)
values
    ('pos','pos reward','rztLu8OZF5xL7TVUJVgHWIXtDC9LuDdwe4H/iXN9g+OGKLF9KNY8AQzCq6e1TOUButUThc9JOIY6hF8ppoU7AidL5SbvX4oRf4WH0tY+y7QRjxqjuscb1+RXnn99PgO4R+qsqvncVCFcmuHhSUvi1PTX1XZyviOx7V26+XoBP7U=',NOW(),NOW());

insert into t_service_key
(service,sub_service,encrypted_key,created_at,updated_at)
values
    ('pos','stake token','Ue0THMXv3m604QOnN6H/eJ+WTcBUHnKf1fpsYyaQE6PkVvVtVpNQCjy/9wRhNPVJhPdkUPJ/AsLyQFQGPahwnGIPzzXN9oCE1vDM5uoJbSFZrPdZ2+kiyKvLDMvQcBWi3Qf65tq0llORRT9OygvntL9DJe+2H6qlfkHPdNatWus=',NOW(),NOW());

insert into t_service_key
(service,sub_service,encrypted_key,created_at,updated_at)
values
    ('pos','stake reward','N7hCCM3Cresa7MTEBI8LGJY3pigGKCwy59OLHLX7DhDqf4tCJFxIxoBgEmDtqlIxyWVpyDhEVjn7qQNer01a7rrMTsSZMPO1bDTs8Z17gmkJf4nRUEe/CDO396ywRql+Fyrf4soIU4YENJd8lm+hZ/+2ABmi7Bj74+ClQzGkOak=',NOW(),NOW());

insert into t_service_key
(service,sub_service,encrypted_key,created_at,updated_at)
values
    ('pos','stake leader reward','al2M3Q+0pu2FDePJvtnaCE4mtPZnTck1BXGcSCZtWgiSPXOveWaSuMKo6i4qapQcHqdMdWp6i3aPjkx/R96aC/AF5Je13r7Lmokpf7/LlxcoZzt+nJxtEmVlTDS1jful93CCIBjp0Yx+b0k1Cf4jWx6FQu9KdkzRqXBIoggoHuw=',NOW(),NOW());

-- 业务交易扫描表
delete from t_tx_scan_info;

INSERT INTO public.t_tx_scan_info
(service, sub_service, native_account, pda_account, until_tx_id, before_tx_id, slot, created_at, updated_at)
VALUES ('reward', 'take token', 'GmKsGRytiVoeMZGmBVCWPcUzJGHVqcvzhP5K9cstdr3E', 'EHEu46gQMTFw1ieiok5XVYLV9MrKUFyk6sRLDjUpEQAd', '5mgtNggkqkKUTu4WweeArYTq55epKxof1WE8KH25c1k5QSA1esrNa99pLyFHuixpemoUj9pRPKtejzmbM1aGHWbm', '', '454355929', NOW(),NOW());

INSERT INTO public.t_tx_scan_info
(service, sub_service, native_account, pda_account, until_tx_id, before_tx_id, slot, created_at, updated_at)
VALUES ('reward', 'give token', 'AjhUm6o9eV2xV9G2ZPH3pSb8DhTAjb27MDrTkKMrDVVZ', '6zimN4MMo5CJW7VwhD3nAdXt6EpVoc5V7SCnZiS1kLPd', '3c1DoVsbMeYjzVjo4Ju4VYzwgWwXJQWUF4DRirN8138eCEiPNxMCHpnxJ1a7NHUPLZWdMXpu5oGAnhD324LLxFAQ', '', '452939864', NOW(),NOW());

INSERT INTO public.t_tx_scan_info
(service, sub_service, native_account, pda_account, until_tx_id, before_tx_id, slot, created_at, updated_at)
VALUES
    ('reward', 'lottery', 'GmKsGRytiVoeMZGmBVCWPcUzJGHVqcvzhP5K9cstdr3E', 'EHEu46gQMTFw1ieiok5XVYLV9MrKUFyk6sRLDjUpEQAd', '5mgtNggkqkKUTu4WweeArYTq55epKxof1WE8KH25c1k5QSA1esrNa99pLyFHuixpemoUj9pRPKtejzmbM1aGHWbm', '', '454355929', NOW(),NOW());

INSERT INTO public.t_tx_scan_info
(service, sub_service, native_account, pda_account, until_tx_id, before_tx_id, slot, created_at, updated_at)
VALUES
    ('reward', 'reward code', '584AMuM1HkV4wRMMVPZuZy9g9mZbSAcTF7QiBrHJaZFE', '7AqjQqMB6WGUd6Tc84b81GDRVz65pcCiEePcZZtMBjSm', '2pJZsh4R9WX6dpgrn9NZEbFAyXPadvsn1USg2VKSSUHyzDBRG7eqVUdynJfFJ35Ke8dmLgDWgsWQohb2LX1aqLpY', '', '448363883', NOW(),NOW());

INSERT INTO public.t_tx_scan_info
("service", "sub_service", "native_account", "pda_account", "until_tx_id", "before_tx_id", "slot", "created_at", "updated_at")
VALUES
    ('reward', 'campaign quote', 'Gacw8xMnWtdSefjTqFhutA6yEdhymDThN7p95vRrThms', '2NVji8RvQAFhg4YJKuxqhdMjWMLJmWbKm5MBvSJmTUHL', '4Vteqp2SFBsq4EJF3dKvGJqQxL3CubrtnyE3pA6DCWaRZPoUJ8wekR7rDTWA4wHwiAKvECLD5BHxRJf49CU4MGuv', '', '451413105', NOW(),NOW());

INSERT INTO public.t_tx_scan_info
(service, sub_service, native_account, pda_account, until_tx_id, before_tx_id, slot, created_at, updated_at)
VALUES
    ('pos', 'pos reward', 'C2E7K1fDUzpihX77xMNnhYNidRHkWnLMTRejWRvfjkDH', '2q66HpBbncGSy3iTMoJhhFZFeSsc6JofwV8xoRdGxtdK', '2ggyqAnaxhtbTZiURqjivXrHvpssg1QDKwvZtXiRTprucHfEjp2jJJnEic76PFc5L6CPQA9C6zaJscuYYZs382gT', '', '450913917', NOW(),NOW());

INSERT INTO public.t_tx_scan_info
(service, sub_service, native_account, pda_account, until_tx_id, before_tx_id, slot, created_at, updated_at)
VALUES ('pos', 'market buy token', 'HtNfUbDaBamCBPWCFiESkXpewvwVLkwrSWRjjV8FNT7i', 'HtNfUbDaBamCBPWCFiESkXpewvwVLkwrSWRjjV8FNT7i', '5EZ6uHLA7b9dkpULbdGaEj5NrxzU3cZm18p74ak2kWMQz4b8EuYuzuqPfrCVqgxcrpTcGSApAejMpKHwXJyoSTwN', '', '412913094', NOW(),NOW());

INSERT INTO public.t_tx_scan_info
(service, sub_service, native_account, pda_account, until_tx_id, before_tx_id, slot, created_at, updated_at)
VALUES ('pos', 'stake token', 'CyLTEgvmqVF9dPJkT6bMgccfXL7G26EXRAM9FEuP5ki6', 'CyLTEgvmqVF9dPJkT6bMgccfXL7G26EXRAM9FEuP5ki6', '3BUhdNDeSNFRXLCF4Y3fnVUXE4VXTwXguXPZ6pZteaQqZ9RqZhRTUazXbPUZ6kCtMBXXBsNEaXyUacvL4icqV8mH', '', '454568185', NOW(),NOW());

INSERT INTO public.t_tx_scan_info
(service, sub_service, native_account, pda_account, until_tx_id, before_tx_id, slot, created_at, updated_at)
VALUES ('pos', 'stake reward', 'H5WmBY45gxP8rj7gecLXsv6yNHqHFXNH4Acmp2U9E2Tb', 'ERYPoieDaHoh9jz1Gbmi9whHLd1QyWPKANvtZTnECufw', '4ev21ymUwMR9EJAWUB58bBZzgT9KfaX3WB4G1tssu5MoM9UwKmL8CU8bfdYZGjracdyChCUE3rvUZLCeBdnj6ud1', '', '454329612', NOW(),NOW());

INSERT INTO public.t_tx_scan_info
(service, sub_service, native_account, pda_account, until_tx_id, before_tx_id, slot, created_at, updated_at)
VALUES ('pos', 'stake leader reward', '2yRkofKW7xKRbN79MHKGX8HFyuHZEtTJDhTwAQjJHnMX', 'HqAp7uFZAckGA9jKFCZXg7rq1GX9CB2kwpGBKoq5geg6', '29dZDL5PFTaBYdqUz7E3gFBJsLzMsHckgwSiga2y1c2RBL95t2RL4y6PRepaZkfQcqxQhRCpe2RKPDn7UUyL9U7p', '', '453414783',NOW(),NOW());

-- reward服务初始化数据
delete from t_level_dist;
insert into t_level_dist(dist_level)values(2);

delete from t_level_ratio;
insert into t_level_ratio(dist_level,ratio)values(1,10);
insert into t_level_ratio(dist_level,ratio)values(2,1);

delete from t_discount_rate;
insert into t_discount_rate(rate)values(1.25);

-- take token 配置
delete from t_take_token_config;
insert into public.t_take_token_config(invite_code,reward_account,cost_account,amount,invite_amount,cost_fee_rate,max_cost_fee,is_default,reward_inviter,invited,created_at,updated_at)values(NULL,'GmKsGRytiVoeMZGmBVCWPcUzJGHVqcvzhP5K9cstdr3E', '6MeXfYMhXpQSz3fqHtEa72V1XgKG7WGsECDy9jEv9e2K', 500000, 1500000, 200,400,true,true,true,NOW(),NOW());

-- give token 配置
delete from t_give_token_config;
insert into public.t_give_token_config(reward_account,cost_account,reward_rate,max_valid_reward,valid_rate,created_at,updated_at)values('AjhUm6o9eV2xV9G2ZPH3pSb8DhTAjb27MDrTkKMrDVVZ','6MeXfYMhXpQSz3fqHtEa72V1XgKG7WGsECDy9jEv9e2K',200,4000000,300,NOW(),NOW());

-- reward code
delete from t_reward_code_config;
INSERT INTO public.t_reward_code_config(reward_account,cost_account)VALUES('584AMuM1HkV4wRMMVPZuZy9g9mZbSAcTF7QiBrHJaZFE','6MeXfYMhXpQSz3fqHtEa72V1XgKG7WGsECDy9jEv9e2K');

delete from t_reward_code_fee;
INSERT INTO public.t_reward_code_fee(amount,fee_rate)VALUES(5000000,8);
INSERT INTO public.t_reward_code_fee(amount,fee_rate)VALUES(1000000,20);
INSERT INTO public.t_reward_code_fee(amount,fee_rate)VALUES(500000,30);

-- campaign 配置
delete from t_campaign_quote_config;
INSERT INTO t_campaign_quote_config(reward_account,cost_account,quote_rate,cost_rate)VALUES('2NVji8RvQAFhg4YJKuxqhdMjWMLJmWbKm5MBvSJmTUHL','6MeXfYMhXpQSz3fqHtEa72V1XgKG7WGsECDy9jEv9e2K',500,17);


-- pos奖励配置
delete from t_pos_reward_config;
INSERT INTO t_pos_reward_config(reward_account,cost_account,cost_fee_rate,max_cost_fee,quote_token_amount,created_at,updated_at)VALUES('C2E7K1fDUzpihX77xMNnhYNidRHkWnLMTRejWRvfjkDH','6MeXfYMhXpQSz3fqHtEa72V1XgKG7WGsECDy9jEv9e2K',200,40000,500000,NOW(),NOW());

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
insert into public.t_stake_fix_rate_config(min_amount,stake_type,fix_rate,individual_rate,created_at,updated_at)values(100000,0,70,100,now(),now());
insert into public.t_stake_fix_rate_config(min_amount,stake_type,fix_rate,individual_rate,created_at,updated_at)values(100000,1,100,100,now(),now());

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

