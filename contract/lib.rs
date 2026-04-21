#![allow(clippy::result_large_err)]

use anchor_lang::prelude::*;

use anchor_spl::{
    associated_token::AssociatedToken,
    token::{self, Mint, Token, TokenAccount, Transfer, transfer}
};

use anchor_lang::solana_program::clock::Clock;

declare_id!("CyLTEgvmqVF9dPJkT6bMgccfXL7G26EXRAM9FEuP5ki6");

pub mod constants{
    pub const VAULT_SEED: &[u8] = b"vault";
    pub const STAKE_INFO_SEED: &[u8] = b"stake_info";
    pub const TOKEN_SEED: &[u8] = b"token";
    pub const CONFIG_SEED: &[u8] = b"config";

    // 全局常量：最大质押记录数，可以在这里修改
    pub const MAX_STAKE_RECORDS: usize = 10;
}

#[program]
pub mod staking {
    use super::*;

    pub fn initialize(ctx: Context<Initialize>, allowed_mint: Pubkey) -> Result<()> {
        msg!("Initialize");

        let config = &mut ctx.accounts.config;
        config.allowed_mint = allowed_mint;
        config.authority = ctx.accounts.signer.key();

        // 根据代币精度计算最小质押数量
        let decimals = ctx.accounts.mint.decimals;
        config.min_stake_amount = 100000u64
            .checked_mul(10u64.pow(decimals as u32))
            .ok_or(ErrorCode::CalculationError)?;

        msg!("Initialized with allowed mint: {} and min stake amount: {}",
             config.allowed_mint, config.min_stake_amount);
        msg!("Token decimals: {}, Base amount: 1,000,000", decimals);
        Ok(())
    }

    // 质押功能 - 每次创建新记录
    pub fn stake(ctx: Context<Stake>, amount: u64, stake_type: u8) -> Result<()> {
        // 验证管理员签名
        if ctx.accounts.admin_signer.key() != ctx.accounts.config.authority {
            return Err(ErrorCode::Unauthorized.into());
        }

        // 验证质押代币是否被允许
        if ctx.accounts.mint.key() != ctx.accounts.config.allowed_mint {
            return Err(ErrorCode::InvalidToken.into());
        }

        // 验证质押类型
        if stake_type != 0 && stake_type != 1 {
            return Err(ErrorCode::InvalidStakeType.into());
        }

        // 动态获取精度并计算实际转移数量
        let decimals = ctx.accounts.mint.decimals;
        let transfer_amount = amount.checked_mul(10u64.pow(decimals as u32))
            .ok_or(ErrorCode::CalculationError)?;

        // 验证质押数量是否达到最小要求
        if transfer_amount < ctx.accounts.config.min_stake_amount {
            return Err(ErrorCode::InsufficientStakeAmount.into());
        }

        let clock = Clock::get()?;
        let current_slot = clock.slot;

        // 计算锁定期
        // 3分钟 3 * 60 * 1000 / 400 = 450
        // 5分钟 5 * 60 * 1000 / 400 = 750
        // 180天 180 * 24 * 60 * 60 * 1000 / 400 = 38880000
        // 360天 360 * 24 * 60 * 60 * 1000 / 400 = 77760000
        let lock_period_slots = match stake_type {
            0 => 38880000, // 180天
            1 => 77760000, // 360天
            _ => return Err(ErrorCode::InvalidStakeType.into()),
        };

        let stake_end_slot = current_slot.checked_add(lock_period_slots)
            .ok_or(ErrorCode::CalculationError)?;

        let stake_info = &mut ctx.accounts.stake_info_account;

        // 如果是新创建的质押信息账户，设置用户钱包地址
        if stake_info.user_wallet == Pubkey::default() {
            stake_info.user_wallet = ctx.accounts.signer.key();
        }

        // 检查是否有可用的记录位置（只查找完全空的记录，不替换过期记录）
        let available_position = stake_info.stakes.iter()
            .position(|record| record.staked_amount == 0);

        if let Some(position) = available_position {
            // 创建新的质押记录
            let new_record = StakeRecord {
                stake_type,
                staked_amount: amount,
                stake_start_slot: current_slot,
                stake_end_slot,
            };

            stake_info.stakes[position] = new_record;

            msg!("Created new stake record at position: {}, type: {}, amount: {}",
                 position, stake_type, amount);
        } else {
            // 没有可用位置，直接报错
            msg!("All stake positions are occupied. Maximum stake records reached: {}", constants::MAX_STAKE_RECORDS);
            return Err(ErrorCode::MaxStakeRecordsReached.into());
        }

        transfer(
            CpiContext::new(
                ctx.accounts.token_program.to_account_info(),
                Transfer {
                    from: ctx.accounts.user_token_account.to_account_info(),
                    to: ctx.accounts.stake_account.to_account_info(),
                    authority: ctx.accounts.signer.to_account_info(),
                }
            ),
            transfer_amount
        )?;

        msg!("Staked {} tokens (type: {}, with {} decimals), total transfer: {}",
             amount, stake_type, decimals, transfer_amount);
        Ok(())
    }

