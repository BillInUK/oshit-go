INSERT INTO t_pos_reward_config(reward_account,cost_account,cost_fee_rate,max_cost_fee,quote_token_amount,created_at,updated_at)VALUES('C2E7K1fDUzpihX77xMNnhYNidRHkWnLMTRejWRvfjkDH','6MeXfYMhXpQSz3fqHtEa72V1XgKG7WGsECDy9jEv9e2K',200,40000,500000,NOW(),NOW());

INSERT INTO t_pos_star_level_rule(amount,group_amount,star_level,rate,created_at,updated_at)VALUES(1000000,10000000,1,10,NOW(),NOW());
INSERT INTO t_pos_star_level_rule(amount,group_amount,star_level,rate,created_at,updated_at)VALUES(2000000,20000000,2,20,NOW(),NOW());
INSERT INTO t_pos_star_level_rule(amount,group_amount,star_level,rate,created_at,updated_at)VALUES(3000000,30000000,3,30,NOW(),NOW());
INSERT INTO t_pos_star_level_rule(amount,group_amount,star_level,rate,created_at,updated_at)VALUES(4000000,40000000,4,40,NOW(),NOW());
INSERT INTO t_pos_star_level_rule(amount,group_amount,star_level,rate,created_at,updated_at)VALUES(5000000,50000000,5,50,NOW(),NOW());

INSERT INTO t_pos_mission_config (reward_type,starred,rate,created_at,updated_at) VALUES (1, false, 10.00, NOW(),NOW());
INSERT INTO t_pos_mission_config (reward_type,starred,rate,created_at,updated_at) VALUES (2, false, 10.00, NOW(),NOW());
INSERT INTO t_pos_mission_config (reward_type,starred,rate,created_at,updated_at) VALUES (3, false, 10.00, NOW(),NOW());
INSERT INTO t_pos_mission_config (reward_type,starred,rate,created_at,updated_at) VALUES (1, true,  10.00, NOW(),NOW());
INSERT INTO t_pos_mission_config (reward_type,starred,rate,created_at,updated_at) VALUES (2, true,  10.00, NOW(),NOW());
INSERT INTO t_pos_mission_config (reward_type,starred,rate,created_at,updated_at) VALUES (3, true,  10.00, NOW(),NOW());
INSERT INTO t_pos_mission_config (reward_type,starred,rate,created_at,updated_at) VALUES (0, false, 15.00, NOW(),NOW());
INSERT INTO t_pos_mission_config (reward_type,starred,rate,created_at,updated_at) VALUES (0, true,  15.00, NOW(),NOW());