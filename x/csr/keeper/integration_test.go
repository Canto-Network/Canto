package keeper_test

import (
	_ "embed" // embed compiled smart contract
	"encoding/json"
	"math"
	"math/big"

	. "github.com/onsi/ginkgo/v2"

	"github.com/Canto-Network/Canto/v7/contracts"
	"github.com/Canto-Network/Canto/v7/testutil"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/evmos/ethermint/crypto/ethsecp256k1"

	evmtypes "github.com/evmos/ethermint/x/evm/types"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// / Load in all of the test contracts that will be used in the integration tests
//
//go:embed test_contracts/compiled_contracts/factoryContract.json
var factoryContractJson []byte // nolint: golint
var factoryContract evmtypes.CompiledContract

//go:embed test_contracts/compiled_contracts/csrSmartContract.json
var csrSmartContractJson []byte // nolint: golint
var csrSmartContract evmtypes.CompiledContract

var turnstileContract = contracts.TurnstileContract

var _ = Describe("CSR Distribution : ", Ordered, func() {
	var (
		// Variables pertaining to account that deploys smart contract
		deployerAddress    sdk.AccAddress
		deployerEVMAddress common.Address

		// Variables pertaining to user that interacts with smart contracts
		userKey     *ethsecp256k1.PrivKey
		userAddress sdk.AccAddress

		// Variables to track the state of CSR
		turnstileAddress common.Address
		csrShares        sdkmath.LegacyDec
		csrContracts     map[uint64][]string
		revenueByNFT     map[uint64]*big.Int

		// EVM transaction inputs
		amount    *big.Int
		gasLimit  uint64
		gasPrice  *big.Int
		gasFeeCap *big.Int
		gasTipCap *big.Int
		accesses  *ethtypes.AccessList
	)

	BeforeAll(func() {
		s.SetupTest()

		// Compile the contracts
		err := json.Unmarshal(factoryContractJson, &factoryContract)
		s.Require().NoError(err)
		err = json.Unmarshal(csrSmartContractJson, &csrSmartContract)
		s.Require().NoError(err)

		// Initial balances for the account
		initAmount := sdkmath.NewInt(int64(math.Pow10(18) * 4))
		initBalance := sdk.NewCoins(sdk.NewCoin(s.denom, initAmount))

		// Set up account that will be used to deploy smart contracts
		_, deployerAddress = GenerateKey()
		testutil.FundAccount(s.app.BankKeeper, s.ctx, deployerAddress, initBalance)
		acc := s.app.AccountKeeper.NewAccountWithAddress(s.ctx, deployerAddress)
		s.app.AccountKeeper.SetAccount(s.ctx, acc)
		deployerEVMAddress = common.BytesToAddress(deployerAddress.Bytes())

		// Set up account that will be used to interact with smart contract
		userKey, userAddress = GenerateKey()
		testutil.FundAccount(s.app.BankKeeper, s.ctx, userAddress, initBalance)
		acc = s.app.AccountKeeper.NewAccountWithAddress(s.ctx, userAddress)
		s.app.AccountKeeper.SetAccount(s.ctx, acc)
		s.Commit()

		// Retrieve the turnstile address
		turnstileAddress, _ = s.app.CSRKeeper.GetTurnstile(s.ctx)
		csrShares = s.app.CSRKeeper.GetParams(s.ctx).CsrShares

		// Set up internal contract -> nft mapping for testing purposes
		csrContracts = make(map[uint64][]string)
		revenueByNFT = make(map[uint64]*big.Int)

		// Set EVM Parameters
		amount = nil
		gasLimit = 10000000
		gasPrice = big.NewInt(1000000000)
		gasFeeCap = nil
		gasTipCap = nil
		accesses = nil
	})

	Context("Testing EVM Hook", func() {
		It("it should not register an EOA as smart contract", func() {
			data, _ := turnstileContract.ABI.Pack("register", common.BytesToAddress(userAddress.Bytes()))

			// This call will make the turnstile address register the test contract to itself
			EVMTX(userKey, &turnstileAddress, amount, gasLimit, gasPrice, gasFeeCap, gasTipCap, data, accesses)
			s.Commit()

			// CSR object should have been created and set in store
			_, found := s.app.CSRKeeper.GetCSR(s.ctx, 0)
			s.Require().False(found)
		})
		It("it should register a smart contract (non-factory deployed)", func() {

			// Deploys the contract directly to the EVM state (does not hit the postTxProcessing hooks)
			contractAddress, err := s.app.CSRKeeper.DeployContract(s.ctx, csrSmartContract, &turnstileAddress)
			s.Require().NoError(err)
			s.Commit()

			// This call will make the turnstile address register the test contract to itself
			data, err := csrSmartContract.ABI.Pack("register", deployerEVMAddress)
			s.Require().NoError(err)

			// Register the smart contract
			res, err := EVMTX(userKey, &contractAddress, amount, gasLimit, gasPrice, gasFeeCap, gasTipCap, data, accesses)
			s.Require().NoError(err)
			s.Commit()

			// Track contracts added to NFT
			csrContracts[1] = append(csrContracts[1], contractAddress.String())

			// CSR object should have been created and set in store
			csr, found := s.app.CSRKeeper.GetCSR(s.ctx, 1)
			s.Require().True(found)

			// Calculate the expected revenue for the transaction
			gasUsed := uint64(0)
			for _, txResult := range res.TxResults {
				gasUsed += uint64(txResult.GasUsed)
			}
			expectedFee := CalculateExpectedFee(gasUsed, gasPrice, csrShares).BigInt()
			revenueByNFT[1] = expectedFee

			// Check CSR obj values
			CheckCSRValues(*csr, 1, csrContracts[1], 1, revenueByNFT[1])
		})
		It("it should not re-register a smart contract", func() {
			data, err := csrSmartContract.ABI.Pack("assign", big.NewInt(1))
			s.Require().NoError(err)

			// Assign the smart contract
			contractAddress := common.HexToAddress(csrContracts[1][0])
			EVMTX(userKey, &contractAddress, amount, gasLimit, gasPrice, gasFeeCap, gasTipCap, data, accesses)
			s.Commit()

			// CSR object should have been created and set in store
			csr, found := s.app.CSRKeeper.GetCSR(s.ctx, 1)
			s.Require().True(found)

			// Check CSR obj values
			CheckCSRValues(*csr, 1, csrContracts[1], 1, revenueByNFT[1])
		})
		It("it should register a contract deployed by a smart contract factory", func() {
			// Deploys the factory contract directly to the EVM state (does not hit the postTxProcessing hooks)
			factoryContractAddress, err := s.app.CSRKeeper.DeployContract(s.ctx, factoryContract, &turnstileAddress)
			s.Require().NoError(err)
			s.Commit()

			// Check that the NFT is not registered beforehand
			_, found := s.app.CSRKeeper.GetCSR(s.ctx, 2)
			s.Require().False(found)

			// Register will create a new NFT (2) and deploy a smart contract
			data, err := factoryContract.ABI.Pack("register", deployerEVMAddress)
			s.Require().NoError(err)

			EVMTX(userKey, &factoryContractAddress, amount, gasLimit, gasPrice, gasFeeCap, gasTipCap, data, accesses)
			s.Commit()

			// CSR object should have been created and set in store
			csr, found := s.app.CSRKeeper.GetCSR(s.ctx, 2)
			s.Require().True(found)

			s.Require().Equal(csr.Txs, uint64(0))
			s.Require().Equal(len(csr.Contracts), 1)
			s.Require().Equal(csr.Revenue.BigInt(), big.NewInt(0))
		})
	})
})
