-- 质押AMM配置
insert into public.t_stake_amm_config(quote_token,public_key)values('SOL','46uzvWDstrwNtEpBSFrcVPx4ZaTMDpjarQYWpq82Z58p');
-- 质押池配置
insert into public.t_stake_token_pool(source_account,from_token_account)values('Raydium','Gh6MjRrJFBU9HcYMYKBbaD8fX1dv3CjGtDDhtVThD9v3');
-- 质押每日固定利息
insert into public.t_stake_fix_rate_config(min_amount,stake_type,fix_rate,individual_rate,created_at,updated_at)values(100000,0,70,100,now(),now());
insert into public.t_stake_fix_rate_config(min_amount,stake_type,fix_rate,individual_rate,created_at,updated_at)values(100000,1,100,100,now(),now());

-- 邀请奖励级别
insert into public.t_stake_invite_dist(dist_level)values(2);

-- 邀请奖励每个级别的奖励费率
insert into public.t_stake_invite_rate(dist_level,rate)values(1,10);
insert into public.t_stake_invite_rate(dist_level,rate)values(2,5);

-- 质押星级配置
insert into public.t_stake_star_level_rule(amount,group_amount,star_level,rate,created_at,updated_at)values(100000000,0,1,8,now(),now());
insert into public.t_stake_star_level_rule(amount,group_amount,star_level,rate,created_at,updated_at)values(200000000,0,2,16,now(),now());
insert into public.t_stake_star_level_rule(amount,group_amount,star_level,rate,created_at,updated_at)values(300000000,0,3,24,now(),now());
insert into public.t_stake_star_level_rule(amount,group_amount,star_level,rate,created_at,updated_at)values(400000000,0,4,32,now(),now());
insert into public.t_stake_star_level_rule(amount,group_amount,star_level,rate,created_at,updated_at)values(500000000,0,5,40,now(),now());
insert into public.t_stake_star_level_rule(amount,group_amount,star_level,rate,created_at,updated_at)values(600000000,0,6,48,now(),now());

-- 质押白名单地址
insert into public.t_stake_star_whitelist(native_account,star_level,rate)values('FrWcQiSAYQGDFToxbYBCCvarzb6SumXCUmT4XBb53xAp',6,48);

-- 质押奖励发放配置表
insert into public.t_stake_reward_config
(program_id,reward_account,cost_account,quote_token_amount,cost_fee_rate,created_at,updated_at)
values
    ('As9Z52f8Sioqr22KpS4xdzrhicwGwAu6x5SxVaHfvLws','H5WmBY45gxP8rj7gecLXsv6yNHqHFXNH4Acmp2U9E2Tb','6MeXfYMhXpQSz3fqHtEa72V1XgKG7WGsECDy9jEv9e2K',500000,110,now(),now());

-- stake奖励明细表

-- 质押奖励领取记录表

-- stake 每日快照表

-- 质押记录表

-- 购买token记录表

-- 插入区域经理配置
insert into public.t_stake_leader_reward_config(reward_account)values('2yRkofKW7xKRbN79MHKGX8HFyuHZEtTJDhTwAQjJHnMX');

-- 总区域经理表
insert into public.t_stake_total_leader(native_account,stake_share,created_at,updated_at)values('G6xxsFzFHPhLCUg4Qq8aun2EcQ3hTcLvwb6pUWMsVBKa',7,now(),now());
insert into public.t_stake_total_leader(native_account,stake_share,created_at,updated_at)values('CH6nEGuiYF5kenkavr4KMLEvY6DP7cKgQKrh9t2UiHX9',3,now(),now());

-- 区域经理表
insert into public.t_stake_leader(native_account,leader_level,up_leader,created_at,updated_at)values('FgPU2MGLkX278ZD2XhNaJBwVd7q2CjHa9Y9Em3u1s7Ur',2,null,now(),now());
insert into public.t_stake_leader(native_account,leader_level,up_leader,created_at,updated_at)values('DGMNe2KYDB8fbxdcxipHwmzMqJk34dMZxdWxguZC4xrC',2,null,now(),now());

insert into public.t_stake_leader(native_account,leader_level,up_leader,created_at,updated_at)values('D8Dibj91XjosaCjiWbk2zHQvvuFDTUvvjEeT351Z7fK9',1,null,now(),now());
insert into public.t_stake_leader(native_account,leader_level,up_leader,created_at,updated_at)values('9ti8HrjakPuRmFa2Q6Lj7UopKuCuS2gSGLNk5umZ7X8v',1,null,now(),now());
insert into public.t_stake_leader(native_account,leader_level,up_leader,created_at,updated_at)values('4tJJv3RiQfwL4vdRsdYJK12K1yoirw3b2n1RTZxKacwn',1,'FgPU2MGLkX278ZD2XhNaJBwVd7q2CjHa9Y9Em3u1s7Ur',now(),now());
insert into public.t_stake_leader(native_account,leader_level,up_leader,created_at,updated_at)values('GVfKwyUXYuVhh65v8Q232dLhdm51eBfsLZPGAsWK6npk',1,'FgPU2MGLkX278ZD2XhNaJBwVd7q2CjHa9Y9Em3u1s7Ur',now(),now());
insert into public.t_stake_leader(native_account,leader_level,up_leader,created_at,updated_at)values('EzehTETE4o8Jua5pCAUk7AHV5Zt3QL1a2D9vfE1ii6HK',1,'DGMNe2KYDB8fbxdcxipHwmzMqJk34dMZxdWxguZC4xrC',now(),now());

-- 区域经理奖励明细表

-- 区域经理奖励领取表

