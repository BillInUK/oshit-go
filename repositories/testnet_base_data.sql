
-- 基础配置
insert into t_system_config(env)values(1);

insert into t_aws_config
(access_key_id,secret_access_key,region)
values
    ('AKIAUBM64K3O3GFSHN7W','7eGqmuBxHO8x82s8gU6pjQRVtU4bkCgE5XC8qTvI','ap-southeast-1');

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



-- 邀请关系建立
INSERT INTO t_native_account_info (native_account, token_account, invite_code, created_at, updated_at)
VALUES
    -- total-leader-1
    ('G6xxsFzFHPhLCUg4Qq8aun2EcQ3hTcLvwb6pUWMsVBKa', '7rqmwZE6beTSeWkNmDGW9VwmTk11bSXbDPYtEc7mcWkK', 'NYTGAGvx', NOW(),
     NOW()),
    -- total-leader-2
    ('CH6nEGuiYF5kenkavr4KMLEvY6DP7cKgQKrh9t2UiHX9', '6gKkNXf3tRFfVM4wbTEdMvzR2LrxtQXJeDdoMkGZCRww', 'e6MUAr3s', NOW(),
     NOW()),
    -- leader-1
    ('FgPU2MGLkX278ZD2XhNaJBwVd7q2CjHa9Y9Em3u1s7Ur', '6VWit9rTCwhPRNPGYxXWhKoiyV3Nzhrwdk3nRXELN1RA', 'rFbGpgF3', NOW(),
     NOW()),
    -- leader-2
    ('DGMNe2KYDB8fbxdcxipHwmzMqJk34dMZxdWxguZC4xrC', 'BNGhpT8iR1bMaFyoMyNvTSHNCqMFEyLLCEPNqAPqhgt3', 'yQwcUQFP', NOW(),
     NOW()),
    -- leader-3
    ('D8Dibj91XjosaCjiWbk2zHQvvuFDTUvvjEeT351Z7fK9', '9XXEPVJgDhVLj2rDfxMXtrgN3ZiRrPtXLDqS9otHnsKG', 'CVxAag8f', NOW(),
     NOW()),
    -- leader-4
    ('9ti8HrjakPuRmFa2Q6Lj7UopKuCuS2gSGLNk5umZ7X8v', 'DyCFdRXUn1tT6LTKPa1Pjiwu7jJB9caLHVVptgp4mCRB', 'SMADVYQt', NOW(),
     NOW()),
    -- leader-5
    ('4tJJv3RiQfwL4vdRsdYJK12K1yoirw3b2n1RTZxKacwn', 'Abk6fS8ct61yS7FZtFPpsqpdCrt9eE4eAibTAu86Sivq', 'KwzYVn8p', NOW(),
     NOW()),
    -- leader-6
    ('GVfKwyUXYuVhh65v8Q232dLhdm51eBfsLZPGAsWK6npk', 'HmxtvFya2zuczkmFc1zF3keQJwPpGRZBdTKk7quzK1tm', 'uvsR9u2E', NOW(),
     NOW()),
    -- leader-7
    ('EzehTETE4o8Jua5pCAUk7AHV5Zt3QL1a2D9vfE1ii6HK', '3pe6cR8nW56HSgFq4VR5oeN4kCE1cB4g1CTUaTwsXVEk', 'SqVvXTpM', NOW(),
     NOW()),
    -- A1
    ('5D4MWh35wxUcY1hBsm5GwuippPL2UBmfDnfkC8MeqxcN', 'FJCn2oL4vAS4mrvKaa4o4tVsCCw49sYpW98SV1K8uQ37', 'bWb8GNKJ', NOW(),
     NOW()),
    -- A2
    ('27htRMGeQ4HV32SPHsJrpndZn1zwmF2kiPqmABcHgehx', '8CfzyaVrqkjZ7MhnYh25Pu1wMvbuNSXGVtUS7StDrq7v', '4NydFEhr', NOW(),
     NOW()),
    -- B1
    ('JAZtFeZfLeeVtWS4vrruCpTa5LdASRDJuMe7yLKbkJk', 'BZjb6u28DoJxpVBiJ87o2UEuwXBo9CXSUS1UTa622vRZ', 'C9N5RpxT', NOW(),
     NOW()),
    -- B2
    ('6HLScqNL4EQWLk8DTcB4hXUrHjDkVbeP2a3Sc5VtHozM', '8ZAMNbJVuyt2W6dV5BnUxT3pm6Us7pZbkHf1zHWjiKJ4', 'juT7Ts2Q', NOW(),
     NOW()),
    -- B3
    ('DTSwNFAzBMTBZbiH1XdzHEpyWSCpVM8hWU6chwKNqJMN', 'AXUnhNJJr7wRt1horZvyHpZp1xe1ajzU1xb8yTpnqPwj', 'SnEQ96xF', NOW(),
     NOW()),
    -- C1
    ('5zp784qrSY1AQXZDaekTi2UESthYfhj122AvLGDRk4UN', 'EnrA1vfEBmdaq3vKa4bieiXMvyvEuUNGcCtZafAWotDM', 'xmfs2UDN', NOW(),
     NOW()),
    -- C2
    ('Ctd4WzJJUC5jzC5AZ6fyYivqAC1JmwnqLBwb7V8GsZHn', '7bVntmeiFJgdrBsi5dZUL53c5FqG5U2ixXTzcUx8YrJN', '3rDPxSbE', NOW(),
     NOW()),
    -- D1
    ('56aGGSM1n3LuxW5zMibVJQV4GwgN9Fma2DQ4vkgwHx3', '3XJBrZqvK9HmXP81ic5Ez9AhGGbTQbQYtsCVmDWhYdm1', 'GvjG5drs', NOW(),
     NOW()),
    -- D2
    ('CiLPZdhd2tt8eE779kMhBNt351PSEFXrp7hUGSQMjzBN', '12GqPrmQthbcM7m79cj7Vz1bvNLkHz6jhdtCYtYAAFyK', 'SgQgzYBU', NOW(),
     NOW()),
    -- E1
    ('4xfHevk5u4bgLj2pgrr4o9MsQtFuuaGRjoLcu6LcKBpA', 'FZDt4BMGy4cSPQKdYg4soUnFuRUC9ZwhN8i8uDf5pjJw', 'Gh2Zg79S', NOW(),
     NOW()),
    -- E2
    ('D4xnHLUKFT9ruLcVz3nLcw31PgK37HrwrP5uXSNgU58t', 'E7MCuLzz8rh5HkVMFzY6WRz7xYBeTfu7eivgjJsrKvHz', '4jdnRxrN', NOW(),
     NOW()),
    -- F1
    ('CkWqUi158vkRLJ9EGtRe1GeA7vM78VTjmchrh4PpPqLJ', 'xTSZBy5T8Udzpw9JRMRcsZJrHwxAE53eLVpWknWdjBx', '9rNWHVrF', NOW(),
     NOW()),
    -- F2
    ('99f8Mz2fVrVPAizcNwmJeXvvrTi1DbMwEgrqXmNL7Udo', 'Co2NTx6XWmNek6n4u3EPRLLWunregDMiEYXi9p1VVKDT', 'ffAk9Qce', NOW(),
     NOW());

