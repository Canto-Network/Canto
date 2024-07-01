package types

import (
	"strings"

	errorsmod "cosmossdk.io/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

const (
	ProposalTypeLendingMarket string = "Lending-Market"
	ProposalTypeTreasury      string = "Treasury"
	MaxDescriptionLength      int    = 1000
	MaxTitleLength            int    = 140
)

var (
	_ govv1beta1.Content = &LendingMarketProposal{}
	_ govv1beta1.Content = &TreasuryProposal{}
)

// Register Compound Proposal type as a valid proposal type in goveranance module
func init() {
	govv1beta1.RegisterProposalType(ProposalTypeLendingMarket)
	govv1beta1.RegisterProposalType(ProposalTypeTreasury)
}

func NewLendingMarketProposal(title, description string, m *LendingMarketMetadata) govv1beta1.Content {
	return &LendingMarketProposal{
		Title:       title,
		Description: description,
		Metadata:    m,
	}
}

func NewTreasuryProposal(title, description string, tm *TreasuryProposalMetadata) govv1beta1.Content {
	return &TreasuryProposal{
		Title:       title,
		Description: description,
		Metadata:    tm,
	}
}

func (*TreasuryProposal) ProposalRoute() string { return RouterKey }

func (*TreasuryProposal) ProposalType() string {
	return ProposalTypeTreasury
}

func (*LendingMarketProposal) ProposalRoute() string { return RouterKey }

func (*LendingMarketProposal) ProposalType() string {
	return ProposalTypeLendingMarket
}

func (lm *LendingMarketProposal) ValidateBasic() error {
	if err := govv1beta1.ValidateAbstract(lm); err != nil {
		return err
	}

	m := lm.GetMetadata()

	cd, vals, sigs := len(m.GetCalldatas()), len(m.GetValues()), len(m.GetSignatures())

	if cd != vals {
		return errorsmod.Wrapf(govtypes.ErrInvalidProposalContent, "proposal array arguments must be same length")
	}

	if vals != sigs {
		return errorsmod.Wrapf(govtypes.ErrInvalidProposalContent, "proposal array arguments must be same length")
	}
	return nil
}

func (tp *TreasuryProposal) ValidateBasic() error {
	if err := govv1beta1.ValidateAbstract(tp); err != nil {
		return err
	}

	tm := tp.GetMetadata()
	s := strings.ToLower(tm.GetDenom())

	if s != "canto" && s != "note" {
		return errorsmod.Wrapf(govtypes.ErrInvalidProposalContent, "%s is not a valid denom string", tm.GetDenom())
	}

	return nil
}

func (tp *TreasuryProposal) FromTreasuryToLendingMarket() *LendingMarketProposal {
	m := tp.GetMetadata()

	lm := LendingMarketMetadata{
		Account:    []string{m.GetRecipient()},
		PropId:     m.GetPropID(),
		Values:     []uint64{m.GetAmount()},
		Calldatas:  nil,
		Signatures: []string{m.GetDenom()},
	}

	return &LendingMarketProposal{
		Title:       tp.GetTitle(),
		Description: tp.GetDescription(),
		Metadata:    &lm,
	}
}
