package keeper

import (
	"encoding/binary"

	"github.com/Canto-Network/Canto/v6/x/csr/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
)

// Returns a CSR object given an NFT ID. If the ID is invalid, i.e. it does not
// exist, then GetCSR will return (nil, false). Otherwise (csr, true).
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

// Returns the NFT ID associated with a smart contract address. If the smart contract address
// entered does belong to some NFT, then it will return (id, true), otherwise (0, false).
func (k Keeper) GetNFTByContract(ctx sdk.Context, address string) (uint64, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixContract)
	bz := store.Get([]byte(address))
	if len(bz) == 0 {
		return 0, false
	}
	nftId := BytesToUInt64(bz)
	return nftId, true
}

// Set CSR will place a new or updated CSR into the store with the
// key being the NFT ID and the value being the marshalled CSR object (bytes).
// Additionally, every single smart contract address entered will get mapped to its
// NFT ID for easy reads and writes later.
func (k Keeper) SetCSR(ctx sdk.Context, csr types.CSR) {
	// Marshal the CSR object into a byte string
	bz, _ := csr.Marshal()
	// Convert the NFT ID to bytes
	nftId := UInt64ToBytes(csr.Id)

	// Sets the id of the NFT to the CSR object itself
	storeCSR := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixCSR)
	storeCSR.Set(nftId, bz)

	// Add a new key, value pair in the store mapping the contract to NFT ID
	contracts := csr.Contracts
	storeContracts := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixContract)
	for _, contract := range contracts {
		storeContracts.Set([]byte(contract), nftId)
	}
}

// Retrieves the deployed Turnstile Address from state if found.
func (k Keeper) GetTurnstile(ctx sdk.Context) (common.Address, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixAddrs)
	// retrieve state object at TurnstileKey
	bz := store.Get(types.TurnstileKey)
	if len(bz) == 0 {
		return common.Address{}, false
	}
	return common.BytesToAddress(bz), true
}

// Sets the deployed Turnstile Address to state.
func (k Keeper) SetTurnstile(ctx sdk.Context, turnstile common.Address) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixAddrs)
	store.Set(types.TurnstileKey, turnstile.Bytes())
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
