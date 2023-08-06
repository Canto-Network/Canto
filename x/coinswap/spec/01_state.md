<!--
order: 1
-->

# State

## Params

Params is a module-wide configuration structure that stores system parameters and defines overall functioning of the token module.

```go
type Params struct {
    Fee                    sdk.Dec
    PoolCreationFee        sdk.Coin  
    TaxRate                sdk.Dec 
    MaxStandardCoinPerPool sdk.Int   
    MaxSwapAmount          sdk.Coins 
}
```

## Pool

Pool stores information about the liquidity pool.

```go
type Pool struct {
    Id                  string  // id of the pool
    StandardDenom       string  // denom of base coin of the pool
    CounterpartyDenom   string  // denom of counterparty coin of the pool
    EscrowAddress       string  // escrow account for deposit tokens
    LptDenom            string  // denom of the liquidity pool coin
}
```
