<!-- order: 4 -->

# Messages

## Chunk

### MsgLiquidStake

Liquid stake with an amount of native tokens. A liquid staker is expected to receive ls tokens at the current mint rate.

```go
type MsgLiquidStake struct {
	DelegatorAddress    string
	Amount              types.Coin
}
```

**msg is failed if:**

- `msg.Amount` is not bond denom
- `msg.Amount` is not multiple of ChunkSize tokens
- If there are no empty slot
- The balance of msg sender(=Delegator) does not have enough amount of coins for `msg.Amount`

### MsgLiquidUnstake

Liquid unstake with an amount of native tokens which is expected to sent to unstaker when unstaking is done. 
The liquid unstake request will be queued until the upcoming Epoch and will initiate the unstaking process.

```go
type MsgLiquidUnstake struct {
	DelegatorAddress string
	Amount sdk.Coin // (How many chunks to be unstaked?) x chunk.size
}
```

**msg is failed if:**

- `msg.Amount` is not bond denom
- `msg.Amount` is not multiple of ChunkSize tokens
- The balance of msg sender(=Delegator) does not have enough amount of ls tokens for corresponding value of `msg.Amount`

## Insurance

### MsgProvideInsurance

Provide insurance to cover slashing penalties for chunks and to receive commission. 
* 9% of chunk size tokens is recommended for the `msg.Amount`.
* 7% is minimum collateral for the chunk size tokens. If the collateral is less than 7%, the insurance will be unpaired and the provider will not receive commission.
* The fee rate + Validator(msg.ValidatorAddress)'s fee rate must be less than 50%.

```go
type MsgProvideInsurance struct {
	ProviderAddress string
	ValidatorAddress string
	Amount types.Coin
	FeeRate staking_types.Dec
}
```

**msg is failed if:**

- `msg.Amount` is not bond denom
- `msg.Amount` must be bigger than minimum collateral (7% of chunk size tokens)
- `msg.ValidatorAddress` is not valid validator
- `msg.FeeRate` + Validator(msg.ValidatorAddress).Commission.Rate >= 0.5 (50%)

### MsgCancelProvideInsurance

Cancel insurance provision. Only pairing insurance can be canceled.

```go
type MsgCancelInsuranceProvide struct {
	ProviderAddress string
	Id uint64 
}
```

**msg is failed if:**

- There are no pairing insurance with given `msg.Id`
- Provider of Insurance with given id is different with `msg.ProviderAddress`

### MsgWithdrawInsurance

Create a pending insurance request for withdrawal or immediately withdraw all its commissions and collaterals when it is unpaired insurance. 
If it is not unpaired, then withdrawal will be triggered during the upcoming Epoch.

```go
type MsgWithdrawInsurance struct {
	ProviderAddress string
	Id uint64 
}
```

**msg is failed if:**

- There are no paired, unpairing or unpaired insurance with given `msg.Id`
- Provider of Insurance with given id is different with `msg.ProviderAddress`

### MsgWithdrawInsuranceCommission

Provider can withdraw accumulated commission from the insurance fee pool at any time. 
Providers can also withdraw their commission by using `MsgWithdrawInsurance` for unpaired insurance.

```go
type MsgWithdrawInsuranceCommission struct {
	ProviderAddress string
	Id uint64 
}
```

**msg is failed if:**

- Provider of Insurance with given id is different with `msg.ProviderAddress`

### MsgDepositInsurance

Provider can deposit native tokens into insurance at any time. 
Providers who are concerned that the insurance may not be sufficient, causing it to become unpaired and unable to earn commissions, can use this message.

```go
type MsgDepositInsurance struct {
	ProviderAddress string
	Id uint64 
	Amount sdk.Coin
}
```

**msg is failed if:**

- There are no insurance with given `msg.Id`
- Provider of Insurance with given id is different with `msg.ProviderAddress`
- `msg.Amount` is not bond denom

### MsgClaimDiscountedReward

Requester can withdraw accumulated reward from the reward pool at any time with discounted price.
How much to get rewards is calculated by `msg.Amount` and discounted mint rate. (maximum discount rate is 3%)

```go
type MsgClaimDiscountedReward struct {
	RequesterAddress string
	Amount sdk.Coin
	minimumDiscountRate sdk.Dec
}
```

**msg is failed if:**

- `msg.Amount` is not liquid bond denom
- current discount rate is lower than `msg.MinimumDiscountRate`