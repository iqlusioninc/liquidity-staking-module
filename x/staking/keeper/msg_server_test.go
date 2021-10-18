package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	simapp "github.com/iqlusioninc/liquidity-staking-module/app"
	"github.com/iqlusioninc/liquidity-staking-module/x/staking/keeper"
	"github.com/iqlusioninc/liquidity-staking-module/x/staking/teststaking"
	"github.com/iqlusioninc/liquidity-staking-module/x/staking/types"
)

// TODO: modify test for modified TokenizeShares and Redeem process
func TestTokenizeSharesAndRedeemTokens(t *testing.T) {
	_, app, ctx := createTestInput(t)

	addrs := simapp.AddTestAddrs(app, ctx, 2, app.StakingKeeper.TokensFromConsensusPower(ctx, 10000))
	addrAcc1, addrAcc2 := addrs[0], addrs[1]
	addrVal1, addrVal2 := sdk.ValAddress(addrAcc1), sdk.ValAddress(addrAcc2)

	pubKeys := simapp.CreateTestPubKeys(2)
	pk1, pk2 := pubKeys[0], pubKeys[1]

	// Create Validators and Delegation
	val1 := teststaking.NewValidator(t, addrVal1, pk1)
	app.StakingKeeper.SetValidator(ctx, val1)
	app.StakingKeeper.SetValidatorByPowerIndex(ctx, val1)

	val2 := teststaking.NewValidator(t, addrVal2, pk2)
	app.StakingKeeper.SetValidator(ctx, val2)
	app.StakingKeeper.SetValidatorByPowerIndex(ctx, val2)

	delTokens := app.StakingKeeper.TokensFromConsensusPower(ctx, 20)
	err := delegateCoinsFromAccount(ctx, app, addrAcc2, delTokens, val1)
	require.NoError(t, err)

	// apply TM updates
	applyValidatorSetUpdates(t, ctx, app.StakingKeeper, -1)

	_, found := app.StakingKeeper.GetDelegation(ctx, addrAcc2, addrVal1)
	require.True(t, found, "delegation not found after delegate")

	msgServer := keeper.NewMsgServerImpl(app.StakingKeeper)
	resp, err := msgServer.TokenizeShares(sdk.WrapSDKContext(ctx), &types.MsgTokenizeShares{
		DelegatorAddress: addrAcc2.String(),
		ValidatorAddress: addrVal1.String(),
		Amount:           sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), delTokens),
	})
	require.NoError(t, err)
	_, found = app.StakingKeeper.GetDelegation(ctx, addrAcc2, addrVal1)
	require.False(t, found, "delegation found after tokenize share")
	shareToken := app.BankKeeper.GetBalance(ctx, addrAcc2, resp.Amount.Denom)
	require.Equal(t, resp.Amount, shareToken)
	validator, found := app.StakingKeeper.GetValidator(ctx, addrVal1)
	require.True(t, found, true, "validator not found")
	require.Equal(t, validator.ShareTokens, resp.Amount.Amount)

	msgServer.RedeemTokens(sdk.WrapSDKContext(ctx), &types.MsgRedeemTokensforShares{
		DelegatorAddress: addrAcc2.String(),
		Amount:           resp.Amount,
	})
	require.NoError(t, err)
	delegation, found := app.StakingKeeper.GetDelegation(ctx, addrAcc2, addrVal1)
	require.True(t, found, "delegation not found after redeem tokens")
	require.Equal(t, delegation.DelegatorAddress, addrAcc2.String())
	require.Equal(t, delegation.ValidatorAddress, addrVal1.String())
	require.Equal(t, delegation.Shares, delTokens.ToDec())
	shareToken = app.BankKeeper.GetBalance(ctx, addrAcc2, resp.Amount.Denom)
	require.Equal(t, shareToken.Amount, sdk.ZeroInt())
	validator, found = app.StakingKeeper.GetValidator(ctx, addrVal1)
	require.True(t, found, true, "validator not found")
	require.Equal(t, validator.ShareTokens, sdk.ZeroInt())
}

func TestTransferTokenizeShareRecord(t *testing.T) {
	_, app, ctx := createTestInput(t)

	addrs := simapp.AddTestAddrs(app, ctx, 3, app.StakingKeeper.TokensFromConsensusPower(ctx, 10000))
	addrAcc1, addrAcc2, valAcc := addrs[0], addrs[1], addrs[2]
	addrVal := sdk.ValAddress(valAcc)

	pubKeys := simapp.CreateTestPubKeys(1)
	pk := pubKeys[0]

	val := teststaking.NewValidator(t, addrVal, pk)
	app.StakingKeeper.SetValidator(ctx, val)
	app.StakingKeeper.SetValidatorByPowerIndex(ctx, val)

	// apply TM updates
	applyValidatorSetUpdates(t, ctx, app.StakingKeeper, -1)

	msgServer := keeper.NewMsgServerImpl(app.StakingKeeper)

	err := app.StakingKeeper.AddTokenizeShareRecord(ctx, types.TokenizeShareRecord{
		Id:              1,
		Owner:           addrAcc1.String(),
		ShareTokenDenom: "share_token_denom",
		ModuleAccount:   "module_account",
		Validator:       val.String(),
	})
	require.NoError(t, err)

	_, err = msgServer.TransferTokenizeShareRecord(sdk.WrapSDKContext(ctx), &types.MsgTransferTokenizeShareRecord{
		TokenizeShareRecordId: 1,
		Sender:                addrAcc1.String(),
		NewOwner:              addrAcc2.String(),
	})
	require.NoError(t, err)

	record, err := app.StakingKeeper.GetTokenizeShareRecord(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, record.Owner, addrAcc2.String())
}
