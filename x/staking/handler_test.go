package staking_test

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
	simapp "github.com/iqlusioninc/liquidity-staking-module/app"
	"github.com/iqlusioninc/liquidity-staking-module/x/staking/teststaking"
	stakingtypes "github.com/iqlusioninc/liquidity-staking-module/x/staking/types"
)

func bootstrapHandlerGenesisTest(t *testing.T, power int64, numAddrs int, accAmount math.Int) (*simapp.SimApp, sdk.Context, []sdk.AccAddress, []sdk.ValAddress) {
	_, app, ctx := getBaseSimappWithCustomKeeper(t)

	addrDels, addrVals := generateAddresses(app, ctx, numAddrs, accAmount)

	amt := app.StakingKeeper.TokensFromConsensusPower(ctx, power)
	totalSupply := sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), amt.MulRaw(int64(len(addrDels)))))

	notBondedPool := app.StakingKeeper.GetNotBondedPool(ctx)

	// set non bonded pool balance
	app.AccountKeeper.SetModuleAccount(ctx, notBondedPool)
	require.NoError(t, testutil.FundModuleAccount(app.BankKeeper, ctx, notBondedPool.GetName(), totalSupply))
	return app, ctx, addrDels, addrVals
}

func TestTokenizeShares(t *testing.T) {
	initPower := int64(1000)

	testCases := []struct {
		name      string
		delIndex  int64
		valIndex  int64
		amount    math.Int
		isSuccess bool
		expStatus stakingtypes.BondStatus
		expJailed bool
	}{
		{
			"tokenize shares for less than self delegation",
			0, 0,
			sdk.NewInt(10000),
			true,
			stakingtypes.Bonded,
			false,
		},
		{
			"tokenize shares for more than self delegation",
			0, 0,
			sdk.TokensFromConsensusPower(initPower+1, sdk.DefaultPowerReduction),
			false,
			stakingtypes.Bonded,
			false,
		},
		{
			"tokenize share for full self delegation",
			0, 0,
			sdk.TokensFromConsensusPower(50, sdk.DefaultPowerReduction),
			true,
			stakingtypes.Bonded,
			false,
		},
		{
			"tokenize shares for less than delegation",
			1, 0,
			sdk.NewInt(1000),
			true,
			stakingtypes.Bonded,
			false,
		},
		{
			"tokenize shares for more than delegation",
			1, 0,
			sdk.NewInt(20000),
			false,
			stakingtypes.Bonded,
			false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(*testing.T) {
			app, ctx, delAddrs, valAddrs := bootstrapHandlerGenesisTest(t, initPower, 3, sdk.TokensFromConsensusPower(initPower, sdk.DefaultPowerReduction))
			val1 := valAddrs[0]
			del2 := delAddrs[1]
			tstaking := teststaking.NewHelper(t, ctx, app.StakingKeeper)

			// set staking params
			params := app.StakingKeeper.GetAllParams(ctx)
			params.MaxValidators = 2
			app.StakingKeeper.SetParams(ctx, params)

			// add validators
			tstaking.CreateValidatorWithValPower(val1, PKs[0], 50, true)

			// call it to update validator status to bonded
			_, err := app.StakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx)
			require.NoError(t, err)

			// delegate tokens to the validator
			tstaking.Delegate(del2, val1, sdk.NewInt(10000))

			del := delAddrs[tc.delIndex]
			val := valAddrs[tc.valIndex]

			tstaking.TokenizeShares(del, val, sdk.NewCoin(sdk.DefaultBondDenom, tc.amount), del, tc.isSuccess)

			if tc.isSuccess {
				// call it to update validator status automatically
				_, err := app.StakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx)
				require.NoError(t, err)

				tstaking.CheckValidator(val, tc.expStatus, tc.expJailed)
			}
		})
	}
}

func TestRedeemTokensforShares(t *testing.T) {
	initPower := int64(1000)

	testCases := []struct {
		name      string
		amount    math.Int
		isSuccess bool
	}{
		{
			"redeem full shares",
			sdk.NewInt(10000),
			true,
		},
		{
			"redeem partial shares",
			sdk.NewInt(1000),
			true,
		},
		{
			"redeem zero shares",
			sdk.NewInt(0),
			false,
		},
		{
			"redeem more than shares",
			sdk.NewInt(20000),
			false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(*testing.T) {
			app, ctx, delAddrs, valAddrs := bootstrapHandlerGenesisTest(t, initPower, 3, sdk.TokensFromConsensusPower(initPower, sdk.DefaultPowerReduction))
			val1 := valAddrs[0]
			del2 := delAddrs[1]
			tstaking := teststaking.NewHelper(t, ctx, app.StakingKeeper)

			// set staking params
			params := app.StakingKeeper.GetAllParams(ctx)
			params.MaxValidators = 2
			app.StakingKeeper.SetParams(ctx, params)

			// add validators
			tstaking.CreateValidatorWithValPower(val1, PKs[0], 50, true)

			// call it to update validator status to bonded
			_, err := app.StakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx)
			require.NoError(t, err)

			// delegate tokens to the validator
			tstaking.Delegate(del2, val1, sdk.NewInt(10000))

			// tokenize shares
			tstaking.TokenizeShares(del2, val1, sdk.NewInt64Coin(sdk.DefaultBondDenom, 10000), del2, true)

			// get tokenize share record
			record, err := app.StakingKeeper.GetTokenizeShareRecord(ctx, 1)
			require.NoError(t, err)

			// redeem share
			tstaking.RedeemTokensForShares(del2, sdk.NewCoin(record.GetShareTokenDenom(), tc.amount), tc.isSuccess)
		})
	}
}

func TransferTokenizeShareRecord(t *testing.T) {
	initPower := int64(1000)

	testCases := []struct {
		name      string
		recordID  uint64
		oldOwner  int64
		newOwner  int64
		isSuccess bool
	}{
		{
			"transfer to other",
			1,
			2, 1,
			true,
		},
		{
			"self transfer",
			1,
			2, 2,
			true,
		},
		{
			"transfer non-existent",
			2,
			2, 2,
			false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(*testing.T) {
			app, ctx, delAddrs, valAddrs := bootstrapHandlerGenesisTest(t, initPower, 3, sdk.TokensFromConsensusPower(initPower, sdk.DefaultPowerReduction))
			val1 := valAddrs[0]
			del2 := delAddrs[1]
			tstaking := teststaking.NewHelper(t, ctx, app.StakingKeeper)

			// set staking params
			params := app.StakingKeeper.GetAllParams(ctx)
			params.MaxValidators = 2
			app.StakingKeeper.SetParams(ctx, params)

			// add validators
			tstaking.CreateValidatorWithValPower(val1, PKs[0], 50, true)

			// call it to update validator status to bonded
			_, err := app.StakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx)
			require.NoError(t, err)

			// delegate tokens to the validator
			tstaking.Delegate(del2, val1, sdk.NewInt(10000))

			// tokenize shares
			tstaking.TokenizeShares(del2, val1, sdk.NewInt64Coin(sdk.DefaultBondDenom, 10000), del2, true)

			// redeem share
			tstaking.TranserTokenizeShareRecord(tc.recordID, delAddrs[tc.oldOwner], delAddrs[tc.newOwner], tc.isSuccess)
		})
	}
}
