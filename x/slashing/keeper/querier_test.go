package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	simapp "github.com/iqlusioninc/liquidity-staking-module/app"
	"github.com/iqlusioninc/liquidity-staking-module/x/slashing/keeper"
	"github.com/iqlusioninc/liquidity-staking-module/x/slashing/testslashing"
	"github.com/iqlusioninc/liquidity-staking-module/x/slashing/types"
)

func TestNewQuerier(t *testing.T) {
	app := simapp.Setup(t, false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	app.SlashingKeeper.SetParams(ctx, testslashing.TestParams())
	legacyQuerierCdc := codec.NewAminoCodec(app.LegacyAmino())
	querier := keeper.NewQuerier(app.SlashingKeeper, legacyQuerierCdc.LegacyAmino)

	query := abci.RequestQuery{
		Path: "",
		Data: []byte{},
	}

	_, err := querier(ctx, []string{types.QueryParameters}, query)
	require.NoError(t, err)
}

func TestQueryParams(t *testing.T) {
	cdc := codec.NewLegacyAmino()
	legacyQuerierCdc := codec.NewAminoCodec(cdc)
	app := simapp.Setup(t, false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	app.SlashingKeeper.SetParams(ctx, testslashing.TestParams())

	querier := keeper.NewQuerier(app.SlashingKeeper, legacyQuerierCdc.LegacyAmino)

	query := abci.RequestQuery{
		Path: "",
		Data: []byte{},
	}

	var params types.Params

	res, err := querier(ctx, []string{types.QueryParameters}, query)
	require.NoError(t, err)

	err = cdc.UnmarshalJSON(res, &params)
	require.NoError(t, err)
	require.Equal(t, app.SlashingKeeper.GetParams(ctx), params)
}
