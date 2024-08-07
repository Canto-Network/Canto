package keeper_test

import (
	"context"
	"math/big"

	errorsmod "cosmossdk.io/errors"
	erc20keeper "github.com/Canto-Network/Canto/v8/x/erc20/keeper"
	erc20types "github.com/Canto-Network/Canto/v8/x/erc20/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	"github.com/stretchr/testify/mock"

	tmbytes "github.com/cometbft/cometbft/libs/bytes"

	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"

	"github.com/Canto-Network/Canto/v8/x/onboarding/types"
)

var _ types.TransferKeeper = &MockTransferKeeper{}

// MockTransferKeeper defines a mocked object that implements the TransferKeeper
// interface. It's used on tests to abstract the complexity of IBC transfers.
// NOTE: Bank keeper logic is not mocked since we want to test that balance has
// been updated for sender and recipient.
type MockTransferKeeper struct {
	mock.Mock
	bankkeeper.Keeper
}

func (m *MockTransferKeeper) GetDenomTrace(ctx sdk.Context, denomTraceHash tmbytes.HexBytes) (transfertypes.DenomTrace, bool) {
	args := m.Called(mock.Anything, denomTraceHash)
	return args.Get(0).(transfertypes.DenomTrace), args.Bool(1)
}

func (m *MockTransferKeeper) SendTransfer(
	ctx sdk.Context,
	sourcePort,
	sourceChannel string,
	token sdk.Coin,
	sender sdk.AccAddress,
	receiver string,
	timeoutHeight clienttypes.Height,
	timeoutTimestamp uint64,
) error {
	args := m.Called(mock.Anything, sourcePort, sourceChannel, token, mock.Anything, mock.Anything, mock.Anything, mock.Anything)

	err := m.SendCoinsFromAccountToModule(ctx, sender, transfertypes.ModuleName, sdk.Coins{token})
	if err != nil {
		return err
	}

	return args.Error(0)
}

type MockErc20Keeper struct {
	mock.Mock
	erc20keeper erc20keeper.Keeper
	bankKeeper  bankkeeper.Keeper
}

func NewMockErc20Keeper(ek erc20keeper.Keeper, bk bankkeeper.Keeper) *MockErc20Keeper {
	return &MockErc20Keeper{erc20keeper: ek, bankKeeper: bk}
}

func (m *MockErc20Keeper) ConvertCoin(
	goCtx context.Context,
	msg *erc20types.MsgConvertCoin,
) (*erc20types.MsgConvertCoinResponse, error) {
	sender, _ := sdk.AccAddressFromBech32(msg.Sender)

	coins := sdk.Coins{msg.Coin}
	err := m.bankKeeper.SendCoinsFromAccountToModule(goCtx, sender, types.ModuleName, coins)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to escrow coins")
	}
	argsMock := m.Called(goCtx, msg)
	return nil, argsMock.Error(1)
}

func (m *MockErc20Keeper) GetTokenPairID(ctx sdk.Context, token string) []byte {
	return m.erc20keeper.GetTokenPairID(ctx, token)
}

func (m *MockErc20Keeper) GetTokenPair(ctx sdk.Context, id []byte) (erc20types.TokenPair, bool) {
	return m.erc20keeper.GetTokenPair(ctx, id)
}

func (m *MockErc20Keeper) BalanceOf(ctx sdk.Context, abi abi.ABI, contract, account common.Address) *big.Int {
	return m.erc20keeper.BalanceOf(ctx, abi, contract, account)
}

func (m *MockErc20Keeper) CallEVM(
	ctx sdk.Context,
	abi abi.ABI,
	from, contract common.Address,
	commit bool,
	method string,
	args ...interface{},
) (*evmtypes.MsgEthereumTxResponse, error) {
	argsMock := m.Called(ctx, abi, from, contract, commit, method, args)
	return nil, argsMock.Error(1)
}
