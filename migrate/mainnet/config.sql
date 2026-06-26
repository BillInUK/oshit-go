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
    ('reward','reward code','xfMYopaKia2LYHsppTwRqNrs1YyY9PSu43abnNTtJif','','reward','ServiceTransaction',0,0,true,true,now(),now());

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
    ('pos','stake leader reward','9eXCGjngWmLdFwcy2NiUUp1aAR7TcfTUV7G7TQURWn8a','','Stake','ServiceTransaction',0,0,true,true,now(),now());

insert into t_service_info
    (service,sub_service,address,webhook,mq_group,mq_topic,hook_type,tx_source,confirm,multi_sign,created_at,updated_at)
values
    ('pos','market buy token','HtNfUbDaBamCBPWCFiESkXpewvwVLkwrSWRjjV8FNT7i','','Stake','ServiceTransaction',1,0,true,true,now(),now());

insert into t_service_info
    (service,sub_service,address,webhook,mq_group,mq_topic,hook_type,tx_source,confirm,multi_sign,created_at,updated_at)
values
    ('gs-relay','contribution','3qc8hmZtFyLmJaLbK88W89PDoJq1D5GuKodi4Lh4yVzU','','gs-relay-activity','ServiceTransaction',1,0,true,true,now(),now());

insert into t_service_info
    (service,sub_service,address,webhook,mq_group,mq_topic,hook_type,tx_source,confirm,multi_sign,created_at,updated_at)
values
    ('gs-relay','payout','3qc8hmZtFyLmJaLbK88W89PDoJq1D5GuKodi4Lh4yVzU','','gs-relay-activity','ServiceTransaction',1,0,true,true,now(),now());

insert into t_service_info
    (service,sub_service,address,webhook,mq_group,mq_topic,hook_type,tx_source,confirm,multi_sign,created_at,updated_at)
values
    ('gs-relay','refund','3qc8hmZtFyLmJaLbK88W89PDoJq1D5GuKodi4Lh4yVzU','','gs-relay-activity','ServiceTransaction',1,0,true,true,now(),now());


-- 业务私钥
delete from t_service_key;

insert into t_service_key
    (service,sub_service,encrypted_key,created_at,updated_at)
values
    ('reward','take token','746dNbN1e0oMQLzYpdIfO37uyroQOCttAc2fQTNAFZXQ0VKtkz12ciixCqKoGknYyAY74ZVAy7s8mmAPQjoClGUdjL9ONab5xoaSOtd5lxyiHEeeo8jPA5nWpLvLt5g30dxyS0EohRCfKtqWoEnbOt6/Q51iENtFtPegT50lZTw=',NOW(),NOW());

insert into t_service_key
    (service,sub_service,encrypted_key,created_at,updated_at)
values
    ('reward','give token','y6EmkLyZRb1jMSXJ4W35k11ZKwvQJrOsyNAGTN0LoZ985G273vyo79By8eS+hFmQlcjUPGk4f+G2ElTpu9kWNLFqal37iBLIjYWk+eC7LtFB9ec3m/qrHDmVUD08sl/eri0bZ8pO7JdXB2XsFvNL22A2Q6Tf5xDcoFwLHZQiWjw=',NOW(),NOW());

insert into t_service_key
    (service,sub_service,encrypted_key,created_at,updated_at)
values
    ('reward','lottery','XsBnQmadWrlwKeduYyWfn3jNtttinMIxd42tKTY0y6dVuL/tyPJeK2mxvP6aJ6bKkObJTlXizHHtCNFmgiitbC4+ztc5at8+vBJ23N8hUYRQTUfvHxvfHY791tqTYEpWEvi3a6m1vgffOk1+Xfi9kzgWNaZpBGEzfG9BQGnLmfg=',NOW(),NOW());

insert into t_service_key
    (service,sub_service,encrypted_key,created_at,updated_at)
values
    ('reward','reward code','7rP31Vv7EM8gXJbFbQYix5uqJvfLa2KH9XiYXQ52HRfe1qWtg8k9NLeuyJ4IhJh0So00qT4JkALwZ5QrRKZ/62TKWsdFLRXTa/S5ixrQLH6GyUwyuCjayxWtHlZ29vhLwSmCJ/JBf9DhOZX2DfAHbwWHdOw/9lal9RLb+OeTcuE=',NOW(),NOW());

insert into t_service_key
    (service,sub_service,encrypted_key,created_at,updated_at)
