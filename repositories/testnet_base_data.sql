
-- 基础配置
insert into t_system_config(env)values(1);

insert into t_chain_config
(chain,rpc_url,wss_url,decimals,symbol)
values
    ('SOL','https://solitary-solitary-brook.solana-devnet.quiknode.pro/59ff9976f07ec18f5fceb2766ebecbb9b2247bc8/','wss://solitary-solitary-brook.solana-devnet.quiknode.pro/59ff9976f07ec18f5fceb2766ebecbb9b2247bc8/',9,'SOL');

insert into t_token_config
(name,symbol,decimals,mint)
values
    ('OShit','OShit',3,'wtnrTujJqBRUknLRhQQcUSwzAzx8LvcxKXEuBwvFnJM');

insert into t_fee_tolerance(max_less_rate)values(0.05);

insert into t_user_wallet_rpc_config
(chain,rpc_url,wss_url)
values
    ('SOL','https://solitary-solitary-brook.solana-devnet.quiknode.pro/59ff9976f07ec18f5fceb2766ebecbb9b2247bc8/','wss://solitary-solitary-brook.solana-devnet.quiknode.pro/59ff9976f07ec18f5fceb2766ebecbb9b2247bc8/');

-- 业务配置
insert into t_service_info
(service,sub_service,address,webhook,mq_group,mq_topic,hook_type,confirm,multi_sign,created_at,updated_at)
values
    ('Reward','TakeToken','GmKsGRytiVoeMZGmBVCWPcUzJGHVqcvzhP5K9cstdr3E','','Reward','ServiceTransaction',0,true,true,now(),now());

insert into t_service_info
(service,sub_service,address,webhook,mq_group,mq_topic,hook_type,confirm,multi_sign,created_at,updated_at)
values
    ('Reward','GiveToken','6zimN4MMo5CJW7VwhD3nAdXt6EpVoc5V7SCnZiS1kLPd','','Reward','ServiceTransaction',0,true,true,now(),now());

insert into t_service_info
(service,sub_service,address,webhook,mq_group,mq_topic,hook_type,confirm,multi_sign,created_at,updated_at)
values
    ('Reward','Lottery','GmKsGRytiVoeMZGmBVCWPcUzJGHVqcvzhP5K9cstdr3E','','Reward','ServiceTransaction',0,true,true,now(),now());

insert into t_service_info
(service,sub_service,address,webhook,mq_group,mq_topic,hook_type,confirm,multi_sign,created_at,updated_at)
values
    ('Reward','ExchangeToken','2NVji8RvQAFhg4YJKuxqhdMjWMLJmWbKm5MBvSJmTUHL','','Reward','ServiceTransaction',0,true,true,now(),now());

insert into t_service_info
(service,sub_service,address,webhook,mq_group,mq_topic,hook_type,confirm,multi_sign,created_at,updated_at)
values
    ('Pos','PosReward','C2E7K1fDUzpihX77xMNnhYNidRHkWnLMTRejWRvfjkDH','','Pos','ServiceTransaction',0,true,true,now(),now());

insert into t_service_info
(service,sub_service,address,webhook,mq_group,mq_topic,hook_type,confirm,multi_sign,created_at,updated_at)
values
    ('Stake','StakeReward','H5WmBY45gxP8rj7gecLXsv6yNHqHFXNH4Acmp2U9E2Tb','','Stake','ServiceTransaction',0,true,true,now(),now());

insert into t_service_info
(service,sub_service,address,webhook,mq_group,mq_topic,hook_type,confirm,multi_sign,created_at,updated_at)
values
    ('Stake','StakeToken','CyLTEgvmqVF9dPJkT6bMgccfXL7G26EXRAM9FEuP5ki6','','Stake','ServiceTransaction',0,true,true,now(),now());

