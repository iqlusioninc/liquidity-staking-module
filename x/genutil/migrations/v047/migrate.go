package v046

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/x/genutil/types"
	v046staking "github.com/cosmos/cosmos-sdk/x/staking/types"
	v2 "github.com/iqlusioninc/liquidity-staking-module/x/staking/migrations/v2"
	v3 "github.com/iqlusioninc/liquidity-staking-module/x/staking/migrations/v3"
)

// Migrate migrates exported state from v0.46 to a v0.47 genesis state.
func Migrate(appState types.AppMap, clientCtx client.Context) types.AppMap {
	// Migrate x/staking.
	if appState[v2.ModuleName] != nil {
		// unmarshal relative source genesis application state
		var old v2.GenesisState
		clientCtx.Codec.MustUnmarshalJSON(appState[v2.ModuleName], &old)

		// delete deprecated x/staking genesis state
		delete(appState, v2.ModuleName)

		// Migrate relative source genesis application state and marshal it into
		// the respective key.
		new, err := v3.MigrateJSON(old)
		if err != nil {
			panic(err)
		}
		appState[v046staking.ModuleName] = clientCtx.Codec.MustMarshalJSON(&new)
	}

	return appState
}
