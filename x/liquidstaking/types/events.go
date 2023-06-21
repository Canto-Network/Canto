package types

const (
	EventTypeMsgLiquidStake                 = TypeMsgLiquidStake
	EventTypeMsgLiquidUnstake               = TypeMsgLiquidUnstake
	EventTypeMsgProvideInsurance            = TypeMsgProvideInsurance
	EventTypeMsgCancelProvideInsurance      = TypeMsgCancelProvideInsurance
	EventTypeMsgDepositInsurance            = TypeMsgDepositInsurance
	EventTypeMsgWithdrawInsurance           = TypeMsgWithdrawInsurance
	EventTypeMsgWithdrawInsuranceCommission = TypeMsgWithdrawInsuranceCommission
	EventTypeMsgClaimDiscountedReward       = TypeMsgClaimDiscountedReward

	AttributeKeyChunkIds                       = "chunk_ids"
	AttributeKeyInsuranceId                    = "insurance_id"
	AttributeKeyDelegator                      = "delegator"
	AttributeKeyRequester                      = "requester"
	AttributeKeyInsuranceProvider              = "insurance_provider"
	AttributeKeyWithdrawnInsuranceCommission   = "withdrawn_insurance_commission"
	AttributeKeyNewShares                      = "new_shares"
	AttributeKeyLsTokenMintedAmount            = "lstoken_minted_amount"
	AttributeKeyEscrowedLsTokens               = "escrowed_lstokens"
	AttributeKeyClaimTokens                    = "claim_tokens"
	AttributeKeyDiscountedMintRate             = "discounted_mint_rate"
	AttributeKeyWithdrawInsuranceRequestQueued = "withdraw_insurance_request_queued"

	AttributeValueCategory = ModuleName
)
