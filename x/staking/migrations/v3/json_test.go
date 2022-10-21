package v3_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v2 "github.com/iqlusioninc/liquidity-staking-module/x/staking/migrations/v2"
	v3 "github.com/iqlusioninc/liquidity-staking-module/x/staking/migrations/v3"
	stakingtypes "github.com/iqlusioninc/liquidity-staking-module/x/staking/types"
)

func TestMigrateJSON(t *testing.T) {
	encodingConfig := simapp.MakeTestEncodingConfig()
	clientCtx := client.Context{}.
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithTxConfig(encodingConfig.TxConfig).
		WithCodec(encodingConfig.Codec)

	val, oldValidator := genOldValidator(t)
	oldState := v2.GenesisState{
		Params:         v2.Params{},
		LastTotalPower: sdk.NewInt(1),
		Validators:     []v2.Validator{oldValidator},
		Delegations: []v2.Delegation{
			{
				DelegatorAddress: val.Address.String(),
				ValidatorAddress: sdk.ValAddress(val.Address).String(),
				Shares:           sdk.NewDec(1),
			},
		},
		LastValidatorPowers:  []stakingtypes.LastValidatorPower{},
		UnbondingDelegations: []stakingtypes.UnbondingDelegation{},
		Redelegations:        []stakingtypes.Redelegation{},
		Exported:             true,
	}

	newState, err := v3.MigrateJSON(oldState)
	require.NoError(t, err)

	bz, err := clientCtx.Codec.MarshalJSON(&newState)
	require.NoError(t, err)

	// Indent the JSON bz correctly.
	var jsonObj map[string]interface{}
	err = json.Unmarshal(bz, &jsonObj)
	require.NoError(t, err)
	indentedBz, err := json.MarshalIndent(jsonObj, "", "\t")
	require.NoError(t, err)

	fmt.Println("indentedBz", string(indentedBz))
	// Make sure about new param MinCommissionRate.
	expected := `{
	"delegations": [
		{
			"delegator_address": "AE99B794C105768EB8A0FB4682CB6D4DB9F2EF5B",
			"exempt": false,
			"shares": "1.000000000000000000",
			"validator_address": "cosmosvaloper146vm09xpq4mgaw9qldrg9jmdfkul9m6mfm9gc9"
		}
	],
	"exported": true,
	"last_tokenize_share_record_id": "0",
	"last_total_power": "1",
	"last_validator_powers": [],
	"params": {
		"bond_denom": "",
		"exemption_factor": "-1.000000000000000000",
		"historical_entries": 0,
		"max_entries": 0,
		"max_validators": 0,
		"min_commission_rate": "0.000000000000000000",
		"unbonding_time": "0s"
	},
	"redelegations": [],
	"tokenize_share_records": [],
	"unbonding_delegations": [],
	"validators": [
		{
			"commission": {
				"commission_rates": {
					"max_change_rate": "0.000000000000000000",
					"max_rate": "0.000000000000000000",
					"rate": "0.000000000000000000"
				},
				"update_time": "1970-01-01T00:00:00Z"
			},
			"consensus_pubkey": {
				"@type": "/cosmos.crypto.ed25519.PubKey",
				"key": "k863wifbfIjCxEDhhEa0VF4R5Qcpm9byH1QjTDSKBSc="
			},
			"delegator_shares": "1.000000000000000000",
			"description": {
				"details": "",
				"identity": "",
				"moniker": "",
				"security_contact": "",
				"website": ""
			},
			"jailed": false,
			"operator_address": "cosmosvaloper146vm09xpq4mgaw9qldrg9jmdfkul9m6mfm9gc9",
			"status": "BOND_STATUS_BONDED",
			"tokens": "1000000",
			"total_exempt_shares": "0.000000000000000000",
			"total_tokenized_shares": "0.000000000000000000",
			"unbonding_height": "0",
			"unbonding_time": "1970-01-01T00:00:00Z"
		}
	]
}`
	require.Equal(t, expected, string(indentedBz))
}
