<!-- order: 7 -->

# Events

The `liquidstaking` module emits the following events:

## Handlers

### MsgLiquidStake

| Type         | Attribute Key         | Attribute Value          |
|--------------|-----------------------|--------------------------|
| liquid_stake | chunk_ids             | {commaSeparatedChunkIds} |
| liquid_stake | delegator             | {msg.DelegatorAddress}   |
| liquid_stake | amount                | {msg.Amount}             |
| liquid_stake | new_shares            | {newShares}              |
| liquid_stake | lstoken_minted_amount | {lsTokenMintAmount}      |
| message      | module                | liquidstaking            |
| message      | action                | liquid_stake             |
| message      | sender                | {senderAddress}          |

### MsgLiquidUnstake

| Type           | Attribute Key     | Attribute Value          |
|----------------|-------------------|--------------------------|
| liquid_unstake | chunk_ids         | {commaSeparatedChunkIds} |
| liquid_unstake | delegator         | {msg.DelegatorAddress}   |
| liquid_unstake | amount            | {msg.Amount}             |
| liquid_unstake | escrowed_lstokens | {escrowedLsTokens}       |
| message        | module            | liquidstaking            |
| message        | action            | liquid_unstake           |
| message        | sender            | {senderAddress}          |


### MsgProvideInsurance

| Type              | Attribute Key      | Attribute Value       |
|-------------------|--------------------|-----------------------|
| provide_insurance | insurance_id       | {insurance.Id}        |
| provide_insurance | insurance_provider | {msg.ProviderAddress} |
| provide_insurance | amount             | {msg.Amount}          |
| message           | module             | liquidstaking         |
| message           | action             | provide_insurance     |
| message           | sender             | {senderAddress}       |

### MsgCancelProvideInsurance

| Type                     | Attribute Key      | Attribute Value          |
|--------------------------|--------------------|--------------------------|
| cancel_provide_insurance | insurance_id       | {insurance.Id}           |
| cancel_provide_insurance | insurance_provider | {msg.ProviderAddress}    |
| message                  | module             | liquidstaking            |
| message                  | action             | cancel_provide_insurance |
| message                  | sender             | {senderAddress}          |

### MsgDepositInsurance

| Type              | Attribute Key      | Attribute Value       |
|-------------------|--------------------|-----------------------|
| deposit_insurance | insurance_id       | {insurance.Id}        |
| deposit_insurance | insurance_provider | {msg.ProviderAddress} |
| deposit_insurance | amount             | {msg.Amount}          |
| message           | module             | liquidstaking         |
| message           | action             | deposit_insurance     |
| message           | sender             | {senderAddress}       |

### MsgWithdrawInsurance

| Type                          | Attribute Key                          | Attribute Value       |
|-------------------------------|----------------------------------------|-----------------------|
| withdraw_insurance_commission | insurance_id                           | {insurance.Id}        |
| withdraw_insurance_commission | insurance_provider                     | {msg.ProviderAddress} |
| withdraw_insurance_commission | withdraw_insurance_request_queued      | {queued}              |
| message                       | module                                 | liquidstaking         |
| message                       | action                                 | withdraw_insurance    |
| message                       | sender                                 | {senderAddress}       |

### MsgWithdrawInsuranceCommission

| Type               | Attribute Key                         | Attribute Value                 |
|--------------------|---------------------------------------|---------------------------------|
| withdraw_insurance | insurance_id                          | {insurance.Id}                  |
| withdraw_insurance | insurance_provider                    | {msg.ProviderAddress}           |
| withdraw_insurance | withdrawn_insurance_commission        | {allBalancesOfInsuranceFeePool} |
| message            | module                                | liquidstaking                   |
| message            | action                                | withdraw_insurance_commission   |
| message            | sender                                | {senderAddress}                 |


### MsgClaimDiscountedReward

| Type                    | Attribute Key        | Attribute Value         |
|-------------------------|----------------------|-------------------------|
| claim_discounted_reward | requester            | {msg.RequesterAddress}  |
| claim_discounted_reward | amount               | {msg.Amount}            |
| claim_discounted_reward | claim_tokens         | {claim}                 |
| claim_discounted_reward | discounted_mint_rate | {discountedMintRate}    |
| message                 | module               | liquidstaking           |
| message                 | action               | claim_discounted_reward |
| message                 | sender               | {senderAddress}         |
