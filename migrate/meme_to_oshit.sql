-- 数据迁移脚本：从 meme_db 迁移到 oshit_db
-- 注意：请在执行前备份数据！
-- 执行步骤：
-- 1. 创建新的数据库结构（使用 oshit_db.sql）
-- 2. 运行本迁移脚本将数据从旧表复制到新表
-- 3. 验证数据完整性
-- 4. 删除旧表（可选）

INSERT INTO t_airdrop_record (record_id, brand, token_symbol, token_mint_account, air_drop_token_account, air_drop_native_account, receipt_token_account, receipt_native_account, amount, air_drop_tx_id, create_time, update_time)
SELECT "RecordId", "Brand", "TokenSymbol", "TokenMintAccount", "AirDropTokenAccount", "AirDropNativeAccount", "ReceiptTokenAccount", "ReceiptNativeAccount", "Amount", "AirDropTxId", "CreateTime", "UpdateTime"
FROM t_sol_airdrop_record;


INSERT INTO t_campaign_exchange_score_rule (record_id, decimal, token_mint_account, reward_token_account, reward_native_account, rate, dex_native_account, cost_rate, create_time, update_time)
SELECT "RecordId", "Decimal", "TokenMintAccount", "RewardTokenAccount", "RewardNativeAccount", "Rate", "DexNativeAccount", "CostRate", "CreateTime", "UpdateTime"
FROM t_sol_campaign_exchange_score_rule;


INSERT INTO t_determine_invite_record (record_id, brand, token_symbol, inviter_token_account, inviter_native_account, invitee_token_account, invitee_native_account, transfer_tx_id, invite_channel, create_time, update_time, level)
SELECT "RecordId", "Brand", "TokenSymbol", "InviterTokenAccount", "InviterNativeAccount", "InviteeTokenAccount", "InviteeNativeAccount", "TransferTxId", "InviteChannel", "CreateTime", "UpdateTime", "Level"
FROM t_sol_determine_invite_record;


INSERT INTO t_exchange_campaign_score_to_token_record (record_id, token_mint_account, reward_token_account, reward_native_account, receipt_token_account, receipt_native_account, provider, user_id, exchange_tx_id, score_flow_id, score_tx_id, amount, score, state, ref_block_hash, last_valid_block_height, create_time, update_time)
SELECT "RecordId", "TokenMintAccount", "RewardTokenAccount", "RewardNativeAccount", "ReceiptTokenAccount", "ReceiptNativeAccount", "Provider", "UserId", "ExchangeTxId", "ScoreFlowId", "ScoreTxId", "Amount", "Score", "State", "RefBlockHash", "LastValidBlockHeight", "CreateTime", "UpdateTime"
FROM t_sol_exchange_campaign_score_to_token_record;


INSERT INTO t_fee_statistics (record_id, slot, transaction_index, block_hash, transaction_id, compute_unit_price, compute_unit_limit, units_consumed, fee, create_time, update_time)
SELECT "RecordId", "Slot", "TransactionIndex", "BlockHash", "TransactionId", "ComputeUnitPrice", "ComputeUnitLimit", "UnitsConsumed", "Fee", "CreateTime", "UpdateTime"
FROM t_sol_fee_statistics;


INSERT INTO t_fund_flow_old (record_id, brand, token_symbol, is_token, from_native_account, to_native_account, tx_id, direction, service_type, flow_type, decimals, amount, create_time, update_time)
SELECT "RecordId", "Brand", "TokenSymbol", "IsToken", "FromNativeAccount", "ToNativeAccount", "TxId", "Direction", "ServiceType", "FlowType", "Decimals", "Amount", "CreateTime", "UpdateTime"
FROM t_sol_fund_flow_old;


INSERT INTO t_game_buy_property_record (record_id, product, user_id, property_id, tx_id, brand, token_symbol, price, quantity, amount, state, error_code, error_msg, create_time, update_time, ref_block_hash, last_valid_block_height)
SELECT "RecordId", "Product", "UserId", "PropertyId", "TxId", "Brand", "TokenSymbol", "Price", "Quantity", "Amount", "State", "ErrorCode", "ErrorMsg", "CreateTime", "UpdateTime", "RefBlockHash", "LastValidBlockHeight"
FROM t_sol_game_buy_property_record;