INSERT INTO t_invite_relation (inviter, invitee, channel, level, tx_id, created_at, updated_at)
VALUES
    -- leader-5 -> A1
    ('4tJJv3RiQfwL4vdRsdYJK12K1yoirw3b2n1RTZxKacwn', '5D4MWh35wxUcY1hBsm5GwuippPL2UBmfDnfkC8MeqxcN', 'TakeToken', 1,
     '2VdRaYZPPPpjEnXScb2tCfuPihobamjte1qSYvvwFp6N8M5PB2JJP42xXQHcs44Yx1fG4srDzEMBThYNEjkTm9T', NOW(), NOW()),
    -- leader-5 -> B2
    ('4tJJv3RiQfwL4vdRsdYJK12K1yoirw3b2n1RTZxKacwn', '6HLScqNL4EQWLk8DTcB4hXUrHjDkVbeP2a3Sc5VtHozM', 'TakeToken', 1,
     '5RWY2kqTYPEwkUZCQV43dJ5UvBTyRYVtVou3Za9MLBSs1VLE5vEquAyAhKWNm3vpqcHHdfzJ5fK2fYizr3DcAtP', NOW(), NOW()),
    -- leader-7 -> A2
    ('EzehTETE4o8Jua5pCAUk7AHV5Zt3QL1a2D9vfE1ii6HK', '27htRMGeQ4HV32SPHsJrpndZn1zwmF2kiPqmABcHgehx', 'TakeToken', 1,
     '47pzPdyCxDti9XagX88UYENJFtZcAjhwrV7JNvKuuNwvGDw9iv6tdcLyRr7dKMPuKHeGNn4sFH9EJgtmcQjwZhv', NOW(), NOW()),
    -- A1 -> B1
    ('5D4MWh35wxUcY1hBsm5GwuippPL2UBmfDnfkC8MeqxcN', 'JAZtFeZfLeeVtWS4vrruCpTa5LdASRDJuMe7yLKbkJk', 'TakeToken', 1,
     '1v7vL8ZnapTCf9Y3hUkYxTXEK7uwj1o563fHGxZ9ET8jg4Q6nZd35eBxVo6EKhFgxzhx95z4n7fKEzyhgFYkfUE', NOW(), NOW()),
    -- A2 -> B3
    ('27htRMGeQ4HV32SPHsJrpndZn1zwmF2kiPqmABcHgehx', 'DTSwNFAzBMTBZbiH1XdzHEpyWSCpVM8hWU6chwKNqJMN', 'TakeToken', 1,
     'XMcTJzVttvy5MHhAmx49vVSXcmsrfsrD7vZmABRLMm9mCmrDvdmhwnmVEEyKyRbm67md1KwgRnTjWXGWTMomK84', NOW(), NOW()),
    -- B1 -> C1
    ('JAZtFeZfLeeVtWS4vrruCpTa5LdASRDJuMe7yLKbkJk', '5zp784qrSY1AQXZDaekTi2UESthYfhj122AvLGDRk4UN', 'TakeToken', 1,
     '5qa6PtyFbyVVrNiRk9hEPxVUrhq6Fs1fFj9JhWWHKFpQqQUGybWwDuPeLqScrKBCXbzx24YLT7ktxne1ioxThvp', NOW(), NOW()),
    -- B3 -> C2
    ('DTSwNFAzBMTBZbiH1XdzHEpyWSCpVM8hWU6chwKNqJMN', 'Ctd4WzJJUC5jzC5AZ6fyYivqAC1JmwnqLBwb7V8GsZHn', 'TakeToken', 1,
     '3MHMFVcaBE6jssTJJFPuEUkKZyRppirhZWcMAebk84QvX23XfLGckKTnTYFZ7CckQ6A1PiSXYUXwLbmPPrzRK32', NOW(), NOW()),
    -- C1 -> D1
    ('5zp784qrSY1AQXZDaekTi2UESthYfhj122AvLGDRk4UN', '56aGGSM1n3LuxW5zMibVJQV4GwgN9Fma2DQ4vkgwHx3', 'TakeToken', 1,
     '3b9iAXmRtyMb8VBwLM5BhggsT2JzJE5iQoQfGTwCHeRijzMoxBbMH21oN5uMo1Mfs2jvKmD2onZQy4KKGv1DFxt', NOW(), NOW()),
    -- C2 -> D2
    ('Ctd4WzJJUC5jzC5AZ6fyYivqAC1JmwnqLBwb7V8GsZHn', 'CiLPZdhd2tt8eE779kMhBNt351PSEFXrp7hUGSQMjzBN', 'TakeToken', 1,
     '5uZFsvEWLoWUqP2bxrqDQmVs8qYddeN4xpvafcVePadTtPgsFCAnhkPj2YwDqQoVQaj84PHacUXxGXgmmwXLk1r', NOW(), NOW()),
    -- D1 -> E1
    ('56aGGSM1n3LuxW5zMibVJQV4GwgN9Fma2DQ4vkgwHx3', '4xfHevk5u4bgLj2pgrr4o9MsQtFuuaGRjoLcu6LcKBpA', 'TakeToken', 1,
     '42jwZR1d3iZYe1tP9yrXsHtrLfFUTkYKmzSByRszetuiwmCpu3SVojTWwMUKdNxfX8NDjvfP1YTC1s9x9yrYbGk', NOW(), NOW()),
    -- D1 -> E2
    ('56aGGSM1n3LuxW5zMibVJQV4GwgN9Fma2DQ4vkgwHx3', 'D4xnHLUKFT9ruLcVz3nLcw31PgK37HrwrP5uXSNgU58t', 'TakeToken', 1,
     '2ZoV9oHbjawm6oMWo5PdnckFKFDQ9Zy8kwoznrJr2zBBWNgvUS8djf4ttiMhf9nzMA8bpMFKxvov51sQ5VCdvZ3', NOW(), NOW()),
    -- E1 -> F2
    ('4xfHevk5u4bgLj2pgrr4o9MsQtFuuaGRjoLcu6LcKBpA', '99f8Mz2fVrVPAizcNwmJeXvvrTi1DbMwEgrqXmNL7Udo', 'TakeToken', 1,
     '3w9PDV3SWRhp1Tzs2prJt4zDucqPYBSU2pJaJ1EL3fTXPLbCSWuibLDw1SxPaUYUeii9WCFR1QEk3j6jsKqWmWi', NOW(), NOW()),
    -- E2 -> F1
    ('D4xnHLUKFT9ruLcVz3nLcw31PgK37HrwrP5uXSNgU58t', 'CkWqUi158vkRLJ9EGtRe1GeA7vM78VTjmchrh4PpPqLJ', 'TakeToken', 1,
     '5NNTJUEsvjiDYkJDi3UjFEiaLKjT2KLycUi6HxFZBEVA8YHVicGwCXzgBjVtJvHgY3xRzBMKezaaTYz2YMtBEX2', NOW(), NOW());


-- 重新计算邀请层级关系
-- 创建一个递归查询，计算层级关系
with recursive invite_hierarchy as (
    -- 基层的邀请人，默认 DistLevel 为 1
    select
        record_id,
        inviter,
        invitee,
        1 as level
    from
        t_invite_relation
    where
        inviter not in (select invitee from t_invite_relation)

    union all

    -- 递归查找下一级邀请关系
    select
        r.record_id,
        r.inviter,
        r.invitee,
        ih.level + 1 as level
    from
        t_invite_relation r
            inner join
        invite_hierarchy ih
        on
            r.inviter = ih.invitee
)
-- 更新原表中的 DistLevel 字段
update t_invite_relation as t
set level = ih.level
from invite_hierarchy ih
where t.record_id = ih.record_id;