insert into t_service_info
(service,sub_service,address,webhook,mq_group,mq_topic,hook_type,confirm,multi_sign,created_at,updated_at)
values
    ('Stake','UnStakeToken','CyLTEgvmqVF9dPJkT6bMgccfXL7G26EXRAM9FEuP5ki6','','Stake','ServiceTransaction',0,true,true,now(),now());

insert into t_service_info
(service,sub_service,address,webhook,mq_group,mq_topic,hook_type,confirm,multi_sign,created_at,updated_at)
values
    ('Stake','ReStakeToken','CyLTEgvmqVF9dPJkT6bMgccfXL7G26EXRAM9FEuP5ki6','','Stake','ServiceTransaction',0,true,true,now(),now());

-- 业务私钥
insert into t_service_key
(service,sub_service,encrypted_key,created_at,updated_at)
values
    ('Reward','TakeToken','Ue0THMXv3m604QOnN6H/eJ+WTcBUHnKf1fpsYyaQE6PkVvVtVpNQCjy/9wRhNPVJhPdkUPJ/AsLyQFQGPahwnGIPzzXN9oCE1vDM5uoJbSFZrPdZ2+kiyKvLDMvQcBWi3Qf65tq0llORRT9OygvntL9DJe+2H6qlfkHPdNatWus=',NOW(),NOW());

insert into t_service_key
(service,sub_service,encrypted_key,created_at,updated_at)
values
    ('Reward','GiveToken','CNshWghLiZv3hEFznuZo24A9aVmh8tEe0Ivxsf7uRxz+BILC1OkD+izKs6N10gYtVA6RHh9WOX0oqm5HLDsEkuD8CYbra9UTdDxpv9PewRBS4yRGdIOJpyPY4jMNtKGlBPaB5Xi/xgyOVJXabXJhDhPsxisGRY205/g803XTK/w=',NOW(),NOW());

insert into t_service_key
(service,sub_service,encrypted_key,created_at,updated_at)
values
    ('Reward','Lottery','Ue0THMXv3m604QOnN6H/eJ+WTcBUHnKf1fpsYyaQE6PkVvVtVpNQCjy/9wRhNPVJhPdkUPJ/AsLyQFQGPahwnGIPzzXN9oCE1vDM5uoJbSFZrPdZ2+kiyKvLDMvQcBWi3Qf65tq0llORRT9OygvntL9DJe+2H6qlfkHPdNatWus=',NOW(),NOW());

insert into t_service_key
(service,sub_service,encrypted_key,created_at,updated_at)
values
    ('Reward','ExchangeToken','B85fDrCcJmQWHAxSZ+3G0cAUCkTPp+b8R5OH/CtR6OLcoOboHk1k0+DvjM+BIn1rmY2bM19WdM0N4XOZSS5o1Rb8AgrT+ggCcRmGYYUyAHSIj7fnLCMkR62MV3KAVne5ziIoyb+hXQ6jlfDyq+ScO56Vykst5sYkSG7/v9AMoLQ=',NOW(),NOW());

insert into t_service_key
(service,sub_service,encrypted_key,created_at,updated_at)
values
    ('Pos','PosReward','rztLu8OZF5xL7TVUJVgHWIXtDC9LuDdwe4H/iXN9g+OGKLF9KNY8AQzCq6e1TOUButUThc9JOIY6hF8ppoU7AidL5SbvX4oRf4WH0tY+y7QRjxqjuscb1+RXnn99PgO4R+qsqvncVCFcmuHhSUvi1PTX1XZyviOx7V26+XoBP7U=',NOW(),NOW());

insert into t_service_key
(service,sub_service,encrypted_key,created_at,updated_at)
values
    ('Stake','StakeReward','N7hCCM3Cresa7MTEBI8LGJY3pigGKCwy59OLHLX7DhDqf4tCJFxIxoBgEmDtqlIxyWVpyDhEVjn7qQNer01a7rrMTsSZMPO1bDTs8Z17gmkJf4nRUEe/CDO396ywRql+Fyrf4soIU4YENJd8lm+hZ/+2ABmi7Bj74+ClQzGkOak=',NOW(),NOW());

