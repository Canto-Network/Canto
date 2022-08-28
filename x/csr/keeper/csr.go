package keeper

import (
	"github.com/Canto-Network/Canto/v2/x/csr/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
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
	// retrieve store / iterator
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.GetKeyPrefixPoolContracts(poolAddress.Bytes()))
	defer iterator.Close()
	// iterate over all contracts in storage and return them
	for ; iterator.Valid(); iterator.Next() {
		bz := iterator.Value()
		contract := string(bz)
		contracts = append(contracts, contract)
	}

	return contracts
}

// Set CSR, sets the updated CSR object to be indexed by the pool address
// Set the deployer by poolAddress, and the list of contracts
// Assume this will only be called on instantiation of CSR
func (k Keeper) SetCSR(ctx sdk.Context, csr types.CSR) {
	// first set the deployer of this CSR
	poolAddr := sdk.MustAccAddressFromBech32(csr.CsrPool.PoolAddress)
	k.SetDeployer(ctx, sdk.MustAccAddressFromBech32(csr.Deployer), poolAddr)
	// next for all contracts in the CSR, check if there are any that are left to be set
	contracts := csr.Contracts
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.GetKeyPrefixPoolContracts(poolAddr))
	for _, contract := range contracts {
		store.Set([]byte(contract), []byte(contract))
	}
}

// SetDeployer sets the deployer of the CSR to be indexed by the PoolAddress of the CSR,
// assumes that the poolAddress and deployer address are both valid sdk Addresses
func (k Keeper) SetDeployer(ctx sdk.Context, deployer, poolAddress sdk.AccAddress) {
	// retrieve store
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixCSRPoolDeployer)
	// set state
	store.Set(poolAddress.Bytes(), deployer.Bytes())
}

// SetContract appends to the current list of contracts for a CSR, the address passed to
// the function. The contract is indexable only by poolAddress + contractAddress.
func (k Keeper) SetContract(ctx sdk.Context, addr common.Address, poolAddress sdk.AccAddress) {
	// get prefix for thisw specific contract address (which pool is this located in)
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.GetKeyPrefixPoolContracts(poolAddress))
	store.Set(addr.Bytes(), addr.Bytes())
}
