package types_test

import (
	"testing"

	"github.com/Canto-Network/Canto/v6/x/liquidstaking/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
)

type msgTestSuite struct {
	suite.Suite
}

func TestMsgTestSuite(t *testing.T) {
	suite.Run(t, new(msgTestSuite))
}

func (suite *msgTestSuite) TestMsgLiquidStake() {
	delegator := sdk.AccAddress("1")
	stakingCoin := sdk.NewCoin("token", sdk.NewInt(1))

	tcs := []struct {
		desc        string
		expectedErr string
		msg         *types.MsgLiquidStake
	}{
		{
			"happy case",
			"",
			types.NewMsgLiquidStake(delegator.String(), stakingCoin),
		},
		{
			"fail: empty address",
			"invalid delegator address : empty address string is not allowed",
			types.NewMsgLiquidStake("", stakingCoin),
		},
		{
			"fail: zero amount",
			"staking amount must not be zero: invalid request",
			types.NewMsgLiquidStake(delegator.String(), sdk.NewCoin("token", sdk.ZeroInt())),
		},
		{
			"fail: minus amount",
			"negative coin amount: -1",
			types.NewMsgLiquidStake(delegator.String(), sdk.Coin{
				Denom:  "token",
				Amount: sdk.ZeroInt().Sub(sdk.OneInt()),
			}),
		},
	}

	for _, tc := range tcs {
		suite.Run(tc.desc, func() {
			suite.IsType(&types.MsgLiquidStake{}, tc.msg)
			suite.Equal(types.TypeMsgLiquidStake, tc.msg.Type())
			suite.Equal(types.RouterKey, tc.msg.Route())
			suite.Equal(
				sdk.MustSortJSON(types.ModuleCdc.MustMarshalJSON(tc.msg)),
				tc.msg.GetSignBytes(),
			)

			err := tc.msg.ValidateBasic()
			if tc.expectedErr == "" {
				suite.Nil(err)
				signers := tc.msg.GetSigners()
				suite.Len(signers, 1)
				suite.Equal(tc.msg.GetDelegator(), signers[0])
			} else {
				suite.EqualError(err, tc.expectedErr)
			}
		})
	}
}

func (suite *msgTestSuite) TestMsgLiquidUnstake() {
	delegator := sdk.AccAddress("1")
	stakingCoin := sdk.NewCoin("token", sdk.NewInt(1))

	tcs := []struct {
		desc        string
		expectedErr string
		msg         *types.MsgLiquidUnstake
	}{
		{
			"happy case",
			"",
			types.NewMsgLiquidUnstake(delegator.String(), stakingCoin),
		},
		{
			"fail: empty address",
			"invalid delegator address : empty address string is not allowed",
			types.NewMsgLiquidUnstake("", stakingCoin),
		},
		{
			"fail: zero amount",
			"unstaking amount must not be zero: invalid request",
			types.NewMsgLiquidUnstake(delegator.String(), sdk.NewCoin("token", sdk.ZeroInt())),
		},
		{
			"fail: minus amount",
			"negative coin amount: -1",
			types.NewMsgLiquidUnstake(delegator.String(), sdk.Coin{
				Denom:  "token",
				Amount: sdk.ZeroInt().Sub(sdk.OneInt()),
			}),
		},
	}

	for _, tc := range tcs {
		suite.Run(tc.desc, func() {
			suite.IsType(&types.MsgLiquidUnstake{}, tc.msg)
			suite.Equal(types.TypeMsgLiquidUnstake, tc.msg.Type())
			suite.Equal(types.RouterKey, tc.msg.Route())
			suite.Equal(
				sdk.MustSortJSON(types.ModuleCdc.MustMarshalJSON(tc.msg)),
				tc.msg.GetSignBytes(),
			)

			err := tc.msg.ValidateBasic()
			if tc.expectedErr == "" {
				suite.Nil(err)
				signers := tc.msg.GetSigners()
				suite.Len(signers, 1)
				suite.Equal(tc.msg.GetDelegator(), signers[0])
			} else {
				suite.EqualError(err, tc.expectedErr)
			}
		})
	}
}

