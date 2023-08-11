<!-- order: 2 -->

# State

## Chunk

All state transitions of Chunk occur at EndBlock and when an epoch is reached, except when the `MsgLiquidStake`.

```go
type Chunk struct {
  Id uint64 // Unique id increased by 1
  PairedInsuranceId uint64 
  UnpairingInsuranceId uint64 
  Status ChunkStatus 
}
```

A **chunk** has the following status:

1. `Pairing`: This status indicates that the chunk is ready to be paired with an insurance.
2. `Paired`: A chunk is paired with an insurance that has the lowest fee rate. 
The fee rate is determined by the sum of the insurance fee rate set by the insurance provider and 
the commission fee rate set by the validator designated by the insurance provider.
3. `Unpairing`: A paired chunk enters this status when paired insurance is started to be withdrawn or 
it's balance <= 5.75%(double_sign_fraction + down_time_fraction) or the validator becomes invalid(e.g. tombstoned).
  * 5.75%(double_sign_fraction + down_time_fraction) is guaranteed when slashing param change is limited through antehandler. we already have this mechanism. 
please check **[Param Change Ante Handlers](10_param_change_ante_handlers.md#param-change-ante-handlers).** for detail.
4. `UnpairingForUnstaking`: When a delegator (also known as a liquid staker) sends a `MsgLiquidUnstake`, it is queued as a `UnpairingForUnstakingChunkInfo`. 
At the end of the epoch, the actual undelegation is triggered and the chunk enters this state. 
Once the unbonding period is over in next epoch, the tokens corresponding chunk size are returned to the delegator's account and the associated chunk object is removed.

## Insurance

An insurance object is created when Insurance Provider sends valid `MsgInsuranceProvide`.

Most state transition of Insurance occurs at EndBlock and an epoch is reached.

```go
type Insurance struct {
  Id uint64 // Unique id increased by 1
  ValidatorAddress string 
  ProviderAddress string // An address of Insurance Provider
  FeeRate staking_types.Dec 
  ChunkId uint64 // Id of the chunk for which the insurance has a duty
  Status InsuranceStatus 
}
```

An **insurance** has the following status:

1. `Pairing`: This is the default status of an insurance when an insurance provider sends a `MsgInsuranceProvide`. 
This status indicates that the insurance is ready to be paired with a chunk. When an empty slot is available and either `msgLiquidStake` is received or 
`pairing` chunks have been created in the recent epoch, the insurance with the lowest fee will be paired with the chunk. 
Only pairing insurances can be canceled using `MsgCancelInsuranceProvide`.
2. `Paired`: An insurance is paired with a chunk. While the insurance is in this status, it serves as a form of protection for the chunk 
by insuring it against unexpected loss that may occur due to validator slashing. 
This ensures that the chunk remains same size and maximize its staking rewards.
3. `Unpairing`: A paired insurance enters this status when it has no longer has enough balance (=5.75% of chunk size tokens) to cover slashing penalties, 
when the validator is tombstoned, or when the paired chunk is started to be undelegated by `MsgLiquidUnstake`. 
At the next epoch, unpairing will be unpaired or pairing if it is still valid.
4. `UnpairingForWithdrawal`: A paired insurance enters this status when there was a queued `WithdrawInsuranceRequest` at the epoch.
5. `Unpaired`: `Unpairing` insurances from previous epoch can enter this status. `Unpaired` insurance can be withdrawn immediately by `MsgWithdrawInsurance`.

## UnpairingForUnstakingChunkInfo

It is created when msgServer receives `MsgLiquidUnstake` for paired chunk. 
The actual unbonding is started at **[Handle Queued Liquid Unstakes](06_end_block.md#handle-queued-liquid-unstakes).**

It is removed **[Cover slashing and handle mature unbondings](06_end_block.md#cover-slashing-and-handle-mature-unbondings)** when chunk unbonding is finished.

```go
type UnpairingForUnstakingChunkInfo struct {
  ChunkId uint64 // Which chunk is tracked by this obj
  DelegatorAddress string // Who requests MsgLiquidUnstake
  // How much lstokens will be burned when unbonding finished
  EscrowedLsTokens sdk.Coin 
}
```

## WithdrawInsuranceRequest

It is created when msgServer got `MsgWithdrawInsurance`

```go
type WithdrawInsuranceRequest struct {
  InsuranceId uint64 // Which insurance is requested for withdrawal
}
```

## RedelegationInfo

It is created when re-delegation for chunk happen between insurances pointing to different validators at epoch.

```go
type RedelegationInfo struct {
  ChunkId uint64 // Which chunk is in re-delegation
  CompletionTime time.Time // When re-delegation will be finished
}
```

This will be consumed at **Handle Queued Withdraw Insurance Requests** when Epoch is reached.

## NetAmountStateEssentials (in-memory only)

NetAmountStateEssentials contains items that are essential for executing the core logic of the liquid staking module (e.g. `MsgLiquidStake` and `MsgLiquidUnstake`).

This state is in-memory only and is not persisted to the database. Every time module need it, it is calculated from the latest state.

**NetAmount** is the sum of the following items

- **reward module account’s native token(e.g. `acanto`) balance**
- **sum of all chunk balance**
  - the chunk balance will only be as much as the balance accumulated from delegation rewards between epochs. 
    at the end of each epoch, the cumulated chunk balance will be transferred to the reward module account.
  - when insurance is withdrawn and there are no candidate insurances, the chunk balance can be the same as the chunk size in tokens.
- **sum of all tokens corresponding delegation shares of paired chunks**
  - total amount of native tokens currently delegated
  - insurance coverage also included which means even if there were a slashing so token value of delegation shares is less than chunk size value,
    the value will be the same as the chunk size value if insurance can cover the slashing penalty.
- **sum of all remaining rewards of all chunks delegations**
  - the remaining reward for each chunk is calculated as follows:
    ```
    rest = del_reward - insurance_commission
    remaining = rest x (1 - dynamic_fee_rate)
    ``` 
- **sum of all unbonding balances of chunks**
  - total amount of native tokens currently in un-delegating
  - insurance coverage also included which means even if there were a slashing so unbonding balance is less than chunk size value, 
    the balance will be the same as the chunk size value if insurance can cover the slashing penalty.

**NetAmountBeforeModuleFee** is almost the same as **NetAmount** except that it does not deduct module fee rate from delegation rewards.
This value is used when calculating the utilization ratio. 

**MintRate** is the rate that is calculated from total supply of ls tokens divided by NetAmount.

- LsTokenTotalSupply / NetAmount

Depending on the equation, the value transformation between native tokens and ls tokens can be calculated as follows:

- NativeTokenToLsToken: `nativeTokenAmount * lsTokenTotalSupply / NetAmount` with truncations
- LsTokenToNativeToken: `lsTokenAmount * NetAmount / LsTokenTotalSupply` with truncations

```go
// NetAmountStateEssentials is a subset of NetAmountState which is used for
// core logics. Insurance related fields are excluded, because they are not used
// in core logics(e.g. calculating mint rate).
type NetAmountStateEssentials struct {
  // Calculated by (total supply of ls tokens) / NetAmount
  MintRate sdk.Dec 
  // Total supply of ls tokens
  // e.g. 100 ls tokens minted -> 10 ls tokens burned, then total supply is 90
  // ls tokens
  LsTokensTotalSupply sdk.Int 
  // Calculated by reward module account's native token balance + all
  // all chunk's native token balance + sum of token values of all chunk's
  // delegation shares + sum of all remaining rewards of paired chunks since
  // last Epoch + all unbonding delegation tokens of unpairing chunks
  NetAmount sdk.Dec 
  // The token amount worth of all delegation shares of all paired chunks
  // (slashing applied amount)
  TotalLiquidTokens sdk.Int 
  // Balance of reward module account
  RewardModuleAccBalance sdk.Int 
  // Fee rate applied when deduct module fee at epoch
  FeeRate sdk.Dec 
  // Utilization ratio
  UtilizationRatio sdk.Dec 
  // How many chunks which can be created left?
  RemainingChunkSlots sdk.Int 
  // Discount rate applied when withdraw rewards
  DiscountRate sdk.Dec 
  // --- Chunk related fields
  // The number of paired chunks
  NumPairedChunks sdk.Int 
  // Current chunk size tokens
  ChunkSize sdk.Int 
  // Total delegation shares of all paired chunks
  TotalDelShares sdk.Dec 
  // The cumulative reward of all chunks delegations from the last distribution
  TotalRemainingRewards sdk.Dec 
  // Sum of the balances of all chunks.
  // Note: Paired chunks can be pairing status for various reasons (such as lack
  // of insurance). In such cases, the delegated native tokens returns to the
  // balance of DerivedAddress(Chunk.Id) after un-bonding period is finished.
  TotalChunksBalance sdk.Int 
  // The sum of unbonding balance of all chunks in Unpairing or
  // UnpairingForUnstaking
  TotalUnbondingChunksBalance sdk.Int 
}
```

## NetAmountState (in-memory only)

NetAmountState includes every field in NetAmountStateEssentials and additional fields related to Insurance 
which is not used by core logics but provided for a query.

The following code just shows the fields that are not included in NetAmountStateEssentials.

```go
// NetAmountState is type for net amount raw data and mint rate, This is a value
// that depends on the several module state every time, so it is used only for
// calculation and query and is not stored in kv.
type NetAmountState struct {
  // (... all fields in NetAmountStateEssential)	
  
  // --- Insurance related fields
  // The sum of all insurances' amount (= DerivedAddress(Insurance.Id).Balance)
  TotalInsuranceTokens sdk.Int 
  // The sum of all paired insurances' amount (= 
  //DerivedAddress(Insurance.Id).Balance)
  TotalPairedInsuranceTokens sdk.Int
  // The sum of all unpairing insurances' amount (=
  // DerivedAddress(Insurance.Id).Balance)
  TotalUnpairingInsuranceTokens sdk.Int 
  // The cumulative commissions of all insurances
  TotalRemainingInsuranceCommissions sdk.Dec 	
}
```

# Store

**The key retrieves liquid bond denom**

- LiquidBondDenomKey: `[]byte{0x01} -> ProtocolBuffer(string)`

**The key retrieves the latest chunk id**

- LastChunkIdKey: `[]byte{0x02} -> ProtocolBuffer(uint64)`

**The key retrieves the latest insurance id**

- LastInsuranceIdKey: `[]byte{0x03} -> ProtocolBuffer(uint64)`

**The key retrieves the chunk with given id**

- ChunkKey: `[]byte{0x04} | Chunk.Id -> ProtocolBuffer(Chunk)`

**The key retrieves the insurance with given id**

- InsuranceKey: `[]byte{0x05} | Insurance.Id -> ProtocolBuffer(Insurance)`

**The key retrieves the withdraw insurance request**

- WithdrawInsuranceRequestKey: `[]byte{0x06} | Insurance.Id -> ProtocolBuffer(WithdrawInsuranceReuqest)`

**The key retrieves the unpairing for unstaking chunk info**

- UnpairingForUnstakingChunkInfoKey: `[]byte{0x07} | Chunk.Id -> ProtocolBuffer(UnpairingForUnstakingChunkInfo)`

**The key retrieves the redelegation info**

- RedelegationInfoKey: `[]byte{0x08} | Chunk.Id -> ProtocolBuffer(RedelegationInfo)`

**The key retrieves the epoch**

- EpochKey: `[]byte{0x09} -> ProtocolBuffer(Epoch)`
