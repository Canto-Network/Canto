<!--
order: 4
-->

# Parameters
The `x/onboarding` module contains the following parameters:

| Key                    | Type         | Default Values                  |
|:-----------------------|:-------------|:--------------------------------|
| EnableOnboarding       | bool         | true                            |
| AutoSwapThreshold      | string (int) | "4000000000000000000" // 4canto |
| WhitelistedChannels    | string[]     | ["channel-0"]                   |

### EnableOnboarding
The EnableOnboarding parameter toggles Onboarding IBC middleware. When the parameter is disabled, it will disable the auto swap and convert.

### AutoSwapThreshold
The AutoSwapThreshold parameter is the threshold of the amount of canto to be swapped. When the balance of canto is less than the threshold, the auto swap will be triggered.

### WhitelistedChannels
The WhitelistedChannels parameter is the list of channels that will be whitelisted for the auto swap and convert. When the channel is not in the list, the auto swap and convert will be disabled.
