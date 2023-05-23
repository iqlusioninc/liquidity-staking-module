package simulation_test

import (
	"math/rand"
	"testing"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	simapp "github.com/iqlusioninc/liquidity-staking-module/app"
	"github.com/iqlusioninc/liquidity-staking-module/x/distribution/simulation"
)

func TestProposalContents(t *testing.T) {
	app := simapp.Setup(t, false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	// initialize parameters
	s := rand.NewSource(1)
	r := rand.New(s)

	accounts := simtypes.RandomAccounts(r, 3)

	// execute ProposalContents function
	weightedProposalContent := simulation.ProposalContents()
	require.Len(t, weightedProposalContent, 1)

	w0 := weightedProposalContent[0]

	// tests w0 interface:
	require.Equal(t, simulation.OpWeightMsgDeposit, w0.AppParamsKey())
	require.Equal(t, simulation.DefaultWeightTextProposal, w0.DefaultWeight())

	amount := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1)), sdk.NewCoin("atoken", sdk.NewInt(2)))

	feePool := app.DistrKeeper.GetFeePool(ctx)
	feePool.CommunityPool = sdk.NewDecCoinsFromCoins(amount...)
	app.DistrKeeper.SetFeePool(ctx, feePool)

	content := w0.ContentSimulatorFn()(r, ctx, accounts)

	require.Equal(t, "fyzeOcbWwNbeHVIkPZBSpYuLyYggwexjxusrBqDOTtGTOWeLrQKjLxzIivHSlcxgdXhhuTSkuxKGLwQvuyNhYFmBZHeAerqyNEUz", content.GetDescription())
	require.Equal(t, "GqiQWIXnku", content.GetTitle())
	require.Equal(t, "gov", content.ProposalRoute())
	require.Equal(t, "Text", content.ProposalType())
}
