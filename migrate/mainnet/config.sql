-- 基础配置
delete from t_system_config;
insert into t_system_config(env)values(1);

-- aws配置
delete from t_aws_config;
insert into t_aws_config
(access_key_id,secret_access_key,region)
values
    ('ENC~qZzWVnJKhsibFFb6sbwUfrudldqC7aZ7G0dudOAzpAgRRQa7xsy/dNrZo1I24ArQjBLafJQQdoJPVN5oV7xBUw==','ENC~o27D5AfD3bXjMfL8UM1kU9Qa+s0AsllVHwM6frrOYG/xcR9z5tb4v79FuMlUqxq9zoc66Cu4+L/TqEIRwUIdUQlsMLjxPK3LK3GimRC0jw0=','ap-southeast-1');

-- rpc 配置
INSERT INTO public.t_rpc_endpoint
    (scope, provider, endpoint, api_key, wss_endpoint, wss_api_key, weight)
VALUES
    ('env', 'helius', 'https://mainnet.helius-rpc.com', 'ENC~N7DDSiL65pIERl8FFuSWg1XrSmFk6I64zLwXDR3SGwJliWO3wZUsYervWKVLZoHrHI4BGLqbIU6C/rbGHG0aqjhJyrXm7LMh8/WOjwfzYCY=', 'wss://mainnet.helius-rpc.com', 'ENC~N7DDSiL65pIERl8FFuSWg1XrSmFk6I64zLwXDR3SGwJliWO3wZUsYervWKVLZoHrHI4BGLqbIU6C/rbGHG0aqjhJyrXm7LMh8/WOjwfzYCY=', 1),
    ('mainnet', 'helius', 'https://mainnet.helius-rpc.com', 'ENC~N7DDSiL65pIERl8FFuSWg1XrSmFk6I64zLwXDR3SGwJliWO3wZUsYervWKVLZoHrHI4BGLqbIU6C/rbGHG0aqjhJyrXm7LMh8/WOjwfzYCY=', 'wss://mainnet.helius-rpc.com', 'ENC~N7DDSiL65pIERl8FFuSWg1XrSmFk6I64zLwXDR3SGwJliWO3wZUsYervWKVLZoHrHI4BGLqbIU6C/rbGHG0aqjhJyrXm7LMh8/WOjwfzYCY=', 1);

delete from t_chain_config;
insert into t_chain_config
(chain_name,decimals,symbol)
values
    ('solana',9,'SOL');

delete from t_token_config;
insert into t_token_config
(token_name,token_symbol,decimals,mint)
values
    ('OShit','OShit',3,'ShitJuMfPKCQU7LedLERFYapDta7CCdKExPWX2gETRH');

delete from t_fee_tolerance;
insert into t_fee_tolerance(max_less_rate)values(0.05);

-- 业务配置
delete from t_service_info;

insert into t_service_info
    (service,sub_service,address,webhook,mq_group,mq_topic,hook_type,tx_source,confirm,multi_sign,created_at,updated_at)
values
    ('reward','take token','4fLi61UtmjtKXJHL1aKb1cdUv3KZfCVnQv45hPN8h3Sf','','reward','ServiceTransaction',0,0,true,true,now(),now());

insert into t_service_info
    (service,sub_service,address,webhook,mq_group,mq_topic,hook_type,tx_source,confirm,multi_sign,created_at,updated_at)
values
    ('reward','give token','5xaFm7Xj2kFvjFmW3kFU2A9Ush6Hx3mqK86cYo6aJsbY','','reward','ServiceTransaction',0,0,true,true,now(),now());

insert into t_service_info
    (service,sub_service,address,webhook,mq_group,mq_topic,hook_type,tx_source,confirm,multi_sign,created_at,updated_at)
values
    ('reward','lottery','4fLi61UtmjtKXJHL1aKb1cdUv3KZfCVnQv45hPN8h3Sf','','reward','ServiceTransaction',0,0,true,true,now(),now());

insert into t_service_info
    (service,sub_service,address,webhook,mq_group,mq_topic,hook_type,tx_source,confirm,multi_sign,created_at,updated_at)
values
    ('reward','reward code','3WwtCT8m2rrjTHK1hgL47eBNZd1KNZQZNvCZhiVsCbN3','','reward','ServiceTransaction',0,0,true,true,now(),now());

insert into t_service_info
    (service,sub_service,address,webhook,mq_group,mq_topic,hook_type,tx_source,confirm,multi_sign,created_at,updated_at)
values
    ('reward','campaign quote','4e1QkZ6mM3bXskMqESVKDs1GRgPGEyVQyPXjWM8GoTjS','','reward','ServiceTransaction',0,0,true,true,now(),now());

insert into t_service_info
    (service,sub_service,address,webhook,mq_group,mq_topic,hook_type,tx_source,confirm,multi_sign,created_at,updated_at)
