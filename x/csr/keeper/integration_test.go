package keeper_test

import (
	_ "embed" // embed compiled smart contract
	"encoding/json"
	"math"
	"math/big"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Canto-Network/Canto/v2/app"
	"github.com/Canto-Network/Canto/v2/contracts"
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
	abci "github.com/tendermint/tendermint/abci/types"
)

var contractCode = "608060405234801561001057600080fd5b507f5202c943f7605429e15ba3fff7a2230f7bd9f3bcdf60a56ec9fe0f156c8d538f3360405161004091906100eb565b60405180910390a1610119565b600082825260208201905092915050565b7f636f6e7472616374206372656174656400000000000000000000000000000000600082015250565b600061009460108361004d565b915061009f8261005e565b602082019050919050565b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b60006100d5826100aa565b9050919050565b6100e5816100ca565b82525050565b6000604082019050818103600083015261010481610087565b905061011360208301846100dc565b92915050565b6102fe806101286000396000f3fe608060405234801561001057600080fd5b50600436106100365760003560e01c806358641a8e1461003b578063aa67735414610057575b600080fd5b610055600480360381019061005091906101f4565b610073565b005b610071600480360381019061006c9190610234565b6100e2565b005b8173ffffffffffffffffffffffffffffffffffffffff1663f5165863826040518263ffffffff1660e01b81526004016100ac9190610283565b600060405180830381600087803b1580156100c657600080fd5b505af11580156100da573d6000803e3d6000fd5b505050505050565b8173ffffffffffffffffffffffffffffffffffffffff16634420e486826040518263ffffffff1660e01b815260040161011b91906102ad565b600060405180830381600087803b15801561013557600080fd5b505af1158015610149573d6000803e3d6000fd5b505050505050565b600080fd5b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b600061018182610156565b9050919050565b61019181610176565b811461019c57600080fd5b50565b6000813590506101ae81610188565b92915050565b600067ffffffffffffffff82169050919050565b6101d1816101b4565b81146101dc57600080fd5b50565b6000813590506101ee816101c8565b92915050565b6000806040838503121561020b5761020a610151565b5b60006102198582860161019f565b925050602061022a858286016101df565b9150509250929050565b6000806040838503121561024b5761024a610151565b5b60006102598582860161019f565b925050602061026a8582860161019f565b9150509250929050565b61027d816101b4565b82525050565b60006020820190506102986000830184610274565b92915050565b6102a781610176565b82525050565b60006020820190506102c2600083018461029e565b9291505056fea26469706673582212205d27edb55199dfa50646e855a8f6110f0b31ea78685521dba532340807af5e6f64736f6c63430008100033"

//go:embed test_contracts/compiled_contracts/csrTest.json
var csrTestContractJson []byte // nolint: golint

var turnstileContract = contracts.TurnstileContract.ABI