INSERT INTO t_game_exchange_prize_record (record_id, product, user_id, pack_id, prize_id, amount, tx_id, ref_block_hash, last_valid_block_height, state, create_time, update_time)
SELECT "RecordId", "Product", "UserId", "PackId", "PrizeId", "Amount", "TxId", "RefBlockHash", "LastValidBlockHeight", "State", "CreateTime", "UpdateTime"
FROM t_sol_game_exchange_prize_record;


INSERT INTO t_game_pack_account (product, user_id, pack_id, pack_name, create_time, update_time)
SELECT "Product", "UserId", "PackId", "PackName", "CreateTime", "UpdateTime"
FROM t_sol_game_pack_account;


INSERT INTO t_game_prize_account (product, user_id, pack_id, prize_id, amount, version, create_time, update_time)
SELECT "Product", "UserId", "PackId", "PrizeId", "Amount", "Version", "CreateTime", "UpdateTime"
FROM t_sol_game_prize_account;


INSERT INTO t_game_prize_info (product, prize_id, prize_name, token_mint_account, decimals, default, swap_amount, create_time, update_time)
SELECT "Product", "PrizeId", "PrizeName", "TokenMintAccount", "Decimals", "Default", "SwapAmount", "CreateTime", "UpdateTime"
FROM t_sol_game_prize_info;


INSERT INTO t_game_property_account (product, user_id, property_id, quantity, create_time, update_time, remark)
SELECT "Product", "UserId", "PropertyId", "Quantity", "CreateTime", "UpdateTime", "Remark"
FROM t_sol_game_property_account;


INSERT INTO t_game_property_info (product, property_id, property_name, brand, token_symbol, decimals, price)
SELECT "Product", "PropertyId", "PropertyName", "Brand", "TokenSymbol", "Decimals", "Price"
FROM t_sol_game_property_info;


INSERT INTO t_game_property_record (record_id, product, user_id, property_id, quantity, action_type, remark, create_time, update_time)
SELECT "RecordId", "Product", "UserId", "PropertyId", "Quantity", "ActionType", "Remark", "CreateTime", "UpdateTime"
FROM t_sol_game_property_record;


INSERT INTO t_game_rank_top (product, user_id, point)
SELECT "Product", "UserId", "Point"
FROM t_sol_game_rank_top;


INSERT INTO t_game_receipt_info (product, brand, token_symbol, native_account, token_account, create_time, update_time)
SELECT "Product", "Brand", "TokenSymbol", "NativeAccount", "TokenAccount", "CreateTime", "UpdateTime"
FROM t_sol_game_receipt_info;


INSERT INTO t_game_register_claim_record (record_id, product, native_account, token_account, user_name, amount, tx_id, ref_block_hash, last_valid_block_height, state, create_time, update_time)
SELECT "RecordId", "Product", "NativeAccount", "TokenAccount", "UserName", "Amount", "TxId", "RefBlockHash", "LastValidBlockHeight", "State", "CreateTime", "UpdateTime"
FROM t_sol_game_register_claim_record;


INSERT INTO t_game_register_info (product, user_id, native_account, token_account, user_name, state, create_time, update_time)
SELECT "Product", "UserId", "NativeAccount", "TokenAccount", "UserName", "State", "CreateTime", "UpdateTime"
FROM t_sol_game_register_info;


INSERT INTO t_game_role_account (product, user_id, role_id, purchased, vote_amount, create_time, update_time)
SELECT "Product", "UserId", "RoleId", "Purchased", "VoteAmount", "CreateTime", "UpdateTime"
FROM t_sol_game_role_account;


INSERT INTO t_game_role_info (product, role_id, role_name, brand, token_symbol, decimals, price, create_time, update_time)
SELECT "Product", "RoleId", "RoleName", "Brand", "TokenSymbol", "Decimals", "Price", "CreateTime", "UpdateTime"
FROM t_sol_game_role_info;


INSERT INTO t_game_unlock_role_record (record_id, product, user_id, role_id, tx_id, brand, token_symbol, decimals, price, state, error_code, error_msg, create_time, update_time, ref_block_hash, last_valid_block_height)
SELECT "RecordId", "Product", "UserId", "RoleId", "TxId", "Brand", "TokenSymbol", "Decimals", "Price", "State", "ErrorCode", "ErrorMsg", "CreateTime", "UpdateTime", "RefBlockHash", "LastValidBlockHeight"
FROM t_sol_game_unlock_role_record;