values
    ('pos','pos reward','H61Yvf3aZxW56hBGDNxT4cyguj98gSZMowzetYhc3TCB','','pos','ServiceTransaction',0,0,true,true,now(),now());

insert into t_service_info
    (service,sub_service,address,webhook,mq_group,mq_topic,hook_type,tx_source,confirm,multi_sign,created_at,updated_at)
values
    ('pos','stake token','89GUQB4BSn1G9ezeXMnDPm7bTPngNAcVxMn53VEpWYaA','','Stake','ServiceTransaction',0,0,true,true,now(),now());

insert into t_service_info
    (service,sub_service,address,webhook,mq_group,mq_topic,hook_type,tx_source,confirm,multi_sign,created_at,updated_at)
values
    ('pos','stake reward','EgFd9DYSF6y6sSPDXzpzQsi2Qo4srCJCqDeSHsUpveej','','Stake','ServiceTransaction',0,0,true,true,now(),now());

insert into t_service_info
    (service,sub_service,address,webhook,mq_group,mq_topic,hook_type,tx_source,confirm,multi_sign,created_at,updated_at)
values
    ('pos','stake leader reward','CuQ885ndc1jzTWE5d21GTGVPLJvGxfRBVywwvUaRCjTj','','Stake','ServiceTransaction',0,0,true,true,now(),now());

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
    ('reward','take token','fp2aXI50Gn9L+DFv+23PtzUmQ/gnFo5zT3l3vSdpbNtG91Sf1F155/TaQUGDZDpng9XVEJALcbP1zGZvfBYFWBTCSKUF0eCHGjd4BXPNtGV/g+/ozels210NcJZ7Yw/uyr4wjjQZ8MLo3XmBayeC7WfSiHjhTveXubxXDx84A+g=',NOW(),NOW());

insert into t_service_key
    (service,sub_service,encrypted_key,created_at,updated_at)
values
    ('reward','give token','qLoVZIgBQbNfhjoIs8553sQthOH6Jg3bcPXSz4NQK3wAh+x9w6r57vxFeUU6sG7IyLxFFXMy6h7/57jkWIiJdEkw+HD9xlEq7OuHxzAHw5V7InKdWJdaLh1DgsnJzu7M6SpLPLAe9+9tW4zyhet/WEWC1e/rBf9phh0mO5vi2Ks=',NOW(),NOW());

insert into t_service_key
    (service,sub_service,encrypted_key,created_at,updated_at)
values
    ('reward','lottery','fp2aXI50Gn9L+DFv+23PtzUmQ/gnFo5zT3l3vSdpbNtG91Sf1F155/TaQUGDZDpng9XVEJALcbP1zGZvfBYFWBTCSKUF0eCHGjd4BXPNtGV/g+/ozels210NcJZ7Yw/uyr4wjjQZ8MLo3XmBayeC7WfSiHjhTveXubxXDx84A+g=',NOW(),NOW());

insert into t_service_key
    (service,sub_service,encrypted_key,created_at,updated_at)
values
    ('reward','reward code','JhypgxF6A4NGrkf+qlLVZ7Y+f3dlT9V6QeGuI5X1dIJCBr/+T+H2gZjexktCWY0ZrDpKBdB9q4loSR8eHL5fAphrWIDC92oNvh1qWiaR/Ot5WAkyUhXH9eCd7EQxvNlLV4xX01EMeOaWLLKvkaQutyJb3i4Xlogpc5JyNVBB4pk=',NOW(),NOW());

insert into t_service_key
    (service,sub_service,encrypted_key,created_at,updated_at)
values
    ('reward','campaign quote','bXwH8yoaIQCO13Nt71akdcaMpUiRStamxsJg5lOQWDwI0Bl4pYuvKm2aaRRGxj6Pk+WR6dEqfdEm+j6ZlG+dgXBpMXLQCn5wV3QPzfwt3p2znYxh1bRkZb8okQzhAhA5UmcluVXUgFr5EtSIOltc7lgL1UjgBhFscUrQkKDO4s4=',NOW(),NOW());

insert into t_service_key
    (service,sub_service,encrypted_key,created_at,updated_at)
values
    ('pos','pos reward','L0IMgCTkXeqtoZGPYBvD7iedzO1ksEv+L78tgIZ0hmU/8dZqtW2HksiKV2gPV45DIATMflXgEISfbv0fTeOl5gqDJ3FzW3K+feYAruJ6DMBb8/MrttpobaTEd4aHx8SGaMgdNhFRets0/tAJ184DY9eH6XwGIvF7dsxXk14do68=',NOW(),NOW());

insert into t_service_key
    (service,sub_service,encrypted_key,created_at,updated_at)
