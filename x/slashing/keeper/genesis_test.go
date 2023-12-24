package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	simapp "github.com/iqlusioninc/liquidity-staking-module/app"
	"github.com/iqlusioninc/liquidity-staking-module/x/slashing/testslashing"
	"github.com/iqlusioninc/liquidity-staking-module/x/slashing/types"
)

func TestExportAndInitGenesis(t *testing.T) {
	app := simapp.Setup(t, false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	app.SlashingKeeper.SetParams(ctx, testslashing.TestParams())

	addrDels := simapp.AddTestAddrsIncremental(app, ctx, 2, app.StakingKeeper.TokensFromConsensusPower(ctx, 200))

	info1 := types.NewValidatorSigningInfo(sdk.ConsAddress(addrDels[0]), int64(4), int64(3),
		time.Now().UTC().Add(100000000000), false, int64(10))
	info2 := types.NewValidatorSigningInfo(sdk.ConsAddress(addrDels[1]), int64(5), int64(4),
		time.Now().UTC().Add(10000000000), false, int64(10))

	app.SlashingKeeper.SetValidatorSigningInfo(ctx, sdk.ConsAddress(addrDels[0]), info1)
	app.SlashingKeeper.SetValidatorSigningInfo(ctx, sdk.ConsAddress(addrDels[1]), info2)
	genesisState := app.SlashingKeeper.ExportGenesis(ctx)

	require.Equal(t, genesisState.Params, testslashing.TestParams())
	require.Len(t, genesisState.SigningInfos, 2)
	require.Equal(t, genesisState.SigningInfos[0].ValidatorSigningInfo, info1)

	// Tombstone validators after genesis shouldn't effect genesis state
	app.SlashingKeeper.Tombstone(ctx, sdk.ConsAddress(addrDels[0]))
	app.SlashingKeeper.Tombstone(ctx, sdk.ConsAddress(addrDels[1]))

	ok := app.SlashingKeeper.IsTombstoned(ctx, sdk.ConsAddress(addrDels[0]))
	require.True(t, ok)

	newInfo1, _ := app.SlashingKeeper.GetValidatorSigningInfo(ctx, sdk.ConsAddress(addrDels[0]))
	require.NotEqual(t, info1, newInfo1)
	// Initialise genesis with genesis state before tombstone

	app.SlashingKeeper.InitGenesis(ctx, app.StakingKeeper, genesisState)

	// Validator isTombstoned should return false as GenesisState is initialised
	ok = app.SlashingKeeper.IsTombstoned(ctx, sdk.ConsAddress(addrDels[0]))
	require.False(t, ok)

	newInfo1, ok = app.SlashingKeeper.GetValidatorSigningInfo(ctx, sdk.ConsAddress(addrDels[0]))
	require.True(t, ok)
	newInfo2, ok := app.SlashingKeeper.GetValidatorSigningInfo(ctx, sdk.ConsAddress(addrDels[1]))
	require.True(t, ok)
	require.Equal(t, info1, newInfo1)
	require.Equal(t, info2, newInfo2)
}
