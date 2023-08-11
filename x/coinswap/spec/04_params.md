<!--
order: 4
-->

# Parameters

The coinswap module contains the following parameters:

| Key                    | Type         | Default value                                                                                                                                                                                                                                                                                                              |
|:-----------------------|:-------------|:---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Fee                    | string (dec) | "0.0"                                                                                                                                                                                                                                                                                                                      |
| PoolCreationFee        | sdk.Coin     | "0acanto"                                                                                                                                                                                                                                                                                                                  |
| TaxRate                | string (dec) | "0.0"                                                                                                                                                                                                                                                                                                                      |
| MaxStandardCoinPerPool | string (int) | "10000000000000000000000"                                                                                                                                                                                                                                                                                                  |
| MaxSwapAmount          | sdk.Coins    | [{"denom":"ibc/17CD484EE7D9723B847D95015FA3EBD1572FD13BC84FB838F55B18A57450F25B","amount":"10000000"},{"denom":"ibc/4F6A2DEFEA52CD8D90966ADCB2BD0593D3993AB0DF7F6AEB3EFD6167D79237B0","amount":"10000000"},{"denom":"ibc/DC186CA7A8C009B43774EBDC825C935CABA9743504CE6037507E6E5CCE12858A","amount":"100000000000000000"}] |

### Fee
Swap fee rate for swap. In this version, swap fees aren't paid upon swap orders directly. Instead, pool just adjust pool's quoting prices to reflect the swap fees.

### PoolCreationFee
Fee paid for to create a pool. This fee prevents spamming and is collected in the fee collector.

### TaxRate
Community tax rate for pool creation fee. This tax is collected in the fee collector.

### MaxStandardCoinPerPool
Maximum amount of standard coin per pool. This parameter is used to prevent pool from being too large.

### MaxSwapAmount
Maximum amount of swap amount. This parameter is used to prevent swap from being too large. It is also used as whitelist for pool creation.