func (suite *msgTestSuite) TestMsgProvideInsurance() {
	provider := sdk.AccAddress("1")
	validator := sdk.ValAddress("1")
	stakingCoin := sdk.NewCoin("token", sdk.NewInt(1))
	tenPercent := sdk.NewDecWithPrec(10, 2)

	tcs := []struct {
		desc        string
		expectedErr string
		msg         *types.MsgProvideInsurance
	}{
		{
			"happy case",
			"",
			types.NewMsgProvideInsurance(provider.String(), validator.String(), stakingCoin, tenPercent),
		},
		{
			"fail: empty provider address",
			"invalid provider address : empty address string is not allowed",
			types.NewMsgProvideInsurance("", validator.String(), stakingCoin, tenPercent),
		},
		{
			"fail: empty validator address",
			"invalid validator address : empty address string is not allowed",
			types.NewMsgProvideInsurance(provider.String(), "", stakingCoin, tenPercent),
		},
		{
			"fail: zero amount",
			"collateral amount must not be zero: invalid request",
			types.NewMsgProvideInsurance(provider.String(), validator.String(), sdk.NewCoin("token", sdk.ZeroInt()), tenPercent),
		},
		{
			"fail: minus amount",
			"negative coin amount: -1",
			types.NewMsgProvideInsurance(provider.String(), validator.String(), sdk.Coin{
				Denom:  "token",
				Amount: sdk.ZeroInt().Sub(sdk.OneInt()),
			}, tenPercent),
		},
		{
			"fail: empty rate",
			"fee rate must not be nil",
			types.NewMsgProvideInsurance(provider.String(), validator.String(), stakingCoin, sdk.Dec{}),
		},
	}

	for _, tc := range tcs {
		suite.Run(tc.desc, func() {
			suite.IsType(&types.MsgProvideInsurance{}, tc.msg)
			suite.Equal(types.TypeMsgProvideInsurance, tc.msg.Type())
			suite.Equal(types.RouterKey, tc.msg.Route())
			suite.Equal(
				sdk.MustSortJSON(types.ModuleCdc.MustMarshalJSON(tc.msg)),
				tc.msg.GetSignBytes(),
			)

			err := tc.msg.ValidateBasic()
			if tc.expectedErr == "" {
				suite.Nil(err)
				signers := tc.msg.GetSigners()
				suite.Len(signers, 1)
				suite.Equal(tc.msg.GetProvider(), signers[0])
			} else {
				suite.EqualError(err, tc.expectedErr)
			}
		})
	}
}

func (suite *msgTestSuite) TestMsgCancelProvideInsurance() {
	provider := sdk.AccAddress("1")

	tcs := []struct {
		desc        string
		expectedErr string
		msg         *types.MsgCancelProvideInsurance
	}{
		{
			"happy case",
			"",
			types.NewMsgCancelProvideInsurance(provider.String(), 1),
		},
		{
			"fail: empty provider address",
			"invalid provider address : empty address string is not allowed",
			types.NewMsgCancelProvideInsurance("", 1),
		},
		{
			"fail: invalid insurance id",
			"invalid insurance id: 0: invalid request",
			types.NewMsgCancelProvideInsurance(provider.String(), 0),
		},
	}

	for _, tc := range tcs {
		suite.Run(tc.desc, func() {
			suite.IsType(&types.MsgCancelProvideInsurance{}, tc.msg)
			suite.Equal(types.TypeMsgCancelProvideInsurance, tc.msg.Type())
			suite.Equal(types.RouterKey, tc.msg.Route())
			suite.Equal(
				sdk.MustSortJSON(types.ModuleCdc.MustMarshalJSON(tc.msg)),
				tc.msg.GetSignBytes(),
			)

			err := tc.msg.ValidateBasic()
			if tc.expectedErr == "" {
				suite.Nil(err)
				signers := tc.msg.GetSigners()
				suite.Len(signers, 1)
				suite.Equal(tc.msg.GetProvider(), signers[0])
			} else {
				suite.EqualError(err, tc.expectedErr)
			}
		})
	}
}