    // 解除质押 - 只能解除过期的质押
    pub fn unstake(ctx: Context<DeStake>, stake_index: u8) -> Result<()> {
        // 新增：验证管理员签名
        if ctx.accounts.admin_signer.key() != ctx.accounts.config.authority {
            return Err(ErrorCode::Unauthorized.into());
        }
        let stake_info: &mut Account<'_, StakeInfo> = &mut ctx.accounts.stake_info_account;

        // 验证索引范围
        if stake_index as usize >= constants::MAX_STAKE_RECORDS {
            return Err(ErrorCode::InvalidStakeIndex.into());
        }

        let clock = Clock::get()?;
        let current_slot = clock.slot;

        msg!("Unstake attempt - Index: {}, Current slot: {}", stake_index, current_slot);

        let stake_record = &stake_info.stakes[stake_index as usize];

        // 检查记录是否存在
        if stake_record.staked_amount == 0 {
            msg!("No stake record found at index: {}", stake_index);
            return Err(ErrorCode::NoStakingRecord.into());
        }

        // 检查锁定期是否结束
        if current_slot < stake_record.stake_end_slot {
            let remaining_slots = stake_record.stake_end_slot - current_slot;
            let remaining_seconds = remaining_slots * 400 / 1000; // 估算剩余秒数
            msg!("Lock period not ended. Remaining: {} slots (~{} seconds)",
                remaining_slots, remaining_seconds);
            return Err(ErrorCode::LockPeriodNotEnded.into());
        }

        msg!("Lock period ended. Proceeding with unstake...");

        let stake_amount: u64 = ctx.accounts.stake_account.amount;

        // 计算应该提取的数量（考虑代币精度）
        let decimals = ctx.accounts.mint.decimals;
        let expected_withdraw_amount = stake_record.staked_amount
            .checked_mul(10u64.pow(decimals as u32))
            .ok_or(ErrorCode::CalculationError)?;

        msg!("Expected withdraw amount: {} (with {} decimals)", expected_withdraw_amount, decimals);
        msg!("Stake account balance: {}", stake_amount);

        // 验证质押账户中的代币数量足够
        if stake_amount < expected_withdraw_amount {
            msg!("Insufficient stake balance: required {}, available {}",
                expected_withdraw_amount, stake_amount);
            return Err(ErrorCode::InsufficientStakeBalance.into());
        }

        let staker = ctx.accounts.signer.key();
        let bump = ctx.bumps.stake_account;
        let signer: &[&[&[u8]]] = &[&[constants::TOKEN_SEED, staker.as_ref(), &[bump]]];

        // 执行代币转移
        transfer(
            CpiContext::new_with_signer(
                ctx.accounts.token_program.to_account_info(),
                Transfer {
                    from: ctx.accounts.stake_account.to_account_info(),
                    to: ctx.accounts.user_token_account.to_account_info(),
                    authority: ctx.accounts.stake_account.to_account_info(),
                },
                signer
            ),
            expected_withdraw_amount,
        )?;

        msg!("Successfully unstaked amount: {} (index: {})", stake_record.staked_amount, stake_index);

        // 清除质押记录
        stake_info.stakes[stake_index as usize].staked_amount = 0;

        msg!("Unstake completed for index: {}", stake_index);
        Ok(())
    }

