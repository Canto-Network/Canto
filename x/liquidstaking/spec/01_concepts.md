<!-- order: 1 -->

# Concept

Protocols using the PoS(Proof-of-Stake) consensus mechanism usually require token owners to stake their tokens on the network 
in order to participate in the governance process. 
At this time, the user's token is locked and loses its potential utility. 
Liquid staking is a staking method that can avoid this loss of capital efficiency. 
In other words, liquid staking allows holders to earn staking rewards while still being able to trade or use their assets as needed, 
whereas traditional staking typically requires locking up assets for a fixed period of time.

Basically, in liquid staking, a new token(lsToken) is minted as proof of staking the native token, 
and the lsToken is circulated in the market instead of native token. 
In order for lsToken to be fully fungible, the relationship between the staking status of the native token used in liquid staking and the minted lsToken must be fungible. 
In other words, regardless of which validator the user chooses for liquid staking, 
the reward accumulated in lsToken and the risk of lsToken must be the same.
However, as is well known, each validator inevitably has differences in node operating ability, security level, and required fee rate, 
so the reward or risk of staking varies depending on which validator the user chooses. 
To solve this problem, we would like to propose our own unique liquid staking that has features such as 
**insurance**, **fee-rate competition**, **reward distribution**, and **reward withdrawal**.

## Insurance

Insurance protects against the potential loss of funds from the slashing of staked tokens. 
In simpler terms, the risk of loss from slashing is transferred to the insurance provider, 
ensuring that the initially staked tokens through the liquid staking module are always protected. 

As previously mentioned, each validator carries a varying level of risk for slashing. 
By transferring this risk to the insurance provider, the user’s choice of validator becomes no longer important. 
This means that the minted lsToken is independent of the user's choice and is completely fungible in terms of risk for slashing.

## Fee-rate competition

**Insurance providers** charge a fee for the protection of staked tokens. 
Tokens can only be staked through the liquid staking module if the corresponding **insurance** is in place. 
This requirement for **insurance** incentives **insurance providers** to charge high fees for their services. 
However, an increase in **insurance** costs decreases the return of liquid staking, which in turn reduces the motivation for users to use it. 
Therefore, it is necessary to prevent **insurance providers** from raising fees arbitrarily, we achieved this through fee rate competition.

Fee rate competition allows liquid staking only for slots whose fee rate fall within a certain rank as determined by the governance.
The fee rate here is the total of the **insurance** fee rate required by the **insurance provider** and the commission fee rate set by the validator selected by the **insurance provider**. 
In this case, both the validator and the **insurance provider** are discouraged from raising the fee rate excessively, 
because if the fee rate is set too high and the amount to be staked is not allocated, no profit will be made.

## Reward distribution

All **active chunks** have the same size of tokens(= hard coded amount: **250K**). **An active chunk** is paired with **insurance** and 
has its own Delegation object in the `staking` module, which earns rewards for every inflation epoch as set by the Inflation module.

All delegation rewards are collected at every `liquidstaking` module epoch, and the duration of the epoch must be the same as `staking` module's unbonding period. 
These rewards are then collected to a **reward module account.**

### Dynamic Fee Rate

liquidstaking module takes **a fee calculated based on utilization ratio** before delegation rewards go to the reward pool
Delegation reward are distributed as follows:
1. insurance take commission 
2. fee is burned
3. rest of the delegation reward goes to the reward pool

The fee is calculated as follows: `fee = (delegation reward - insurance commission) * feeRate`

**Fee rate** is calculated based on **utilization ratio** and **fee rate parameters** set by the governance.
* u (= utilization ratio) = `NetAmountBeforeModuleFee / total supply of native token` (for `NetAmountBeforeModuleFee`, please refer to [02_state.md](02_state.md#netamountstate-in-memory-only))
* if u < softCap then, **fee rate =** `r0`
* if softCap <= u <= optimal then, **fee rate =** `r0 + ((u - softcap) / (optimal - softcap) x slope1)`
* if optimal < u <= hardCap then, **fee rate =** `r0 + slope1 + ((min(u, hardcap) - optimal) / (hardcap - optimal) x slope2)`

An explanation of the parameters used in the above formula can be found in [09_params.md](09_params.md). 

Fee rate is calculated at the beginning of the epoch and applied to every chunk delegation rewards.
Calculated fee with fee rate is burned and the rest of the delegation reward goes to the reward pool.

## Reward withdrawal at discounted price

The rewards accumulated on the **reward module account** can be withdrawn by anyone who has lsToken, at a discounted price.

The discount rate is calculated as follows: `discount rate = reward module account's balance / NetAmount`

The cap is 3% so the discount rate cannot exceed 3%. This value is a parameter that can be changed by governance.