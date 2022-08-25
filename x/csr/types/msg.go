package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	ethermint "github.com/evmos/ethermint/types"
	"golang.org/x/sync/errgroup"
)

var (
	_ sdk.Msg = &MsgRegisterCSR{}
)

const (
	TypeMsgRegisterCSR = "register_csr"
)

// method to create a new instance of msgRegisterCSR
func NewMsgRegisterCSR(
	deployer sdk.AccAddress,
	nftsupply uint,
	allocations map[string]uint64, //mapping between Bech32 AccAddress
	contracts []string,
	nonces []*UIntArray,
) *MsgRegisterCSR {
	// if there are no allocations, then set the sole allocation to the deployer
	if len(allocations) == 0 {
		allocations[deployer.String()] = uint64(nftsupply) // all nfts go to deployer
	}
	// return address of the newly constructed MsgRegisterCSR
	return &MsgRegisterCSR{
		Deployer:    deployer.String(), // eth address of deployer
		NftSupply:   uint64(nftsupply),
		Allocations: allocations,
		Contracts:   contracts,
		Nonces:      nonces,
	}
}

// route to csr msg_router (method to implement sdk.Msg interface)
func (msg MsgRegisterCSR) Route() string { return RouterKey }

// type of the msgRegisterCSR
func (msg MsgRegisterCSR) Type() string { return TypeMsgRegisterCSR }

// sdk.Msg Validate Basic Method
func (msg MsgRegisterCSR) ValidateBasic() error {
	// error group for concurrent validation methods
	eg := &errgroup.Group{}
	// check that the deployer is a valid evm address
	if _, err := sdk.AccAddressFromBech32(msg.Deployer); err != nil {
		return sdkerrors.Wrapf(err, "MsgRegisterCSR: ValidateBasic: invalid sdk address: %s", msg.Deployer)
	}
	// check that the NFTSupply are non-zero
	if msg.NftSupply < uint64(1) {
		return sdkerrors.Wrapf(ErrInvalidNFTSupply, "MsgRegisterCSR: ValidateBasic: invalid NFT Supply: %d", msg.NftSupply)
	}

	noncesLen := len(msg.Nonces)
	contractsLen := len(msg.Contracts)
	// fail if array of UintArray is not as long as the array of contracts
	if noncesLen != contractsLen {
		return sdkerrors.Wrapf(ErrInvalidArity, "MsgRegisterCSR: ValidateBase: invalid length of nonces/contracts: nonces: %d, contracts: %d", noncesLen, contractsLen)
	}

	// concurrently run this method validation along
	eg.Go(msg.CheckAllocations)
	// concurrently validate all contracts in the msg
	eg.Go(msg.CheckContracts)
	//concurrently validate all nonces in the array
	eg.Go(msg.CheckNonces)
	// fail on the first error returned from the group
	if err := eg.Wait(); err != nil {
		return sdkerrors.Wrap(err, "MsgRegisterCSR: ValidateBasic: error in validation")
	}

	return nil
}

// GetSignBytes encodes msg for signing
func (msg *MsgRegisterCSR) GetSignBytes() []byte {
	return sdk.MustSortJSON(AminoCdc.MustMarshalJSON(msg))
}

// GetSigners defines the address to sign this message (must be deployer, )
func (msg *MsgRegisterCSR) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(msg.Deployer)}
}

// check that the msg is fully allocated (used in validate basic) and that the addresses allocated to are valid cosmos addresses
func (msg *MsgRegisterCSR) CheckAllocations() error {
	sumAlloc := uint64(0)
	// check that msg's allocations are exactly equal to NFTSupply
	for addr, alloc := range msg.Allocations {
		// check that sdk addresses given are valid
		if _, err := sdk.AccAddressFromBech32(addr); err != nil {
			return sdkerrors.Wrapf(err, "MsgRegisterCSR::CheckAllocations: invalid sdk address: %s", addr)
		}
		sumAlloc += alloc
	}
	if sumAlloc != msg.NftSupply {
		return sdkerrors.Wrapf(ErrMisMatchedAllocations, "MsgRegisterCSR::CheckAllocations: invalid NFT allocation: expected: %d, got: %d", msg.NftSupply, sumAlloc)
	}
	return nil
}

// check that all contracts registered non-zero and correctly formatted, called from ValidateBasic
func (msg *MsgRegisterCSR) CheckContracts() error {
	// check that none of the contract addresses are the zero address or empty
	for _, addr := range msg.Contracts {
		// check for zero-address or invalid address format
		if err := ethermint.ValidateNonZeroAddress(addr); err != nil {
			return sdkerrors.Wrapf(err, "MsgRegisterCSR::CheckContracts: invalid evm address: %s", addr)
		}
	}
	return nil
}

// check that all nonces registered are not less than 1, and that the given contracts match
func (msg *MsgRegisterCSR) CheckNonces() error {
	// check that all of the nonces registered are not less than 1
	for _, arr := range msg.Nonces {
		for _, nonce := range arr.Value {
			// if nonce is zero or negative throw error
			if nonce < uint64(1) {
				return sdkerrors.Wrapf(ErrInvalidNonce, "MsgRegisterCSR::CheckAllocations: invalid nonce: %d", nonce)
			}
		}
	}
	return nil
}
