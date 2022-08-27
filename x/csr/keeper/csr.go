package keeper

import (
	"github.com/Canto-Network/Canto/v2/x/csr/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) GetCSR(ctx sdk.Context, poolAddress sdk.AccAddress) (*types.CSR, bool) {
	deployer, found := k.GetDeployer(ctx, poolAddress)

	if !found {
		return nil, false
	}

	// If there was no deployer, this means that the CSR pool was never registered
	// i.e. there are also no smart contracts for this pool
	contracts := k.GetContracts(ctx, poolAddress)

	return &types.CSR{
		Deployer:  deployer.String(),
		Contracts: contracts,
		CsrPool: &types.CSRPool{
			CsrNfts:     []*types.CSRNFT{},
			NftSupply:   1,
			PoolAddress: poolAddress.String(),
		},
	}, true
}

// Returns the deployer of a CSR pool given the pool address.
func (k Keeper) GetDeployer(ctx sdk.Context, poolAddress sdk.AccAddress) (sdk.AccAddress, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixCSRPoolDeployer)
	bz := store.Get(poolAddress.Bytes())
	if len(bz) == 0 {
		return nil, false
	}
	return sdk.AccAddress(bz), true
}

// Returns all of the contracts deploying to a csr pool
func (k Keeper) GetContracts(ctx sdk.Context, poolAddress sdk.AccAddress) []string {
	contracts := []string{}

	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.GetKeyPrefixPoolContracts(poolAddress.Bytes()))
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		bz := iterator.Value()
		contract := string(bz)
		contracts = append(contracts, contract)
	}

	return contracts
}