INSERT INTO t_game_vote_record (record_id, product, user_id, role_id, quantity, state, remark, create_time, update_time)
SELECT "RecordId", "Product", "UserId", "RoleId", "Quantity", "State", "Remark", "CreateTime", "UpdateTime"
FROM t_sol_game_vote_record;


INSERT INTO t_native_account_info (record_id, brand, token_symbol, native_account, token_account, invite_code, create_time, update_time)
SELECT "RecordId", "Brand", "TokenSymbol", "NativeAccount", "TokenAccount", "InviteCode", "CreateTime", "UpdateTime"
FROM t_sol_native_account_info;


INSERT INTO t_official_give_token_record (record_id, brand, token_symbol, token_mint_account, reward_token_account, reward_native_account, receipt_token_account, receipt_native_account, dex_native_account, reward_tx_id, amount, dex_fee, use_invite_code, invite_code, create_time, update_time, state, ref_block_hash, last_valid_block_height, determine_invite)
SELECT "RecordId", "Brand", "TokenSymbol", "TokenMintAccount", "RewardTokenAccount", "RewardNativeAccount", "ReceiptTokenAccount", "ReceiptNativeAccount", "DexNativeAccount", "RewardTxId", "Amount", "DexFee", "UseInviteCode", "InviteCode", "CreateTime", "UpdateTime", "State", "RefBlockHash", "LastValidBlockHeight", "DetermineInvite"
FROM t_sol_official_give_token_record;


INSERT INTO t_official_give_token_reward_record (record_id, brand, token_symbol, token_mint_account, reward_token_account, reward_native_account, receipt_token_account, receipt_native_account, dex_native_account, reward_tx_id, amount, dex_fee, use_invite_code, invite_code, create_time, update_time)
SELECT "RecordId", "Brand", "TokenSymbol", "TokenMintAccount", "RewardTokenAccount", "RewardNativeAccount", "ReceiptTokenAccount", "ReceiptNativeAccount", "DexNativeAccount", "RewardTxId", "Amount", "DexFee", "UseInviteCode", "InviteCode", "CreateTime", "UpdateTime"
FROM t_sol_official_give_token_reward_record;


INSERT INTO t_official_give_token_reward_rule (record_id, invite_code, brand, token_symbol, decimal, token_mint_account, reward_token_account, reward_native_account, dex_native_account, amount, invite_amount, dex_fee_rate, max_dex_fee, interval, default, reward_inviter, determine_invite, create_time, update_time)
SELECT "RecordId", "InviteCode", "Brand", "TokenSymbol", "Decimal", "TokenMintAccount", "RewardTokenAccount", "RewardNativeAccount", "DexNativeAccount", "Amount", "InviteAmount", "DexFeeRate", "MaxDexFee", "Interval", "Default", "RewardInviter", "DetermineInvite", "CreateTime", "UpdateTime"
FROM t_sol_official_give_token_reward_rule;


INSERT INTO t_official_transfer_token_reward_record (record_id, brand, token_symbol, token_mint_account, from_token_account, from_native_account, to_token_account, to_native_account, reward_amount, tx_id, create_time, update_time)
SELECT "RecordId", "Brand", "TokenSymbol", "TokenMintAccount", "FromTokenAccount", "FromNativeAccount", "ToTokenAccount", "ToNativeAccount", "RewardAmount", "TxId", "CreateTime", "UpdateTime"
FROM t_sol_official_transfer_token_reward_record;


INSERT INTO t_official_transfer_token_reward_rule (brand, token_symbol, decimal, token_mint_account, reward_token_account, reward_native_account, dex_native_account, reward_rate, reward_valid_address_rate, max_reward_per_tx, max_valid_address_reward_per_tx, dex_fee_rate, max_dex_fee, interval, create_time, update_time)
SELECT "Brand", "TokenSymbol", "Decimal", "TokenMintAccount", "RewardTokenAccount", "RewardNativeAccount", "DexNativeAccount", "RewardRate", "RewardValidAddressRate", "MaxRewardPerTx", "MaxValidAddressRewardPerTx", "DexFeeRate", "MaxDexFee", "Interval", "CreateTime", "UpdateTime"
FROM t_sol_official_transfer_token_reward_rule;


