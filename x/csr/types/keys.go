package types

const (
	// ModuleName defines the module name
	ModuleName = "csr"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName
)

const (
	// nft id -> csr
	prefixCSR = iota + 1
	// nft id -> owner
	prefixOwner
	// contract -> nft id
	prefixContract
)

// KVStore key prefixes
var (
	KeyPrefixCSR      = []byte{prefixCSR}
	KeyPrefixOwner    = []byte{prefixOwner}
	KeyPrefixContract = []byte{prefixContract}
)