values
    ('pos','stake token','yN3U4PDyjRQg/S7ca/Wa+gZ4qfiztb6kxksRlwiOGPX4yjVmVyCG0QYDAH51awiF8RvkLJjEseL3YDJEZUCr3reAqOcjFhhSwggl5aSWCg00ttkU+nlgS+O9Z3aAoAbXZNpGVIpsRMAf+5Ii7Oqslf7OfnMO6YUbiZeTMts1aaU=',NOW(),NOW());

insert into t_service_key
    (service,sub_service,encrypted_key,created_at,updated_at)
values
    ('pos','stake reward','YDOe0++F4CK10+Lih4QpO0a6MNIouIN6oUexLBcEGHC4QK4zZvTLs2NTbMAa4zT8CK7cQRB5TIYDCx57n3sWHD4VIyjBoPoFmmJvu2+e+uQUv+VX4RtxbaYxVH7ot5rt/yKD8cF23lG+S/iavh5jD9MOQ5AHHZpmyW5h9YIATFc=',NOW(),NOW());

insert into t_service_key
    (service,sub_service,encrypted_key,created_at,updated_at)
values
    ('pos','stake leader reward','TlfmRrO93wUoFUmCtqVGWoDb79ivGYjx68ZSA15O/8BYOTqwljEQki7UpmSNQ0eSmExoayYirRg+hUptNNUU3YU3v3lpDnr5d1o+WLNfNYd26aWpzxoBoVWJhtAErs0FuF8+DddhYwbzyp2nhKNzvWvqNIZq1k4iA/Ymw8Slr94=',NOW(),NOW());

insert into t_service_key
    (service,sub_service,encrypted_key,created_at,updated_at)
values
    ('gs-relay','contribution','vSIx1nk/mc5I2caCC3JZ66OOychaleBkz6mPduAXm9YuO8CUszhwnuyVaczUKg3QuzIBani5ahqGao8jObY/RdY7xaXur4vyP9n/kkF6yIKHu+RteGpi6uwqgmFlcLumwkWR93eLrWXaNtjSawq1PWyskDaKFJmz18R34lKExwc=',NOW(),NOW());

insert into t_service_key
    (service,sub_service,encrypted_key,created_at,updated_at)
values
    ('gs-relay','payout','vSIx1nk/mc5I2caCC3JZ66OOychaleBkz6mPduAXm9YuO8CUszhwnuyVaczUKg3QuzIBani5ahqGao8jObY/RdY7xaXur4vyP9n/kkF6yIKHu+RteGpi6uwqgmFlcLumwkWR93eLrWXaNtjSawq1PWyskDaKFJmz18R34lKExwc=',NOW(),NOW());

insert into t_service_key
    (service,sub_service,encrypted_key,created_at,updated_at)
values
    ('gs-relay','refund','vSIx1nk/mc5I2caCC3JZ66OOychaleBkz6mPduAXm9YuO8CUszhwnuyVaczUKg3QuzIBani5ahqGao8jObY/RdY7xaXur4vyP9n/kkF6yIKHu+RteGpi6uwqgmFlcLumwkWR93eLrWXaNtjSawq1PWyskDaKFJmz18R34lKExwc=',NOW(),NOW());


-- 业务交易扫描表
delete from t_tx_scan_info;

INSERT INTO public.t_tx_scan_info
    (service, sub_service, native_account, pda_account, until_tx_id, before_tx_id, slot, created_at, updated_at)
VALUES
    ('reward', 'take token', '4fLi61UtmjtKXJHL1aKb1cdUv3KZfCVnQv45hPN8h3Sf', '4LQDhhNW9Nqyo7DTBDszs97RSQwSr648hsbGt85hw5Cf', '5AJ5QqLtSnt3drpXR636BKfzQJC1gYLAH6FeCFnfdVsAmRTeAaEEdSj5RFyKx8EygpGs6y4ph7VAvLqX1wzWMWuU', '', '426019453', NOW(),NOW());

INSERT INTO public.t_tx_scan_info
    (service, sub_service, native_account, pda_account, until_tx_id, before_tx_id, slot, created_at, updated_at)
VALUES
    ('reward', 'give token', '5xaFm7Xj2kFvjFmW3kFU2A9Ush6Hx3mqK86cYo6aJsbY', '9PEVsQ6Fnfgui5wpsXPSHa3L8byc3UvgZSc193ShUUzK', '5s6onUbLB8F5KTRfWagHScR76S2dG2RNhZXpPLvwBcb4MEQ7yVXKNgjxUedp6cxJmcwBQ9JCa7to1cTt9LPQVYnk', '', '423759492', NOW(),NOW());

INSERT INTO public.t_tx_scan_info
    (service, sub_service, native_account, pda_account, until_tx_id, before_tx_id, slot, created_at, updated_at)