    // 重新质押 - 对已到期的记录原地续期，不移动token
    pub fn restake(ctx: Context<ReStake>, stake_index: u8, stake_type: u8) -> Result<()> {
        // 验证管理员签名
        if ctx.accounts.admin_signer.key() != ctx.accounts.config.authority {
            return Err(ErrorCode::Unauthorized.into());
        }

        // 验证质押类型
        if stake_type != 0 && stake_type != 1 {
            return Err(ErrorCode::InvalidStakeType.into());
        }

        // 验证索引范围
        if stake_index as usize >= constants::MAX_STAKE_RECORDS {
            return Err(ErrorCode::InvalidStakeIndex.into());
        }

        let clock = Clock::get()?;
        let current_slot = clock.slot;

        let stake_info = &mut ctx.accounts.stake_info_account;
        let record = &stake_info.stakes[stake_index as usize];

        // 检查记录是否存在
        if record.staked_amount == 0 {
            msg!("No stake record found at index: {}", stake_index);
            return Err(ErrorCode::NoStakingRecord.into());
        }

        // 检查记录是否已到期
        if current_slot < record.stake_end_slot {
            let remaining_slots = record.stake_end_slot - current_slot;
            msg!("Lock period not ended. Remaining: {} slots", remaining_slots);
            return Err(ErrorCode::LockPeriodNotEnded.into());
        }

        // 计算新的锁定期
        let lock_period_slots = match stake_type {
            0 => 450, // 3分钟
            1 => 750, // 5分钟
            _ => return Err(ErrorCode::InvalidStakeType.into()),
        };

        let new_stake_end_slot = current_slot.checked_add(lock_period_slots)
            .ok_or(ErrorCode::CalculationError)?;

        // 原地更新记录，不移动token
        let record = &mut stake_info.stakes[stake_index as usize];
        record.stake_type = stake_type;
        record.stake_start_slot = current_slot;
        record.stake_end_slot = new_stake_end_slot;

        msg!("Restaked at index: {}, type: {}, amount: {}, new end slot: {}",
             stake_index, stake_type, record.staked_amount, new_stake_end_slot);
        Ok(())
    }

    // 只读函数 - 获取质押信息（不需要签名）
    pub fn get_stake_info(ctx: Context<GetStakeInfo>) -> Result<StakeInfoResponse> {
        let stake_info = &ctx.accounts.stake_info_account;
        let clock = Clock::get()?;
        let current_slot = clock.slot;

        // 返回所有质押记录，包括活跃的和过期的
        let all_stakes: Vec<StakeRecordWithIndex> = stake_info.stakes.iter()
            .enumerate()
            .map(|(index, record)| StakeRecordWithIndex {
                index: index as u8,
                stake_record: record.clone(),
                is_expired: record.is_expired(current_slot),
                is_active: record.staked_amount > 0,
            })
            .collect();

        Ok(StakeInfoResponse {
            user_wallet: stake_info.user_wallet,
            stakes: all_stakes,
            current_slot,
            max_stake_records: constants::MAX_STAKE_RECORDS as u8,
        })
    }

