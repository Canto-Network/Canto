package keeper

import (
	"encoding/binary"

	"github.com/Canto-Network/Canto/v2/x/csr/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Returns the CSR object given an NFT, returns nil if the NFT id has no
// corresponding CSR object
func (k Keeper) GetCSR(ctx sdk.Context, nftId uint64) (*types.CSR, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixCSR)
	key := UInt64ToBytes(nftId)

	bz := store.Get(key)
	if len(bz) == 0 {
		return nil, false
	}

	csr := &types.CSR{}
	csr.Unmarshal(bz)
	return csr, true
}

// Returns the NFT id a given smart contract corresponds to if it exists in the store
func (k Keeper) GetNFTByContract(ctx sdk.Context, address string) (uint64, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixContract)
	bz := store.Get([]byte(address))
	if len(bz) == 0 {
		return 0, false
	}
	nftId := BytesToUInt64(bz)
	return nftId, true
}

// Returns all of the CSRs an account is the owner of
func (k Keeper) GetCSRsByOwner(ctx sdk.Context, account string) []uint64 {
	csrs := make([]uint64, 0)
	// retrieve store / iterator
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.KeyPrefixOwner)
	defer iterator.Close()
	// iterate over all contracts in storage and return them
	for ; iterator.Valid(); iterator.Next() {
		bz := iterator.Value()
		owner := string(bz[:])

		if owner == account {
			nftId := BytesToUInt64(iterator.Key()[1:])
			csrs = append(csrs, nftId)
		}
	}

	return csrs
}

// Set CSR will place a new or update CSR type into the store with the
// key being the NFT id and the value being the marshalled CSR object
// This will also allow mapping the owner to NFT id and contract to NFT id
// for fast read and writes.
func (k Keeper) SetCSR(ctx sdk.Context, csr types.CSR) {
	// Marshal the CSR object into a byte string
	bz, _ := csr.Marshal()

	// Convert the NFT id to bytes so we can store properly
	nftId := UInt64ToBytes(csr.Id)

	// Sets the id of the NFT to the CSR object itself
	storeCSR := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixCSR)
	storeCSR.Set(nftId, bz)

	// Add a new key, value pair in the store mapping owner to nft id
	k.SetCSROwner(ctx, csr.Id, csr.Owner)

	// Add a new key, value pair in the store mapping the contract to NFT id
	contracts := csr.Contracts
	storeContracts := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixContract)
	for _, contract := range contracts {
		storeContracts.Set([]byte(contract), nftId)
	}
}

// Sets the owner of the CSR object (denoated by id) to a new account
// ONLY CHANGES THE STORE NOT THE CSR
func (k Keeper) SetCSROwner(ctx sdk.Context, id uint64, account string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixOwner)
	key := UInt64ToBytes(id)
	store.Set(key, []byte(account))
}

// Converts a uint64 to a []byte
func UInt64ToBytes(number uint64) []byte {
	bz := make([]byte, 8)
	binary.LittleEndian.PutUint64(bz, number)
	return bz
}

// Converts a []byte into a uint64
func BytesToUInt64(bz []byte) uint64 {
	return uint64(binary.LittleEndian.Uint64(bz))
}
