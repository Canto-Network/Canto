package keeper_test

import (
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"
	epochstypes "github.com/Canto-Network/Canto/v7/x/epochs/types"
	"github.com/Canto-Network/Canto/v7/x/inflation/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankKeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	epochNumber int64
	skipped     uint64
	provision   sdkmath.LegacyDec
	found       bool
)

var _ = Describe("Inflation", Ordered, func() {
	BeforeEach(func() {
		s.SetupTest()
	})

	Describe("Commiting a block", func() {

		Context("with inflation param enabled", func() {
			BeforeEach(func() {
				params := s.app.InflationKeeper.GetParams(s.ctx)
				params.EnableInflation = true
				s.app.InflationKeeper.SetParams(s.ctx, params)
			})

			Context("before an epoch ends", func() {
				BeforeEach(func() {
					s.CommitAfter(time.Minute)    // Start Epoch
					s.CommitAfter(time.Hour * 23) // End Epoch
				})
				It("should not allocate funds to the community pool", func() {
					feePool, err := s.app.DistrKeeper.FeePool.Get(s.ctx)
					s.Require().NoError(err)
					fmt.Println("Community Pool balance before epoch end: ", feePool.CommunityPool.AmountOf(denomMint))
					Expect(feePool.CommunityPool.IsZero()).To(BeTrue())
				})
			})

			Context("after an epoch ends", func() {
				BeforeEach(func() {
					s.CommitAfter(time.Minute)    // Start Epoch
					s.CommitAfter(time.Hour * 25) // End Epoch
				})
				It("should allocate staking provision funds to the community pool", func() {
					feePool, err := s.app.DistrKeeper.FeePool.Get(s.ctx)
					s.Require().NoError(err)
					provision, _ := s.app.InflationKeeper.GetEpochMintProvision(s.ctx)
					params := s.app.InflationKeeper.GetParams(s.ctx)

					distributionStaking := params.InflationDistribution.StakingRewards
					expectedStaking := provision.Mul(distributionStaking)

					staking := s.app.AccountKeeper.GetModuleAddress("fee_collector")
					stakingBal := s.app.BankKeeper.GetAllBalances(s.ctx, staking)
					// fees distributed
					Expect(feePool.CommunityPool.AmountOf(denomMint).Equal(expectedStaking)).To(BeTrue())
					Expect(stakingBal.AmountOf(denomMint).Equal(sdkmath.NewInt(0))).To(BeTrue())
				})
			})
		})

		Context("with inflation param disabled", func() {
			BeforeEach(func() {
				params := s.app.InflationKeeper.GetParams(s.ctx)
				params.EnableInflation = false
				s.app.InflationKeeper.SetParams(s.ctx, params)
			})

			Context("after the network was offline for several days/epochs", func() {
				BeforeEach(func() {
					s.CommitAfter(time.Minute)        // start initial epoch
					s.CommitAfter(time.Hour * 24 * 5) // end epoch after several days
				})
				When("the epoch start time has not caught up with the block time", func() {
					BeforeEach(func() {
						// commit next 3 blocks to trigger afterEpochEnd let EpochStartTime
						// catch up with BlockTime
						s.CommitAfter(time.Second * 6)
						s.CommitAfter(time.Second * 6)
						s.CommitAfter(time.Second * 6)

						epochInfo, found := s.app.EpochsKeeper.GetEpochInfo(s.ctx, epochstypes.DayEpochID)
						s.Require().True(found)
						epochNumber = epochInfo.CurrentEpoch

						skipped = s.app.InflationKeeper.GetSkippedEpochs(s.ctx)

						s.CommitAfter(time.Second * 6) // commit next block
					})
					It("should increase the epoch number ", func() {
						epochInfo, _ := s.app.EpochsKeeper.GetEpochInfo(s.ctx, epochstypes.DayEpochID)
						Expect(epochInfo.CurrentEpoch).To(Equal(epochNumber + 1))
					})
					It("should not increase the skipped epochs number", func() {
						skippedAfter := s.app.InflationKeeper.GetSkippedEpochs(s.ctx)
						Expect(skippedAfter).To(Equal(skipped + 1))
					})
				})

				When("the epoch start time has caught up with the block time", func() {
					BeforeEach(func() {
						// commit next 4 blocks to trigger afterEpochEnd hook several times
						// and let EpochStartTime catch up with BlockTime
						s.CommitAfter(time.Second * 6)
						s.CommitAfter(time.Second * 6)
						s.CommitAfter(time.Second * 6)
						s.CommitAfter(time.Second * 6)

						epochInfo, found := s.app.EpochsKeeper.GetEpochInfo(s.ctx, epochstypes.DayEpochID)
						s.Require().True(found)
						epochNumber = epochInfo.CurrentEpoch

						skipped = s.app.InflationKeeper.GetSkippedEpochs(s.ctx)

						s.CommitAfter(time.Second * 6) // commit next block
					})
					It("should not increase the epoch number ", func() {
						epochInfo, _ := s.app.EpochsKeeper.GetEpochInfo(s.ctx, epochstypes.DayEpochID)
						Expect(epochInfo.CurrentEpoch).To(Equal(epochNumber))
					})
					It("should not increase the skipped epochs number", func() {
						skippedAfter := s.app.InflationKeeper.GetSkippedEpochs(s.ctx)
						Expect(skippedAfter).To(Equal(skipped))
					})

					When("epoch number passes epochsPerPeriod + skippedEpochs and inflation re-enabled", func() {
						BeforeEach(func() {
							params := s.app.InflationKeeper.GetParams(s.ctx)
							params.EnableInflation = true
							s.app.InflationKeeper.SetParams(s.ctx, params)

							epochInfo, _ := s.app.EpochsKeeper.GetEpochInfo(s.ctx, epochstypes.DayEpochID)
							epochNumber := epochInfo.CurrentEpoch // 6

							epochsPerPeriod := int64(1)
							s.app.InflationKeeper.SetEpochsPerPeriod(s.ctx, epochsPerPeriod)
							skipped := s.app.InflationKeeper.GetSkippedEpochs(s.ctx)
							s.Require().Equal(epochNumber, epochsPerPeriod+int64(skipped))

							provision, found = s.app.InflationKeeper.GetEpochMintProvision(s.ctx)
							s.Require().True(found)
							fmt.Println(provision)

							s.CommitAfter(time.Hour * 23) // commit before next full epoch
							provisionAfter, _ := s.app.InflationKeeper.GetEpochMintProvision(s.ctx)
							s.Require().Equal(provisionAfter, provision)

							s.CommitAfter(time.Hour * 2) // commit after next full epoch
						})

						It("should recalculate the EpochMintProvision", func() {
							provisionAfter, _ := s.app.InflationKeeper.GetEpochMintProvision(s.ctx)
							// fmt.Println("provisionAfter: ", provisionAfter)
							Expect(provisionAfter).ToNot(Equal(provision))
							fmt.Println("provision after: ", provisionAfter)
							Expect(provisionAfter).To(Equal(sdkmath.LegacyMustNewDecFromStr("10597826200000000000000000.000000000000000000")))
						})
					})
				})
			})
		})
	})
})
var v stakingtypes.Validator
var _ = Describe("Inflation", Ordered, func() {
	BeforeEach(func() {
		s.clearValidatorsAndInitPool(1000)
		valAddrs := MakeValAccts(1)
		pk := GenKeys(1)
		// instantiate validator
		v, err := stakingtypes.NewValidator(valAddrs[0].String(), pk[0].PubKey(), stakingtypes.Description{})
		s.Require().NoError(err)
		s.Require().Equal(stakingtypes.Unbonded, v.Status)
		// Increment Validator balance + power Index
		tokens := s.app.StakingKeeper.TokensFromConsensusPower(s.ctx, 1000)
		v, _ = v.AddTokensFromDel(tokens)
		// set validator in state
		s.app.StakingKeeper.SetValidator(s.ctx, v)
		s.app.StakingKeeper.SetValidatorByPowerIndex(s.ctx, v)
		//update validator set
		_, err = s.app.StakingKeeper.ApplyAndReturnValidatorSetUpdates(s.ctx) // failing bc validator tokens are not enough
		s.Require().NoError(err)
		v, err = s.app.StakingKeeper.GetValidator(s.ctx, valAddrs[0])
		s.Require().NoError(err)
		s.Require().Equal(stakingtypes.Bonded, v.Status)
		// set consAddress
		s.consAddress = sdk.GetConsAddress(pk[0].PubKey())
		s.SetupTest()
	})
	Context("Expect the validator consAddress to be the block proposer address", func() {
		BeforeEach(func() {
			params := s.app.InflationKeeper.GetParams(s.ctx)
			params.EnableInflation = true
			s.app.InflationKeeper.SetParams(s.ctx, params)
		})
		It("Commit a block and check that proposer address is the address of the suite consAddress", func() {
			// commit
			s.CommitAfter(time.Minute)
			header := s.ctx.BlockHeader()
			s.Require().Equal(sdk.AccAddress(s.consAddress), sdk.AccAddress(header.ProposerAddress))
		})
		It("Commit Block Before Epoch and check rewards", func() {
			s.CommitAfter(time.Minute)
			valBal := s.app.BankKeeper.GetAllBalances(s.ctx, sdk.AccAddress(sdk.AccAddress(s.consAddress)))
			Expect(valBal.AmountOf(denomMint).Equal(sdkmath.NewInt(0))).To(BeTrue())
		})
		It("Commit block after Epoch and balance will be Epoch Mint Provision", func() {
			provision, _ := s.app.InflationKeeper.GetEpochMintProvision(s.ctx)
			s.CommitAfter(time.Minute)
			s.CommitAfter(time.Hour * 25) // epoch will have ended

			valAddr, _ := sdk.ValAddressFromBech32(v.OperatorAddress)
			valBal, err := s.app.DistrKeeper.GetValidatorCurrentRewards(s.ctx, valAddr)
			s.Require().NoError(err)
			Expect(valBal.Rewards.AmountOf(denomMint).Equal(provision)).To(BeFalse())
		})
	})
})