values
    ('reward','campaign quote','siRYTm5Pe5HIIDpK3AEaiG0eJr7Mdzy9tCCFJ6MO/NLKa9Qds+pQIBJeMuV6ITWumvrpoaQwyBixO9kUSmxxP2WwF14KaUVH9XbLLwTLz4EsS+17fbRJoiVDJ/cOWNYhj2WnnA1TmfgEYI+M8F9dwagw17Gpq1HJz069BOVEQWA=',NOW(),NOW());

insert into t_service_key
    (service,sub_service,encrypted_key,created_at,updated_at)
values
    ('pos','pos reward','8ZZURDLAJO4ZEts6J/xIk0vMQOeP7pVugftSnKnam4OW/szhio46HJgwCsAgOyLGDZXovK0t6mpKk87mMXaQFTRC1KJwppfGVFp9mpRF/ZRukaLYBIPYDYOlUvgQTExMvDN0lUH2TlEiOW+97PvdeJMqWb8/0IPmn0s510YFsKg=',NOW(),NOW());

insert into t_service_key
    (service,sub_service,encrypted_key,created_at,updated_at)
values
    ('pos','stake token','JDLO6JRXvY6gzeSECPYZTF0VRsAAHBcdN7mv5R5n4s4FmI2b5/9OGrI8EnYiRGSgs/eLVm+ef+KLrN6+zr3HwvvtO6Q2Num1XHNhuoh5F+DGyZ2yTgpq7RbJwAwo+MHfbaekuxqHI6Pwie1HuiDvoLaj8yuDFxbTB0CNZAkEUlI=',NOW(),NOW());

insert into t_service_key
    (service,sub_service,encrypted_key,created_at,updated_at)
values
    ('pos','stake reward','VoSBFzxk5M8ZahbBeLjZTJ8esRpBISSA06gF3+w0iLAn643WR8ZODA8XRIArHviqoi+VfFMe1LZzvkeE9wyVG59j5DoOYrgvnWC2x7w+y9cf3f/QDnlnptIi7O7pChGbYE9+NROuQOU7NfxVLCOrJ12k2Bgr2Ak/lkNKkhLneAY=',NOW(),NOW());

insert into t_service_key
    (service,sub_service,encrypted_key,created_at,updated_at)
values
    ('pos','stake leader reward','tTOwffx8nydnEJ8FbKuhEjZXVnP3ezEBsOh6avbsf7JKRgmFlO4dtFXmnehgI1ClxhsK202434MxIl02H+bK71NUTvCViv3tI9WaMlZRBjcE2qW31SNbV8yQbszKz83WtiYqkmpBqHHa7I1Nf2Jxp8lSEeSr5OgXd02aay68ihY=',NOW(),NOW());

insert into t_service_key
    (service,sub_service,encrypted_key,created_at,updated_at)
values
    ('gs-relay','contribution','ds6Q5dlqcO2UfgolwsG5erC/+joXE6TlbKYstmYSkoNsf4QqW4V+XkGNcTVZbXOzwZY/gAR+jCF02Oo1cNUzZAE02c/8ZzuykNRn3zWrJhBb/MnrGV0oUIOnsdGFUcat7pdteccIsvXGBAe+JrmmhTKHTfoxnMj0EaXanIr1SN8=',NOW(),NOW());

insert into t_service_key
    (service,sub_service,encrypted_key,created_at,updated_at)
values
    ('gs-relay','payout','ds6Q5dlqcO2UfgolwsG5erC/+joXE6TlbKYstmYSkoNsf4QqW4V+XkGNcTVZbXOzwZY/gAR+jCF02Oo1cNUzZAE02c/8ZzuykNRn3zWrJhBb/MnrGV0oUIOnsdGFUcat7pdteccIsvXGBAe+JrmmhTKHTfoxnMj0EaXanIr1SN8=',NOW(),NOW());

insert into t_service_key
    (service,sub_service,encrypted_key,created_at,updated_at)
values
    ('gs-relay','refund','ds6Q5dlqcO2UfgolwsG5erC/+joXE6TlbKYstmYSkoNsf4QqW4V+XkGNcTVZbXOzwZY/gAR+jCF02Oo1cNUzZAE02c/8ZzuykNRn3zWrJhBb/MnrGV0oUIOnsdGFUcat7pdteccIsvXGBAe+JrmmhTKHTfoxnMj0EaXanIr1SN8=',NOW(),NOW());


-- 业务交易扫描表
delete from t_tx_scan_info;

INSERT INTO public.t_tx_scan_info
    (service, sub_service, native_account, pda_account, until_tx_id, before_tx_id, slot, created_at, updated_at)
