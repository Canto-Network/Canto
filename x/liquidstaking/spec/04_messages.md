<!--
order: 4
-->

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
- The balance of msg sender(=Delegator) does not have enough amount of coins for `msg.Amount`

### MsgLiquidUnstake

Liquid unstake with an amount of native tokens which is expected to sent to unstaker when unstaking is done. The liquid unstake request will be queued until the upcoming Epoch and will initiate the unstaking process.

```go
type MsgLiquidUnstake struct {
	DelegatorAddress string
	Amount sdk.Coin // (How many chunks to be unstaked?) x chunk.size
}
```

**msg is failed if:**

- `msg.Amount` is not bond denom
- `msg.Amount` is not multiple of ChunkSize tokens
- The balance of msg sender(=Delegator) does not have enough amount of ls tokens for corresponding value of `msg.Amount`

## Insurance

### MsgProvideInsurance

Provide insurance to cover slashing penalties for chunks and to receive commission.

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
- `msg.Amount` must be bigger than minimum coverage
- `msg.ValidatorAddress` is not valid validator

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
- The insurance is not pairing insurance.
- Provider of Insurance with given id is different with `msg.ProviderAddress`

### MsgWithdrawInsurance

Create a pending insurance request for withdrawal. The withdrawal will start during the upcoming Epoch.

```go
type MsgWithdrawInsurance struct {
	ProviderAddress string
	Id uint64 
}
```

**msg is failed if:**

- There are no paired or unpaired insurance with given `msg.Id`
- Provider of Insurance with given id is different with `msg.ProviderAddress`
- The insurance is not paired or unpaired.

### MsgWithdrawInsuranceCommission

Provider can withdraw accumulated commission from the insurance fee pool at any time. Providers can also withdraw their commission by using `MsgWithdrawInsurance` for unpaired insurance.

```go
type MsgWithdrawInsuranceCommission struct {
	ProviderAddress string
	Id uint64 
}
```

**msg is failed if:**

- Provider of Insurance with given id is different with `msg.ProviderAddress`

### DepositInsurance

Provider can deposit native tokens into insurance at any time. Providers who are concerned that the insurance may not be sufficient, causing it to become unpaired and unable to earn commissions, can use this message.

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