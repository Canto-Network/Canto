<!-- order: 6 -->

# EndBlock

The end block logic is executed at the end of each epoch.

## Distribute Reward

- for all paired chunks
  - withdraw delegation rewards
    - chunk balance increases
  - distribute rewards
    - send insurance commission from chunk
      - insurance commission: `(balance of chunk) x insurance.FeeRate`
    - burn fees calculated by `fee rate x (balance of chunk - insurance commission`) (For more details, please check the `CalcDynamicFeeRate` in `dynamic_fee_rate.go` )
    - send rest of the chunk balance to the reward pool

## Cover slashing and handle mature unbondings

### For all unpairing for unstake chunks

- calculate the penalty
  - penalty: `(chunk size tokens) - (balance of chunk)`
- if the penalty value is positive
  - if the unpairing insurance can cover the penalty
    - the unpairing insurance sends penalty to the chunk
  - if the unpairing insurance cannot cover the penalty
    - then unpairing insurance sends the penalty to the reward pool
    - refund lsTokens equivalent to the penalty amount from the lsToken escrow account
      - refund amount: `(penalty / (chunk size tokens)) x (ls tokens to burn)`
- complete the unpairing insurance's duty (the penalty is already covered)
- burn all remaining escrowed lsTokens
- send all of chunk's balances to the unstaker
- delete the tracking object (`UnpairingForUnstakingChunkInfo`) and the chunk

### For all unpairing chunks

- calculate the penalty
  - penalty: `(chunk size tokens) - (balance of chunk)`
- if the penalty value is positive
  - if the unpairing insurance can cover the penalty
    - the unpairing insurance sends the penalty to the chunk
  - if the unpairing insurance cannot cover the penalty
    - the unpairing insurance sends its remaining balance to the reward pool
- complete the insurance's duty (the penalty is already covered)
- if the chunk is damaged (the unpairing insurance could not fully cover the penalty)
  - send all chunk balances to the reward pool (damaaged chunk is not valid anymore)
  - delete the chunk
  - if the unpairing insurance's fee pool is empty, delete the unpairing insurance
- if the chunk is still valid
  - set the chunk's status to `Pairing`

### For all paired chunks

- calculate the penalty
  - penalty: `(chunk size tokens) - (token values of chunk del shares)`
- if the penalty value is positive
  - if the chunk is re-paired at the previous epoch
    - if a double sign slashing evidence, created before the previous epoch, is found
      - the unpairing insurance sends the penalty to the chunk
      - the chunk delegates additional tokens
      - deduct covered amount from penalty
  - if penalty value is bigger than the balance of paired insurance (cannot fully cover the penalty)
    - un-pair and un-delegate chunk (`Paired` → `Unpairing`)
    - set the paired insurance's status to `Unpairing`
  - if the penalty is less than or equal to the balance of the paired insurance (able to cover the penalty).
    - send the penalty from the insurance to the chunk
    - the chunk delegates additional shares corresponding to the covered penalty
- if the paired insurance balance is less than 5.75% of the chunk size after covering the penalty and if undelegate is not started
  - un-pair and undelegate the chunk (`Paired` → `Unpairing`)
  - set the paired insurance's status to `Unpairing`
- if the validator is not valid
  - un-pair the chunk and the insurance (`Paired` → `Unpairing` for both the chunk and the insurance)
- if the chunk has both `chunk.PairedInsuranceId` and `chunk.UnpairingInsuranceId` value (re-paring occured in the previous epoch)
  - set the `chunk.PairedInsuranceId` to 0
  - if the insurance is still valid, set the insurance status to `Pairing`
  - otherwise, set the insurance's status to `Unpaired`

## Remove Deletable Redelegation Infos

- For all `RedelegationInfo`s
  - if it is matured, delete the object.

## Handle Queued Liquid Unstakes

- For all `UnpairingForUnstakingChunkInfos`
  - retrieve a chunk from  `info.chunkId`
  - un-pair and un-delegate chunk if the chunk status is `Paired`
    - set the paired insurance's status to `Unpairing`
    - set the chunk's status to `UnpairingForUnstaking`

