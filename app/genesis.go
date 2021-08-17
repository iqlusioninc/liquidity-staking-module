package app

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// The genesis state of the blockchain is represented here as a map of raw json
// messages key'd by a identifier string.
// The identifier is used to determine which module genesis information belongs
// to so it may be appropriately routed during init chain.
// Within this application default genesis information is retrieved from
// the ModuleBasicManager which populates json from each BasicModule
// object provided to it during init.
type GenesisState map[string]json.RawMessage

// NewDefaultGenesisState generates the default state for the application.
func NewDefaultGenesisState(cdc codec.JSONMarshaler) GenesisState {
	genesis := ModuleBasics.DefaultGenesis(cdc)
	mintGenesis := mintGenesisState()
	stakingGenesis := stakingGenesisState()
	govGenesis := govGenesisState()

	genesis["mint"] = cdc.MustMarshalJSON(mintGenesis)
	genesis["staking"] = cdc.MustMarshalJSON(stakingGenesis)
	genesis["gov"] = cdc.MustMarshalJSON(govGenesis)

	return genesis
}

// stakingGenesisState returns the default genesis state for the staking module, replacing the
// bond denom from stake to ubtsg
func stakingGenesisState() *stakingtypes.GenesisState {
	return &stakingtypes.GenesisState{
		Params: stakingtypes.NewParams(
			stakingtypes.DefaultUnbondingTime,
			stakingtypes.DefaultMaxValidators,
			stakingtypes.DefaultMaxEntries,
			0,
			sdk.DefaultBondDenom,
		),
	}
}

func govGenesisState() *govtypes.GenesisState {
	return govtypes.NewGenesisState(
		1,
		govtypes.NewDepositParams(
			sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, govtypes.DefaultMinDepositTokens)),
			govtypes.DefaultPeriod,
		),
		govtypes.NewVotingParams(govtypes.DefaultPeriod),
		govtypes.NewTallyParams(govtypes.DefaultQuorum, govtypes.DefaultThreshold, govtypes.DefaultVetoThreshold),
	)
}

func mintGenesisState() *minttypes.GenesisState {
	return &minttypes.GenesisState{
		Params: minttypes.NewParams(
			sdk.DefaultBondDenom,
			sdk.NewDecWithPrec(13, 2),
			sdk.NewDecWithPrec(20, 2),
			sdk.NewDecWithPrec(7, 2),
			sdk.NewDecWithPrec(67, 2),
			uint64(60*60*8766/5), // assuming 5 second block times
		),
	}
}
