package keeper_test

import (
	"fmt"
	"math"
	"math/big"

	"github.com/Canto-Network/Canto/v2/app"
	"github.com/Canto-Network/Canto/v2/testutil"
	"github.com/Canto-Network/Canto/v2/x/csr/types"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/evmos/ethermint/crypto/ethsecp256k1"
	"github.com/evmos/ethermint/encoding"
	"github.com/evmos/ethermint/tests"

	evmtypes "github.com/evmos/ethermint/x/evm/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	. "github.com/onsi/ginkgo/v2"
	abci "github.com/tendermint/tendermint/abci/types"
)

var contractCode = "608060405234801561001057600080fd5b5061001f61002460201b60201c565b610129565b7f142f41d272585cc7a6eae3dbcac228c0151c4c458c743eddab11b2c2fbac73553360405161005391906100fb565b60405180910390a1565b600082825260208201905092915050565b7f75706461746564206576656e7400000000000000000000000000000000000000600082015250565b60006100a4600d8361005d565b91506100af8261006e565b602082019050919050565b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b60006100e5826100ba565b9050919050565b6100f5816100da565b82525050565b6000604082019050818103600083015261011481610097565b905061012360208301846100ec565b92915050565b610175806101386000396000f3fe608060405234801561001057600080fd5b506004361061002b5760003560e01c80637b0cb83914610030575b600080fd5b61003861003a565b005b7f142f41d272585cc7a6eae3dbcac228c0151c4c458c743eddab11b2c2fbac7355336040516100699190610111565b60405180910390a1565b600082825260208201905092915050565b7f75706461746564206576656e7400000000000000000000000000000000000000600082015250565b60006100ba600d83610073565b91506100c582610084565b602082019050919050565b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b60006100fb826100d0565b9050919050565b61010b816100f0565b82525050565b6000604082019050818103600083015261012a816100ad565b90506101396020830184610102565b9291505056fea26469706673582212203907ed7d0b543881f2292494961d2548a3b5e14fac4b6823dbb85069899ea63364736f6c63430008100033"

var _ = Describe("CSR Distribution : ", Ordered, func() {
	// feeCollectorAddress := s.app.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName)
	denom := s.denom

	// account initial balances
	initAmount := sdk.NewInt(int64(math.Pow10(18) * 4))
	initBalance := sdk.NewCoins(sdk.NewCoin(denom, initAmount))

	var (
		deployerKey     *ethsecp256k1.PrivKey
		userKey         *ethsecp256k1.PrivKey
		deployerAddress sdk.AccAddress
		userAddress     sdk.AccAddress
		params          types.Params
		contractAddress common.Address
	)

	BeforeAll(func() {
		s.SetupTest()

		params = s.app.CSRKeeper.GetParams(s.ctx)
		params.EnableCsr = true
		s.app.CSRKeeper.SetParams(s.ctx, params)

		// setup deployer account
		deployerKey, deployerAddress = generateKey()
		testutil.FundAccount(s.app.BankKeeper, s.ctx, deployerAddress, initBalance)

		// setup account interacting with registered contracts
		userKey, userAddress = generateKey()
		testutil.FundAccount(s.app.BankKeeper, s.ctx, userAddress, initBalance)
		acc := s.app.AccountKeeper.NewAccountWithAddress(s.ctx, userAddress)
		s.app.AccountKeeper.SetAccount(s.ctx, acc)
		s.Commit()

		// deploy a factory
		contractAddress = deployContract(deployerKey, contractCode)
		s.Commit()

	})

	Context("testing init", func() {
		It("It should exist", func() {
			gasPrice := big.NewInt(1000000000)

			dataSig := []byte("emitEvent()")
			data := crypto.Keccak256Hash(dataSig)

			contractInteract(userKey, &contractAddress, gasPrice, nil, nil, data.Bytes()[:4], nil)
			s.Commit()
		})
	})

})

func generateKey() (*ethsecp256k1.PrivKey, sdk.AccAddress) {
	address, priv := tests.NewAddrKey()
	return priv.(*ethsecp256k1.PrivKey), sdk.AccAddress(address.Bytes())
}

func deployContract(priv *ethsecp256k1.PrivKey, contractCode string) common.Address {
	from := common.BytesToAddress(priv.PubKey().Address().Bytes())
	data := common.Hex2Bytes(contractCode)
	nonce := getNonce(from.Bytes())

	s.app.Erc20Keeper.CallEVMWithData(s.ctx, from, nil, data, true)

	contractAddress := crypto.CreateAddress(from, nonce)
	acc := s.app.EvmKeeper.GetAccountWithoutBalance(s.ctx, contractAddress)

	s.Require().NotEmpty(acc)
	s.Require().True(acc.IsContract())
	return contractAddress
}

func getNonce(addressBytes []byte) uint64 {
	return s.app.EvmKeeper.GetNonce(
		s.ctx,
		common.BytesToAddress(addressBytes),
	)
}

func contractInteract(
	priv *ethsecp256k1.PrivKey,
	contractAddr *common.Address,
	gasPrice *big.Int,
	gasFeeCap *big.Int,
	gasTipCap *big.Int,
	data []byte,
	accesses *ethtypes.AccessList,
) abci.ResponseDeliverTx {
	msgEthereumTx := buildEthTx(priv, contractAddr, gasPrice, gasFeeCap, gasTipCap, data, accesses)
	res := deliverEthTx(priv, msgEthereumTx)
	return res
}

func buildEthTx(
	priv *ethsecp256k1.PrivKey,
	to *common.Address,
	gasPrice *big.Int,
	gasFeeCap *big.Int,
	gasTipCap *big.Int,
	data []byte,
	accesses *ethtypes.AccessList,
) *evmtypes.MsgEthereumTx {
	chainID := s.app.EvmKeeper.ChainID()
	from := common.BytesToAddress(priv.PubKey().Address().Bytes())
	fmt.Println(from)
	nonce := getNonce(from.Bytes())
	gasLimit := uint64(100000)
	msgEthereumTx := evmtypes.NewTx(
		chainID,
		nonce,
		to,
		nil,
		gasLimit,
		gasPrice,
		gasFeeCap,
		gasTipCap,
		data,
		accesses,
	)
	msgEthereumTx.From = from.String()
	return msgEthereumTx
}

func deliverEthTx(priv *ethsecp256k1.PrivKey, msgEthereumTx *evmtypes.MsgEthereumTx) abci.ResponseDeliverTx {
	bz := prepareEthTx(priv, msgEthereumTx)
	req := abci.RequestDeliverTx{Tx: bz}
	res := s.app.BaseApp.DeliverTx(req)
	return res
}

func prepareEthTx(priv *ethsecp256k1.PrivKey, msgEthereumTx *evmtypes.MsgEthereumTx) []byte {
	// Sign transaction
	err := msgEthereumTx.Sign(s.ethSigner, tests.NewSigner(priv))
	s.Require().NoError(err)

	// Assemble transaction from fields
	encodingConfig := encoding.MakeConfig(app.ModuleBasics)
	txBuilder := encodingConfig.TxConfig.NewTxBuilder()
	tx, err := msgEthereumTx.BuildTx(txBuilder, s.app.EvmKeeper.GetParams(s.ctx).EvmDenom)
	s.Require().NoError(err)

	// Encode transaction by default Tx encoder and broadcasted over the network
	txEncoder := encodingConfig.TxConfig.TxEncoder()
	bz, err := txEncoder(tx)
	s.Require().NoError(err)

	return bz
}
