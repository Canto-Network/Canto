<!-- order: 9 -->

# Parameters

The `liquidstaking` module contains the following parameters:

| Param          | Type           | Default                                       |
|----------------|----------------|-----------------------------------------------|  
| DynamicFeeRate | DynamicFeeRate | (Please take a look the following section.)   |

## DynamicFeeRate

| Param      | Type             | Default                      |
|------------|------------------|------------------------------|  
| R0         | string (sdk.Dec) | "0.000000000000000000" (0%)  |
| USoftCap   | string (sdk.Dec) | "0.050000000000000000" (5%)  |
| UHardCap   | string (sdk.Dec) | "0.100000000000000000" (10%) |
| UOptimal   | string (sdk.Dec) | "0.090000000000000000" (9%)  |
| Slope1     | string (sdk.Dec) | "0.100000000000000000" (10%) |
| Slope2     | string (sdk.Dec) | "0.400000000000000000" (40%) |
| MaxFeeRate | string (sdk.Dec) | "0.500000000000000000" (50%) |

### R0

Minimum fee rate.

### USoftCap

SoftCap for utilization ratio. If U is below softcap, fee rate is R0.

### UHardCap

HardCap for utilization ratio. U cannot bigger than hardcap.

### UOptimal

Optimal utilization ratio.

### Slope1

If the current utilization ratio is below optimal, the fee rate increases at a slow pace.

### Slope2

If the current utilization ratio is above optimal, the fee rate increases at a faster pace.

### MaxFeeRate

Maximum fee rate. Fee rate cannot exceed this value.