func (suite *msgTestSuite) TestMsgDepositInsurance() {
	provider := sdk.AccAddress("1")
	insuranceId := uint64(1)
	amount := sdk.NewCoin("token", sdk.NewInt(1))

	tcs := []struct {
		desc        string
		expectedErr string
		msg         *types.MsgDepositInsurance
	}{
		{
			"happy case",
			"",
			types.NewMsgDepositInsurance(provider.String(), insuranceId, amount),
		},
		{
			"fail: empty provider address",
			"invalid provider address : empty address string is not allowed",
			types.NewMsgDepositInsurance("", insuranceId, amount),
		},
		{
			"fail: invalid insurance id",
			"invalid insurance id: 0: invalid request",
			types.NewMsgDepositInsurance(provider.String(), 0, amount),
		},
		{
			"fail: zero amount",
			"deposit amount must not be zero: invalid request",
			types.NewMsgDepositInsurance(provider.String(), insuranceId, sdk.NewCoin("token", sdk.ZeroInt())),
		},
		{
			"fail: minus amount",
			"negative coin amount: -1",
			types.NewMsgDepositInsurance(provider.String(), insuranceId, sdk.Coin{
				Denom:  "token",
				Amount: sdk.ZeroInt().Sub(sdk.OneInt()),
			}),
		},
	}

	for _, tc := range tcs {
		suite.Run(tc.desc, func() {
			suite.IsType(&types.MsgDepositInsurance{}, tc.msg)
			suite.Equal(types.TypeMsgDepositInsurance, tc.msg.Type())
			suite.Equal(types.RouterKey, tc.msg.Route())
			suite.Equal(
				sdk.MustSortJSON(types.ModuleCdc.MustMarshalJSON(tc.msg)),
				tc.msg.GetSignBytes(),
			)

			err := tc.msg.ValidateBasic()
			if tc.expectedErr == "" {
				suite.Nil(err)
				signers := tc.msg.GetSigners()
				suite.Len(signers, 1)
				suite.Equal(tc.msg.GetProvider(), signers[0])
			} else {
				suite.EqualError(err, tc.expectedErr)
			}
		})
	}
}

func (suite *msgTestSuite) TestMsgWithdrawInsurance() {
	provider := sdk.AccAddress("1")
	insuranceId := uint64(1)

	tcs := []struct {
		desc        string
		expectedErr string
		msg         *types.MsgWithdrawInsurance
	}{
		{
			"happy case",
			"",
			types.NewMsgWithdrawInsurance(provider.String(), insuranceId),
		},
		{
			"fail: empty provider address",
			"invalid provider address : empty address string is not allowed",
			types.NewMsgWithdrawInsurance("", insuranceId),
		},
		{
			"fail: invalid insurance id",
			"invalid insurance id: 0: invalid request",
			types.NewMsgWithdrawInsurance(provider.String(), 0),
		},
	}

	for _, tc := range tcs {
		suite.Run(tc.desc, func() {
			suite.IsType(&types.MsgWithdrawInsurance{}, tc.msg)
			suite.Equal(types.TypeMsgWithdrawInsurance, tc.msg.Type())
			suite.Equal(types.RouterKey, tc.msg.Route())
			suite.Equal(
				sdk.MustSortJSON(types.ModuleCdc.MustMarshalJSON(tc.msg)),
				tc.msg.GetSignBytes(),
			)

			err := tc.msg.ValidateBasic()
			if tc.expectedErr == "" {
				suite.Nil(err)
				signers := tc.msg.GetSigners()
				suite.Len(signers, 1)
				suite.Equal(tc.msg.GetProvider(), signers[0])
			} else {
				suite.EqualError(err, tc.expectedErr)
			}
		})
	}
}

