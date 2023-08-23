<!-- order: 2 -->

# State

## Chunk

```go
type Chunk struct {
    Id                   uint64 // Unique id increased by 1
    PairedInsuranceId    uint64
    UnpairingInsuranceId uint64
    Status               ChunkStatus
}
```
A chunk object is created when token holder sends valid `MsgLiquidStake` and empty slot and a `Pairing` insurance are available.

A **chunk** has the following status:

1. `Paired`: This is the default status of a chunk and this status indicates that a chunk is paired with an insurance that has the lowest fee rate and staked. 
The fee rate is determined by the sum of the insurance fee rate set by the insurance provider and 
the commission fee rate set by the validator designated by the insurance provider.
2. `Unpairing`: A paired chunk enters an `Unpairing` status when paired insurance begins to be withdrawn, its balance becomes less than 5.75% of a chunk's size, or the validator becomes invalid (e.g., tombstoned). The 5.75% represents the minimum amount of tokens required to cover both downtime slashing and double signing slashing penalties once.
    * The calculation of 5.75% involves the sum of the `SlashFractionDoubleSign` and `SlashFractionDowntime` parameters. Modifying these parameters while the `liquidstaking` module is operational can introduce unforeseen risks. To mitigate this, changes to the slashing parameters are restricted via antehandlers. 
   For more information, please refer to the details provided in the **[Param Change Ante Handlers](10_ante_handlers.md#param-change-ante-handlers)**.
3. `UnpairingForUnstaking`: When a delegator (also referred to as a liquid staker) submits a MsgLiquidUnstake, the request is enqueued as UnpairingForUnstakingChunkInfo. 
At the conclusion of the epoch, the actual undelegation process is initiated, causing the chunk to transition into this state. Following the completion of the unbonding period in the subsequent epoch, tokens equivalent to the chunk's size are restored to the delegator's account, and the related chunk object is subsequently deleted.
Once the unbonding period is over in next epoch, the tokens corresponding chunk size are returned to the delegator's account and the associated chunk object is removed.
4. `Pairing`: This status indicates that the chunk is ready to be paired again with a new insurance after `unparing` process is completed.



## Insurance

An insurance object is created when Insurance Provider sends valid `MsgInsuranceProvide`. The message is valid only when the collateral assets of the insurance are equal to or greater than 7% of the minimum chunk size.

```go
type Insurance struct {
    Id               uint64 // Unique id increased by 1
    ValidatorAddress string
    ProviderAddress  string // An address of Insurance Provider
    FeeRate          staking_types.Dec
    ChunkId          uint64 // Id of the chunk for which the insurance has a duty
    Status           InsuranceStatus
}
```

An **insurance** has the following status:

1. `Pairing`: This is the initial status of an insurance when an insurance provider dispatches a `MsgInsuranceProvide`. 
This state signifies the insurance's readiness for pairing with a chunk. When an unoccupied slot becomes accessible, 
and either a `msgLiquidStake` is received or `Pairing` chunks have been established in the preceding epoch, 
the insurance having the lowest fee (including validator commission) will be matched with the chunk. Pairing insurances can be canceled through the utilization of `MsgCancelInsuranceProvide` before it is paired with a chunk.
2. `Paired`: An insurance is paired with a chunk. While the insurance holds this status, 
it functions as a safeguard for the chunk, offering coverage against undesirable losses that could arise from validator slashing. 
This guarantees the chunk's size remains unchanged and optimizes its staking rewards.
3. `Unpairing`: A paired insurance enters this status when its balance is no longer sufficient (less than 5.75% of chunk size tokens) to offset slashing penalties, 
when the validator becomes tombstoned, or when the associated chunk is initiated for undelegation through `MsgLiquidUnstake`. 
In the following epoch, the insurance will either remain paired or undergo unpairing, contingent upon its ongoing validity.
4. `UnpairingForWithdrawal`: A paired insurance transitions to this state when a queued `WithdrawInsuranceRequest` exists in the epoch.
5. `Unpaired`: `Unpairing` insurances from previous epoch can enter this status. `Unpaired` insurance can be withdrawn immediately by `MsgWithdrawInsurance` or The pairing status is automatically triggered when the following conditions are met:
    - The insurance is pointing to a valid validator (bonded validators).
    - Insurance balance is equal to or greater than 7% of the minimum chunk size.
    - No `WithdrawInsuranceRequest` is queued in the current epoch.

## UnpairingForUnstakingChunkInfo

This object is created when msgServer receives `MsgLiquidUnstake` for a paired chunk. 
The actual unbonding process is started on an upcoming epoch (**[Handle Queued Liquid Unstakes](06_end_block.md#handle-queued-liquid-unstakes)**).

The unstaking request does not take place immediately; it is initiated within the upcoming epoch and the actual unstaking occurs after the unbonding period has elapsed. During the unbonding period, changes in the chunk size may occur (if the insurance is unable to cover all penalties, the chunk size may decrease). In such cases, a portion of the escrowed lsTokens must be refunded. Therefore, the associated object serves to track the quantity of escrowed lsTokens when an unstaking request is made.

```go
type UnpairingForUnstakingChunkInfo struct {
    ChunkId          uint64 // Which chunk is tracked by this obj
    DelegatorAddress string // Who requests MsgLiquidUnstake
    // How much lstokens will be burned when unbonding finished
    EscrowedLsTokens sdk.Coin
}
```
It is removed when the chunk unbonding is finished (**[Cover slashing and handle mature unbondings](06_end_block.md#cover-slashing-and-handle-mature-unbondings)**).


## WithdrawInsuranceRequest

It is created when msgServer got `MsgWithdrawInsurance`

```go
type WithdrawInsuranceRequest struct {
	InsuranceId uint64 // Which insurance is requested for withdrawal
}
```

## RedelegationInfo

It is created when re-delegation for chunk happens between insurances pointing to different validators at epoch.
This situation happens when there's a more appealing validator and insurance pair on an epoch. The chunk keeps its paired status while being redelegated to a new validator.
When the chunk is undergoing redelegation, a separate logic (**[Cover redelegation penalty](05_begin_block.md#Cover-Redelegation-Penalty)**) is followed to ensure that the insurance covers any penalties. Therefore, the object is used to track whether the chunk is being redelegated or not.





```go
type RedelegationInfo struct {
    ChunkId        uint64    // Which chunk is in re-delegation
    CompletionTime time.Time // When re-delegation will be finished
}
```

This will be consumed at **Handle Queued Withdraw Insurance Requests** when an epoch is reached.

## NetAmountStateEssentials (in-memory only)

NetAmountStateEssentials includes crucial elements required for executing the fundamental operations of the `liquidstaking` module, such as `MsgLiquidStake` and `MsgLiquidUnstake`.

This state resides solely in memory and is not stored in the database. Whenever the module requires the value, it is computed using the most recent state.

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

**NetAmountBeforeModuleFee** is nearly identical to **NetAmount**, with the distinction that it doesn't subtract the module fee rate from delegation rewards. This value is employed when calculating the utilization ratio.

**MintRate** is a rate derived from the total supply of ls tokens divided by NetAmount:
- LsTokenTotalSupply / NetAmount

Based on the equation, the conversion between native tokens and lsTokens can be calculated as follows:
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

The following code displays the fields not encompassed by NetAmountStateEssentials, but present within NetAmountState, and relates to Insurance. These additional fields are not employed by the core logic but are included for querying purposes.

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
