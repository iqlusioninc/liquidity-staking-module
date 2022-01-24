package distribution_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	simapp "github.com/iqlusioninc/liquidity-staking-module/app"
	"github.com/iqlusioninc/liquidity-staking-module/x/distribution"
	"github.com/iqlusioninc/liquidity-staking-module/x/distribution/types"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

// test msg registration
func TestWithdrawTokenizeShareRecordReward(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	h := distribution.NewHandler(app.DistrKeeper)
	delAddr1 = sdk.AccAddress(delPk1.Address())

	res, err := h(ctx, &types.MsgWithdrawTokenizeShareRecordReward{
		OwnerAddress: delAddr1.String(),
	})
	require.NoError(t, err)
	require.NotNil(t, res)
}
