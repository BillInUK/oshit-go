insert into t_mainnet_rpc_config
(chain_name,rpc_url,wss_url)
values
    ('solana','https://mainnet.helius-rpc.com/?api-key=a4309444-6229-433a-a89f-3fbe85f5f043','wss://mainnet.helius-rpc.com/?api-key=a4309444-6229-433a-a89f-3fbe85f5f043');

-- 业务配置
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
VALUES ('pos', 'stake token', 'As9Z52f8Sioqr22KpS4xdzrhicwGwAu6x5SxVaHfvLws', 'As9Z52f8Sioqr22KpS4xdzrhicwGwAu6x5SxVaHfvLws', '3BUhdNDeSNFRXLCF4Y3fnVUXE4VXTwXguXPZ6pZteaQqZ9RqZhRTUazXbPUZ6kCtMBXXBsNEaXyUacvL4icqV8mH', '', '454568185', NOW(),NOW());

INSERT INTO public.t_tx_scan_info
(service, sub_service, native_account, pda_account, until_tx_id, before_tx_id, slot, created_at, updated_at)
VALUES ('pos', 'stake reward', 'H5WmBY45gxP8rj7gecLXsv6yNHqHFXNH4Acmp2U9E2Tb', 'ERYPoieDaHoh9jz1Gbmi9whHLd1QyWPKANvtZTnECufw', '4ev21ymUwMR9EJAWUB58bBZzgT9KfaX3WB4G1tssu5MoM9UwKmL8CU8bfdYZGjracdyChCUE3rvUZLCeBdnj6ud1', '', '454329612', NOW(),NOW());

INSERT INTO public.t_tx_scan_info
(service, sub_service, native_account, pda_account, until_tx_id, before_tx_id, slot, created_at, updated_at)
VALUES ('pos', 'stake leader reward', '2yRkofKW7xKRbN79MHKGX8HFyuHZEtTJDhTwAQjJHnMX', 'HqAp7uFZAckGA9jKFCZXg7rq1GX9CB2kwpGBKoq5geg6', '29dZDL5PFTaBYdqUz7E3gFBJsLzMsHckgwSiga2y1c2RBL95t2RL4y6PRepaZkfQcqxQhRCpe2RKPDn7UUyL9U7p', '', '453414783',NOW(),NOW());

insert into public.t_stake_leader_reward_config(reward_account)values('2yRkofKW7xKRbN79MHKGX8HFyuHZEtTJDhTwAQjJHnMX');

update t_chain_config set chain_name='solana';
update t_stake_amm_config set quote_token='solana';
update t_user_wallet_rpc_config set chain_name='solana';