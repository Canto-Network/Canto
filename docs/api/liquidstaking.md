---
Title: Liquidstaking
Description: A high-level overview of what gRPC-gateway REST routes are supported in the liquidstaking module.
---

# Liquidstaking Module

## Synopsis

This document provides a high-level overview of what gRPC-gateway REST routes are supported in the liquidstaking module.
To set up a local testing environment, you should run [init_testnet.sh](https://github.com/b-harvest/Canto/blob/liquidstaking-module/init_testnet.sh)

## gRPC-gateway REST Routes


++https://github.com/Canto-Network/Canto/blob/main/proto/canto/liquidstaking/v1/query.proto
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

## Params

Query the current liquidstaking parameters information.

Example Request

```bash
http://localhost:1317/canto/liquidstaking/v1/params
```

Example Response

```json
{
  "params": {
    "dynamic_fee_rate": {
      "r0": "0.000000000000000000",
      "u_soft_cap": "0.050000000000000000",
      "u_hard_cap": "0.100000000000000000",
      "u_optimal": "0.090000000000000000",
      "slope1": "0.100000000000000000",
      "slope2": "0.400000000000000000",
      "max_fee_rate": "0.500000000000000000"
    }
  }
}
```

## Epoch

Query the epoch information.

Example Request

```bash
http://localhost:1317/canto/liquidstaking/v1/epoch
```

Example Response

```json
{
  "epoch": {
    "current_number": "648",
    "start_time": "2060-10-01T01:34:14.723955Z",
    "duration": "1814400s",
    "start_height": "3235"
  }
}
```

## Chunks

Query chunks.

Usage

```bash
http://localhost:1317/canto/liquidstaking/v1/chunks
```

Example Response

```json
{
  "chunks": [
    {
      "chunk": {
        "id": "1",
        "paired_insurance_id": "4",
        "unpairing_insurance_id": "0",
        "status": "CHUNK_STATUS_PAIRED"
      },
      "derived_address": "canto14zq9dj3mde6kwl7302zxcf2nv83m3k3qj9cq3k"
    },
    {
      "chunk": {
        "id": "2",
        "paired_insurance_id": "7",
        "unpairing_insurance_id": "0",
        "status": "CHUNK_STATUS_PAIRED"
      },
      "derived_address": "canto15r7jycu6dsljrrngnuez8ytpk8sey3awyleeht"
    }
  ],
  "pagination": {
    "next_key": null,
    "total": "2"
  }
}
```

## Chunk

Query a chunk by id.

Example Request

```bash
http://localhost:1317/canto/liquidstaking/v1/chunks/1
```

Example Response

```json
{
  "chunk": {
    "id": "1",
    "paired_insurance_id": "4",
    "unpairing_insurance_id": "0",
    "status": "CHUNK_STATUS_PAIRED"
  },
  "derived_address": "canto14zq9dj3mde6kwl7302zxcf2nv83m3k3qj9cq3k"
}
```

## Insurances

Query insurances.

Example Request

```bash
http://localhost:1317/canto/liquidstaking/v1/insurances
```

Example Response

```json
{
  "insurances": [
    {
      "insurance": {
        "id": "1",
        "validator_address": "cantovaloper1xjlslz2vl7v6gu807fmfw8ae7726q9pf84kzqs",
        "provider_address": "canto1xjlslz2vl7v6gu807fmfw8ae7726q9pf9t3x34",
        "fee_rate": "0.100000000000000000",
        "chunk_id": "0",
        "status": "INSURANCE_STATUS_UNPAIRED"
      },
      "derived_address": "canto1p6qg4xu665ld3l8nr72z0vpsujf0s9ekhfjhuv",
      "fee_pool_address": "canto1fy0mcah0tcedpyqyz423mefdxh7zqz4g2lu8jf"
    },
    {
      "insurance": {
        "id": "2",
        "validator_address": "cantovaloper1xjlslz2vl7v6gu807fmfw8ae7726q9pf84kzqs",
        "provider_address": "canto1xjlslz2vl7v6gu807fmfw8ae7726q9pf9t3x34",
        "fee_rate": "0.100000000000000000",
        "chunk_id": "0",
        "status": "INSURANCE_STATUS_UNPAIRED"
      },
      "derived_address": "canto1hk5wgk3js5uqymxppawk87tv0j0fnc3pefcex4",
      "fee_pool_address": "canto1a3f65vrngauvsj066067qsjh068hgxezpdr6rg"
    }
  ],
  "pagination": {
    "next_key": null,
    "total": "2"
  }
}
```


## Insurance

Query an Insurance by id.

Example Request

```bash
http://localhost:1317/canto/liquidstaking/v1/insurances/4
```

Example Response

```json
{
  "insurance": {
    "id": "4",
    "validator_address": "cantovaloper1xjlslz2vl7v6gu807fmfw8ae7726q9pf84kzqs",
    "provider_address": "canto1xjlslz2vl7v6gu807fmfw8ae7726q9pf9t3x34",
    "fee_rate": "0.100000000000000000",
    "chunk_id": "1",
    "status": "INSURANCE_STATUS_PAIRED"
  },
  "derived_address": "canto1my633g6sqx9fr4szzxuj70zutmsd78zymhv5kf",
  "fee_pool_address": "canto1sdl4z9y8x59979qjx8ut9zyndsux9sld0s6kcv"
}
```

## WithdrawInsuranceRequests

Query WithdrawInsuranceRequests.

Example Request

```bash
http://localhost:1317/canto/liquidstaking/v1/withdraw_insurance_requests
```

Example Response

```json
{
  "withdraw_insurance_requests": [
    {
      "insurance_id": "7"
    }
  ],
  "pagination": {
    "next_key": null,
    "total": "1"
  }
}
```

## WithdrawInsuranceRequest

Query a WithdrawInsuranceRequest by insurance id.

Example Request

```bash
http://localhost:1317/canto/liquidstaking/v1/insurances/5/withdraw_insurance_requests
```

Example Response

```json
{
  "withdraw_insurance_request": {
    "insurance_id": "5"
  }
}
```

## UnpairingForUnstakingChunkInfos

Query UnpairingForUnstakingChunkInfos.

Example Request

```bash
http://localhost:1317/canto/liquidstaking/v1/unpairing_for_unstaking_chunk_infos 
```

Example Response

```json
{
  "unpairing_for_unstaking_chunk_info": {
    "chunk_id": "2",
    "delegator_address": "canto1xjlslz2vl7v6gu807fmfw8ae7726q9pf9t3x34",
    "escrowed_lstokens": {
      "denom": "lscanto",
      "amount": "240214408039107442750000"
    }
  }
}
```

## UnpairingForUnstakingChunkInfo

Query an UnpairingForUnstakingChunkInfo by chunk id.

Example Request

```bash
http://localhost:1317/canto/liquidstaking/v1/chunks/2/unpairing_for_unstaking_chunk_infos
```

Example Response

```json
{
  "unpairing_for_unstaking_chunk_info": {
    "chunk_id": "2",
    "delegator_address": "canto1xjlslz2vl7v6gu807fmfw8ae7726q9pf9t3x34",
    "escrowed_lstokens": {
      "denom": "lscanto",
      "amount": "240214408039107442750000"
    }
  }
}
```

## RedelegationInfos

Query RedelegationInfos.

Example Request

```bash
http://localhost:1317/canto/liquidstaking/v1/redelegation_infos
```

Example Response

```json
{
  "redelegation_infos": [
    {
      "chunk_id": "1",
      "completion_time": "2030-09-08T06:25:48.694135Z"
    }
  ],
  "pagination": {
    "next_key": null,
    "total": "1"
  }
}
```

## RedelegationInfo

Query RedelegationInfo by chunk id.

Example Request

```bash
http://localhost:1317/canto/liquidstaking/v1/chunks/redelegation_infos
```

Example Response

```json
{
  "redelegation_info": {
    "chunk_id": "1",
    "completion_time": "2030-09-08T06:25:48.694135Z"
  }
}
```

## ChunkSize

Query ChunkSize.

Example Request

```bash
curl http://localhost:1317/canto/liquidstaking/v1/chunk_size
```

Example Response

```json
{
  "chunk_size": {
    "denom": "acanto",
    "amount": "250000000000000000000000"
  }
}
```

## MinimumCollateral

Query MinimumCollateral.

Example Request

```bash
curl http://localhost:1317/canto/liquidstaking/v1/minimum_collateral
```

Example Response

```jso

```json
{
  "minimum_collateral": {
    "denom": "acanto",
    "amount": "17500000000000000000000.000000000000000000"
  }
}
```

## States

Query net amount state.

Example Request

```bash
curl http://localhost:1317/canto/liquidstaking/v1/states
```

Example Response

```json
{
  "net_amount_state": {
    "mint_rate": "0.000000000000000000",
    "ls_tokens_total_supply": "0",
    "net_amount": "0.000000000000000000",
    "total_liquid_tokens": "0",
    "reward_module_acc_balance": "0",
    "fee_rate": "0.000000000000000000",
    "utilization_ratio": "0.000000000000000000",
    "remaining_chunk_slots": "1220",
    "discount_rate": "0.000000000000000000",
    "num_paired_chunks": "0",
    "chunk_size": "250000000000000000000000",
    "total_del_shares": "0.000000000000000000",
    "total_remaining_rewards": "0.000000000000000000",
    "total_chunks_balance": "0",
    "total_unbonding_chunks_balance": "0",
    "total_insurance_tokens": "0",
    "total_paired_insurance_tokens": "0",
    "total_unpairing_insurance_tokens": "0",
    "total_remaining_insurance_commissions": "0.000000000000000000"
  }
}
```
