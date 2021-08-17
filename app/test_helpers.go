package app

import (
	"encoding/json"
	"fmt"

	bam "github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankcli "github.com/cosmos/cosmos-sdk/x/bank/client/cli"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"
	"github.com/tendermint/tmlibs/cli"
)

func setup(withGenesis bool, invCheckPeriod uint) (*LiquidStaking, GenesisState) {
	db := dbm.NewMemDB()
	encCdc := MakeEncodingConfig()
	app := New(log.NewNopLogger(), db, nil, true, map[int64]bool{}, DefaultNodeHome, invCheckPeriod, encCdc, simapp.EmptyAppOptions{})
	if withGenesis {
		return app, NewDefaultGenesisState(encCdc.Marshaler)
	}
	return app, GenesisState{}
}

// Setup initializes a new SimApp. A Nop logger is set in SimApp.
func Setup(isCheckTx bool) *LiquidStaking {
	app, genesisState := setup(!isCheckTx, 5)
	if !isCheckTx {
		// init chain must be called to stop deliverState from being nil
		stateBytes, err := json.MarshalIndent(genesisState, "", " ")
		if err != nil {
			panic(err)
		}

		// Initialize the chain
		app.InitChain(
			abci.RequestInitChain{
				Validators:      []abci.ValidatorUpdate{},
				ConsensusParams: simapp.DefaultConsensusParams,
				AppStateBytes:   stateBytes,
			},
		)
	}

	return app
}

func SimAppConstructor(val network.Validator) servertypes.Application {
	return New(
		val.Ctx.Logger, dbm.NewMemDB(), nil, true, make(map[int64]bool),
		val.Ctx.Config.RootDir, 0, MakeEncodingConfig(), simapp.EmptyAppOptions{},
		bam.SetPruning(storetypes.NewPruningOptionsFromString(val.AppConfig.Pruning)),
		bam.SetMinGasPrices(val.AppConfig.MinGasPrices),
	)
}

func NewConfig() network.Config {
	cfg := network.DefaultConfig()
	encCfg := MakeEncodingConfig()
	cfg.Codec = encCfg.Marshaler
	cfg.TxConfig = encCfg.TxConfig
	cfg.LegacyAmino = encCfg.Amino
	cfg.InterfaceRegistry = encCfg.InterfaceRegistry
	cfg.AppConstructor = SimAppConstructor
	cfg.GenesisState = NewDefaultGenesisState(cfg.Codec)
	cfg.BondDenom = sdk.DefaultBondDenom
	cfg.MinGasPrices = fmt.Sprintf("0.000006%s", sdk.DefaultBondDenom)
	return cfg
}

func QueryBalanceExec(clientCtx client.Context, address string, denom string, extraArgs ...string) (testutil.BufferWriter, error) {
	args := []string{
		address,
		fmt.Sprintf("--%s=%s", bankcli.FlagDenom, denom),
		fmt.Sprintf("--%s=json", cli.OutputFlag),
	}
	args = append(args, extraArgs...)

	return clitestutil.ExecTestCLICmd(clientCtx, bankcli.GetBalancesCmd(), args)
}
