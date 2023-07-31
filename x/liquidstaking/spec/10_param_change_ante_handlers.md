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