func (s *KeeperTestSuite) clearValidatorsAndInitPool(power int64) {
	amt := s.app.StakingKeeper.TokensFromConsensusPower(s.ctx, power)
	notBondedPool := s.app.StakingKeeper.GetNotBondedPool(s.ctx)
	bondDenom, err := s.app.StakingKeeper.BondDenom(s.ctx)
	s.Require().NoError(err)
	totalSupply := sdk.NewCoins(sdk.NewCoin(bondDenom, amt))
	s.app.AccountKeeper.SetModuleAccount(s.ctx, notBondedPool)
	err = FundModuleAccount(s.app.BankKeeper, s.ctx, notBondedPool.GetName(), totalSupply)
	s.Require().NoError(err)
}

func FundModuleAccount(bk bankKeeper.Keeper, ctx sdk.Context, recipient string, amount sdk.Coins) error {
	if err := bk.MintCoins(ctx, types.ModuleName, amount); err != nil {
		panic(err)
	}
	return bk.SendCoinsFromModuleToModule(ctx, types.ModuleName, recipient, amount)
}

func MakeValAccts(numAccts int) []sdk.ValAddress {
	addrs := make([]sdk.ValAddress, numAccts)
	for i := 0; i < numAccts; i++ {
		pk := ed25519.GenPrivKey().PubKey()
		addrs[i] = sdk.ValAddress(sdk.AccAddress(pk.Address()))
	}
	return addrs
}

func GenKeys(numKeys int) []*ed25519.PrivKey {
	pks := make([]*ed25519.PrivKey, numKeys)
	for i := 0; i < numKeys; i++ {
		pks[i] = ed25519.GenPrivKey()
	}
	return pks
}
