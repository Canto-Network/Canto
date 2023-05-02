package types

func NewWithdrawInsuranceRequest(
	insuranceId uint64,
) WithdrawInsuranceRequest {
	return WithdrawInsuranceRequest{
		InsuranceId: insuranceId,
	}
}

func (wir *WithdrawInsuranceRequest) Validate(insuranceMap map[uint64]Insurance) error {
	insurance, ok := insuranceMap[wir.InsuranceId]
	if !ok {
		return ErrNotFoundWithdrawInsuranceRequestInsuranceId
	}
	if insurance.Status != INSURANCE_STATUS_PAIRED {
		return ErrInvalidInsuranceStatus
	}
	return nil
}
