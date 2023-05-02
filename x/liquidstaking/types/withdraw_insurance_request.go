package types

func NewWithdrawInsuranceRequest(
	insuranceId uint64,
) WithdrawInsuranceRequest {
	return WithdrawInsuranceRequest{
		InsuranceId: insuranceId,
	}
}