var _ = Describe("CSR Distribution : ", Ordered, func() {
	// feeCollectorAddress := s.app.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName)
	denom := s.denom

	// account initial balances
	initAmount := sdk.NewInt(int64(math.Pow10(18) * 4))
	initBalance := sdk.NewCoins(sdk.NewCoin(denom, initAmount))

	var (
		deployerKey      *ethsecp256k1.PrivKey
		userKey          *ethsecp256k1.PrivKey
		deployerAddress  sdk.AccAddress
		userAddress      sdk.AccAddress
		params           types.Params
		turnstileAddress common.Address
		contractAddress  common.Address
		testContract     evmtypes.CompiledContract
	)

	BeforeAll(func() {
		s.SetupTest()

		params = s.app.CSRKeeper.GetParams(s.ctx)
		params.EnableCsr = true
		s.app.CSRKeeper.SetParams(s.ctx, params)
		s.Commit()

		// setup deployer account
		deployerKey, deployerAddress = generateKey()
		testutil.FundAccount(s.app.BankKeeper, s.ctx, deployerAddress, initBalance)

		// setup account interacting with registered contracts
		userKey, userAddress = generateKey()
		testutil.FundAccount(s.app.BankKeeper, s.ctx, userAddress, initBalance)
		acc := s.app.AccountKeeper.NewAccountWithAddress(s.ctx, userAddress)
		s.app.AccountKeeper.SetAccount(s.ctx, acc)
		s.Commit()

		// deploy the turnstile
		turnstileAddress, _ = s.app.CSRKeeper.GetTurnstile(s.ctx)
		s.Commit()

		// Deploy a test smart contract
		contractAddress = deployContract(deployerKey, contractCode)
		s.Commit()
	})

	Context("testing event handler", func() {
		It("it should not register an EOA as smart contract", func() {
			data, _ := turnstileContract.Pack("register", common.BytesToAddress(userAddress.Bytes()))
			gasPrice := big.NewInt(1000000000)

			// This call will make the turnstile address register the test contract to itself
			contractInteract(userKey, &turnstileAddress, gasPrice, nil, nil, data, nil)
			s.Commit()

			// CSR object should have been created and set in store
			_, found := s.app.CSRKeeper.GetCSR(s.ctx, 0)
			Expect(found).To(Equal(false))
		})

		It("it should register a smart contract", func() {
			// Register event embedded in an test smart contract
			json.Unmarshal(csrTestContractJson, &testContract)

			data, _ := testContract.ABI.Pack("register", turnstileAddress, common.BytesToAddress(userAddress.Bytes()))
			gasPrice := big.NewInt(1000000000)

			// This call will make the turnstile address register the test contract to itself
			contractInteract(userKey, &contractAddress, gasPrice, nil, nil, data, nil)
			s.Commit()

			// CSR object should have been created and set in store
			csr, found := s.app.CSRKeeper.GetCSR(s.ctx, 1)
			Expect(found).To(Equal(true))
			Expect(csr.Id).To(Equal(uint64(1)))
			Expect(csr.Account).ToNot(Equal(nil))
		})

		It("it should not register the same smart contract", func() {
			// Register event embedded in an test smart contract
			json.Unmarshal(csrTestContractJson, &testContract)

			data, _ := testContract.ABI.Pack("register", turnstileAddress, common.BytesToAddress(userAddress.Bytes()))
			gasPrice := big.NewInt(1000000000)

			// This call will make the turnstile address register the test contract to itself
			contractInteract(userKey, &contractAddress, gasPrice, nil, nil, data, nil)
			s.Commit()

			// CSR object should have been created and set in store
			csr, found := s.app.CSRKeeper.GetCSR(s.ctx, 1)
			Expect(found).To(Equal(true))
			Expect(len(csr.Contracts)).To(Equal(1))
		})

		It("it should register multiple NFTs", func() {
			// Register event embedded in an test smart contract
			json.Unmarshal(csrTestContractJson, &testContract)

			contractAddress2 := deployContract(deployerKey, contractCode)

			data, _ := testContract.ABI.Pack("register", turnstileAddress, common.BytesToAddress(userAddress.Bytes()))
			gasPrice := big.NewInt(1000000000)

			// This call will make the turnstile address register the test contract to itself
			contractInteract(userKey, &contractAddress2, gasPrice, nil, nil, data, nil)
			s.Commit()

			// CSR object should have been created and set in store
			csr, found := s.app.CSRKeeper.GetCSR(s.ctx, 2)
			Expect(found).To(Equal(true))
			Expect(csr.Id).To(Equal(uint64(2)))
			Expect(csr.Account).ToNot(Equal(nil))
		})

		It("it should register new contracts to existing NFTs", func() {
			// Register event embedded in an test smart contract
			json.Unmarshal(csrTestContractJson, &testContract)

			contractAddress3 := deployContract(deployerKey, contractCode)

			data, _ := testContract.ABI.Pack("update", turnstileAddress, uint64(1))
			gasPrice := big.NewInt(1000000000)

			// This call will make the turnstile address register the test contract to itself
			contractInteract(userKey, &contractAddress3, gasPrice, nil, nil, data, nil)
			s.Commit()

			// CSR object should have been created and set in store
			csr, found := s.app.CSRKeeper.GetCSR(s.ctx, 1)
			Expect(found).To(Equal(true))
			Expect(csr.Id).To(Equal(uint64(1)))
			Expect(len(csr.Contracts)).To(Equal(2))
			Expect(csr.Account).ToNot(Equal(nil))
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
