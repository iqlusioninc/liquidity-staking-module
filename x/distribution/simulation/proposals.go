package simulation

import (
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/iqlusioninc/liquidity-staking-module/x/distribution/keeper"
	"github.com/iqlusioninc/liquidity-staking-module/x/distribution/types"
)

// OpWeightSubmitCommunitySpendProposal app params key for community spend proposal
const OpWeightSubmitCommunitySpendProposal = "op_weight_submit_community_spend_proposal"

// ProposalContents defines the module weighted proposals' contents
func ProposalContents(k keeper.Keeper) []simtypes.WeightedProposalContent {
	return []simtypes.WeightedProposalContent{
		simulation.NewWeightedProposalContent(
			OpWeightSubmitCommunitySpendProposal,
			SimulateCommunityPoolSpendProposalContent(k),
		),
	}
}

// SimulateCommunityPoolSpendProposalContent generates random community-pool-spend proposal content
func SimulateCommunityPoolSpendProposalContent(k keeper.Keeper) simtypes.ContentSimulatorFn {
	return func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) simtypes.Content {
		simAccount, _ := simtypes.RandomAcc(r, accs)

		balance := k.GetFeePool(ctx).CommunityPool
		if balance.Empty() {
			return nil
		}

		denomIndex := r.Intn(len(balance))
		amount, err := simtypes.RandPositiveInt(r, balance[denomIndex].Amount.TruncateInt())
		if err != nil {
			return nil
		}

		return types.NewCommunityPoolSpendProposal(
			simtypes.RandStringOfLength(r, 10),
			simtypes.RandStringOfLength(r, 100),
			simAccount.Address,
			sdk.NewCoins(sdk.NewCoin(balance[denomIndex].Denom, amount)),
		)
	}
}
