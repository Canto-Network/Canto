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
* SignedBlocksWindow, MinSignedPerWindow, DowntimeJailDuration are not allowed to be decreased.
  * If we decrease these params, the slashing penalty can be increased.

* SlashFractionDoubleSign, SlashFractionDowntime are not allowed to be increased.
  * If we increase these params, the slashing penalty can be increased.

### Staking module
* UnbondingTime or BondDenom are not allowed to change.