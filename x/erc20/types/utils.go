package types

import (
	"fmt"
	"math/big"
	"math/rand"
	"regexp"
	"strings"

	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/tendermint/tendermint/crypto"
)

const (
	// (?m)^(\d+) remove leading numbers
	reLeadingNumbers = `(?m)^(\d+)`
	// ^[^A-Za-z] forces first chars to be letters
	// [^a-zA-Z0-9/-] deletes special characters
	reDnmString = `^[^A-Za-z]|[^a-zA-Z0-9/-]`
)

func removeLeadingNumbers(str string) string {
	re := regexp.MustCompile(reLeadingNumbers)
	return re.ReplaceAllString(str, "")
}

func removeSpecialChars(str string) string {
	re := regexp.MustCompile(reDnmString)
	return re.ReplaceAllString(str, "")
}

// recursively remove every invalid prefix
func removeInvalidPrefixes(str string) string {
	if strings.HasPrefix(str, "ibc/") {
		return removeInvalidPrefixes(str[4:])
	}
	if strings.HasPrefix(str, "erc20/") {
		return removeInvalidPrefixes(str[6:])
	}
	return str
}

// SanitizeERC20Name enforces 128 max string length, deletes leading numbers
// removes special characters  (except /)  and spaces from the ERC20 name
func SanitizeERC20Name(name string) string {
	name = removeLeadingNumbers(name)
	name = removeSpecialChars(name)
	if len(name) > 128 {
		name = name[:128]
	}
	name = removeInvalidPrefixes(name)
	return name
}

// EqualMetadata checks if all the fields of the provided coin metadata are equal.
func EqualMetadata(a, b banktypes.Metadata) error {
	if a.Base == b.Base && a.Description == b.Description && a.Display == b.Display && a.Name == b.Name && a.Symbol == b.Symbol {
		if len(a.DenomUnits) != len(b.DenomUnits) {
			return fmt.Errorf("metadata provided has different denom units from stored, %d ≠ %d", len(a.DenomUnits), len(b.DenomUnits))
		}

		for i, v := range a.DenomUnits {
			if (v.Exponent != b.DenomUnits[i].Exponent) || (v.Denom != b.DenomUnits[i].Denom) || !EqualStringSlice(v.Aliases, b.DenomUnits[i].Aliases) {
				return fmt.Errorf("metadata provided has different denom unit from stored, %s ≠ %s", a.DenomUnits[i], b.DenomUnits[i])
			}
		}

		return nil
	}
	return fmt.Errorf("metadata provided is different from stored")
}

// EqualStringSlice checks if two string slices are equal.
func EqualStringSlice(aliasesA, aliasesB []string) bool {
	if len(aliasesA) != len(aliasesB) {
		return false
	}

	for i := 0; i < len(aliasesA); i++ {
		if aliasesA[i] != aliasesB[i] {
			return false
		}
	}

	return true
}

// DeriveAddress derives an address with the given address length type, module name, and
func DeriveAddress(moduleName, name string) sdk.AccAddress {
	return sdk.AccAddress(crypto.AddressHash([]byte(moduleName + name)))
}

// RandomInt returns a random integer in the half-open interval [min, max).
func RandomInt(r *rand.Rand, min, max sdk.Int) sdk.Int {
	return min.Add(sdk.NewIntFromBigInt(new(big.Int).Rand(r, max.Sub(min).BigInt())))
}

// RandomDec returns a random decimal in the half-open interval [min, max).
func RandomDec(r *rand.Rand, min, max sdk.Dec) sdk.Dec {
	return min.Add(sdk.NewDecFromBigIntWithPrec(new(big.Int).Rand(r, max.Sub(min).BigInt()), sdk.Precision))
}

// GenAndDeliverTx generates a transactions and delivers it.
func GenAndDeliverTx(txCtx simulation.OperationInput, fees sdk.Coins, gas uint64) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
	account := txCtx.AccountKeeper.GetAccount(txCtx.Context, txCtx.SimAccount.Address)
	tx, err := helpers.GenTx(
		txCtx.TxGen,
		[]sdk.Msg{txCtx.Msg},
		fees,
		gas,
		txCtx.Context.ChainID(),
		[]uint64{account.GetAccountNumber()},
		[]uint64{account.GetSequence()},
		txCtx.SimAccount.PrivKey,
	)

	if err != nil {
		return simtypes.NoOpMsg(txCtx.ModuleName, txCtx.MsgType, "unable to generate mock tx"), nil, err
	}

	_, _, err = txCtx.App.Deliver(txCtx.TxGen.TxEncoder(), tx)
	if err != nil {
		return simtypes.NoOpMsg(txCtx.ModuleName, txCtx.MsgType, "unable to deliver tx"), nil, err
	}

	return simtypes.NewOperationMsg(txCtx.Msg, true, "", txCtx.Cdc), nil, nil
}

// GenAndDeliverTxWithFees generates a transaction with given fee and delivers it.
func GenAndDeliverTxWithFees(txCtx simulation.OperationInput, gas uint64, fees sdk.Coins) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
	account := txCtx.AccountKeeper.GetAccount(txCtx.Context, txCtx.SimAccount.Address)
	spendable := txCtx.Bankkeeper.SpendableCoins(txCtx.Context, account.GetAddress())

	var err error

	_, hasNeg := spendable.SafeSub(txCtx.CoinsSpentInMsg)
	if hasNeg {
		return simtypes.NoOpMsg(txCtx.ModuleName, txCtx.MsgType, "message doesn't leave room for fees"), nil, err
	}

	if err != nil {
		return simtypes.NoOpMsg(txCtx.ModuleName, txCtx.MsgType, "unable to generate fees"), nil, err
	}
	return GenAndDeliverTx(txCtx, fees, gas)
}

func GenRandomCoinMetadata(r *rand.Rand) banktypes.Metadata {
	randDescription := simtypes.RandStringOfLength(r, 10)
	randTokenBase := "a" + simtypes.RandStringOfLength(r, 4)
	randSymbol := strings.ToUpper(simtypes.RandStringOfLength(r, 4))

	validMetadata := banktypes.Metadata{
		Description: randDescription,
		Base:        randTokenBase,
		// NOTE: Denom units MUST be increasing
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    randTokenBase,
				Exponent: 0,
			},
			{
				Denom:    randTokenBase[1:],
				Exponent: uint32(18),
			},
		},
		Name:    randTokenBase,
		Symbol:  randSymbol,
		Display: randTokenBase,
	}

	return validMetadata
}
