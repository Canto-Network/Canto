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
	EventTypeDelegate                       = "delegate"

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
	AttributeKeyCompletionTime                 = "completion_time"
	AttributeKeyValidator                      = "validator"
	AttributeKeySrcValidator                   = "src_validator"
	AttributeKeyDstValidator                   = "dst_validator"

	AttributeValueCategory                                  = ModuleName
	AttributeValueReasonNotEnoughPairedInsCoverage          = "not_enough_paired_insurance_coverage"
	AttributeValueReasonPairedInsBalUnderDoubleSignSlashing = "paired_insurance_coverage_is_under_double_sign_slashing"
	AttributeValueReasonPairedInsCoverPenalty               = "paired_insurance_cover_penalty"
	AttributeValueReasonUnpairingInsCoverPenalty            = "unpairing_insurance_cover_penalty"
	AttributeValueReasonPairingChunkPaired                  = "pairing_chunk_paired"
	AttributeValueReasonNoCandidateIns                      = "no_candidate_insurance"
)
