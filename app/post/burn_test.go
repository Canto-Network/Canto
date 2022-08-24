package posthandler_test

import (
	feemarkettypes "github.com/evmos/ethermint/x/feemarket/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// test that base fee is burned for Eth Txs When feeBurn is enabled

// test that base fee is not burned for non-eth Txs

// test that base fee is not burned for eth Txs when the feemarket is not enabled

// test that for all possible instances of the feemarket params, the baseFeeBurning is correct

type args struct {
	feeGenState *feemarkettypes.GenesisState
	isEthTx     bool
}

func (suite *PostTestSuite) TestBaseFeeBurn() {
	testCases := []struct { 
		name 		string,
		args 		args,
		gasConsumed uint64,
		amtBurned   uint64,
		isErr 		bool,
	}{
		{
			"base fee is burned for eth txs when feeburn is enabled", 
			&feemarkettypes.GenesisState{
				NoBaseFee: false, // no base fee is false, indicating that we will indeed be burning base fee for this tx
				BaseFeeChangeDenominator: 1, // base fee change denom will be 1 in this case
				ElasticityMultiplier: 1, 
				EnableHeight: 1000,
				BaseFee: sdk.Int(100), 
				MinGasPrice sdk.Dec()
			},
			0,
			1000, 
			false,
		},
	}
}
