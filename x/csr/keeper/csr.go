package keeper

import (
	"encoding/binary"

	"github.com/Canto-Network/Canto/v2/x/csr/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
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

// Set CSR will place a new or update CSR type into the store with the
// key being the NFT id and the value being the marshalled CSR object
// We also map the smart contract to the correct NFT for easy reads/writes
func (k Keeper) SetCSR(ctx sdk.Context, csr types.CSR) {
	// Marshal the CSR object into a byte string
	bz, _ := csr.Marshal()

	// Convert the NFT id to bytes so we can store properly
	nftId := UInt64ToBytes(csr.Id)

	// Sets the id of the NFT to the CSR object itself
	storeCSR := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixCSR)
	storeCSR.Set(nftId, bz)

	// Add a new key, value pair in the store mapping the contract to NFT id
	contracts := csr.Contracts
	storeContracts := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixContract)
	for _, contract := range contracts {
		storeContracts.Set([]byte(contract), nftId)
	}
}

// sets the deployed Turnstile Address to state
func (k Keeper) SetTurnstile(ctx sdk.Context, turnstile common.Address) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixAddrs)
	store.Set(types.TurnstileKey, turnstile.Bytes())
}

// retrieves the deployed Turnstile Address from state
// returns the address and a boolean representing the success of the retrieval
func (k Keeper) GetTurnstile(ctx sdk.Context) (common.Address, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixAddrs)
	// retrieve state object at TurnstileKey
	bz := store.Get(types.TurnstileKey)
	if len(bz) == 0 {
		return common.Address{}, false
	}
	// if something was found, convert byte to address and return true
	return common.BytesToAddress(bz), true
}

// sets the deployed CSRNFT to state
func (k Keeper) SetCSRNFT(ctx sdk.Context, csrnft common.Address) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixAddrs)
	store.Set(types.CSRNFTKey, csrnft.Bytes())
}

// gets the deployed CSRNFT address to state
func (k Keeper) GetCSRNFT(ctx sdk.Context) (common.Address, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixAddrs)
	bz := store.Get(types.CSRNFTKey)
	if len(bz) == 0 {
		return common.Address{}, false
	}
	return common.BytesToAddress(bz), true
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