INSERT INTO t_pos_mission_config (record_id, reward_type, starred, rate, create_time, update_time)
SELECT "RecordId", "RewardType", "Starred", "Rate", "CreateTime", "UpdateTime"
FROM t_sol_pos_mission_config;


INSERT INTO t_pos_retweet_config (user_id, retweet_id, create_time, update_time)
SELECT "UserID", "RetweetId", "CreateTime", "UpdateTime"
FROM t_sol_pos_retweet_config;


INSERT INTO t_pos_reward (record_id, group_id, native_account, star_level, base, rate, reward_amount, reward_type, state, starred, pending, day, create_time, update_time)
SELECT "RecordId", "GroupId", "NativeAccount", "StarLevel", "Base", "Rate", "RewardAmount", "RewardType", "State", "Starred", "Pending", "Day", "CreateTime", "UpdateTime"
FROM t_sol_pos_reward;


INSERT INTO t_pos_reward_claim_record (record_id, reward_ids, tx_id, ref_block_hash, last_valid_block_height, state, create_time, update_time)
SELECT "RecordId", "RewardIds", "TxId", "RefBlockHash", "LastValidBlockHeight", "State", "CreateTime", "UpdateTime"
FROM t_sol_pos_reward_claim_record;


INSERT INTO t_pos_reward_rule (record_id, brand, token_symbol, decimal, token_mint_account, reward_token_account, reward_native_account, dex_native_account, dex_fee_rate, max_dex_fee, quote_token_amount, create_time, update_time)
SELECT "RecordId", "Brand", "TokenSymbol", "Decimal", "TokenMintAccount", "RewardTokenAccount", "RewardNativeAccount", "DexNativeAccount", "DexFeeRate", "MaxDexFee", "QuoteTokenAmount", "CreateTime", "UpdateTime"
FROM t_sol_pos_reward_rule;


INSERT INTO t_pos_snap_shot (record_id, native_account, amount, star_level, rate, range_base, day, create_time, update_time)
SELECT "RecordId", "NativeAccount", "Amount", "StarLevel", "Rate", "RangeBase", "Day", "CreateTime", "UpdateTime"
FROM t_sol_pos_snap_shot;


INSERT INTO t_pos_social_media_mission_record (record_id, provider, user_id, user_name, type, state, day, create_time, update_time)
SELECT "RecordId", "Provider", "UserID", "UserName", "Type", "State", "Day", "CreateTime", "UpdateTime"
FROM t_sol_pos_social_media_mission_record;


INSERT INTO t_pos_star_level_config (record_id, native_account, star_level, rate, create_time, update_time)
SELECT "RecordId", "NativeAccount", "StarLevel", "Rate", "CreateTime", "UpdateTime"
FROM t_sol_pos_star_level_config;


INSERT INTO t_pos_star_level_rule (record_id, amount, group_amount, star_level, rate, create_time, update_time)
SELECT "RecordId", "Amount", "GroupAmount", "StarLevel", "Rate", "CreateTime", "UpdateTime"
FROM t_sol_pos_star_level_rule;


INSERT INTO t_pos_twitter_oauth_2_config (user_id, client_id, client_secret, api_key, api_secret, bearer_token, access_token, access_token_secret, web_redirect_url, o_auth_call_back_url, create_time, update_time)
SELECT "UserID", "ClientId", "ClientSecret", "ApiKey", "ApiSecret", "BearerToken", "AccessToken", "AccessTokenSecret", "WebRedirectURL", "OAuthCallBackURL", "CreateTime", "UpdateTime"
FROM t_sol_pos_twitter_oauth2_config;


INSERT INTO t_qn_fee (id, slot, low_avg, medium_avg, high_avg, create_time, update_time)
SELECT "Id", "Slot", "LowAvg", "MediumAvg", "HighAvg", "CreateTime", "UpdateTime"
FROM t_sol_qn_fee;


INSERT INTO t_scan_info (brand, token_symbol, token_mint_account, create_token_tx_id, until_tx_id, before_tx_id, native_account, token_account)
SELECT "Brand", "TokenSymbol", "TokenMintAccount", "CreateTokenTxId", "UntilTxId", "BeforeTxId", "NativeAccount", "TokenAccount"
FROM t_sol_scan_info;


