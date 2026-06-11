-- 基础配置
delete from t_system_config;
insert into t_system_config(env)values(1);

-- aws配置
delete from t_aws_config;
insert into t_aws_config
(access_key_id,secret_access_key,region)
values
    ('ENC~sqgSgTyY0NVv6pekNSkwHQrkYlKWOkg5oMKAX+hfaIZOwuG80Ky54xC7JS4WB938ZWU+lKm/m9WEtr10wPjKyQ==','ENC~CRD3d4mB6UQ4SG10t4qOEfp/Zp88hWWvsAo6ktxD1qwG2j82H6ULNy8zsGTTl6e+7gIgpHz1DFcW+/+VQ3XPOZzm1rptBvnFKHJm/JWZdVU=','ap-southeast-1');

-- rpc 配置
INSERT INTO public.t_rpc_endpoint
    (scope, provider, endpoint, api_key, wss_endpoint, wss_api_key, weight)
VALUES
    ('env', 'helius', 'https://devnet.helius-rpc.com', 'ENC~At1eC+h77eD+A1jXTu0Acf8Ug9fLbV5CiB8lYl7y5agOo76Pdh5Zy8RfyvPdp5gMIMV5SAq1bLIUoE67+65ORvHmUitZ/PUIkaUgo0s9XbU=', 'wss://devnet.helius-rpc.com', 'ENC~At1eC+h77eD+A1jXTu0Acf8Ug9fLbV5CiB8lYl7y5agOo76Pdh5Zy8RfyvPdp5gMIMV5SAq1bLIUoE67+65ORvHmUitZ/PUIkaUgo0s9XbU=', 1),
    ('mainnet', 'helius', 'https://mainnet.helius-rpc.com', 'ENC~At1eC+h77eD+A1jXTu0Acf8Ug9fLbV5CiB8lYl7y5agOo76Pdh5Zy8RfyvPdp5gMIMV5SAq1bLIUoE67+65ORvHmUitZ/PUIkaUgo0s9XbU=', 'wss://mainnet.helius-rpc.com', 'ENC~At1eC+h77eD+A1jXTu0Acf8Ug9fLbV5CiB8lYl7y5agOo76Pdh5Zy8RfyvPdp5gMIMV5SAq1bLIUoE67+65ORvHmUitZ/PUIkaUgo0s9XbU=', 1);

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
    ('reward','take token','WJE59ZzX8cDcJRuNYEmI3kMJpxAVuyii3ARmy6DI4yYKKcteKxhGmaXP+NJQG/lesWpAzaHOfl8wIKqOdrCph45btkAddrVWFMHVJwPk0hXdb3dGws/BuAFRBORnnWwZPkUl1URuYZuK4T79FG48rYFaT9I+gxq/MDNAzxxuiBk=',NOW(),NOW());

insert into t_service_key
    (service,sub_service,encrypted_key,created_at,updated_at)
values
    ('reward','give token','d8l6eQhkw3jR3TNFaUo7pV9SQFvYNRz7QLcq8D5ySSWoItmEwFg/cPvZvdRT3mPkNd61pC6JXpqgSn7ZTpmoTMqS7UJaMm4otAYMDj7iW0NwoW9viTU6UmGVBWv15GqJbhqhWO/AC91BVOxOuzuawSmGf8zLFcEgv9kbKf+Delo=',NOW(),NOW());

insert into t_service_key
    (service,sub_service,encrypted_key,created_at,updated_at)
values
    ('reward','lottery','kMnsZAst1FXB0TnMUkoYv9AzAJB2hlhM45RjOhAgBTkZxwcqVgZIEftrzm2hvKKQ/7aR5jyKcHJzIiyn3XPJfrgsTWJmzUxdGHDtD+omqerLxwQ3SLYazuB0qw4LVo8/tw7nNF72CZAoQ79HYNk90sxoBpi5qNjAeizUDnuSpoY=',NOW(),NOW());

insert into t_service_key
    (service,sub_service,encrypted_key,created_at,updated_at)
values
    ('reward','reward code','HaYXzvtD/JTiCGa0HM+9WEkn/TQRKqPTu2e5pqpwprb3Rm8xrMuwI6FaotVc53Mp6BV8vNCiAO1nL5cLv9VGmLTfJREW1A+PWdPZCKm52OYHFaXnLHWgmA846FjfhoNQIotbXMX/+3b5PgNBdo52qFAz8P0lY2+1Ob0vsoOUlnE=',NOW(),NOW());

insert into t_service_key
    (service,sub_service,encrypted_key,created_at,updated_at)
