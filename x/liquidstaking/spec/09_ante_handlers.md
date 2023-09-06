<!-- order: 9 -->

# Param Change Ante Handlers

The `liquidstaking` module operates in close conjunction with the parameters of the `slashing` and `staking` modules. For instance, constants like `MinimumCollateral` are derived from the `slashing` parameters. Additionally, the epoch period in the `liquidstaking` module aligns with the unbonding period in the `staking` module.

To mitigate unforeseen risks, it becomes imperative to enforce constraints on parameter modifications within the slashing and staking modules.

## Notes

During the process of upgrading the cosmos-sdk version, it is crucial to verify whether any alternative methods for adjusting module parameters have been introduced. Currently, the sole avenue for modifying a parameter is via a param change proposal. Therefore, it is of utmost importance to conduct a thorough review to confirm that no additional mechanisms have been introduced.
## Param Change Limit Decorator 

### Slashing module
At present, when managing paired chunks, the `liquidstaking` module verifies whether the balance of the paired insurance is greater than or equal to 5.75% of the chunk size tokens. If this condition is not met, the paired chunk initiates the unbonding process. The calculation of the 5.75% threshold is derived from the existing slashing parameters.
* 5%: SlashFractionDoubleSign
* 0.75%: SlashFractionDowntime

The significance of 5.75% lies in safeguarding the security of the `liquidstaking` module. Consequently, it becomes imperative to enforce limitations on any alterations to the slashing parameters.
* `SignedBlocksWindow`, `MinSignedPerWindow`, `DowntimeJailDuration` are not allowed to be decreased: reducing these parameters could lead to an increase in the slashing penalty

* `SlashFractionDoubleSign`, `SlashFractionDowntime` are not allowed to be increased: increasing these parameters could result in an escalation of the slashing penalty.

  
### Staking module
* `UnbondingTime` or `BondDenom` are not allowed to change: the `liquidstaking` module's epoch period is identical to the `staking` module's unbonding period. Therefore, any changes to the unbonding period could lead to a mismatch between the two modules' epoch periods, resulting in the `liquidstaking` module's failure to function properly.

# Validation Commission Change Ante Handler

The `liquidstaking` module sorts validator-insurance pairs in ascending order of the combined insurance fee and validator commission. Only pairs within a specific ranking can participate in liquidstaking. As a result, validators and insurers voluntarily lower their fee rates to engage in a fee rate competition mechanism, aiming to secure more delegation rewards.
The logic for calculating the ranking takes place during each epoch. Therefore, malicious validators can manipulate the ranking by setting a low commission rate just before the epoch and then increasing it significantly using the MsgEditValidator message after the epoch has passed. This can render the natural fee rate competition meaningless.


## ValCommissionChangeLimitDecorator
To prevent this, the `liquidstaking` module imposes a restriction on the frequency of commission rate changes.
It only accepts `MsgEditValidator` when current block time is within a certain time window (23 hours and 50 minutes before the upcoming epoch).
`staking` module have 24 hours limit for continuous `MsgEditValidator`, therefore validators can change their commission rates only once per epoch.
