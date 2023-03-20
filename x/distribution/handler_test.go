package distribution_test

import (
	"testing"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simapp "github.com/iqlusioninc/liquidity-staking-module/app"
	"github.com/iqlusioninc/liquidity-staking-module/x/distribution"
	"github.com/iqlusioninc/liquidity-staking-module/x/distribution/types"
	"github.com/stretchr/testify/require"
)

// test msg registration
func TestWithdrawTokenizeShareRecordReward(t *testing.T) {
	app := simapp.Setup(t, false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	h := distribution.NewHandler(app.DistrKeeper)
	delAddr1 = sdk.AccAddress(delPk1.Address())

	res, err := h(ctx, &types.MsgWithdrawAllTokenizeShareRecordReward{
		OwnerAddress: delAddr1.String(),
	})
	require.NoError(t, err)
	require.NotNil(t, res)
}
