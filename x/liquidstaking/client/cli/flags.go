package cli

import (
	flag "github.com/spf13/pflag"
)

const (
	FlagChunkStatus      = "status"
	FlagInsuranceStatus  = "status"
	FlagValidatorAddress = "validator-address"
	FlagProviderAddress  = "provider-address"
)

func flagSetChunks() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.String(FlagChunkStatus, "", "The chunk status")

	return fs
}

func flagSetInsurances() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.String(FlagInsuranceStatus, "", "The insurance status")
	fs.String(FlagValidatorAddress, "", "The bech-32 encoded address of the validator")
	fs.String(FlagProviderAddress, "", "The bech-32 encoded address of the provider")

	return fs
}