values
    ('reward','campaign quote','fj0gPICnUwMXyuZReLi70voft/645NSm/Lygx7o5FDL6/bLIvNblFdu5bXrRyTVEeLTMP/MKf4g7FbNwdMRK65aiWH2oLFYZheAblsoDJCaI3LT5qwOJm45V+91TfslDKrqFrh/WBbDuXEPBLhVx540s3nh6Sj8/iXi5a71+8Eo=',NOW(),NOW());

insert into t_service_key
    (service,sub_service,encrypted_key,created_at,updated_at)
values
    ('pos','pos reward','wIyh1brCX3JR8SAnUif1JK7RHJUtXou/2tzxOuP6yg6hNvLTfQkyB5VZjdhhbcv4I1cGMDKHYf4VkpOaAHsHUEZke8arRhnMXNNPxZqahzuC+3hKcq4TD0uVOOb371zIIvVfhbTVP16EkOt2BlKtF/pdtsLJ80WM5Y0+ntXFPIc=',NOW(),NOW());

insert into t_service_key
    (service,sub_service,encrypted_key,created_at,updated_at)
values
    ('pos','stake token','VDkROsiZMlc1TEdcDKtDIQ3svM4K2w8jFFrDQ5ihiHIiHZIXZ7vN2TSit4S/P8BKWG2PyX5ydg5LfUKvDuaKMMYg4zagHy3vxhpFsS6HBGIo1W9kyBD+I1oPxL/08sMIsSb1e3HkE5GXEgViHxUtGdPNr7GpwUemTEudclzZoz0=',NOW(),NOW());

insert into t_service_key
    (service,sub_service,encrypted_key,created_at,updated_at)
values
    ('pos','stake reward','GJemwID1rA5TiZyKv9/JOVcJcByNomw4lyfJR42lJJp1IzLV5fDvW1/VicZm7pCjRvF0tRP0hXbQFAGSvETolJdq7kd+Nv4nsuSCarXvIrwUn5o2G8pEY59geADdTYRTsqQa6dnDguxKleXFeo5WY96VGvy6T3daN1Y3+woTL4s=',NOW(),NOW());

insert into t_service_key
    (service,sub_service,encrypted_key,created_at,updated_at)
values
    ('pos','stake leader reward','zZOeeFk6UpSh35nCqbZMNZWgFqIreqsMlnZAYf344s8PnKyW8d4IRNDrpUxEfNeajgbOolKtov9gbe0Qy/A4b0O7zHQdaBBOZq8gKQzBF1ZXiSSJJK64Y6twSSe9QeIAL9RaX5Kg3B3jj0heEfvv+cQFDqPm/xVk9bj6+XZspQs=',NOW(),NOW());

insert into t_service_key
    (service,sub_service,encrypted_key,created_at,updated_at)
values
    ('gs-relay','contribution','a+S0XiQhYVtQ05IwWBnXQIYe3Lb6ObZu5Do+tWX0lrt/xDPkjlAO8gxi2BSJk4hJs7WCc9Z/wxcJjeoGU0fUy/SGOkkwlShTt2XCJMAyVUUdJEiTfISU43nAHgXsqAlIJICFsj1hGjH17IA8oYVAUs72K8y8MfpAOwuZkS5Wczc=',NOW(),NOW());

insert into t_service_key
    (service,sub_service,encrypted_key,created_at,updated_at)
values
    ('gs-relay','payout','a+S0XiQhYVtQ05IwWBnXQIYe3Lb6ObZu5Do+tWX0lrt/xDPkjlAO8gxi2BSJk4hJs7WCc9Z/wxcJjeoGU0fUy/SGOkkwlShTt2XCJMAyVUUdJEiTfISU43nAHgXsqAlIJICFsj1hGjH17IA8oYVAUs72K8y8MfpAOwuZkS5Wczc=',NOW(),NOW());

insert into t_service_key
    (service,sub_service,encrypted_key,created_at,updated_at)
values
    ('gs-relay','refund','a+S0XiQhYVtQ05IwWBnXQIYe3Lb6ObZu5Do+tWX0lrt/xDPkjlAO8gxi2BSJk4hJs7WCc9Z/wxcJjeoGU0fUy/SGOkkwlShTt2XCJMAyVUUdJEiTfISU43nAHgXsqAlIJICFsj1hGjH17IA8oYVAUs72K8y8MfpAOwuZkS5Wczc=',NOW(),NOW());


-- 业务交易扫描表
delete from t_tx_scan_info;

INSERT INTO public.t_tx_scan_info
    (service, sub_service, native_account, pda_account, until_tx_id, before_tx_id, slot, created_at, updated_at)
