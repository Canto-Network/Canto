<!-- order: 4 -->

# Messages

## Chunk

### MsgLiquidStake

Depositing native tokens that are multiples of the chunk size for liquid staking. 
The liquid staker is anticipated to receive lsTokens at the prevailing mint rate.

```go
type MsgLiquidStake struct {
	DelegatorAddress    string
	Amount              types.Coin // (How many chunks to liquid stake?) x ChunkSize
}
```

Message **fails** if:

- `msg.Amount` is not a bond denom
- `msg.Amount` is not multiple of ChunkSize tokens
- no empty slot or pairing insurance available
- number of chunks to liquid stake is bigger than empty slot or pairing insurance
- balance of msg sender(=delegator) does not have enough amount of native coins for `msg.Amount`

### MsgLiquidUnstake

Submitting an amount of native tokens (multiples of the chunk size) that is projected to be transferred to the unstaker upon the completion of the unstaking process. 
The liquid unstake request will be held in a queue until the next epoch, at which point it will initiate the unstaking procedure.

```go
type MsgLiquidUnstake struct {
	DelegatorAddress string
	Amount sdk.Coin // (How many chunks to be unstaked?) x ChunkSize
}
```

Message fails if:

- `msg.Amount` is not a bond denom
- `msg.Amount` is not multiple of ChunkSize tokens
- no paired chunks available
- number of chunks to liquid unstake is bigger than the number of paired chunks
- balance of msg sender(=delegator) does not have enough amount of lsTokens corresponding value of `msg.Amount`

## Insurance

### MsgProvideInsurance

Provide insurance to cover slashing penalties for chunks and to receive commission. 
* **recommended** to use 9% of the chunk size tokens for the `msg.Amount`.
* **minimum** collateral is 7% of a chunk size.
* Sum of insurance fee rate and corresponding validator's commission rate must be less than 50%.

```go
type MsgProvideInsurance struct {
	ProviderAddress string
	ValidatorAddress string
	Amount types.Coin
	FeeRate staking_types.Dec
}
```

Message fails if:

- `msg.Amount` is not a bond denom
- `msg.Amount` is less than the minimum collateral (7% of chunk size)
- `msg.ValidatorAddress` is not valid validator (e.g., unbonded or tombstoned)
- `msg.FeeRate` + `Validator(msg.ValidatorAddress).Commission.Rate` >= 0.5 (50%)

### MsgCancelProvideInsurance

This message is a request to cancel an insurance provision. It's only possible to cancel pairing insurances.
```go
type MsgCancelInsuranceProvide struct {
	ProviderAddress string
	Id uint64 
}
```

Message fails if:

- no pairing insurance with given `msg.Id` exists
- insurance provider with the provided ID is not the same as `msg.ProviderAddress`.

### MsgWithdrawInsurance

This message is a request to withdraw the collaterals and commissions that have been accumulated. If the insurance status is `Unpaired`, the withdrawal will happen right away. For other statuses, the withdrawal will be initiated in the next epoch.

```go
type MsgWithdrawInsurance struct {
	ProviderAddress string
	Id uint64 
}
```

Message fails if:

- no `Paired` or `Unpaired` insurance with the given `msg.Id`
- insurance provider with the provided ID is not the same as `msg.ProviderAddress`.



### MsgWithdrawInsuranceCommission

This message is a request to withdraw the accumulated commission from the insurance fee pool. The message is processed as soon as the request is received.

```go
type MsgWithdrawInsuranceCommission struct {
	ProviderAddress string
	Id uint64 
}
```

Message fails if:

- no insurance with the given `msg.Id`
- insurance provider with the provided ID is not the same as `msg.ProviderAddress`.

### MsgDepositInsurance

Depositing more native tokens as collaterals into a existing insurance. The message is processed as soon as the request is received.
This message can be employed when the insurance's balance is not sufficient, leading to an unpaired status and blocking commission earnings. 
It serves to avert such circumstances, maintaining the insurance status as either `Paired` or `Pairing`.

```go
type MsgDepositInsurance struct {
    ProviderAddress string
    Id              uint64
    Amount          sdk.Coin
}
```

Message fails if:

- no insurance with the given `msg.Id`
- insurance provider with the provided ID is not the same as `msg.ProviderAddress`.
- `msg.Amount` is not bond denom

### MsgClaimDiscountedReward

This message requests the exchange of lsTokens for native tokens from the reward pool at a reduced rate.
The exchange rate is calcuated by current `MintRate` * `DiscountRate` where `discount rate = reward module account's balance / NetAmount`.
```go
type MsgClaimDiscountedReward struct {
    RequesterAddress    string
    Amount              sdk.Coin
    MinimumDiscountRate sdk.Dec
}
```

Message fails if:

- `msg.Amount` is not a liquid bond denom
- current discount rate is lower than `msg.MinimumDiscountRate` 
- `msg.RequesterAddress` doesn't have enough amount of lsTokens corresponding to the value of `msg.Amount`.