VALUES
    ('reward', 'lottery', '4fLi61UtmjtKXJHL1aKb1cdUv3KZfCVnQv45hPN8h3Sf', '4LQDhhNW9Nqyo7DTBDszs97RSQwSr648hsbGt85hw5Cf', '5AJ5QqLtSnt3drpXR636BKfzQJC1gYLAH6FeCFnfdVsAmRTeAaEEdSj5RFyKx8EygpGs6y4ph7VAvLqX1wzWMWuU', '', '426019453', NOW(),NOW());

INSERT INTO public.t_tx_scan_info
    (service, sub_service, native_account, pda_account, until_tx_id, before_tx_id, slot, created_at, updated_at)
VALUES
    ('reward', 'reward code', '3WwtCT8m2rrjTHK1hgL47eBNZd1KNZQZNvCZhiVsCbN3', 'F1oVN6C1gRDm3NMr5ZWgMTQxaEXZDhxsg51CVGAThSLT', '5gW1EPX3dQJQs3bQkB7ru6GZD4s2pmnGJnsz5Aaoz826UdcEXDwgY5hyRq4XV6Q4zChAE2ckZqFqgkAA5D8BFkyc', '', '425937734', NOW(),NOW());

INSERT INTO public.t_tx_scan_info
    (service, sub_service, native_account, pda_account, until_tx_id, before_tx_id, slot, created_at, updated_at)
VALUES
    ('reward', 'campaign quote', '4e1QkZ6mM3bXskMqESVKDs1GRgPGEyVQyPXjWM8GoTjS', '7JKwZqRsx76an4eMUWUFWwAisuhhE8y5Zv396HNcsRGo', '5TDNYWJMvwWXATKEvjxeyhZxzvSR6wJCQqatBWqnrmj44MU76ETHbzc4eacSCCS4fxSBYSQMMyX6wqWw1N4YRqGK', '', '426005659', NOW(),NOW());

INSERT INTO public.t_tx_scan_info
    (service, sub_service, native_account, pda_account, until_tx_id, before_tx_id, slot, created_at, updated_at)
VALUES
    ('pos', 'pos reward', 'H61Yvf3aZxW56hBGDNxT4cyguj98gSZMowzetYhc3TCB', 'EZGgg2zSGCbuy7KxjoYUvMWGw3mPdavGjFhYmD6sr4zz', '39jfHgZYReJdt5stMUNnGGSWtVuzrRwEfsSfMSYGoqFXSMW8CrbiwEQD52ejeJPFPf2AtkC49tGQ1ybTH4o973kB', '', '426019496', NOW(),NOW());

INSERT INTO public.t_tx_scan_info
    (service, sub_service, native_account, pda_account, until_tx_id, before_tx_id, slot, created_at, updated_at)
VALUES
    ('pos', 'market buy token', 'HtNfUbDaBamCBPWCFiESkXpewvwVLkwrSWRjjV8FNT7i', 'HtNfUbDaBamCBPWCFiESkXpewvwVLkwrSWRjjV8FNT7i', '2ChwRQku5kTsBjysDQsEh8jmW8amwWcq3Wuw1sx8Da48tNSM3twVCXX6nqcGyLcFxW3jPJCmpRLb8AU4SsCqS2ck', '', '426020465', NOW(),NOW());

INSERT INTO public.t_tx_scan_info
    (service, sub_service, native_account, pda_account, until_tx_id, before_tx_id, slot, created_at, updated_at)
VALUES
    ('pos', 'stake token', '89GUQB4BSn1G9ezeXMnDPm7bTPngNAcVxMn53VEpWYaA', '89GUQB4BSn1G9ezeXMnDPm7bTPngNAcVxMn53VEpWYaA', 'mPt1XDV9Yx4zmDmoMFk81mAyPrXppbEbLDKATTrYLSKArpADQeFnvbhVKj9vPkPmm8Z81i2fLiu46CQVK8y69tm', '', '425971442', NOW(),NOW());

INSERT INTO public.t_tx_scan_info
    (service, sub_service, native_account, pda_account, until_tx_id, before_tx_id, slot, created_at, updated_at)
VALUES
    ('pos', 'stake reward', 'EgFd9DYSF6y6sSPDXzpzQsi2Qo4srCJCqDeSHsUpveej', 'AwXCN4QbLoo1q558o7qAfKRDqJeqK3GLCDrH5F5eF1TB', '3s19Sg3HT2x243kvZHSnxXeqtae4uzKNfi16a9xjHubpDiWbx7oR8c34T6okw9QY9zAs1SDgKpLhSKyuMipCVzQY', '', '426019460', NOW(),NOW());

INSERT INTO public.t_tx_scan_info
    (service, sub_service, native_account, pda_account, until_tx_id, before_tx_id, slot, created_at, updated_at)
