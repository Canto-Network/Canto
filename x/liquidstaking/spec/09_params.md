<!-- order: 9 -->

# Parameters

The `liquidstaking` module contains the following parameters:

| Param                | Type           | Default                                      |
|----------------------|----------------|----------------------------------------------|  
| DynamicFeeRate       | DynamicFeeRate | (Please take a look the following section.)  |
| MaximumDiscountRate  | sdk.Dec        | 0.030000000000000000 (3%)                    |

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

There is a Security Cap for UHardCap. The Security Cap is 25%, so even if the parameter is set to a value greater than 25%, the hard capp will not exceed 25%.

### UOptimal

Optimal utilization ratio.

### Slope1

If the current utilization ratio is below optimal, the fee rate increases at a slow pace.

### Slope2

If the current utilization ratio is above optimal, the fee rate increases at a faster pace.

### MaxFeeRate

Maximum fee rate. Fee rate cannot exceed this value.


## MaximumDiscountRate
The cap for the discount rate when claiming accumulated rewards of reward pool. The discount rate cannot exceed this value.

There is a Security Cap for the maximum discount rate. The Security Cap is 10%, so even if the parameter is set to a value greater than 10%, the maximum discount rate will not exceed 10%.