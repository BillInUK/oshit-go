-- reward服务初始化数据
insert into t_level_dist("level")values(2);
insert into t_level_ratio("level","ratio")values(1,10);
insert into t_level_ratio("level","ratio")values(2,1);
insert into t_discount_rate(rate)values(1.25);

-- take token
insert into public.t_take_token_config
(invite_code,decimals,token_mint_account,reward_token_account,reward_native_account,dex_native_account,amount,invite_amount,dex_fee_rate,max_dex_fee,interval,is_default,reward_inviter,invited,created_at,updated_at)
values
    (NULL,3,'wtnrTujJqBRUknLRhQQcUSwzAzx8LvcxKXEuBwvFnJM', 'EHEu46gQMTFw1ieiok5XVYLV9MrKUFyk6sRLDjUpEQAd', 'GmKsGRytiVoeMZGmBVCWPcUzJGHVqcvzhP5K9cstdr3E', '6MeXfYMhXpQSz3fqHtEa72V1XgKG7WGsECDy9jEv9e2K', 500000, 1500000, 200,400,10,true,true,true,NOW(),NOW());

-- give token
insert into public.t_give_token_config
(token_mint_account,decimal,reward_token_account,reward_native_account,dex_native_account,reward_rate,max_valid_reward,valid_rate,created_at,updated_at)
values
    ('wtnrTujJqBRUknLRhQQcUSwzAzx8LvcxKXEuBwvFnJM', 3,'6zimN4MMo5CJW7VwhD3nAdXt6EpVoc5V7SCnZiS1kLPd','AjhUm6o9eV2xV9G2ZPH3pSb8DhTAjb27MDrTkKMrDVVZ','6MeXfYMhXpQSz3fqHtEa72V1XgKG7WGsECDy9jEv9e2K',200,4000000,300,'2025-09-30 11:59:01.7249', '2025-09-30 11:59:01.7249');

-- reward code
INSERT INTO public.t_reward_code_config(reward_account,cost_account)VALUES('584AMuM1HkV4wRMMVPZuZy9g9mZbSAcTF7QiBrHJaZFE','6MeXfYMhXpQSz3fqHtEa72V1XgKG7WGsECDy9jEv9e2K');
INSERT INTO public.t_reward_code_fee(amount,cost_rate)VALUES(5000000,8);
INSERT INTO public.t_reward_code_fee(amount,cost_rate)VALUES(1000000,20);
INSERT INTO public.t_reward_code_fee(amount,cost_rate)VALUES(500000,30);

-- campaign
INSERT INTO t_campaign_exchange_config(reward_account,cost_account,rate,cost_rate)VALUES('2NVji8RvQAFhg4YJKuxqhdMjWMLJmWbKm5MBvSJmTUHL','6MeXfYMhXpQSz3fqHtEa72V1XgKG7WGsECDy9jEv9e2K',500,17);