VALUES
    ('pos', 'stake leader reward', 'CuQ885ndc1jzTWE5d21GTGVPLJvGxfRBVywwvUaRCjTj', 'BxuVu6mqtjsmouawujN4yJQVZfV9snToCoykofi6i75Q', '4LKu2ntP5pWrhNqp3PaKHr7KXccQRCE8dmzMFTLBbSrKAonTkn19cbM7N34Khhb5WKt62WSQzDCpnj3eBkTP25Xh', '', '426016493', NOW(),NOW());

INSERT INTO public.t_tx_scan_info
    (service, sub_service, native_account, pda_account, until_tx_id, before_tx_id, slot, created_at, updated_at)
VALUES
    ('gs-relay', 'contribution', 'AUq3iXbJjBjED6ZN2JqJDdZ3mknd7cXv9HQqNdYuEZ6q', 'E2EauLpUaKMJz5hxJ3eJrwNfrBHgSa7nsDrCLBX2eXsx', '5xJDwFhzR1d9h2BLncVKUUCKdwwYdY38HNKRMiXFPPYRMMkq83Y4AMZJQvCqRnDXJkQSSanb7wEzNWritHCNLvkh', '', '467262544', NOW(),NOW());

INSERT INTO public.t_tx_scan_info
    (service, sub_service, native_account, pda_account, until_tx_id, before_tx_id, slot, created_at, updated_at)
VALUES
    ('gs-relay', 'payout', 'AUq3iXbJjBjED6ZN2JqJDdZ3mknd7cXv9HQqNdYuEZ6q', 'E2EauLpUaKMJz5hxJ3eJrwNfrBHgSa7nsDrCLBX2eXsx', '5xJDwFhzR1d9h2BLncVKUUCKdwwYdY38HNKRMiXFPPYRMMkq83Y4AMZJQvCqRnDXJkQSSanb7wEzNWritHCNLvkh', '', '467262544', NOW(),NOW());

INSERT INTO public.t_tx_scan_info
    (service, sub_service, native_account, pda_account, until_tx_id, before_tx_id, slot, created_at, updated_at)
VALUES
    ('gs-relay', 'refund', 'AUq3iXbJjBjED6ZN2JqJDdZ3mknd7cXv9HQqNdYuEZ6q', 'E2EauLpUaKMJz5hxJ3eJrwNfrBHgSa7nsDrCLBX2eXsx', '5xJDwFhzR1d9h2BLncVKUUCKdwwYdY38HNKRMiXFPPYRMMkq83Y4AMZJQvCqRnDXJkQSSanb7wEzNWritHCNLvkh', '', '467262544', NOW(),NOW());






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
    (NULL,'4fLi61UtmjtKXJHL1aKb1cdUv3KZfCVnQv45hPN8h3Sf', '4DZ3ry8LfaM3kEc6KCL4nDpNpnVX6g4Z6EFvYyNXsiqn', 500000, 1500000, 200,20,75000,true,true,true,NOW(),NOW());

-- give token 配置
delete from public.t_give_token_config;
insert into public.t_give_token_config
    (reward_account,cost_account,max_reward,max_valid_reward,reward_rate,valid_rate,created_at,updated_at)
values
    ('5xaFm7Xj2kFvjFmW3kFU2A9Ush6Hx3mqK86cYo6aJsbY','4DZ3ry8LfaM3kEc6KCL4nDpNpnVX6g4Z6EFvYyNXsiqn',500000,1500000,200,300,NOW(),NOW());

-- lottery 配置
DELETE FROM t_lottery_config;
INSERT INTO t_lottery_config
    (reward_account,cost_account,cost_amount,cost_fee_rate,created_at,updated_at)
VALUES
    ('4fLi61UtmjtKXJHL1aKb1cdUv3KZfCVnQv45hPN8h3Sf', '4DZ3ry8LfaM3kEc6KCL4nDpNpnVX6g4Z6EFvYyNXsiqn',500000,110,NOW(),NOW());

-- reward code
delete from t_reward_code_config;
INSERT INTO public.t_reward_code_config
    (reward_account,cost_account)
VALUES
    ('3WwtCT8m2rrjTHK1hgL47eBNZd1KNZQZNvCZhiVsCbN3','4DZ3ry8LfaM3kEc6KCL4nDpNpnVX6g4Z6EFvYyNXsiqn');

delete from t_reward_code_fee;
INSERT INTO public.t_reward_code_fee(amount,fee_rate)VALUES(5000000,8);
INSERT INTO public.t_reward_code_fee(amount,fee_rate)VALUES(1000000,20);
INSERT INTO public.t_reward_code_fee(amount,fee_rate)VALUES(500000,30);

-- campaign 配置
delete from t_campaign_quote_config;
INSERT INTO t_campaign_quote_config
    (reward_account,cost_account,quote_rate,cost_rate)