## Handle Unprocessed Queued Liquid Unstakes

- For all `UnpairingForUnstakingChunkInfos` (= `info`)
  - retrieve a chunk from `info.chunkId`
  - if the chunk is not `UnpairingForUnstaking`
    - delete `info` and refund `info.EscrowedLsTokens` to `info.DelegatorAddress`

## Handle Queued Withdraw Insurance Requests

- For all `WithdrawInsuranceRequests` (= `req`)
  - retrieve an insurance from `req.InsuranceId`
  - retrieve a chunk from `insurance.ChunkId`
  - if the chunk is `Paired`
    - set the chunk's status to `Unpairing`
    - set the paired insurance id from chunk to 0
    - set `chunk.UnpairingInsuranceId` to `insurance.Id`
  - set the insurance's status to `UnpairingForWithdrawal`
  - delete the `req` object

## Rank Insurances

- get all **re-pairable chunks**, **out insurances**, and **pairedInsuranceMap**
  - re-pairable chunks (re-pairable means can be paired with new insurance):
    - chunks that have one of following status
      - `Pairing` 
      - `Paired`
      - `Unpairing` (without unbonding obj)
  - out insurances:
    - insurances that paired with `Unpairing` chunk which have no unbonding object
      - The most common case for this is withdrawing an insurance.
    - insurances that paired with `Paired` chunk but have invalid validator. 
- create candidate insurances
  - candidate insurance must be in `Pairing` or `Paired` status
  - candidate insurance must have valid validator 
- sort candidate insurances in ascending order, placing the insurance with the lowest fee at the top
- select rank-in insurances and rank-out insurances
  - if re-pairable chunks are more than candidate insurances, all candidate insurances can be ranked in.
    - rank-in insurances: `candidates`
    - rank-out insurances: `out insurances`
  - rank-in insurances: `candidates[:len(rePairableChunks)]`
  - rank-out insurances: paired insurances in `candidates[len(rePairableChunks):]`
- append **out insurances** to rank-out insurance list
- create **newly ranked-in insurances**
  - for insurances in **rank-in insurances** which not exists in **pairedInsuranceMap**
- return **newly ranked-in insurances** and **rank-out insurances**

## RePair Ranked Insurances

- create rank-out insurance chunk map
  - key: insurance id which in **ranked out insurances**
  - value: `Chunk`
- for every insurances in **newly ranked in insurances**
  - if there is a rank-out insurance which has the same validator
    - replace insurance id of chunk with new one because it directs same validator, we don’t have to re-delegate it
      - Rank out insurance becomes `Unpairing` insurance of chunk (`Paired → Unpairing`)
        - if rank out insurance is withdrawing insurance, just keep it as it is 
      - rank in insurance becomes `Paired` insurance of chunk (`Pairing → Paired`)
      - state transition of chunk (`Paired | Unpairing → Paired`) 
      - mark the out insurance as handled
  - if there is no rank out insurance which have same validator
    - add it to **new insurances with different validators**
- make **remained out insurances** (= rank out insurances but not yet handled)
- for insurance in **new insurances with different validators**
  - get all **pairing chunks** (=immediately pariable) and pair them with insurance
- for insurance in **remained out insurances**
  - if there are no new insurances anymore, then break the loop
  - if validator of out insurance (=srcVal) is in Unbonding status, then continue
    - if we rere-delegate chunk's delegation from unbonding validator, 
    then we cannot guarantee the re-delegation ends at the epoch exactly. so we skip.
  - begin re-delegation and create tracking obj if srcVal is not Unbonded.
    - if srcVal is Unbonded, then re-delegation obj in staking module will not be created.
    so we don't need to track it because there will be no re-delegation slashing situation.
  - mark the insurance as handled
- make **rest out insurances** by removing handled insurance from **remained out insurances**
- for insurance in **rest out insurances**
  - un-delegate chunk