VALUES
    ('reward', 'take token', 'GmKsGRytiVoeMZGmBVCWPcUzJGHVqcvzhP5K9cstdr3E', 'EHEu46gQMTFw1ieiok5XVYLV9MrKUFyk6sRLDjUpEQAd', '5WQPkuVLDaJhp5xRKYLPTXBGirLidd5C6iZjV4tit2sMLwPPHBKBPCcwiVsh7xh65ij7QfoX5AfNNqbSv4N2MggY', '', '468657714', NOW(),NOW());

INSERT INTO public.t_tx_scan_info
    (service, sub_service, native_account, pda_account, until_tx_id, before_tx_id, slot, created_at, updated_at)
VALUES
    ('reward', 'give token', 'AjhUm6o9eV2xV9G2ZPH3pSb8DhTAjb27MDrTkKMrDVVZ', '6zimN4MMo5CJW7VwhD3nAdXt6EpVoc5V7SCnZiS1kLPd', '3Nt2sTD6PU8Xp7LhPyAKkAHB9UR3tgDF1KBNwzvucVNzmUEoBMzs7qUzcvJfALb5wHMKcTfwVTZ9hZbf2KhZ3hAJ', '', '468657801', NOW(),NOW());

INSERT INTO public.t_tx_scan_info
    (service, sub_service, native_account, pda_account, until_tx_id, before_tx_id, slot, created_at, updated_at)
VALUES
    ('reward', 'lottery', 'GmKsGRytiVoeMZGmBVCWPcUzJGHVqcvzhP5K9cstdr3E', 'EHEu46gQMTFw1ieiok5XVYLV9MrKUFyk6sRLDjUpEQAd', '5WQPkuVLDaJhp5xRKYLPTXBGirLidd5C6iZjV4tit2sMLwPPHBKBPCcwiVsh7xh65ij7QfoX5AfNNqbSv4N2MggY', '', '468657714', NOW(),NOW());

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
    ('pos', 'market buy token', 'HtNfUbDaBamCBPWCFiESkXpewvwVLkwrSWRjjV8FNT7i', 'HtNfUbDaBamCBPWCFiESkXpewvwVLkwrSWRjjV8FNT7i', '4KLb4yP9sYRxfJ11Kvg9CN8vkbRqmbHtqPpBWoMb7Et8ad1iuQMNPD8fVNRQsFq5dR1t4pBnnfpc5k1fhrYhDtuT', '', '425727261', NOW(),NOW());

INSERT INTO public.t_tx_scan_info
    (service, sub_service, native_account, pda_account, until_tx_id, before_tx_id, slot, created_at, updated_at)
VALUES
    ('pos', 'stake token', 'CyLTEgvmqVF9dPJkT6bMgccfXL7G26EXRAM9FEuP5ki6', 'CyLTEgvmqVF9dPJkT6bMgccfXL7G26EXRAM9FEuP5ki6', '4CATtfwvXKKdwj4exQsg6EobJFhHye8TnbkSQJUa1mPG9zm9YZj66ghKYbH5Tn2M6N8BaCA5kQ6J8VWMrHjzYMAB', '', '468179889', NOW(),NOW());

INSERT INTO public.t_tx_scan_info
    (service, sub_service, native_account, pda_account, until_tx_id, before_tx_id, slot, created_at, updated_at)
VALUES
    ('pos', 'stake reward', 'H5WmBY45gxP8rj7gecLXsv6yNHqHFXNH4Acmp2U9E2Tb', 'ERYPoieDaHoh9jz1Gbmi9whHLd1QyWPKANvtZTnECufw', 'UeVGdCUHKTiPBEsikutc9HUw2mb1dsCy8jKT1Hwz8TPrHXNRK1bznzCkojbkF5ZhHvg2fUMMcc7LDbhpmKBHQTF', '', '468179455', NOW(),NOW());

INSERT INTO public.t_tx_scan_info
    (service, sub_service, native_account, pda_account, until_tx_id, before_tx_id, slot, created_at, updated_at)
VALUES
    ('pos', 'stake leader reward', '2yRkofKW7xKRbN79MHKGX8HFyuHZEtTJDhTwAQjJHnMX', 'HqAp7uFZAckGA9jKFCZXg7rq1GX9CB2kwpGBKoq5geg6', '4EjeyG9GvWGhFVYjJDnXHA7Pije91HoFSfpUoMAg9dW68Xf1jDxCvdB5n3m74rAQM7cXZ1GxL9oad5j3FpjT6urp', '', '467320074', NOW(),NOW());

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