INSERT INTO t_stake_fix_interest_config (record_id, min_amount, stake_type, fix_rate, individual_rate, create_time, update_time)
SELECT "RecordId", "MinAmount", "StakeType", "FixRate", "IndividualRate", "CreateTime", "UpdateTime"
FROM t_sol_stake_fix_interest_config;


INSERT INTO t_stake_invite_dist (level)
SELECT "Level"
FROM t_sol_stake_invite_dist;


INSERT INTO t_stake_invite_rate (level, rate)
SELECT "Level", "Rate"
FROM t_sol_stake_invite_rate;


INSERT INTO t_stake_reward (record_id, group_id, native_account, star_level, base, rate, reward_amount, reward_type, state, starred, pending, day, create_time, update_time, tx_id)
SELECT "RecordId", "GroupId", "NativeAccount", "StarLevel", "Base", "Rate", "RewardAmount", "RewardType", "State", "Starred", "Pending", "Day", "CreateTime", "UpdateTime", "TxId"
FROM t_sol_stake_reward;


INSERT INTO t_stake_reward_claim_record (record_id, reward_ids, tx_id, ref_block_hash, last_valid_block_height, state, create_time, update_time)
SELECT "RecordId", "RewardIds", "TxId", "RefBlockHash", "LastValidBlockHeight", "State", "CreateTime", "UpdateTime"
FROM t_sol_stake_reward_claim_record;


INSERT INTO t_stake_reward_rule (record_id, decimal, token_mint_account, stake_admin, program_id, faucet_token_account, faucet_native_account, reward_token_account, reward_native_account, dex_native_account, quote_token_amount, dex_fee_rate, create_time, update_time)
SELECT "RecordId", "Decimal", "TokenMintAccount", "StakeAdmin", "ProgramId", "FaucetTokenAccount", "FaucetNativeAccount", "RewardTokenAccount", "RewardNativeAccount", "DexNativeAccount", "QuoteTokenAmount", "DexFeeRate", "CreateTime", "UpdateTime"
FROM t_sol_stake_reward_rule;


INSERT INTO t_stake_snap_shot (record_id, native_account, amount, stake_type, day, create_time, update_time)
SELECT "RecordId", "NativeAccount", "Amount", "StakeType", "Day", "CreateTime", "UpdateTime"
FROM t_sol_stake_snap_shot;


INSERT INTO t_stake_star_level_config (record_id, native_account, star_level, rate, create_time, update_time)
SELECT "RecordId", "NativeAccount", "StarLevel", "Rate", "CreateTime", "UpdateTime"
FROM t_sol_stake_star_level_config;


INSERT INTO t_stake_star_level_rule (record_id, amount, group_amount, star_level, rate, create_time, update_time)
SELECT "RecordId", "Amount", "GroupAmount", "StarLevel", "Rate", "CreateTime", "UpdateTime"
FROM t_sol_stake_star_level_rule;


INSERT INTO t_swap_new_token_config (brand, token_symbol, old_token_mint_account, new_token_mint_account, old_decimals, new_decimals, receipt_native_account, receipt_token_account, send_native_account, send_token_account, rate, create_time, update_time)
SELECT "Brand", "TokenSymbol", "OldTokenMintAccount", "NewTokenMintAccount", "OldDecimals", "NewDecimals", "ReceiptNativeAccount", "ReceiptTokenAccount", "SendNativeAccount", "SendTokenAccount", "Rate", "CreateTime", "UpdateTime"
FROM t_sol_swap_new_token_config;


INSERT INTO t_swap_token_record (record_id, input_token_symbol, input_token_mint_account, input_token_decimal, output_token_symbol, output_token_mint_account, output_token_decimal, user_native_account, user_token_account, tx_id, tx_type, input_token_amount, output_token_amount, state, ref_block_hash, last_valid_block_height, create_time, update_time, house_id, batch)
SELECT "RecordId", "InputTokenSymbol", "InputTokenMintAccount", "InputTokenDecimal", "OutputTokenSymbol", "OutputTokenMintAccount", "OutputTokenDecimal", "UserNativeAccount", "UserTokenAccount", "TxId", "TxType", "InputTokenAmount", "OutputTokenAmount", "State", "RefBlockHash", "LastValidBlockHeight", "CreateTime", "UpdateTime", "HouseId", "Batch"
FROM t_sol_swap_token_record;