insert into t_service_key
(service,sub_service,encrypted_key,created_at,updated_at)
values
    ('Stake','StakeToken','Ue0THMXv3m604QOnN6H/eJ+WTcBUHnKf1fpsYyaQE6PkVvVtVpNQCjy/9wRhNPVJhPdkUPJ/AsLyQFQGPahwnGIPzzXN9oCE1vDM5uoJbSFZrPdZ2+kiyKvLDMvQcBWi3Qf65tq0llORRT9OygvntL9DJe+2H6qlfkHPdNatWus=',NOW(),NOW());

-- 业务交易扫描表
insert into t_tx_scan_info
(service,sub_service,native_account,pda_account,until_tx_id,before_tx_id,slot,created_at,updated_at)
values
    ('Reward','TakeToken', 'GmKsGRytiVoeMZGmBVCWPcUzJGHVqcvzhP5K9cstdr3E','EHEu46gQMTFw1ieiok5XVYLV9MrKUFyk6sRLDjUpEQAd','3JGiMGUV9RyACmJc5Mc1mabtSTLqfQASGyq4FXDEWRtj88ErD3XGjY6qnp4K4SqvA95muqfq58ytxkLnWr9SJftj','',451641554,now(),now());

insert into t_tx_scan_info
(service,sub_service,native_account,pda_account,until_tx_id,before_tx_id,slot,created_at,updated_at)
values
    ('Reward','GiveToken', 'AjhUm6o9eV2xV9G2ZPH3pSb8DhTAjb27MDrTkKMrDVVZ','6zimN4MMo5CJW7VwhD3nAdXt6EpVoc5V7SCnZiS1kLPd','36ygSSVqAgcfJPJYrbwDWi6Z7UXc9A7FHpKoJEieAmsy3CrBauZ6y9ertruaGbJohdrWHeJsvE2E71SXVKthSfNX','',448362934,now(),now());

insert into t_tx_scan_info
(service,sub_service,native_account,pda_account,until_tx_id,before_tx_id,slot,created_at,updated_at)
values
    ('Reward','Lottery', 'GmKsGRytiVoeMZGmBVCWPcUzJGHVqcvzhP5K9cstdr3E','EHEu46gQMTFw1ieiok5XVYLV9MrKUFyk6sRLDjUpEQAd','3JGiMGUV9RyACmJc5Mc1mabtSTLqfQASGyq4FXDEWRtj88ErD3XGjY6qnp4K4SqvA95muqfq58ytxkLnWr9SJftj','',451641554,now(),now());

insert into t_tx_scan_info
(service,sub_service,native_account,pda_account,until_tx_id,before_tx_id,slot,created_at,updated_at)
values
    ('Reward','ExchangeToken', 'Gacw8xMnWtdSefjTqFhutA6yEdhymDThN7p95vRrThms','2NVji8RvQAFhg4YJKuxqhdMjWMLJmWbKm5MBvSJmTUHL','4Vteqp2SFBsq4EJF3dKvGJqQxL3CubrtnyE3pA6DCWaRZPoUJ8wekR7rDTWA4wHwiAKvECLD5BHxRJf49CU4MGuv','',451413105,now(),now());

insert into t_tx_scan_info
(service,sub_service,native_account,pda_account,until_tx_id,before_tx_id,slot,created_at,updated_at)
values
    ('Pos','PosReward', 'C2E7K1fDUzpihX77xMNnhYNidRHkWnLMTRejWRvfjkDH','2q66HpBbncGSy3iTMoJhhFZFeSsc6JofwV8xoRdGxtdK','2ggyqAnaxhtbTZiURqjivXrHvpssg1QDKwvZtXiRTprucHfEjp2jJJnEic76PFc5L6CPQA9C6zaJscuYYZs382gT','',450913917,now(),now());