    // 更新最小质押金额（仅管理员）
    pub fn update_min_stake_amount(ctx: Context<UpdateConfig>, new_min_base_amount: u64) -> Result<()> {
        let config = &mut ctx.accounts.config;

        // 验证调用者是否是管理员
        if ctx.accounts.signer.key() != config.authority {
            return Err(ErrorCode::Unauthorized.into());
        }

        // 根据代币精度计算实际最小质押数量
        let decimals = ctx.accounts.mint.decimals;
        config.min_stake_amount = new_min_base_amount
            .checked_mul(10u64.pow(decimals as u32))
            .ok_or(ErrorCode::CalculationError)?;

        msg!("Updated min stake amount to: {} (base units), actual: {} (with {} decimals)",
            new_min_base_amount, config.min_stake_amount, decimals);
        Ok(())
    }
}

#[derive(Accounts)]
pub struct Initialize<'info> {
    #[account(mut)]
    pub signer: Signer<'info>,

    #[account(
        init,
        seeds = [constants::CONFIG_SEED],
        bump,
        payer = signer,
        space = 8 + std::mem::size_of::<Config>()
    )]
    pub config: Account<'info, Config>,

    // 添加mint账户用于获取代币精度
    pub mint: Account<'info, Mint>,

    pub system_program: Program<'info, System>,
}

#[derive(Accounts)]
pub struct Stake<'info>{
    #[account(mut)]
    pub signer: Signer<'info>,

    // 添加管理员签名验证
    pub admin_signer: Signer<'info>,

    #[account(
        seeds = [constants::CONFIG_SEED],
        bump,
    )]
    pub config: Account<'info, Config>,

    #[account(
        init_if_needed,
        seeds = [constants::STAKE_INFO_SEED, signer.key.as_ref()],
        bump,
        payer = signer,
        space = 8 + std::mem::size_of::<StakeInfo>()
    )]
    pub stake_info_account: Account<'info, StakeInfo>,

    #[account(
        init_if_needed,
        seeds =  [constants::TOKEN_SEED, signer.key.as_ref()],
        bump,
        payer = signer,
        token::mint = mint,
        token::authority = stake_account,
    )]
    pub stake_account : Account<'info, TokenAccount>,

    #[account(
        mut,
        associated_token::mint = mint,
        associated_token::authority = signer,
    )]
    pub user_token_account: Account<'info, TokenAccount>,

    pub mint: Account<'info, Mint>,
    pub token_program: Program<'info, Token>,
    pub associated_token_program: Program<'info, AssociatedToken>,
    pub system_program: Program<'info, System>,
}

#[derive(Accounts)]
pub struct DeStake<'info>{
      #[account(mut)]
      pub signer: Signer<'info>,

      // 新增：管理员签名
      pub admin_signer: Signer<'info>,

      // 新增：用于验证 admin_signer 是否是合约 authority
      #[account(
          seeds = [constants::CONFIG_SEED],
          bump,
      )]
      pub config: Account<'info, Config>,

      #[account(
          mut,
          seeds = [constants::STAKE_INFO_SEED, signer.key.as_ref()],
          bump,
      )]
      pub stake_info_account: Account<'info, StakeInfo>,

      #[account(
          mut,
          seeds = [constants::TOKEN_SEED, signer.key.as_ref()],
          bump,
      )]
      pub stake_account: Account<'info, TokenAccount>,

      #[account(
          mut,
          associated_token::mint = mint,
          associated_token::authority = signer,
      )]
      pub user_token_account: Account<'info, TokenAccount>,

      pub mint: Account<'info, Mint>,
      pub token_program: Program<'info, Token>,
      pub associated_token_program: Program<'info, AssociatedToken>,
      pub system_program: Program<'info, System>,
}

#[derive(Accounts)]
pub struct ReStake<'info> {
    #[account(mut)]
    pub signer: Signer<'info>,

    pub admin_signer: Signer<'info>,

    #[account(
        seeds = [constants::CONFIG_SEED],
        bump,
    )]
    pub config: Account<'info, Config>,

    #[account(
        mut,
        seeds = [constants::STAKE_INFO_SEED, signer.key.as_ref()],
        bump,
    )]
    pub stake_info_account: Account<'info, StakeInfo>,

    pub mint: Account<'info, Mint>,
    pub system_program: Program<'info, System>,
}

