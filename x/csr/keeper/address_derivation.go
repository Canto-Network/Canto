package keeper

import (
	"bytes"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/Canto-Network/Canto/v2/x/csr/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

//  

// method to determine contract address using standard contract creation, using only
// bytes and nonce of sender. 
func (k Keeper) DeriveAddress(ctx sdk.Context, initDeployer common.Address, nonces []uint64, salts [][32]byte) (error, common.Address) {
	// fail if the length of the nonces / salts / initcodes
	if len(nonces) != len(salts) {
		return sdkerrors.Wrapf(types.ErrAddressDerivation, 
			"::deriveAddress: invalid length: nonces: %d, salts:%d", 
			len(nonces), len(salts)), 
			common.Address{}
	}

	derivedContract := initDeployer
	// now derive addresses using either CREATE2 or CreateAddress
	for i := 0; i < len(nonces); i++ {
		// there is no salt used to derive this address
		if salts[i] ==  [32]byte{} {
			derivedContract = crypto.CreateAddress(derivedContract, nonces[i])
		} else {
			// otherwise derive contract address through CREATE2 
			// retrieve account
			acct := k.evmKeeper.GetAccountWithoutBalance(ctx, derivedContract)
			// fail if there is no code at this address
			if bytes.Equal(acct.CodeHash,crypto.Keccak256(nil)) {
				return sdkerrors.Wrapf(types.ErrAddressDerivation, 
					"CSR::Keeper::deriveAddress: empty code hash: %s", 
					common.Bytes2Hex(derivedContract.Bytes())),
					common.Address{}
			}
			// set derived address to the contract address derived through CREATE2
			derivedContract = crypto.CreateAddress2(derivedContract, salts[i], acct.CodeHash[:32])
		}
	}

	return nil, derivedContract
}