VALUES
    ('4e1QkZ6mM3bXskMqESVKDs1GRgPGEyVQyPXjWM8GoTjS','4DZ3ry8LfaM3kEc6KCL4nDpNpnVX6g4Z6EFvYyNXsiqn',500,17);

-- pos奖励配置
delete from t_pos_reward_config;
INSERT INTO t_pos_reward_config
    (reward_account,cost_account,cost_fee_rate,max_cost_fee,quote_token_amount,created_at,updated_at)
VALUES
    ('H61Yvf3aZxW56hBGDNxT4cyguj98gSZMowzetYhc3TCB','4DZ3ry8LfaM3kEc6KCL4nDpNpnVX6g4Z6EFvYyNXsiqn',200,40000,500000,NOW(),NOW());

-- pos星级配置
delete from t_pos_star_level_rule;
INSERT INTO t_pos_star_level_rule(amount,group_amount,star_level,rate,created_at,updated_at)VALUES(3000000000,30000000000,1,10,NOW(),NOW());
INSERT INTO t_pos_star_level_rule(amount,group_amount,star_level,rate,created_at,updated_at)VALUES(10000000000,100000000000,2,20,NOW(),NOW());
INSERT INTO t_pos_star_level_rule(amount,group_amount,star_level,rate,created_at,updated_at)VALUES(30000000000,300000000000,3,30,NOW(),NOW());
INSERT INTO t_pos_star_level_rule(amount,group_amount,star_level,rate,created_at,updated_at)VALUES(150000000000,3000000000000,4,40,NOW(),NOW());
INSERT INTO t_pos_star_level_rule(amount,group_amount,star_level,rate,created_at,updated_at)VALUES(1000000000000,20000000000000,5,50,NOW(),NOW());

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

-- pos 质押白名单
delete from t_pos_star_whitelist;
INSERT INTO t_pos_star_whitelist (native_account,star_level,rate) VALUES('G5jNET7pyDyEcNdc9MENSr3q6m3aUp5UvMcufHpdFZpu',6,60);
INSERT INTO t_pos_star_whitelist (native_account,star_level,rate) VALUES('24nkGFjBxxd3LxurZSagMoypNqedfjpeBQmQheTBeRkN',5,50);
INSERT INTO t_pos_star_whitelist (native_account,star_level,rate) VALUES('EMzETjbAjEpc9mosncgySeRUMtayNVLrifsZTwafB9jh',4,40);
INSERT INTO t_pos_star_whitelist (native_account,star_level,rate) VALUES('ERagFfpNQqMZ2838ENKMucdBF8pF4nUmfditxnivhwMb',4,40);
INSERT INTO t_pos_star_whitelist (native_account,star_level,rate) VALUES('8GyzTtq3LBh5pa1PamsuDksngWTWrXJaw5bSeXKAZjAZ',3,30);
INSERT INTO t_pos_star_whitelist (native_account,star_level,rate) VALUES('67nCqDhuwNCToADcNRBhAi4gxBf4dXvAWfS6WqSU33Ct',4,40);

-- 质押AMM配置
delete from t_stake_amm_config;
insert into public.t_stake_amm_config(quote_token,public_key)values('solana','GjkvqFpZ5gqbzYEUAGsn5ozmFgM52JJDgso426DiLXbQ');

-- 质押池配置
delete from t_stake_token_pool;
insert into public.t_stake_token_pool(source_account,from_token_account)values('Raydium','Gh6MjRrJFBU9HcYMYKBbaD8fX1dv3CjGtDDhtVThD9v3');

-- 质押每日固定利息
delete from t_stake_fix_rate_config;
insert into public.t_stake_fix_rate_config(min_amount,stake_type,rate_tier,fix_rate,individual_rate,created_at,updated_at)values(100000000,0,1,70,70,now(),now());
insert into public.t_stake_fix_rate_config(min_amount,stake_type,rate_tier,fix_rate,individual_rate,created_at,updated_at)values(100000000,0,2,63,63,now(),now());
insert into public.t_stake_fix_rate_config(min_amount,stake_type,rate_tier,fix_rate,individual_rate,created_at,updated_at)values(100000000,1,1,100,100,now(),now());
insert into public.t_stake_fix_rate_config(min_amount,stake_type,rate_tier,fix_rate,individual_rate,created_at,updated_at)values(100000000,1,2,90,90,now(),now());

-- 邀请奖励级别
delete from t_stake_invite_dist;
insert into public.t_stake_invite_dist(dist_level)values(2);

-- 邀请奖励每个级别的奖励费率
delete from t_stake_invite_rate;
insert into public.t_stake_invite_rate(dist_level,rate)values(1,10);
insert into public.t_stake_invite_rate(dist_level,rate)values(2,5);

