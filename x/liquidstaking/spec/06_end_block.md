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
  - if re-delegation info exists and info.penalty > 0
    - penalty = penalty - info.penalty (calc only penalty which current unpairing chunk should cover)
    - info.Deletable = true 
- if penalty > 0 
  - if unpairing insurance can cover
    - unpairing insurance send penalty to chunk
  - if unpairing insurance cannot cover
    - unpairing insurance send penalty to reward pool
- complete insurance duty because unpairing insurance already covered penalty
- if chunk got damaged (unpairing insurance could not cover fully)
  - send all chunk balances to reward pool because chunk is not valid anymore.
- else(= chunk is fine)
  - state transition (`Unpairing → Pairing`)

### For all paired chunks

- calc penalty
- if penalty > 0
  - if re-delegation info exists for chunk and info.penalty > 0
    - unpairing insurance send all of its balance to reward pool
    - update penalty to penalty - info.penalty (we updated it because paired insurance doens't have to cover penalty from re-delegation.)
    - state transition of chunk (`Paired → Unpairing`)
    - un-delegate chunk
    - set undelegatedByRedelegationPenalty as true
  - if penalty > balance of insurance (meaning the insurance cannot fully cover it)
    - if undelegatedByRedelegationPenalty is false
      - un-delegate chunk
      - state transition of insurance (`Paired → Unpairing`)
      - state transition of chunk (`Paired → Unpairing`)
  - if penalty ≤ balance of insurance (meaning the insurance can cover it)
    - if undelegatedByRedelegationPenalty is false
      - send penalty to chunk
      - chunk delegate additional shares corresponding penalty
- if insurance is not sufficient after cover penalty
  - state transition of insurance (`Paired → Unpairing`)
  - state transition of chunk (`Paired → Unpairing`)
- if tombstone happened or the validator it is paired is not valid
  - state transition of insurance (`Paired → Unpairing`)
  - state transition of chunk (`Paired → Unpairing`)
- if there was an unpairing insurance came from previous epoch and it is already finished its duty
  - empty unpairing insurance id from chunk

## Remove Deletable Redelegation Infos

- For all re-delegation infos
  - if is is matured and deletable, then remove it.

## Handle Queued Liquid Unstakes

- For all UnpairingForUnstakingChunkInfos (= info)
  - got chunk from info.chunkId
  - un-delegate chunk
  - state transition of insurance (`Paired → Unpairing`)
  - state transition of chunk (`Paired → UnpairingForUnstaking`)

## Handle Queued Withdraw Insurance Requests

- For all withdraw insurance requests (= req)
  - got insurance from req.InsuranceId
  - state transition of insurance (`Paired → UnpairingForWithdrawal`)
  - state transition of chunk (`Paired → Unpairing`)
    - if the status of chunk is `UnpairingForUnstaking`, just keep it as it is
  - delete request

## Rank Insurances

- get all **re-pairable chunks,** **out insurances, and pairedInsuranceMap**
  - condition of re-pairable chunk (re-pairable means can be paired with new insurance)
    - must be one of `Pairing`, `Paired`, or `Unpairing (without unbonding obj)`
  - out insurances are
    - paired with `Unpairing` chunk which have no unbonding obj
      - The most common case for this is withdrawing an insurance.
    - paired with `Paired` chunk but have invalid validator. (sanity check)

- create candidate insurances
  - candidate insurance must be in `Pairing or Paired`
- sort candidate insurances in ascending order, with the cheapest insurance listed first.
- create rank in insurances and rank out insurances
  - rank in insurances: `candidates[:len(rePairableChunks)]`
  - rank out insurances:
    - for those in `candidates[len(rePairableChunks):]`
      - must be `Paired`. others like `Pairing` does not have matched chunk, so it is not rank out, actually.
- append out insurances from get all **re-pairable chunks,out insurances, and pairedInsuranceMap** to **rank out insurances**
- create **newly ranked in insurances**
  - **condition**
    - for those in **rank in insurances** which not exists in **pairedInsuranceMap**
- return **newly ranked in insurances** and **rank out insurances**

## RePair Ranked Insurances

- create rank out insurance chunk map
  - key: insurance id which in **ranked out insurances**
  - value: `Chunk`
- for insurance in **newly ranked in insurances**
  - if there is a rank out insurance which have same validator
    - replace insurance id of chunk with new one because it directs same validator, we don’t have to re-delegate it
      - state transition of rank out insurance (`Paired -> Unpairing`)
        - if rank out insurance is withdrawing insurance, just keep it as it is 
      - state transition of rank in insurance (`Pairing -> Paired`)
      - state transition of chunk (`Paired | Unpairing → Paired`) and update paired and unpairing insurance ids
      - delete matched insurance from **rank out insurances**
  - if there is no rank out insurance which have same validator
    - add it to **new insurances with different validators**
- for **remaining newly ranked in insurances**
  - get all **pairing chunks** which is immediately pariable
  - pair **pairing chunks** with **remaining insurances**
    - delegate chunk
- if there are no remaining **newly ranked in insurances**
  - for **out insurance** in **rank out insurances**
    - un-delegate chunk
- if there are remaining **newly ranked in insurances**
  - for **out insurance** in **rank out insurances**
    - begin re-delegation
      - src validator: from **out insurance**
      - dst validator: from **new insurance**
      - shares: original shares of delegation