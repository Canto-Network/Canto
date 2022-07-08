package cli

import (
	"encoding/json"
	"github.com/Canto-Network/Canto-Testnet-v2/v1/x/unigov/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"io/ioutil"
	"path/filepath"
)

// PARSING METADATA ACCORDING TO PROPOSAL STRUCT IN GOVTYPES TYPE IN UNIGOV

// ParseRegisterCoinProposal reads and parses a ParseRegisterCoinProposal from a file.
func ParseLendingMarketMetadata(cdc codec.JSONCodec, metadataFile string) (types.LendingMarketMetadata, error) {
	propMetaData := types.LendingMarketMetadata{}

	contents, err := ioutil.ReadFile(filepath.Clean(metadataFile))
	if err != nil {
		return propMetaData, err
	}

	// if err = cdc.UnmarshalJSON(contents, &propMetaData); err != nil {
	// 	return propMetaData, err
	// }

	if err = json.Unmarshal(contents, &propMetaData); err != nil {
		return types.LendingMarketMetadata{}, err
	}

	propMetaData.PropId = 0

	return propMetaData, nil
}

func ParseTreasuryMetadata(cdc codec.JSONCodec, metadataFile string) (types.TreasuryProposalMetadata, error) {
	propMetaData := types.TreasuryProposalMetadata{}

	contents, err := ioutil.ReadFile(filepath.Clean(metadataFile))
	if err != nil {
		return propMetaData, err
	}

	// if err = cdc.UnmarshalJSON(contents, &propMetaData); err != nil {
	// 	return propMetaData, err
	// }

	if err = json.Unmarshal(contents, &propMetaData); err != nil {
		return types.TreasuryProposalMetadata{}, err
	}

	propMetaData.PropID = 0

	return propMetaData, nil
}
