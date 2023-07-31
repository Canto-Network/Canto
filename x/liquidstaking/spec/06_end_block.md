<!-- order: 6 -->

# EndBlock

The end block logic is executed at the end of each epoch.

## Distribute Reward

- For all paired chunks
  - withdraw delegation reward
    - chunk balance increased
  - distribute reward
    - send insurance commission from chunk
      - insurance commission: `(balance of chunk) x insurance.FeeRate`
    - burn fee calculated by `fee rate x (balance of chunk - insurance commission)` (Please check the `CalcDynamicFeeRate` in `dynamic_fee_rate.go` for detail.)
    - send rest of chunk balance to reward pool

## Cover slashing and handle mature unbondings

### For all unpairing for unstake chunks

- calc penalty
  - penalty: `(chunk size tokens) - (balance of chunk)`
- if penalty > 0
  - if unpairing insurance can cover
    - unpairing insurance send penalty to chunk
  - if unpairing insurance cannot cover
    - unpairing insurance send penalty to reward pool
    - refund lstokens corresponding penalty from ls token escrow acc
      - refund amount: `(penalty / (chunk size tokens)) x (ls tokens to burn)`
- complete unpairing insurance's duty because it already covered penalty
- burn all escrowed LS tokens, except for those that have been refunded (if any)
- send all of chunk's balances to un-delegator
- delete tracking obj(=`UnpairingForUnstakingChunkInfo`) and chunk

### For all unpairing chunks

- calc penalty
  - penalty: `(chunk size tokens) - (balance of chunk)`
- if penalty > 0 
  - if unpairing insurance can cover
    - unpairing insurance send penalty to chunk
  - if unpairing insurance cannot cover
    - unpairing insurance send penalty to reward pool
- complete insurance duty because unpairing insurance already covered penalty
- if chunk got damaged (unpairing insurance could not cover fully)
  - send all chunk balances to reward pool because chunk is not valid anymore.
  - delete chunk
  - if unpairing insurance's fee pool is empty, then delete unpairing insurance
- else(= chunk is fine)
  - chunk becomes `Pairing`

### For all paired chunks

- calc penalty
  - penalty: `(chunk size tokens) - (token values of chunk del shares)`
- if penalty > 0
  - if chunk is re-paired at previous epoch
    - if there was double sign slashing because of evidence created before previous epoch
      - unpairing insurance send penalty to chunk
      - chunk delegate additional tokens
      - deduct covered amt from penalty
  - if penalty > balance of paired insurance (cannot fully cover it)
    - un-pair and un-delegate chunk (`Paired → Unpairing`)
    - paired insurance becomes `Unpairing`
  - if penalty ≤ balance of paired insurance (can cover it)
    - send penalty to chunk
    - chunk delegate additional shares corresponding penalty
- if paired insurance balance < 5.75% after cover penalty and if undelegate not started
  - un-pair and undelegate chunk (`Paired → Unpairing`)
  - paired insurance becomes `Unpairing`
- if validator is not valid
  - un-pair chunk and insurance (both chunk and insurance `Paired → Unpairing`)
- if there was an unpairing insurance came from previous epoch and it is already finished its duty
  - empty unpairing insurance from chunk
  - if the insurance is still valid (balance and validator are all fine), then it becomes `Pairing`
  - if not, then it becomes `Unpaired`

## Remove Deletable Redelegation Infos

- For all re-delegation infos
  - if it is matured, then delete it.

## Handle Queued Liquid Unstakes

- For all UnpairingForUnstakingChunkInfos (= info)
  - got chunk from info.chunkId
  - if the chunk is not `Paired`, then do nothing and return. 
  - un-pair and un-delegate chunk 
    - paired insurance becomes `Unpairing`
    - chunk becomes `UnpairingForUnstaking`

## Handle Unprocessed Queued Liquid Unstakes

- For all UnpairingForUnstakingChunkInfos (= info)
  - got chunk from info.chunkId
  - if the chunk is not `UnpairingForUnstaking`, then delete info and refund info.EscrowedLsTokens to info.DelegatorAddress

## Handle Queued Withdraw Insurance Requests

- For all WithdrawInsuranceRequests (= req)
  - got insurance from req.InsuranceId
  - insurance must be `Paired` or `Unpairing`
  - got chunk from insurance.ChunkId
  - if the chunk is `Paired`, unpair it 
    - chunk becomes `Unpairing`
    - empty paired insurance id from chunk
    - chunk.UnpairingInsuranceId = insurance.Id
  - insurance becomes `UnpairingForWithdrawal`
  - delete request

## Rank Insurances

- get all **re-pairable chunks**, **out insurances**, and **pairedInsuranceMap**
  - condition of re-pairable chunk (re-pairable means can be paired with new insurance)
    - must be one of `Pairing`, `Paired`, or `Unpairing (without unbonding obj)`
  - out insurances are
    - paired with `Unpairing` chunk which have no unbonding obj
      - The most common case for this is withdrawing an insurance.
    - paired with `Paired` chunk but have invalid validator. 
- create candidate insurances
  - candidate insurance must be in `Pairing` or `Paired` statuses
  - candidate insurance must have valid validator 
- sort candidate insurances in ascending order, with the cheapest insurance listed first
- create rank in insurances and rank out insurances
  - if re-pairable chunks are more than candidate insurances, then all candidates can be rank in.
    - rank in insurances: `candidates`
    - rank out insurances: `out insurances`
  - rank in insurances: `candidates[:len(rePairableChunks)]`
  - rank out insurances: paired insurances in `candidates[len(rePairableChunks):]`
- append out insurances to rank out insurances
- create **newly ranked in insurances**
  - for insurances in **rank in insurances** which not exists in **pairedInsuranceMap**
- return **newly ranked in insurances** and **rank out insurances**

## RePair Ranked Insurances

- create rank out insurance chunk map
  - key: insurance id which in **ranked out insurances**
  - value: `Chunk`
- for insurance in **newly ranked in insurances**
  - if there is a rank out insurance which have same validator
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