insert into t_tx_scan_info
(service,sub_service,native_account,pda_account,until_tx_id,before_tx_id,slot,created_at,updated_at)
values
    ('Stake','StakeReward', 'H5WmBY45gxP8rj7gecLXsv6yNHqHFXNH4Acmp2U9E2Tb','ERYPoieDaHoh9jz1Gbmi9whHLd1QyWPKANvtZTnECufw','5WZuHhsGzzuz8D2u3TE1tjc3G9zQoKqwnprNXGn88xykJjPD6DrSLHjL5ah394Fto8vz5hFuXRDwzfZWRPcboE3A','',453398810,now(),now());

insert into t_tx_scan_info
(service,sub_service,native_account,pda_account,until_tx_id,before_tx_id,slot,created_at,updated_at)
values
    ('Stake','StakeToken', 'CyLTEgvmqVF9dPJkT6bMgccfXL7G26EXRAM9FEuP5ki6','CyLTEgvmqVF9dPJkT6bMgccfXL7G26EXRAM9FEuP5ki6','4SA6A52W24XrAyKSnE5GSzAKoeZHxeBqLWcZDFtqsuUdfLtdLjQjuKUA6JaTEcDU7vmLNMKWmFSUeau6pEFDGdiP','',453414552,now(),now());


-- 初始化各个业务的地址信息，防止解析交易的时候反复查询
-- take token
insert into t_native_account_info(native_account,token_account,invite_code)values('GmKsGRytiVoeMZGmBVCWPcUzJGHVqcvzhP5K9cstdr3E','EHEu46gQMTFw1ieiok5XVYLV9MrKUFyk6sRLDjUpEQAd','8XXZCPrO');
-- give token
insert into t_native_account_info(native_account,token_account,invite_code)values('AjhUm6o9eV2xV9G2ZPH3pSb8DhTAjb27MDrTkKMrDVVZ','6zimN4MMo5CJW7VwhD3nAdXt6EpVoc5V7SCnZiS1kLPd','8XXZCPr1');
-- exchange token
insert into t_native_account_info(native_account,token_account,invite_code)values('2NVji8RvQAFhg4YJKuxqhdMjWMLJmWbKm5MBvSJmTUHL','Gacw8xMnWtdSefjTqFhutA6yEdhymDThN7p95vRrThms','8XXZCPr2');
-- pos reward
insert into t_native_account_info(native_account,token_account,invite_code)values('C2E7K1fDUzpihX77xMNnhYNidRHkWnLMTRejWRvfjkDH','2q66HpBbncGSy3iTMoJhhFZFeSsc6JofwV8xoRdGxtdK','8XXZCPr3');
-- stake reward
insert into t_native_account_info(native_account,token_account,invite_code)values('H5WmBY45gxP8rj7gecLXsv6yNHqHFXNH4Acmp2U9E2Tb','ERYPoieDaHoh9jz1Gbmi9whHLd1QyWPKANvtZTnECufw','8XXZCPr4');

-- 替换成短合约
update t_service_info set address='As9Z52f8Sioqr22KpS4xdzrhicwGwAu6x5SxVaHfvLws' where address='CyLTEgvmqVF9dPJkT6bMgccfXL7G26EXRAM9FEuP5ki6';
update t_tx_scan_info set native_account='As9Z52f8Sioqr22KpS4xdzrhicwGwAu6x5SxVaHfvLws',pda_account='As9Z52f8Sioqr22KpS4xdzrhicwGwAu6x5SxVaHfvLws' where native_account='CyLTEgvmqVF9dPJkT6bMgccfXL7G26EXRAM9FEuP5ki6';
update t_tx_scan_info set until_tx_id='fNm9y9puUxhdshZv8Ft1o9DFNUE18eYQZ1ZBccHNQpVnGRXDRceuXGmQGhqt25Q9quFVoskKz4fHoL2NByF3KBJ',slot=454099465 where native_account='CyLTEgvmqVF9dPJkT6bMgccfXL7G26EXRAM9FEuP5ki6';