#[derive(Accounts)]
pub struct GetStakeInfo<'info> {
    #[account(
        seeds = [constants::STAKE_INFO_SEED, stake_info_owner.key.as_ref()],
        bump,
    )]
    pub stake_info_account: Account<'info, StakeInfo>,
    /// CHECK: 这只是一个只读引用，不需要签名
    pub stake_info_owner: AccountInfo<'info>,
}

#[derive(Accounts)]
pub struct UpdateConfig<'info> {
    #[account(mut)]
    pub signer: Signer<'info>,

    #[account(
        mut,
        seeds = [constants::CONFIG_SEED],
        bump,
    )]
    pub config: Account<'info, Config>,

    // 添加 mint 账户用于获取代币精度
    pub mint: Account<'info, Mint>,
}

#[account]
pub struct Config {
    pub allowed_mint: Pubkey,    // 允许质押的代币地址
    pub authority: Pubkey,       // 管理员地址
    pub min_stake_amount: u64,   // 最小质押数量（考虑精度后的实际数量）
}

#[account]
pub struct StakeInfo {
    pub user_wallet: Pubkey,     // 用户钱包地址
    pub stakes: [StakeRecord; constants::MAX_STAKE_RECORDS], // 使用全局常量定义数组大小
}

#[derive(AnchorSerialize, AnchorDeserialize, Clone, Copy, Debug)]
pub struct StakeRecord {
    pub stake_type: u8,          // 质押类型：0=180天，1=360天
    pub staked_amount: u64,      // 质押的token数量（基础单位）
    pub stake_start_slot: u64,   // 质押开始的slot
    pub stake_end_slot: u64,     // 质押结束的slot
}

impl StakeRecord {
    pub fn is_expired(&self, current_slot: u64) -> bool {
        current_slot >= self.stake_end_slot
    }
}

// 返回给客户端的质押信息响应结构（包含索引信息）
#[derive(AnchorSerialize, AnchorDeserialize, Clone, Debug)]
pub struct StakeInfoResponse {
    pub user_wallet: Pubkey,                    // 用户钱包地址
    pub stakes: Vec<StakeRecordWithIndex>,      // 所有质押记录（包含索引）
    pub current_slot: u64,                      // 当前slot
    pub max_stake_records: u8,                  // 最大质押记录数
}

// 包含索引的质押记录
#[derive(AnchorSerialize, AnchorDeserialize, Clone, Debug)]
pub struct StakeRecordWithIndex {
    pub index: u8,                              // 记录索引
    pub stake_record: StakeRecord,              // 质押记录
    pub is_expired: bool,                       // 是否过期
    pub is_active: bool,                        // 是否活跃（有质押金额）
}

#[error_code]
pub enum ErrorCode{
    #[msg("Tokens are already staked")]
    IsStaked,
    #[msg("No staking record found for this index")]
    NoStakingRecord,
    #[msg("Staking record found but lock period not ended")]
    LockPeriodNotEnded,
    #[msg("No Tokens or too small to stake")]
    NoTokens,
    #[msg("Invalid token for staking")]
    InvalidToken,
    #[msg("Token mismatch during unstake")]
    TokenMismatch,
    #[msg("Unauthorized access")]
    Unauthorized,
    #[msg("Calculation error")]
    CalculationError,
    #[msg("Invalid stake type")]
    InvalidStakeType,
    #[msg("Insufficient stake balance")]
    InsufficientStakeBalance,
    #[msg("The number of staking records has reached the maximum. Please wait for the staking period to end and unstake.")]
    MaxStakeRecordsReached,
    #[msg("Insufficient stake amount. Minimum is 1,000,000 tokens")]
    InsufficientStakeAmount,
    #[msg("Invalid stake index")]
    InvalidStakeIndex,
}