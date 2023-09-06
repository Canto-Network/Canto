<!-- order: 1 -->

# Concept
Protocols utilizing the Proof-of-Stake (PoS) consensus mechanism typically require 
token owners to stake their tokens on the network to participate in the governance 
process. During this time, the user's tokens are locked, resulting in a loss of 
their potential utility. Liquid staking is a staking method that mitigates this 
capital efficiency loss. In essence, liquid staking enables holders to earn staking 
rewards while retaining the ability to trade or use their assets as needed. 
In contrast, traditional staking typically mandates locking up assets for a predetermined 
period.

In the liquid staking process, a new token (lsToken) is minted as evidence of staking 
the native token, and the lsToken is traded on the market in place of the native token. 
To achieve full fungibility of lsToken, there must be fungibility between the staking 
status of the native token used in liquid staking and the minted lsToken. 
In simpler terms, regardless of the validator chosen for liquid staking, 
the rewards accumulated in lsToken and the associated risks must remain consistent. 
However, it's well-known that each validator inherently possesses differences in 
node operating capabilities, security levels, and required fee rates. 
As a result, the rewards and risks of staking vary depending on the chosen validator.

To address this challenge, we propose our distinct liquid staking solution, 
which encompasses features such as **insurance**, **fee-rate competition**, **reward distribution**, 
and **reward withdrawal**.

## Chunk

In liquid staking, staking occurs on a chunk basis rather than individual tokens, with each chunk 
having a fixed size of 250k.

## Insurance

Insurance serves as a safeguard against potential fund losses stemming from the slashing of 
staked tokens. In simpler words, the risk of losing funds due to slashing is shifted to the 
insurance provider, guaranteeing the perpetual protection of the initially staked tokens 
within the liquid staking module. 

As mentioned earlier, each validator carries a distinct level of slashing risk. 
By transferring this risk to the insurance provider, the user's selection of validator 
loses its significance. Consequently, the minted lsToken's independence from the user's 
choice ensures complete fungibility in terms of the risk associated with slashing.


## Fee-rate competition

**Insurance providers** impose a fee for the safeguarding of staked tokens. Tokens are eligible 
for staking through the liquid staking module only when the corresponding insurance coverage 
is in place. This necessity for insurance creates an incentive for insurance providers to levy 
substantial fees for their services. However, an escalation in insurance costs diminishes 
the yield of liquid staking, subsequently reducing users' motivation to utilize it. 
Therefore, preventing arbitrary fee hikes by insurance providers is required, we have incorporated 
the concept of **fee rate competition** into our implementation.

Fee rate competition allows liquid staking exclusively for slots with fee rates falling within 
a specific rank determined by the governance. The fee rate here is the total of the **insurance** 
fee rate required by the **insurance provider** and the commission fee rate set by the validator 
selected by the **insurance provider**. In this context, both the validator and the insurance provider 
are dissuaded from excessively elevating their fee rate. This is because if the commission rate is set too high, 
they will drop in the rankings and won't receive staking rewards, resulting in no profit.

## Reward distribution

All **active chunks** consist of tokens of the same size (= hard coded amount: **250K**). 
An **active chunk** is associated with **insurance** and possesses its dedicated `Delegation` object 
within the `staking` module, acquiring rewards during each inflation epoch as established by 
the `inflation` module.

All delegation rewards are collected at every `liquidstaking` module epoch, and the epoch's duration 
must match the `staking` module's unbonding period. These rewards are then collected to the 
**reward module account.**

### Dynamic Fee Rate

The `liquidstaking` module also applies a fee, which is calculated based on the utilization ratio, 
before delegation rewards are sent to the reward pool. Delegation rewards are distributed in the following manner:

1. insurances take their commission 
2. `liquidstaking` module fee is burned
3. rest of the delegation reward are sent to the reward pool

The `liquidstaking` module fee is calculated as follows: `fee = (delegation reward - insurance commission) * feeRate`

**Fee rate** is calculated based on **utilization ratio** and **fee rate parameters** set by the governance.
* u (= utilization ratio) = `NetAmountBeforeModuleFee / total supply of native token` (for `NetAmountBeforeModuleFee`, please refer to [02_state.md](02_state.md#netamountstate-in-memory-only))
* if u < softCap then, **fee rate =** `r0`
* if softCap <= u <= optimal then, **fee rate =** `r0 + ((u - softcap) / (optimal - softcap) x slope1)`
* if optimal < u <= hardCap then, **fee rate =** `r0 + slope1 + ((u - optimal) / (hardcap - optimal) x slope2)`
* if hardCap < u, then, **fee rate =** `r0 + slope1 +slope2`

An explanation of the parameters used in the above formula can be found in [09_params.md](08_params.md). 

The `liquidstaking` module fee is calculated at the beginning of every epoch and is applied to the delegation rewards of all chunks.
The calculated fee is burned and the rest of the delegation reward goes to the reward pool.

## Reward withdrawal at discounted price

The rewards amassed in **the reward module account** can be withdrawn by anyone possessing lsToken, at a reduced price.

The discount rate is calculated as follows: `discount rate = reward module account's balance / NetAmount`

However, there is a maximum limit set at 10%, which prevents the discount rate from exceeding this threshold. 
This value is a parameter that can be changed through governance; the default value is set at 3%.




