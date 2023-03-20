package distribution_test

import (
	"testing"

	dbm "github.com/cometbft/cometbft-db"
	abcitypes "github.com/cometbft/cometbft/abci/types"
	tmjson "github.com/cometbft/cometbft/libs/json"
	"github.com/cometbft/cometbft/libs/log"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/require"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	simapp "github.com/iqlusioninc/liquidity-staking-module/app"
	"github.com/iqlusioninc/liquidity-staking-module/x/distribution/types"
)

func TestItCreatesModuleAccountOnInitBlock(t *testing.T) {
	db := dbm.NewMemDB()
	encCdc := simapp.MakeTestEncodingConfig()
	app := simapp.NewSimApp(log.NewNopLogger(), db, nil, true, map[int64]bool{}, simapp.DefaultNodeHome, 5, encCdc, simapp.EmptyAppOptions{})

	genesisState := simapp.GenesisStateWithSingleValidator(t, app)
	stateBytes, err := tmjson.Marshal(genesisState)
	require.NoError(t, err)

	app.InitChain(
		abcitypes.RequestInitChain{
			AppStateBytes: stateBytes,
			ChainId:       "test-chain-id",
		},
	)

	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	acc := app.AccountKeeper.GetAccount(ctx, authtypes.NewModuleAddress(types.ModuleName))
	require.NotNil(t, acc)
}
