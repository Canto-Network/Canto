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
	EventTypeBeginLiquidUnstake             = "begin_liquid_unstake"
	EventTypeDeleteQueuedLiquidUnstake      = "delete_queued_liquid_unstake"
	EventTypeBeginWithdrawInsurance         = "begin_withdraw_insurance"
	EventTypeBeginUndelegate                = "begin_undelegate"
	EventTypeRePairedWithNewInsurance       = "re_paired_with_new_insurance"
	EventTypeBeginRedelegate                = "begin_redelegate"

	AttributeKeyChunkId                        = "chunk_id"
	AttributeKeyChunkIds                       = "chunk_ids"
	AttributeKeyInsuranceId                    = "insurance_id"
	AttributeKeyInsuranceIds                   = "insurance_ids"
	AttributeKeyNewInsuranceId                 = "new_insurance_id"
	AttributeKeyOutInsuranceId                 = "out_insurance_id"
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
	AttributeKeyReason                         = "reason"

	AttributeValueCategory                         = ModuleName
	AttributeValueReasonNotEnoughInsuranceCoverage = "not_enough_insurance_coverage"
	AttributeValueReasonInsuranceCoverPenalty      = "insurance_cover_penalty"
	AttributeValueReasonPairingChunkPaired         = "pairing_chunk_paired"
	AttributeValueReasonNoCandidateInsurance       = "no_candidate_insurance"
)
