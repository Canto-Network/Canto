---
Title: Liquidstaking
Description: A high-level overview of how the command-line interfaces (CLI) works for the liquidstaking module.
---

# Liquidstaking Module

## Synopsis

This document provides a high-level overview of how the command line (CLI) interface works for the `liquidstaking` module.
To set up a local testing environment, you should run [init_testnet.sh](https://github.com/b-harvest/Canto/blob/liquidstaking-module/init_testnet.sh)

Note that [jq](https://stedolan.github.io/jq/) is recommended to be installed as it is used to process JSON throughout the document.

## Command Line Interfaces

- [Transaction](#Transaction)
  - [ProvideInsurance](#ProvideInsurance)
  - [CancelProvideInsurance](#CancelProvideInsurance)
  - [LiquidStake](#LiquidStake)
  - [LiquidUnstake](#LiquidUnstake)
  - [DepositInsurance](#DepositInsurance)
  - [WithdrawInsurance](#WithdrawInsurance)
  - [WithdrawInsuranceCommission](#WithdrawInsuranceCommission)
  - [ClaimDiscountedReward](#ClaimDiscountedReward)
- [Query](#Query)
  - [Params](#Params)
  - [Epoch](#Epoch)
  - [Chunks](#Chunks)
  - [Chunk](#Chunk)
  - [Insurances](#Insurances)
  - [Insurance](#Insurance)
  - [WithdrawInsuranceRequests](#WithdrawInsuranceRequests)
  - [WithdrawInsuranceRequest](#WithdrawInsuranceRequest)
  - [UnpairingForUnstakingChunkInfos](#UnpairingForUnstakingChunkInfos)
  - [UnpairingForUnstakingChunkInfo](#UnpairingForUnstakingChunkInfo)
  - [RedelegationInfos](#RedelegationInfos)
  - [RedelegationInfo](#RedelegationInfo)
  - [ChunkSize](#ChunkSize)
  - [MinimumCollateral](#MinimumCollateral)
  - [States](#States)

# Transaction

## ProvideInsurance

Provide insurance.

Usage

```bash
provide-insurance [validator-address] [amount] [fee-rate]
```

| **Argument**      | **Description**                                                                                                          |
|:------------------|:-------------------------------------------------------------------------------------------------------------------------|
| validator-address | the validator address that the insurance provider wants to cover                                                         |
| amount            | amount of collalteral; it must be acanto and amount must be bigger than 7% of ChunkSize(=250K) tokens(9% is recommended) |
| fee-rate          | how much commission will you receive for providing insurance? (fee-rate x chunk's delegation reward) will be commission. |

Example

```bash
# Provide insurance with 9% of ChunkSize collateral and 10% as fee-rate.
cantod tx liquidstaking provide-insurance <validator-address> 22500000000000000000000acanto 0.1 --from key1 --fees 200000acanto  \
--from key1 \
--keyring-backend test \
--fees 200000acanto \
--output json | jq

#
# Tips
# 
# Query validators first you want to cover and copy operator_address of the validator.
# And use that address at <validator-address>
cantod q staking validators
#
# Query chunks
# You can see newly created insurances (initial status of insurance is "Pairing")
cantod q liquidstaking insurances -o json | jq
```

## CancelProvideInsurance

Provide insurance.

Usage

```bash
cancel-provide-insurance [insurance-id]
```

| **Argument** | **Description**                                |
|:-------------|:-----------------------------------------------|
| insurance-id | the id of pairing insurance you want to cancel |

Example

```bash
cantod tx liquidstaking cancel-provide-insurance 3
--from key1 \
--keyring-backend test \
--fees 200000acanto \
--output json | jq

#
# Tips
#
# Query insurances
# If it is succeeded, then you cannot see the insurance with the id in result.
cantod q liquidstaking insurances -o json | jq
```

## LiquidStake

Liquid stake coin.

Usage

```bash
liquid-stake [amount]
```

| **Argument**  | **Description**                                                                                          |
| :------------ |:---------------------------------------------------------------------------------------------------------|
| amount        | amount of coin to stake; it must be acanto and amount must be multiple of ChunkSize(=250K) tokens |

Example

```bash
# Liquid stake 1 chunk (250K tokens)
cantod tx liquidstaking liquid-stake 250000000000000000000000acanto \
--from key1 \
--keyring-backend test \
--fees 3000000acanto \
--gas 3000000 \
--output json | jq

#
# Tips
#
# Query account balances
# If liquid stake succeeded, you can see the newly minted lsToken
cantod q bank balances <address> -o json | jq

# Query chunks
# And you can see newly created chunk with new id
cantod q liquidstaking chunks -o json | jq
```

## LiquidUnstake

Liquid stake coin.

Usage

```bash
liquid-unstake [amount]
```

| **Argument**  | **Description**                                                                                      |
| :------------ |:-----------------------------------------------------------------------------------------------------|
| amount        | amount of coin to un-stake; it must be acanto and amount must be multiple of ChunkSize(=250K) tokens |

Example

```bash
# Liquid unstake 1 chunk (250K tokens)
cantod tx liquidstaking liquid-unstake 250000000000000000000000acanto \
--from key1 \
--keyring-backend test \
--fees 3000000acanto
--gas 3000000 \
--output json | jq

#
# Tips
#
# Query account balances
# If liquid unstake request is accepted, you can see lsToken corresponding msg.Amount is escrowed(=decreased).
# When the actual unstaking process is finished, then you can see unstaked token in your account.
# Notice the newly minted lsToken
cantod q bank balances <address> -o json | jq

# Query your unstaking request
# If your unstake request is accepted, then you can query your unstaking request.
cantod q liquidstaking unpairing-for-unstaking-chunk-infos --queued="true" -o json | jq
```

## DepositInsurance

Deposit more coins to insurance

Usage

```bash
deposit-insurance [insurance-id] [amount]
```

| **Argument** | **Description**                                |
|:-------------|:-----------------------------------------------|
| insurance-id | the id of insurance you want to deposit        |
| amount       | amount of coin to deposit; it must be acanto   |

Example

```bash
# Deposit
cantod tx liquidstaking deposit-insurance 1 22500000000000000000000acanto \
--from key1 \
--keyring-backend test \
--fees 3000000acanto
--gas 3000000 \
--output json | jq

#
# Tips
#
# Query balance of insurance's derived address
# Notice the added token
cantod q bank balances <derived_address> -o json | jq
```

## WithdrawInsurance

Withdraw insurance

Usage

```bash
withdraw-insurance [insurance-id]
```

| **Argument** | **Description**                          |
|:-------------|:-----------------------------------------|
| insurance-id | the id of insurance you want to withdraw |

Example

```bash
# Withdraw insurance 
cantod tx liquidstaking withdraw-insurance 1 \
--from key1 \
--keyring-backend test \
--fees 3000000acanto
--gas 3000000 \
--output json | jq

#
# Tips
#
# Query balance of insurance's derived address
# Notice the added token
cantod q bank balances <derived_address> -o json | jq

# Query your unstaking request
# If your unstake request is accepted, then you can query your unstaking request.
cantod q liquidstaking withdraw-insurance-requests -o json | jq

# If send request to already Unpaired insurance, then insurance is removed from state
# and you got insurance's deposit and its commissions back.
cantod q liquidstaking insurances
cantod q bank balances <provider_address> -o json | jq
```

## WithdrawInsuranceCommission

Withdraw insurance commission

Usage

```bash
withdraw-insurance-commission [insurance-id]
```

| **Argument** | **Description**                                     |
|:-------------|:----------------------------------------------------|
| insurance-id | the id of insurance you want to withdraw commission |

Example

```bash
# Withdraw insurance commission
cantod tx liquidstaking withdraw-insurance-commission 1 \
--from key1 \
--keyring-backend test \
--fees 3000000acanto
--gas 3000000 \
--output json | jq

#
# Tips
#
# Query balance of insurance's feepool address before withdraw
# Notice this balance is decreased after withdraw commission.
cantod q bank balances <fee_pool_address> -o json | jq
cantod q bank balances <provider_address> -o json | jq
```

## ClaimDiscountedReward

Claim discounted reward

Usage

```bash
claim-discounted-reward [amount] [minimum-discount-rate]
```

| **Argument**          | **Description**                                                             |
|:----------------------|:----------------------------------------------------------------------------|
| amount                | amount of coin willing to burn to get discounted reward; it must be lscanto |
| minimum-discount-rate | if current discount rate is lower than this, then msg will be rejected.     |

Example

```bash
# Claim discounted reward
cantod tx liquidstaking claim-discounted-reward 1000lscanto 0.009 \
--from key1 \
--keyring-backend test \
--fees 3000000acanto
--gas 3000000 \
--output json | jq

#
# Tips
#
# Query states
# If it is successful, then you can see decreased reward_module_acc_balance and ls_tokens_total_supply.
# And your acanto balance will be increased.
cantod q liquidstaking states
cantod q bank balances <address> -o json | jq
```


# Query

## Params


Query the current liquidstaking parameters information.

Usage

```bash
params
```

Example

```bash
cantod query liquidstaking params -o json | jq
```

## Epoch

Query the epoch information.

Usage

```bash
epoch
```

Example

```bash
cantod query liquidstaking epoch -o json | jq
```

## Chunks

Query chunks.

Usage

```bash
chunks
```

Example

```bash
# Query all chunks
cantod query liquidstaking chunks -o json | jq
# Query only paired chunks.
cantod query liquidstaking chunks --status CHUNK_STATUS_PAIRED -o json | jq
```

## Chunk

Query a chunk by id.

Usage

```bash
chunks [chunk-id]
```

Example

```bash
cantod query liquidstaking chunk 1 -o json | jq
```

## Insurances

Query insurances.

Usage

```bash
insurances
```

Example

```bash
# Query all insurances
cantod query liquidstaking insurances -o json | jq
# Query only paired insurances.
cantod query liquidstaking insurances --status INSURANCE_STATUS_PAIRED -o json | jq
# Query insurances by provider address.
cantod query liquidstaking insurances --provider-address canto1xjlslz2vl7v6gu807fmfw8ae7726q9pf9t3x34 -o json | jq
# Query paired insurances by provider address and status.
cantod query liquidstaking insurances --provider-address canto1xjlslz2vl7v6gu807fmfw8ae7726q9pf9t3x34 --status INSURANCE_STATUS_PAIRED -o json | jq
```


## Insurance

Query an Insurance by id.

Usage

```bash
insurance [insurance-id] 
```

Example

```bash
cantod query liquidstaking insurance 1 -o json | jq
```

## WithdrawInsuranceRequests

Query WithdrawInsuranceRequests.

Usage

```bash
withdraw-insurance-requests
```

Example

```bash
# Query all withdraw-insurance-requests
cantod query liquidstaking withdraw-insurance-requests -o json | jq
# Query all withdraw-insurance-requests by provider address
cantod query liquidstaking withdraw-insurance-requests --provider-address canto1xjlslz2vl7v6gu807fmfw8ae7726q9pf9t3x34 -o json | jq
```

## WithdrawInsuranceRequest

Query a WithdrawInsuranceRequest by insurance id.

Usage

```bash
withdraw-insurance-request [insurance-id] 
```

Example

```bash
# Query a withdraw-insurance-request by insurance id
cantod query liquidstaking withdraw-insurance-request 3 -o json | jq
```

## UnpairingForUnstakingChunkInfos

Query UnpairingForUnstakingChunkInfos.

Usage

```bash
unpairing-for-unstaking-chunk-infos 
```

Example

```bash
# Query all unpairing-for-unstaking-chunk-infos
cantod query liquidstaking unpairing-for-unstaking-chunk-infos -o json | jq
# Query all queued unpairing-for-unstaking-chunk-infos
cantod query liquidstaking unpairing-for-unstaking-chunk-infos --queued=true -o json | jq
```

## UnpairingForUnstakingChunkInfo

Query an UnpairingForUnstakingChunkInfo by chunk id.

Usage

```bash
unpairing-for-unstaking-chunk-info [chunk-id]
```

Example

```bash
# Query all unpairing-for-unstaking-chunk-info by chunk id
cantod query liquidstaking unpairing-for-unstaking-chunk-info 1 -o json | jq
```

## RedelegationInfos

Query RedelegationInfos.

Usage

```bash
redelegation-infos
```

Example

```bash
# Query redelegation-infos
cantod query liquidstaking redelegation-infos -o json | jq
```

## RedelegationInfo

Query RedelegationInfo by chunk id.

Usage

```bash
redelegation-info [chunk-id]
```

Example

```bash
# Query redelegation-info by chunk id
cantod query liquidstaking redelegation-info 1 -o json | jq
```

## ChunkSize

Query ChunkSize.

Usage

```bash
chunk-size
```

Example

```bash
# Query chunk size
cantod query liquidstaking chunk-size -o json | jq
```

## MinimumCollateral

Query MinimumCollateral.

Usage

```bash
minimum-collateral
```

Example

```bash
# Query minimum collateral  
cantod query liquidstaking minimum-collateral -o json | jq
```

## States

Query net amount state.

Usage

```bash
states
```

Example

```bash
cantod query liquidstaking states -o json | jq
```