VALUES
    ('reward', 'take token', '4fLi61UtmjtKXJHL1aKb1cdUv3KZfCVnQv45hPN8h3Sf', '4LQDhhNW9Nqyo7DTBDszs97RSQwSr648hsbGt85hw5Cf', 'FB6vFNeMQGFmMPcw5E5v4V4GFPCSqvYZeSf77eysoUqm6VUueAsU7mMVh2Q7n6GhH4Jiq4rT5GKEJKajrzNkXNi', '', '428401136', NOW(),NOW());

INSERT INTO public.t_tx_scan_info
    (service, sub_service, native_account, pda_account, until_tx_id, before_tx_id, slot, created_at, updated_at)
VALUES
    ('reward', 'give token', '5xaFm7Xj2kFvjFmW3kFU2A9Ush6Hx3mqK86cYo6aJsbY', '9PEVsQ6Fnfgui5wpsXPSHa3L8byc3UvgZSc193ShUUzK', 'mCvZjNSQtsaFBkiqCzpmHgktL5c2uxUB8ezYWcgSq2eQVJzhwm5Wsw9J3ZU4krPWdbg7CXbjaVSCij5RGypoecB', '', '428288091', NOW(),NOW());

INSERT INTO public.t_tx_scan_info
    (service, sub_service, native_account, pda_account, until_tx_id, before_tx_id, slot, created_at, updated_at)
VALUES
    ('reward', 'lottery', '4fLi61UtmjtKXJHL1aKb1cdUv3KZfCVnQv45hPN8h3Sf', '4LQDhhNW9Nqyo7DTBDszs97RSQwSr648hsbGt85hw5Cf', 'FB6vFNeMQGFmMPcw5E5v4V4GFPCSqvYZeSf77eysoUqm6VUueAsU7mMVh2Q7n6GhH4Jiq4rT5GKEJKajrzNkXNi', '', '428401136', NOW(),NOW());

INSERT INTO public.t_tx_scan_info
    (service, sub_service, native_account, pda_account, until_tx_id, before_tx_id, slot, created_at, updated_at)
VALUES
    ('reward', 'reward code', 'xfMYopaKia2LYHsppTwRqNrs1YyY9PSu43abnNTtJif', 'A1jj3kCE5PdYPL1FKiLXWqquNW9KAEdizwU1a2RC9yzJ', '2gW8Ln9hhrreXYpPxRk6aZZGjH9rCNQfwJ8ayT5p2kebEY6WrTvJUmFN9uefbXrEwU7rZJjFdp5GSQoTfs5mbZVf', '', '428398745', NOW(),NOW());

INSERT INTO public.t_tx_scan_info
    (service, sub_service, native_account, pda_account, until_tx_id, before_tx_id, slot, created_at, updated_at)
VALUES
    ('reward', 'campaign quote', '4e1QkZ6mM3bXskMqESVKDs1GRgPGEyVQyPXjWM8GoTjS', '7JKwZqRsx76an4eMUWUFWwAisuhhE8y5Zv396HNcsRGo', '5UtkG9VC1EkENMQatk6PYy3dtNb2FFHRM6jKEZ8zDiSPER8uxSc6CdWcX4w2FdAvtrf47LEMKCnzxnDPvqJNx1EF', '', '428371092', NOW(),NOW());

INSERT INTO public.t_tx_scan_info
    (service, sub_service, native_account, pda_account, until_tx_id, before_tx_id, slot, created_at, updated_at)
VALUES
    ('pos', 'pos reward', 'H61Yvf3aZxW56hBGDNxT4cyguj98gSZMowzetYhc3TCB', 'EZGgg2zSGCbuy7KxjoYUvMWGw3mPdavGjFhYmD6sr4zz', 'DyY94bThe9hUaV7xnnSCFUhVYY8iktzizoQgnQgkWk8yppk6gARp1S1SBDkFMYNe96hXRDAERMf9XSbodBev5tQ', '', '428400906', NOW(),NOW());

INSERT INTO public.t_tx_scan_info
    (service, sub_service, native_account, pda_account, until_tx_id, before_tx_id, slot, created_at, updated_at)
VALUES
    ('pos', 'market buy token', 'HtNfUbDaBamCBPWCFiESkXpewvwVLkwrSWRjjV8FNT7i', 'HtNfUbDaBamCBPWCFiESkXpewvwVLkwrSWRjjV8FNT7i', 'ddpJLBLd5SiHZoFWMcg2MCepmUN9hbnB4Wf7Vfs71qhWMZVXytcH1Ms7eMaPvg9xjygoPQjqtAUJiMfeoQ81SMf', '', '428401401', NOW(),NOW());

