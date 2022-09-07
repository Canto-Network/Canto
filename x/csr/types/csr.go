package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	ethermint "github.com/evmos/ethermint/types"
)

// Creates a new instance of the CSR object
func NewCSR(owner sdk.AccAddress, contracts []string, id uint64, account sdk.AccAddress) CSR {
	return CSR{
		Owner:     owner.String(),
		Contracts: contracts,
		Id:        id,
		Account:   account.String(),
	}
}

// Validate performs stateless validation of a CSR object
func (csr CSR) Validate() error {
	// Check if the address of the owner is valid
	owner := csr.Owner
	if _, err := sdk.AccAddressFromBech32(owner); err != nil {
		return err
	}

	seenSmartContracts := make(map[string]bool)
	for _, smartContract := range csr.Contracts {
		if err := ethermint.ValidateNonZeroAddress(smartContract); err != nil {
			return sdkerrors.Wrapf(ErrInvalidSmartContractAddress, "CSR::Validate one or more of the entered smart contract address are invalid.")
		}

		if seenSmartContracts[smartContract] {
			return sdkerrors.Wrapf(ErrDuplicateSmartContracts, "CSR::Validate there are duplicate smart contracts in this CSR.")
		}
	}

	// Ensure that there is at least one smart contract in the CSR Pool
	numSmartContracts := len(csr.Contracts)
	if numSmartContracts < 1 {
		return sdkerrors.Wrapf(ErrSmartContractSupply, "CSR::Validate # of smart contracts must be greater than 0 got: %d", numSmartContracts)
	}

	// Ensure that the account address entered is a valid canto address
	account := csr.Account
	if _, err := sdk.AccAddressFromBech32(account); err != nil {
		return err
	}
	return nil
}
