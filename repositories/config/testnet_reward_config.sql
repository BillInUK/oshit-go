-- reward服务初始化数据
insert into t_level_dist(dist_level)values(2);
insert into t_level_ratio(dist_level,ratio)values(1,10);
insert into t_level_ratio(dist_level,ratio)values(2,1);
insert into t_discount_rate(rate)values(1.25);

-- take token 配置
insert into public.t_take_token_config
(invite_code,reward_account,cost_account,amount,invite_amount,cost_fee_rate,max_cost_fee,is_default,reward_inviter,invited,created_at,updated_at)
values
    (NULL,'GmKsGRytiVoeMZGmBVCWPcUzJGHVqcvzhP5K9cstdr3E', '6MeXfYMhXpQSz3fqHtEa72V1XgKG7WGsECDy9jEv9e2K', 500000, 1500000, 200,400,true,true,true,NOW(),NOW());

-- give token 配置
insert into public.t_give_token_config
(reward_account,cost_account,reward_rate,max_valid_reward,valid_rate,created_at,updated_at)
values
    ('AjhUm6o9eV2xV9G2ZPH3pSb8DhTAjb27MDrTkKMrDVVZ','6MeXfYMhXpQSz3fqHtEa72V1XgKG7WGsECDy9jEv9e2K',200,4000000,300,NOW(),NOW());

-- reward code
INSERT INTO public.t_reward_code_config(reward_account,cost_account)VALUES('584AMuM1HkV4wRMMVPZuZy9g9mZbSAcTF7QiBrHJaZFE','6MeXfYMhXpQSz3fqHtEa72V1XgKG7WGsECDy9jEv9e2K');
INSERT INTO public.t_reward_code_fee(amount,fee_rate)VALUES(5000000,8);
INSERT INTO public.t_reward_code_fee(amount,fee_rate)VALUES(1000000,20);
INSERT INTO public.t_reward_code_fee(amount,fee_rate)VALUES(500000,30);

-- campaign 配置
INSERT INTO t_campaign_quote_config
(reward_account,cost_account,quote_rate,cost_rate)
VALUES
    ('2NVji8RvQAFhg4YJKuxqhdMjWMLJmWbKm5MBvSJmTUHL','6MeXfYMhXpQSz3fqHtEa72V1XgKG7WGsECDy9jEv9e2K',500,17);

