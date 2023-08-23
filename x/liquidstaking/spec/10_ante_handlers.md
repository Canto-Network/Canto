<!-- order: 10 -->

# Param Change Ante Handlers

The liquidstaking module works closely with the slashing and staking params.
For example, MinimumCollateral constant is calculated based on the slashing params. And epoch period of liquidstaking module is same with unbonding period of staking module.

To avoid unexpected risks, it is necessary to impose restrictions on parameter changes in the slashing and staking modules.

## Notes

When upgrading the cosmos-sdk version, it is important to check if there are any alternative methods to change a module's parameters. 
Currently, the only way to change a parameter is through a param change proposal, so it is essential to carefully review and ensure that no other methods have been introduced.

## Param Change Limit Decorator 

### Slashing module
Currently, when handle paired chunks, the liquidstaking module checks paired insurance's balance >= 5.75% of chunk size tokens. 
If not, the paired chunk start to be unbonded. 5.75% is calculated based on the current slashing params.
* 5%: SlashFractionDoubleSign
* 0.75%: SlashFractionDowntime

5.75% is very important to ensure the security of the liquidstaking module. So we need to impose restrictions on the slashing param changes.

* SignedBlocksWindow, MinSignedPerWindow, DowntimeJailDuration are not allowed to be decreased.
  * If we decrease these params, the slashing penalty can be increased.

* SlashFractionDoubleSign, SlashFractionDowntime are not allowed to be increased.
  * If we increase these params, the slashing penalty can be increased.

### Staking module
* UnbondingTime or BondDenom are not allowed to change.

# Validation Commission Change Ante Handler

The liquidstaking module has fee rate competition mechanism, so validator have incentive to lower their commission rate to get delegations from liquid staking module.
At every epoch, the liquidstaking module will check the validator commission rate + insurance fee rate and sorts by ascending order(insurance with lower fee rate comes first).

But validator can edit its commission rate at any time by using MsgEditValidator which can make the fee rate competition mechanism meaningless.
To avoid this, we need to impose restrictions on validator commission changes.

## ValCommissionChangeLimitDecorator
It only accepts MsgEditValidator when current block time is within 23 hours and 50 minutes of the next epoch. 
staking module have 24 hours limit for continuous MsgEditValidator, so validator can change only one time in that period.