INSERT INTO t_swap_token_rule (record_id, input_token_symbol, input_token_mint_account, input_token_decimal, output_token_symbol, output_token_mint_account, output_token_decimal, output_token_account, output_native_account, cost_token_account, cost_native_account, rate, create_time, update_time)
SELECT "RecordId", "InputTokenSymbol", "InputTokenMintAccount", "InputTokenDecimal", "OutputTokenSymbol", "OutputTokenMintAccount", "OutputTokenDecimal", "OutputTokenAccount", "OutputNativeAccount", "CostTokenAccount", "CostNativeAccount", "Rate", "CreateTime", "UpdateTime"
FROM t_sol_swap_token_rule;


INSERT INTO t_token_config (brand, token_symbol, decimal, token_mint_account, minter_native_account, air_drop_token_account, air_drop_native_account, reward_token_account, reward_native_account, create_time, update_time)
SELECT "Brand", "TokenSymbol", "Decimal", "TokenMintAccount", "MinterNativeAccount", "AirDropTokenAccount", "AirDropNativeAccount", "RewardTokenAccount", "RewardNativeAccount", "CreateTime", "UpdateTime"
FROM t_sol_token_config;


INSERT INTO t_transfer_checked_record (record_id, token_mint_account, from_token_account, from_native_account, receipt_token_account, receipt_native_account, owner_native_account, transfer_tx_id, instruction_index, transfer_amount, create_time, update_time, state, ref_block_hash, last_valid_block_height)
SELECT "RecordId", "TokenMintAccount", "FromTokenAccount", "FromNativeAccount", "ReceiptTokenAccount", "ReceiptNativeAccount", "OwnerNativeAccount", "TransferTxId", "InstructionIndex", "TransferAmount", "CreateTime", "UpdateTime", "State", "RefBlockHash", "LastValidBlockHeight"
FROM t_sol_transfer_checked_record;


INSERT INTO t_transfer_reward_claim (brand, token_symbol, level, claim_ratio)
SELECT "Brand", "TokenSymbol", "Level", "ClaimRatio"
FROM t_sol_transfer_reward_claim;


INSERT INTO t_transfer_reward_distribution (brand, token_symbol, level)
SELECT "Brand", "TokenSymbol", "Level"
FROM t_sol_transfer_reward_distribution;


INSERT INTO t_transfer_reward_record (record_id, brand, token_symbol, token_mint_account, from_token_account, from_native_account, receipt_token_account, receipt_native_account, transfer_amount, transfer_tx_id, reward_tx_id, state, error_message, create_time, update_time, version, ref_block_hash, last_valid_block_height, reward_state)
SELECT "RecordId", "Brand", "TokenSymbol", "TokenMintAccount", "FromTokenAccount", "FromNativeAccount", "ReceiptTokenAccount", "ReceiptNativeAccount", "TransferAmount", "TransferTxId", "RewardTxId", "State", "ErrorMessage", "CreateTime", "UpdateTime", "Version", "RefBlockHash", "LastValidBlockHeight", "RewardState"
FROM t_sol_transfer_reward_record;


INSERT INTO t_transfer_to_dex_record (record_id, from_native_account, receipt_native_account, tx_id, transfer_amount, create_time, update_time)
SELECT "RecordId", "FromNativeAccount", "ReceiptNativeAccount", "TxId", "TransferAmount", "CreateTime", "UpdateTime"
FROM t_sol_transfer_to_dex_record;


INSERT INTO t_transfer_token_reward_rule (brand, token_symbol, decimal, token_mint_account, reward_token_account, reward_native_account, account_exist_slot, account_exist_buffer_slot, reward_rate, max_reward_per_tx, create_time, update_time)
SELECT "Brand", "TokenSymbol", "Decimal", "TokenMintAccount", "RewardTokenAccount", "RewardNativeAccount", "AccountExistSlot", "AccountExistBufferSlot", "RewardRate", "MaxRewardPerTx", "CreateTime", "UpdateTime"
FROM t_sol_transfer_token_reward_rule;