-- 质押星级配置
delete from t_stake_star_level_rule;
insert into public.t_stake_star_level_rule(amount,group_amount,star_level,rate,created_at,updated_at)values(3000000000,30000000000,1,8,now(),now());
insert into public.t_stake_star_level_rule(amount,group_amount,star_level,rate,created_at,updated_at)values(10000000000,100000000000,2,16,now(),now());
insert into public.t_stake_star_level_rule(amount,group_amount,star_level,rate,created_at,updated_at)values(30000000000,300000000000,3,24,now(),now());
insert into public.t_stake_star_level_rule(amount,group_amount,star_level,rate,created_at,updated_at)values(150000000000,3000000000000,4,32,now(),now());
insert into public.t_stake_star_level_rule(amount,group_amount,star_level,rate,created_at,updated_at)values(1000000000000,20000000000000,5,40,now(),now());
insert into public.t_stake_star_level_rule(amount,group_amount,star_level,rate,created_at,updated_at)values(1500000000000,30000000000000,6,48,now(),now());

-- 质押奖励发放配置表
delete from t_stake_reward_config;
insert into public.t_stake_reward_config
(program_id,reward_account,cost_account,quote_token_amount,cost_fee_rate,created_at,updated_at)
values
    ('89GUQB4BSn1G9ezeXMnDPm7bTPngNAcVxMn53VEpWYaA','EgFd9DYSF6y6sSPDXzpzQsi2Qo4srCJCqDeSHsUpveej','4DZ3ry8LfaM3kEc6KCL4nDpNpnVX6g4Z6EFvYyNXsiqn',500000,110,now(),now());

-- 插入质押白名单
delete from t_stake_star_whitelist;
INSERT INTO t_stake_star_whitelist(native_account,star_level,rate)VALUES('EMzETjbAjEpc9mosncgySeRUMtayNVLrifsZTwafB9jh',4,32);
INSERT INTO t_stake_star_whitelist(native_account,star_level,rate)VALUES('Gwt3n495KMYnwyceHbAvHLGYtcZVKVFX9cJvkts9qUdX',5,40);
INSERT INTO t_stake_star_whitelist(native_account,star_level,rate)VALUES('G5jNET7pyDyEcNdc9MENSr3q6m3aUp5UvMcufHpdFZpu',6,48);
INSERT INTO t_stake_star_whitelist(native_account,star_level,rate)VALUES('DQZMZTmFWstjc1f5pGgr829gggWkcUkFDzFgmwKCY2dM',1,8);
INSERT INTO t_stake_star_whitelist(native_account,star_level,rate)VALUES('2ueDjAgybGjbdbgjKLDnc7KnUr5RVUSQCdLbTZyyFKWZ',1,8);
INSERT INTO t_stake_star_whitelist(native_account,star_level,rate)VALUES('5CVvAyCBiBqGjE1DJQANskmL9W7DVmLRctaVW4j3Rkgj',1,8);
INSERT INTO t_stake_star_whitelist(native_account,star_level,rate)VALUES('EgqrVc7LT4w9nuWTwMZ1bF9E15kiBK6rH2wSH4ywnvmC',2,16);
INSERT INTO t_stake_star_whitelist(native_account,star_level,rate)VALUES('7DLizFp8GTfB2Z6b5C9uEn5mQZc51CSECkc6Bn5ZfmsV',1,8);

-- 插入区域经理配置
delete from t_stake_leader_reward_config;
insert into public.t_stake_leader_reward_config(reward_account)values('CuQ885ndc1jzTWE5d21GTGVPLJvGxfRBVywwvUaRCjTj');

-- 插入总区域经理配置
delete from t_stake_total_leader;
INSERT INTO t_stake_total_leader (native_account, stake_share, created_at, updated_at) VALUES ('Gwt3n495KMYnwyceHbAvHLGYtcZVKVFX9cJvkts9qUdX', 7, '2026-04-05 17:34:19.128958', '2026-04-05 17:34:19.128958');
INSERT INTO t_stake_total_leader (native_account, stake_share, created_at, updated_at) VALUES ('HjvdyZXukoDYFjoPLnwNZaWuUwnzZuHKyj8HgRWMyYMZ', 3, '2026-04-05 17:34:19.130259', '2026-04-05 17:34:19.130259');

