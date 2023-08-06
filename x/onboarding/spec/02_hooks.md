<!--
order: 2
-->

# Hooks

The `x/onboarding` module implements an IBC Middleware in order to swap and convert IBC transferred assets to Canto and ERC20 tokens with `Keeper.OnRecvPacket` callback.

1. A user performs an IBC transfer to the Canto network. This is done using a `FungibleTokenPacket` IBC packet.
2. Check that the onboarding conditions are met and skip to the next middleware if any condition is not satisfied:
   1. onboarding is enabled globally
   2. channel is authorized 
   4. the recipient account is not a module account
3. Check the recipient's Canto balance and if the balance is less than the `AutoSwapThreshold`, swap the assets to Canto. Amount of the swapped Canto is always equal to the `AutoSwapThreshold` and the price is determined by the liquidity pool.
4. Check if the transferred asset is registered in the `x/erc20` module as a ERC20 token pair and the token pair is enabled. If so, convert the remaining assets to ERC20 tokens.
