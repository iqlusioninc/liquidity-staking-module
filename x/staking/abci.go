package staking

import (
	"time"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/iqlusioninc/liquidity-staking-module/x/staking/keeper"
	"github.com/iqlusioninc/liquidity-staking-module/x/staking/types"
)

// BeginBlocker will persist the current header and validator set as a historical entry
// and prune the oldest entry based on the HistoricalEntries parameter
func BeginBlocker(ctx sdk.Context, k keeper.Keeper) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)

	k.TrackHistoricalInfo(ctx)
}

// Called every block, update validator set
func EndBlocker(ctx sdk.Context, k keeper.Keeper) []abci.ValidatorUpdate {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)
	EpochInterval := k.GetParams(ctx).EpochInterval
	if ctx.BlockHeight()%EpochInterval == 0 {
		k.ExecuteEpoch(ctx)
	}

	// run block validator updates for slashed, jailed validators
	return k.BlockValidatorUpdates(ctx)
}