-- 插入区域经理配置
delete from t_stake_leader;
INSERT INTO t_stake_leader(native_account,leader_level,up_leader)VALUES('24nkGFjBxxd3LxurZSagMoypNqedfjpeBQmQheTBeRkN',2,null);
INSERT INTO t_stake_leader(native_account,leader_level,up_leader)VALUES('32GACeJW8bxE9AxNiCHiaJWwhchEYiDgaWtG7swQWPnE',2,null);
INSERT INTO t_stake_leader(native_account,leader_level,up_leader)VALUES('8VLa41iRh4PaALH149LJGao3fXNphH2vE92BcUsSFJmw',2,null);
INSERT INTO t_stake_leader(native_account,leader_level,up_leader)VALUES('FVimu1UZGspPcZ7BKzFvayukYBhEL2ZKdA2SDYmV1fEz',2,null);
INSERT INTO t_stake_leader(native_account,leader_level,up_leader)VALUES('9PqbpkmcxxzzadBtUE3ddXVaHC1t5JN2PYKMNzRh4gUV',2,null);
INSERT INTO t_stake_leader(native_account,leader_level,up_leader)VALUES('3ojkzguhJA2f9JTM8se3gq6CsipsJTedUs5mLfF1QSd7',2,null);

INSERT INTO t_stake_leader(native_account,leader_level,up_leader)VALUES('D5MC5otX5ch1csg723SS38FoSxBoFaqDjw4EqDdzmG8Q',1,'3ojkzguhJA2f9JTM8se3gq6CsipsJTedUs5mLfF1QSd7');
INSERT INTO t_stake_leader(native_account,leader_level,up_leader)VALUES('CQGhfR1xv7cMqVQ2GiLPgf6NXiCnUqhbAqGFFGuE7oD4',1,'3ojkzguhJA2f9JTM8se3gq6CsipsJTedUs5mLfF1QSd7');
INSERT INTO t_stake_leader(native_account,leader_level,up_leader)VALUES('AZ2zpC41iXWXPaK2CgEDf62c8oSWdLVTp8ZcELhjVQWH',1,'3ojkzguhJA2f9JTM8se3gq6CsipsJTedUs5mLfF1QSd7');
INSERT INTO t_stake_leader(native_account,leader_level,up_leader)VALUES('2izq8DYVoHKk1wf636AbiqsaDK6GxoaX1gUZxmawdrgA',1,'3ojkzguhJA2f9JTM8se3gq6CsipsJTedUs5mLfF1QSd7');
INSERT INTO t_stake_leader(native_account,leader_level,up_leader)VALUES('DARnMkBvvwrKqX6niYyfYvJW8JJNNzPwYRixSRgjcYYP',1,'3ojkzguhJA2f9JTM8se3gq6CsipsJTedUs5mLfF1QSd7');

INSERT INTO t_stake_leader(native_account,leader_level,up_leader)VALUES('2gHf58q2Rqq6chGVQKBvYieG4nT4xaR1JhjVkLkLNBNL',1,'9PqbpkmcxxzzadBtUE3ddXVaHC1t5JN2PYKMNzRh4gUV');
INSERT INTO t_stake_leader(native_account,leader_level,up_leader)VALUES('HwNLVyiB5n27PDnHRXtDUJ3BGL5jH4daYgFJuX1CAR1C',1,'9PqbpkmcxxzzadBtUE3ddXVaHC1t5JN2PYKMNzRh4gUV');
INSERT INTO t_stake_leader(native_account,leader_level,up_leader)VALUES('757pDZPAtpUici7HE6kCmy7CXe7MtxqFapkLg2baNs2V',1,'9PqbpkmcxxzzadBtUE3ddXVaHC1t5JN2PYKMNzRh4gUV');
INSERT INTO t_stake_leader(native_account,leader_level,up_leader)VALUES('5Ji4T6WRPCDFx69jpFnSHH3FziwsF63t6Dr3dMG2kTEi',1,'9PqbpkmcxxzzadBtUE3ddXVaHC1t5JN2PYKMNzRh4gUV');
INSERT INTO t_stake_leader(native_account,leader_level,up_leader)VALUES('4vZtkXfxwgKyHfE6FTk1iAWPxDw96ZcLU4yiHwBPvJ6J',1,'9PqbpkmcxxzzadBtUE3ddXVaHC1t5JN2PYKMNzRh4gUV');
INSERT INTO t_stake_leader(native_account,leader_level,up_leader)VALUES('BLHhpwijY7gZsqVjXJtxjpWFtMnXKSGp98yknijuXpn3',1,'9PqbpkmcxxzzadBtUE3ddXVaHC1t5JN2PYKMNzRh4gUV');
INSERT INTO t_stake_leader(native_account,leader_level,up_leader)VALUES('ED6B6XpuJasBSfbSG4sdJQMZnTqsuCTAeGqrx31UsoMk',1,'9PqbpkmcxxzzadBtUE3ddXVaHC1t5JN2PYKMNzRh4gUV');
INSERT INTO t_stake_leader(native_account,leader_level,up_leader)VALUES('Gy2xG5DpcJMKLXcqBMwn4nkgwLUEM7EXJN2D9zhJcKtB',1,'9PqbpkmcxxzzadBtUE3ddXVaHC1t5JN2PYKMNzRh4gUV');