func (suite *msgTestSuite) TestMsgWithdrawInsuranceCommission() {
	provider := sdk.AccAddress("1")
	insuranceId := uint64(1)

	tcs := []struct {
		desc        string
		expectedErr string
		msg         *types.MsgWithdrawInsuranceCommission
	}{
		{
			"happy case",
			"",
			types.NewMsgWithdrawInsuranceCommission(provider.String(), insuranceId),
		},
		{
			"fail: empty provider address",
			"invalid provider address : empty address string is not allowed",
			types.NewMsgWithdrawInsuranceCommission("", insuranceId),
		},
		{
			"fail: invalid insurance id",
			"invalid insurance id: 0: invalid request",
			types.NewMsgWithdrawInsuranceCommission(provider.String(), 0),
		},
	}

	for _, tc := range tcs {
		suite.Run(tc.desc, func() {
			suite.IsType(&types.MsgWithdrawInsuranceCommission{}, tc.msg)
			suite.Equal(types.TypeMsgWithdrawInsuranceCommission, tc.msg.Type())
			suite.Equal(types.RouterKey, tc.msg.Route())
			suite.Equal(
				sdk.MustSortJSON(types.ModuleCdc.MustMarshalJSON(tc.msg)),
				tc.msg.GetSignBytes(),
			)

			err := tc.msg.ValidateBasic()
			if tc.expectedErr == "" {
				suite.Nil(err)
				signers := tc.msg.GetSigners()
				suite.Len(signers, 1)
				suite.Equal(tc.msg.GetProvider(), signers[0])
			} else {
				suite.EqualError(err, tc.expectedErr)
			}
		})
	}
}

func (suite *msgTestSuite) TestMsgClaimDiscountedReward() {
	requester := sdk.AccAddress("1")
	amount := sdk.NewCoin("token", sdk.NewInt(1))
	tenPercent := sdk.NewDecWithPrec(10, 2)

	tcs := []struct {
		desc        string
		expectedErr string
		msg         *types.MsgClaimDiscountedReward
	}{
		{
			"happy case",
			"",
			types.NewMsgClaimDiscountedReward(requester.String(), amount, tenPercent),
		},
		{
			"fail: empty requester address",
			"invalid requester address : empty address string is not allowed",
			types.NewMsgClaimDiscountedReward("", amount, tenPercent),
		},
		{
			"fail: zero amount",
			"maximum allowed ls tokens to pay must not be zero: invalid request",
			types.NewMsgClaimDiscountedReward(requester.String(), sdk.NewCoin("token", sdk.ZeroInt()), tenPercent),
		},
		{
			"fail: minus amount",
			"negative coin amount: -1",
			types.NewMsgClaimDiscountedReward(requester.String(), sdk.Coin{
				Denom:  "token",
				Amount: sdk.ZeroInt().Sub(sdk.OneInt()),
			}, tenPercent),
		},
		{
			"fail: minus discount rate",
			"minimum discount rate must not be negative: invalid request",
			types.NewMsgClaimDiscountedReward(requester.String(), amount, sdk.NewDec(-1)),
		},
		{
			"fail: empty rate",
			"minimum discount rate must not be nil: invalid request",
			types.NewMsgClaimDiscountedReward(requester.String(), amount, sdk.Dec{}),
		},
	}

	for _, tc := range tcs {
		suite.Run(tc.desc, func() {
			suite.IsType(&types.MsgClaimDiscountedReward{}, tc.msg)
			suite.Equal(types.TypeMsgClaimDiscountedReward, tc.msg.Type())
			suite.Equal(types.RouterKey, tc.msg.Route())
			suite.Equal(
				sdk.MustSortJSON(types.ModuleCdc.MustMarshalJSON(tc.msg)),
				tc.msg.GetSignBytes(),
			)

			err := tc.msg.ValidateBasic()
			if tc.expectedErr == "" {
				suite.Nil(err)
				signers := tc.msg.GetSigners()
				suite.Len(signers, 1)
				suite.Equal(tc.msg.GetRequestser(), signers[0])
			} else {
				suite.EqualError(err, tc.expectedErr)
			}
		})
	}
}