INSERT INTO public.t_tx_scan_info
    (service, sub_service, native_account, pda_account, until_tx_id, before_tx_id, slot, created_at, updated_at)
VALUES
    ('pos', 'stake token', '89GUQB4BSn1G9ezeXMnDPm7bTPngNAcVxMn53VEpWYaA', '89GUQB4BSn1G9ezeXMnDPm7bTPngNAcVxMn53VEpWYaA', '3v6aLfi54mduTu6d95hLEy77fm7njS4max5NJ7m9FkLKBpVYDy1cUSzV6JeWvqNgSNsDxAaTP55SZfA1yZAUgKmW', '', '428362770', NOW(),NOW());

INSERT INTO public.t_tx_scan_info
    (service, sub_service, native_account, pda_account, until_tx_id, before_tx_id, slot, created_at, updated_at)
VALUES
    ('pos', 'stake reward', 'EgFd9DYSF6y6sSPDXzpzQsi2Qo4srCJCqDeSHsUpveej', 'AwXCN4QbLoo1q558o7qAfKRDqJeqK3GLCDrH5F5eF1TB', '2zUzjW16DjXrEBS8yjbe86bUYDqVNgSGWwYdjV25BCHTbLnEbPjteySANnyegC5ajqHwXru2R2Tvc62eLvXxvHRX', '', '428400375', NOW(),NOW());

INSERT INTO public.t_tx_scan_info
    (service, sub_service, native_account, pda_account, until_tx_id, before_tx_id, slot, created_at, updated_at)
VALUES
    ('pos', 'stake leader reward', '9eXCGjngWmLdFwcy2NiUUp1aAR7TcfTUV7G7TQURWn8a', 'BYMQBqkHDrtEuZoMzxqcSUhx3yLVEtDZgSCt2kqND77E', '2n1bbriu6ky8k48EX3cQ7FaH1FPTyumJ6upWHSXJPknJG412tznPrCZZmnd6EzTQRTJf7fEfjsKpAodmBBKabf3U', '', '427765444', NOW(),NOW());

INSERT INTO public.t_tx_scan_info
    (service, sub_service, native_account, pda_account, until_tx_id, before_tx_id, slot, created_at, updated_at)
VALUES
    ('gs-relay', 'contribution', '3qc8hmZtFyLmJaLbK88W89PDoJq1D5GuKodi4Lh4yVzU', 'Dprbw9Bab87WK9Z9CsQva8u9U4TyDicZFV2CjMriKcq7', '4qbKD7t1C66rUayaDnj8fdMqXZsRjEXdEp1tj5RdNZXumQu7fQT5aFx1mUmt8YCFsd9sVCNhcU5S35npgFJtyapp', '', '428398838', NOW(),NOW());

INSERT INTO public.t_tx_scan_info
    (service, sub_service, native_account, pda_account, until_tx_id, before_tx_id, slot, created_at, updated_at)
VALUES
    ('gs-relay', 'payout', '3qc8hmZtFyLmJaLbK88W89PDoJq1D5GuKodi4Lh4yVzU', 'Dprbw9Bab87WK9Z9CsQva8u9U4TyDicZFV2CjMriKcq7', '4qbKD7t1C66rUayaDnj8fdMqXZsRjEXdEp1tj5RdNZXumQu7fQT5aFx1mUmt8YCFsd9sVCNhcU5S35npgFJtyapp', '', '428398838', NOW(),NOW());

INSERT INTO public.t_tx_scan_info
    (service, sub_service, native_account, pda_account, until_tx_id, before_tx_id, slot, created_at, updated_at)
VALUES
    ('gs-relay', 'refund', '3qc8hmZtFyLmJaLbK88W89PDoJq1D5GuKodi4Lh4yVzU', 'Dprbw9Bab87WK9Z9CsQva8u9U4TyDicZFV2CjMriKcq7', '4qbKD7t1C66rUayaDnj8fdMqXZsRjEXdEp1tj5RdNZXumQu7fQT5aFx1mUmt8YCFsd9sVCNhcU5S35npgFJtyapp', '', '428398838', NOW(),NOW());




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
    ('xfMYopaKia2LYHsppTwRqNrs1YyY9PSu43abnNTtJif','4DZ3ry8LfaM3kEc6KCL4nDpNpnVX6g4Z6EFvYyNXsiqn');

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
