package distribution

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/iqlusioninc/liquidity-staking-module/x/distribution/keeper"
	"github.com/iqlusioninc/liquidity-staking-module/x/distribution/types"
	govtypes "github.com/iqlusioninc/liquidity-staking-module/x/gov/types"
)

func NewCommunityPoolSpendProposalHandler(k keeper.Keeper) govtypes.Handler {
	return func(ctx sdk.Context, content govtypes.Content) error {
		switch c := content.(type) {
		case *types.CommunityPoolSpendProposal:
			return keeper.HandleCommunityPoolSpendProposal(ctx, k, c)

		default:
			return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized distr proposal content type: %T", c)
		}
	}
}
