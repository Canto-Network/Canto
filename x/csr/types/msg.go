package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	ethermint "github.com/evmos/ethermint/types"
	"golang.org/x/sync/errgroup"
)

var (
	_ sdk.Msg = &MsgRegisterCSR{}
	_ sdk.Msg = &MsgUpdateCSR{}
)

const (
	TypeMsgRegisterCSR = "register_csr"
	TypeMsgUpdateCSR   = "update_csr"
)

type MsgInterface struct {
	msg interface{}
}

// method to create a new instance of msgRegisterCSR
func NewMsgRegisterCSR(
	deployer sdk.AccAddress,
	nftsupply uint64,
	allocations map[string]uint64, //mapping between Bech32 AccAddress
	contracts []string,
	nonces []*UIntArray,
) *MsgRegisterCSR {
	// if there are no allocations, then set the sole allocation to the deployer
	if len(allocations) == 0 {
		allocations[deployer.String()] = nftsupply // all nfts go to deployer
	}
	// return address of the newly constructed MsgRegisterCSR
	return &MsgRegisterCSR{
		Deployer:    deployer.String(), // canto address of deployer
		NftSupply:   nftsupply,
		Allocations: allocations,
		ContractData: &ContractData{
			Contracts: contracts,
			Nonces:    nonces,
		},
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
	// check that the deployer is a valid canto address
	if _, err := sdk.AccAddressFromBech32(msg.Deployer); err != nil {
		return sdkerrors.Wrapf(err, "MsgRegisterCSR: ValidateBasic: invalid sdk address of deployer: %s", msg.Deployer)
	}
	// check that the NFTSupply are non-zero
	if msg.NftSupply < uint64(1) {
		return sdkerrors.Wrapf(ErrInvalidNFTSupply, "MsgRegisterCSR: ValidateBasic: invalid NFT Supply: %d", msg.NftSupply)
	}

	noncesLen := len(msg.ContractData.Nonces)
	contractsLen := len(msg.ContractData.Contracts)
	// fail if array of UintArray is not as long as the array of contracts
	if noncesLen != contractsLen {
		return sdkerrors.Wrapf(ErrInvalidArity, "MsgRegisterCSR: ValidateBase: invalid length of nonces/contracts: nonces: %d, contracts: %d", noncesLen, contractsLen)
	}

	// concurrently run this method validation along
	eg.Go(msg.CheckAllocations)
	// concurrently validate all contracts in the msg
	eg.Go(msg.ContractData.CheckContracts)
	//concurrently validate all nonces in the array
	eg.Go(msg.ContractData.CheckNonces)
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

// method to create a new MsgUpdateCSR message, deployer is the sdk.AccAddress of the deployer of the CSR being updated
// pooladdr is the sdk.AccAddress of the CSR pool being updated
func NewMsgUpdateCSR(
	deployer,
	poolAddress sdk.AccAddress,
	contracts []string,
	nonces []*UIntArray,
) *MsgUpdateCSR {
	return &MsgUpdateCSR{
		Deployer:    deployer.String(),
		PoolAddress: poolAddress.String(),
		ContractData: &ContractData{
			Contracts: contracts,
			Nonces:    nonces,
		},
	}
}

// validateBasic is a method defined in the sdk.Msg interface, executed upon creation and receipt of sdk.Msgs
// performs stateless validation on this message type
func (msg MsgUpdateCSR) ValidateBasic() error {
	// first validate that the deployer's address is not an invalid sdk.AccAddress
	if _, err := sdk.AccAddressFromBech32(msg.Deployer); err != nil {
		return sdkerrors.Wrapf(err, "MsgUpdateCSR: ValidateBasic: invalid sdk address of deployer: %s", msg.Deployer)
	}
	// next check that the PoolAddress of the CSR is not invalid
	if _, err := sdk.AccAddressFromBech32(msg.PoolAddress); err != nil {
		return sdkerrors.Wrapf(err, "MsgUpdateCSR: ValidateBasic: invalid sdk address of CSR pool: %s", msg.PoolAddress)
	}
	// now check that nonces length and contracts length are equal
	noncesLen := len(msg.ContractData.Nonces)
	contractsLen := len(msg.ContractData.Contracts)
	if noncesLen != contractsLen {
		return sdkerrors.Wrapf(ErrInvalidArity, "MsgUpdateCSR: ValidateBasic: invalid length of nonces/contracts: nonces: %d, contracts: %d", noncesLen, contractsLen)
	}
	//initialize errgroup for concurrent processing of contracts and nonces
	eg := &errgroup.Group{}

	// concurrently handle checking of nonces
	eg.Go(msg.ContractData.CheckNonces)
	// concurrently handle checking of contracts
	eg.Go(msg.ContractData.CheckContracts)
	// block until groups finish and return err if non-nil
	if err := eg.Wait(); err != nil {
		return sdkerrors.Wrapf(err, "MsgUpdateCSR:: ValidateBasic: error validating contracts/nonces")
	}

	return nil
}

// returns route for MsgUpdateCSR
func (msg MsgUpdateCSR) Type() string { return TypeMsgUpdateCSR }

// returns route for MsgUpdateCSR
func (msg MsgUpdateCSR) Route() string { return RouterKey }

// getSignBytes returns the serialized bytes of this message to sign, used in determining signature of this message
func (msg *MsgUpdateCSR) GetSignBytes() []byte {
	return sdk.MustSortJSON(AminoCdc.MustMarshalJSON(msg))
}

// GetSigners returns the sdk.AccAddresses whose signatures are needed for this message, in this case, only the CSR deployer's signature is needed
func (msg *MsgUpdateCSR) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(msg.Deployer)}
}

// check that the msg is fully allocated (used in validate basic) and that the addresses allocated to are valid cosmos addresses
func (msg *MsgRegisterCSR) CheckAllocations() error {
	sumAlloc := uint64(0)
	// check that msg's allocations are exactly equal to NFTSupply
	for addr, alloc := range msg.Allocations {
		// check that sdk addresses given are valid
		if _, err := sdk.AccAddressFromBech32(addr); err != nil {
			return sdkerrors.Wrapf(err, "CheckAllocations: invalid sdk address: %s", addr)
		}
		sumAlloc += alloc
	}
	if sumAlloc != msg.NftSupply {
		return sdkerrors.Wrapf(ErrMisMatchedAllocations, "CheckAllocations: invalid NFT allocation: expected: %d, got: %d", msg.NftSupply, sumAlloc)
	}
	return nil
}

// check that all contracts registered non-zero and correctly formatted, called from ValidateBasic
func (contractData *ContractData) CheckContracts() error {
	// check that none of the contract addresses are the zero address or empty
	for _, addr := range contractData.Contracts {
		// check for zero-address or invalid address format
		if err := ethermint.ValidateNonZeroAddress(addr); err != nil {
			return sdkerrors.Wrapf(err, "CheckContracts: invalid evm address: %s", addr)
		}
	}
	return nil
}

// check that all nonces registered are not less than 1, and that the given contracts match
func (contractData *ContractData) CheckNonces() error {
	// check that all of the nonces registered are not less than 1
	for _, arr := range contractData.Nonces {
		for _, nonce := range arr.Value {
			// if nonce is zero or negative throw error
			if nonce < uint64(1) {
				return sdkerrors.Wrapf(ErrInvalidNonce, "CheckAllocations: invalid nonce: %d", nonce)
			}
		}
	}
	return nil
}
