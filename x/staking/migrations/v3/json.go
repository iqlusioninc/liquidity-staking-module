package v3

import (
	v2 "github.com/iqlusioninc/liquidity-staking-module/x/staking/migrations/v2"
	stakingtypes "github.com/iqlusioninc/liquidity-staking-module/x/staking/types"
)

// MigrateJSON accepts exported v0.46 x/stakinng genesis state and migrates it to
// v0.47 x/staking genesis state. The migration includes:
//
// - Add MinCommissionRate param.
func MigrateJSON(oldState v2.GenesisState) (stakingtypes.GenesisState, error) {
	oldState.Params.MinCommissionRate = stakingtypes.DefaultMinCommissionRate

	validators := []stakingtypes.Validator{}
	for _, val := range oldState.Validators {
		validators = append(validators, getMigratedValidator(val))
	}

	delegations := []stakingtypes.Delegation{}
	for _, del := range oldState.Delegations {
		delegations = append(delegations, getMigratedDelegation(del))
	}

	return stakingtypes.GenesisState{
		Params: stakingtypes.Params{
			UnbondingTime:     oldState.Params.UnbondingTime,
			MaxValidators:     oldState.Params.MaxValidators,
			MaxEntries:        oldState.Params.MaxEntries,
			HistoricalEntries: oldState.Params.HistoricalEntries,
			BondDenom:         oldState.Params.BondDenom,
			MinCommissionRate: stakingtypes.DefaultMinCommissionRate,
			ExemptionFactor:   stakingtypes.DefaultExemptionFactor,
		},
		LastTotalPower:            oldState.LastTotalPower,
		LastValidatorPowers:       oldState.LastValidatorPowers,
		Validators:                validators,
		Delegations:               delegations,
		UnbondingDelegations:      oldState.UnbondingDelegations,
		Redelegations:             oldState.Redelegations,
		Exported:                  oldState.Exported,
		TokenizeShareRecords:      []stakingtypes.TokenizeShareRecord{},
		LastTokenizeShareRecordId: 0,
	}, nil
}
