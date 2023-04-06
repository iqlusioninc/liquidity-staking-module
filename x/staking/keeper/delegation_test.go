package keeper_test

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
	sdkstaking "github.com/cosmos/cosmos-sdk/x/staking/types"
	simapp "github.com/iqlusioninc/liquidity-staking-module/app"
	"github.com/iqlusioninc/liquidity-staking-module/x/staking/keeper"
	"github.com/iqlusioninc/liquidity-staking-module/x/staking/teststaking"
	"github.com/iqlusioninc/liquidity-staking-module/x/staking/types"
)

// tests GetLiquidDelegation, GetDelegatorDelegations, SetDelegation, RemoveDelegation, GetDelegatorDelegations
func TestDelegation(t *testing.T) {
	_, app, ctx := createTestInput(t)

	// remove genesis validator delegations
	delegations := app.StakingKeeper.GetAllDelegations(ctx)
	require.Len(t, delegations, 1)

	err := app.StakingKeeper.RemoveDelegation(ctx, types.Delegation{
		ValidatorAddress: delegations[0].ValidatorAddress,
		DelegatorAddress: delegations[0].DelegatorAddress,
	})
	require.NoError(t, err)

	addrDels := simapp.AddTestAddrsIncremental(app, ctx, 3, sdk.NewInt(10000))
	valAddrs := simapp.ConvertAddrsToValAddrs(addrDels)

	// construct the validators
	amts := []math.Int{sdk.NewInt(9), sdk.NewInt(8), sdk.NewInt(7)}
	var validators [3]types.Validator
	for i, amt := range amts {
		validators[i] = teststaking.NewValidator(t, valAddrs[i], PKs[i])
		validators[i], _ = validators[i].AddTokensFromDel(amt)
	}

	validators[0] = keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validators[0], true)
	validators[1] = keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validators[1], true)
	validators[2] = keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validators[2], true)

	// first add a validators[0] to delegate too
	bond1to1 := types.NewDelegation(addrDels[0], valAddrs[0], sdk.NewDec(9), false)

	// check the empty keeper first
	_, found := app.StakingKeeper.GetLiquidDelegation(ctx, addrDels[0], valAddrs[0])
	require.False(t, found)

	// set and retrieve a record
	app.StakingKeeper.SetDelegation(ctx, bond1to1)
	resBond, found := app.StakingKeeper.GetLiquidDelegation(ctx, addrDels[0], valAddrs[0])
	require.True(t, found)
	require.Equal(t, bond1to1, resBond)

	// modify a records, save, and retrieve
	bond1to1.Shares = sdk.NewDec(99)
	app.StakingKeeper.SetDelegation(ctx, bond1to1)
	resBond, found = app.StakingKeeper.GetLiquidDelegation(ctx, addrDels[0], valAddrs[0])
	require.True(t, found)
	require.Equal(t, bond1to1, resBond)

	// add some more records
	bond1to2 := types.NewDelegation(addrDels[0], valAddrs[1], sdk.NewDec(9), false)
	bond1to3 := types.NewDelegation(addrDels[0], valAddrs[2], sdk.NewDec(9), false)
	bond2to1 := types.NewDelegation(addrDels[1], valAddrs[0], sdk.NewDec(9), false)
	bond2to2 := types.NewDelegation(addrDels[1], valAddrs[1], sdk.NewDec(9), false)
	bond2to3 := types.NewDelegation(addrDels[1], valAddrs[2], sdk.NewDec(9), false)
	app.StakingKeeper.SetDelegation(ctx, bond1to2)
	app.StakingKeeper.SetDelegation(ctx, bond1to3)
	app.StakingKeeper.SetDelegation(ctx, bond2to1)
	app.StakingKeeper.SetDelegation(ctx, bond2to2)
	app.StakingKeeper.SetDelegation(ctx, bond2to3)

	// test all bond retrieve capabilities
	resBonds := app.StakingKeeper.GetDelegatorDelegations(ctx, addrDels[0], 5)
	require.Equal(t, 3, len(resBonds))
	require.Equal(t, bond1to1, resBonds[0])
	require.Equal(t, bond1to2, resBonds[1])
	require.Equal(t, bond1to3, resBonds[2])
	resBonds = app.StakingKeeper.GetAllLiquidDelegatorDelegations(ctx, addrDels[0])
	require.Equal(t, 3, len(resBonds))
	resBonds = app.StakingKeeper.GetDelegatorDelegations(ctx, addrDels[0], 2)
	require.Equal(t, 2, len(resBonds))
	resBonds = app.StakingKeeper.GetDelegatorDelegations(ctx, addrDels[1], 5)
	require.Equal(t, 3, len(resBonds))
	require.Equal(t, bond2to1, resBonds[0])
	require.Equal(t, bond2to2, resBonds[1])
	require.Equal(t, bond2to3, resBonds[2])
	allBonds := app.StakingKeeper.GetAllDelegations(ctx)
	require.Equal(t, 6, len(allBonds))
	require.Equal(t, bond1to1, allBonds[0])
	require.Equal(t, bond1to2, allBonds[1])
	require.Equal(t, bond1to3, allBonds[2])
	require.Equal(t, bond2to1, allBonds[3])
	require.Equal(t, bond2to2, allBonds[4])
	require.Equal(t, bond2to3, allBonds[5])

	resVals := app.StakingKeeper.GetDelegatorValidators(ctx, addrDels[0], 3)
	require.Equal(t, 3, len(resVals))
	resVals = app.StakingKeeper.GetDelegatorValidators(ctx, addrDels[1], 4)
	require.Equal(t, 3, len(resVals))

	for i := 0; i < 3; i++ {
		resVal, err := app.StakingKeeper.GetDelegatorValidator(ctx, addrDels[0], valAddrs[i])
		require.Nil(t, err)
		require.Equal(t, valAddrs[i], resVal.GetOperator())

		resVal, err = app.StakingKeeper.GetDelegatorValidator(ctx, addrDels[1], valAddrs[i])
		require.Nil(t, err)
		require.Equal(t, valAddrs[i], resVal.GetOperator())

		resDels := app.StakingKeeper.GetValidatorDelegations(ctx, valAddrs[i])
		require.Len(t, resDels, 2)
	}

	// delete a record
	err = app.StakingKeeper.RemoveDelegation(ctx, bond2to3)
	require.Nil(t, err)
	_, found = app.StakingKeeper.GetLiquidDelegation(ctx, addrDels[1], valAddrs[2])
	require.False(t, found)
	resBonds = app.StakingKeeper.GetDelegatorDelegations(ctx, addrDels[1], 5)
	require.Equal(t, 2, len(resBonds))
	require.Equal(t, bond2to1, resBonds[0])
	require.Equal(t, bond2to2, resBonds[1])

	resBonds = app.StakingKeeper.GetAllLiquidDelegatorDelegations(ctx, addrDels[1])
	require.Equal(t, 2, len(resBonds))

	// delete all the records from delegator 2
	err = app.StakingKeeper.RemoveDelegation(ctx, bond2to1)
	require.Nil(t, err)
	err = app.StakingKeeper.RemoveDelegation(ctx, bond2to2)
	require.Nil(t, err)
	_, found = app.StakingKeeper.GetLiquidDelegation(ctx, addrDels[1], valAddrs[0])
	require.False(t, found)
	_, found = app.StakingKeeper.GetLiquidDelegation(ctx, addrDels[1], valAddrs[1])
	require.False(t, found)
	resBonds = app.StakingKeeper.GetDelegatorDelegations(ctx, addrDels[1], 5)
	require.Equal(t, 0, len(resBonds))
}

// tests Get/Set/Remove UnbondingDelegation
func TestUnbondingDelegation(t *testing.T) {
	_, app, ctx := createTestInput(t)

	delAddrs := simapp.AddTestAddrsIncremental(app, ctx, 2, sdk.NewInt(10000))
	valAddrs := simapp.ConvertAddrsToValAddrs(delAddrs)

	ubd := types.NewUnbondingDelegation(
		delAddrs[0],
		valAddrs[0],
		0,
		time.Unix(0, 0).UTC(),
		sdk.NewInt(5),
	)

	// set and retrieve a record
	app.StakingKeeper.SetUnbondingDelegation(ctx, ubd)
	resUnbond, found := app.StakingKeeper.GetUnbondingDelegation(ctx, delAddrs[0], valAddrs[0])
	require.True(t, found)
	require.Equal(t, ubd, resUnbond)

	// modify a records, save, and retrieve
	ubd.Entries[0].Balance = sdk.NewInt(21)
	app.StakingKeeper.SetUnbondingDelegation(ctx, ubd)

	resUnbonds := app.StakingKeeper.GetUnbondingDelegations(ctx, delAddrs[0], 5)
	require.Equal(t, 1, len(resUnbonds))

	resUnbonds = app.StakingKeeper.GetAllUnbondingDelegations(ctx, delAddrs[0])
	require.Equal(t, 1, len(resUnbonds))

	resUnbond, found = app.StakingKeeper.GetUnbondingDelegation(ctx, delAddrs[0], valAddrs[0])
	require.True(t, found)
	require.Equal(t, ubd, resUnbond)

	// delete a record
	app.StakingKeeper.RemoveUnbondingDelegation(ctx, ubd)
	_, found = app.StakingKeeper.GetUnbondingDelegation(ctx, delAddrs[0], valAddrs[0])
	require.False(t, found)

	resUnbonds = app.StakingKeeper.GetUnbondingDelegations(ctx, delAddrs[0], 5)
	require.Equal(t, 0, len(resUnbonds))

	resUnbonds = app.StakingKeeper.GetAllUnbondingDelegations(ctx, delAddrs[0])
	require.Equal(t, 0, len(resUnbonds))
}

func TestUnbondDelegation(t *testing.T) {
	_, app, ctx := createTestInput(t)

	delAddrs := simapp.AddTestAddrsIncremental(app, ctx, 1, sdk.NewInt(10000))
	valAddrs := simapp.ConvertAddrsToValAddrs(delAddrs)

	startTokens := app.StakingKeeper.TokensFromConsensusPower(ctx, 10)
	notBondedPool := app.StakingKeeper.GetNotBondedPool(ctx)

	require.NoError(t, testutil.FundModuleAccount(app.BankKeeper, ctx, notBondedPool.GetName(), sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), startTokens))))
	app.AccountKeeper.SetModuleAccount(ctx, notBondedPool)

	// create a validator and a delegator to that validator
	// note this validator starts not-bonded
	validator := teststaking.NewValidator(t, valAddrs[0], PKs[0])

	validator, issuedShares := validator.AddTokensFromDel(startTokens)
	require.Equal(t, startTokens, issuedShares.RoundInt())

	_ = keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validator, true)

	delegation := types.NewDelegation(delAddrs[0], valAddrs[0], issuedShares, false)
	app.StakingKeeper.SetDelegation(ctx, delegation)

	bondTokens := app.StakingKeeper.TokensFromConsensusPower(ctx, 6)
	amount, err := app.StakingKeeper.Unbond(ctx, delAddrs[0], valAddrs[0], sdk.NewDecFromInt(bondTokens))
	require.NoError(t, err)
	require.Equal(t, bondTokens, amount) // shares to be added to an unbonding delegation

	delegation, found := app.StakingKeeper.GetLiquidDelegation(ctx, delAddrs[0], valAddrs[0])
	require.True(t, found)
	validator, found = app.StakingKeeper.GetLiquidValidator(ctx, valAddrs[0])
	require.True(t, found)

	remainingTokens := startTokens.Sub(bondTokens)
	require.Equal(t, remainingTokens, delegation.Shares.RoundInt())
	require.Equal(t, remainingTokens, validator.BondedTokens())
}

func TestUnbondingDelegationsMaxEntries(t *testing.T) {
	_, app, ctx := createTestInput(t)

	addrDels := simapp.AddTestAddrsIncremental(app, ctx, 1, sdk.NewInt(10000))
	addrVals := simapp.ConvertAddrsToValAddrs(addrDels)

	startTokens := app.StakingKeeper.TokensFromConsensusPower(ctx, 10)

	bondDenom := app.StakingKeeper.BondDenom(ctx)
	notBondedPool := app.StakingKeeper.GetNotBondedPool(ctx)

	require.NoError(t, testutil.FundModuleAccount(app.BankKeeper, ctx, notBondedPool.GetName(), sdk.NewCoins(sdk.NewCoin(bondDenom, startTokens))))
	app.AccountKeeper.SetModuleAccount(ctx, notBondedPool)

	// create a validator and a delegator to that validator
	validator := teststaking.NewValidator(t, addrVals[0], PKs[0])

	validator, issuedShares := validator.AddTokensFromDel(startTokens)
	require.Equal(t, startTokens, issuedShares.RoundInt())

	validator = keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validator, true)
	require.True(math.IntEq(t, startTokens, validator.BondedTokens()))
	require.True(t, validator.IsBonded())

	delegation := types.NewDelegation(addrDels[0], addrVals[0], issuedShares, false)
	app.StakingKeeper.SetDelegation(ctx, delegation)

	maxEntries := app.StakingKeeper.MaxEntries(ctx)

	oldBonded := app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetBondedPool(ctx).GetAddress(), bondDenom).Amount
	oldNotBonded := app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetNotBondedPool(ctx).GetAddress(), bondDenom).Amount

	// should all pass
	var completionTime time.Time
	for i := uint32(0); i < maxEntries; i++ {
		var err error
		completionTime, err = app.StakingKeeper.Undelegate(ctx, addrDels[0], addrVals[0], sdk.NewDec(1))
		require.NoError(t, err)
	}

	newBonded := app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetBondedPool(ctx).GetAddress(), bondDenom).Amount
	newNotBonded := app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetNotBondedPool(ctx).GetAddress(), bondDenom).Amount
	require.True(math.IntEq(t, newBonded, oldBonded.SubRaw(int64(maxEntries))))
	require.True(math.IntEq(t, newNotBonded, oldNotBonded.AddRaw(int64(maxEntries))))

	oldBonded = app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetBondedPool(ctx).GetAddress(), bondDenom).Amount
	oldNotBonded = app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetNotBondedPool(ctx).GetAddress(), bondDenom).Amount

	// an additional unbond should fail due to max entries
	_, err := app.StakingKeeper.Undelegate(ctx, addrDels[0], addrVals[0], sdk.NewDec(1))
	require.Error(t, err)

	newBonded = app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetBondedPool(ctx).GetAddress(), bondDenom).Amount
	newNotBonded = app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetNotBondedPool(ctx).GetAddress(), bondDenom).Amount

	require.True(math.IntEq(t, newBonded, oldBonded))
	require.True(math.IntEq(t, newNotBonded, oldNotBonded))

	// mature unbonding delegations
	ctx = ctx.WithBlockTime(completionTime)
	_, err = app.StakingKeeper.CompleteUnbonding(ctx, addrDels[0], addrVals[0])
	require.NoError(t, err)

	newBonded = app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetBondedPool(ctx).GetAddress(), bondDenom).Amount
	newNotBonded = app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetNotBondedPool(ctx).GetAddress(), bondDenom).Amount
	require.True(math.IntEq(t, newBonded, oldBonded))
	require.True(math.IntEq(t, newNotBonded, oldNotBonded.SubRaw(int64(maxEntries))))

	oldNotBonded = app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetNotBondedPool(ctx).GetAddress(), bondDenom).Amount

	// unbonding  should work again
	_, err = app.StakingKeeper.Undelegate(ctx, addrDels[0], addrVals[0], sdk.NewDec(1))
	require.NoError(t, err)

	newBonded = app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetBondedPool(ctx).GetAddress(), bondDenom).Amount
	newNotBonded = app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetNotBondedPool(ctx).GetAddress(), bondDenom).Amount
	require.True(math.IntEq(t, newBonded, oldBonded.SubRaw(1)))
	require.True(math.IntEq(t, newNotBonded, oldNotBonded.AddRaw(1)))
}

// Make sure that that the retrieving the delegations doesn't affect the state
func TestGetRedelegationsFromSrcValidator(t *testing.T) {
	_, app, ctx := createTestInput(t)

	addrDels := simapp.AddTestAddrsIncremental(app, ctx, 2, sdk.NewInt(0))
	addrVals := simapp.ConvertAddrsToValAddrs(addrDels)

	rd := types.NewRedelegation(addrDels[0], addrVals[0], addrVals[1], 0,
		time.Unix(0, 0), sdk.NewInt(5),
		sdk.NewDec(5))

	// set and retrieve a record
	app.StakingKeeper.SetRedelegation(ctx, rd)
	resBond, found := app.StakingKeeper.GetRedelegation(ctx, addrDels[0], addrVals[0], addrVals[1])
	require.True(t, found)

	// get the redelegations one time
	redelegations := app.StakingKeeper.GetRedelegationsFromSrcValidator(ctx, addrVals[0])
	require.Equal(t, 1, len(redelegations))
	require.Equal(t, redelegations[0], resBond)

	// get the redelegations a second time, should be exactly the same
	redelegations = app.StakingKeeper.GetRedelegationsFromSrcValidator(ctx, addrVals[0])
	require.Equal(t, 1, len(redelegations))
	require.Equal(t, redelegations[0], resBond)
}

// tests Get/Set/Remove/Has UnbondingDelegation
func TestRedelegation(t *testing.T) {
	_, app, ctx := createTestInput(t)

	addrDels := simapp.AddTestAddrsIncremental(app, ctx, 2, sdk.NewInt(0))
	addrVals := simapp.ConvertAddrsToValAddrs(addrDels)

	rd := types.NewRedelegation(addrDels[0], addrVals[0], addrVals[1], 0,
		time.Unix(0, 0).UTC(), sdk.NewInt(5),
		sdk.NewDec(5))

	// test shouldn't have and redelegations
	has := app.StakingKeeper.HasReceivingRedelegation(ctx, addrDels[0], addrVals[1])
	require.False(t, has)

	// set and retrieve a record
	app.StakingKeeper.SetRedelegation(ctx, rd)
	resRed, found := app.StakingKeeper.GetRedelegation(ctx, addrDels[0], addrVals[0], addrVals[1])
	require.True(t, found)

	redelegations := app.StakingKeeper.GetRedelegationsFromSrcValidator(ctx, addrVals[0])
	require.Equal(t, 1, len(redelegations))
	require.Equal(t, redelegations[0], resRed)

	redelegations = app.StakingKeeper.GetRedelegations(ctx, addrDels[0], 5)
	require.Equal(t, 1, len(redelegations))
	require.Equal(t, redelegations[0], resRed)

	redelegations = app.StakingKeeper.GetAllRedelegations(ctx, addrDels[0], nil, nil)
	require.Equal(t, 1, len(redelegations))
	require.Equal(t, redelegations[0], resRed)

	// check if has the redelegation
	has = app.StakingKeeper.HasReceivingRedelegation(ctx, addrDels[0], addrVals[1])
	require.True(t, has)

	// modify a records, save, and retrieve
	rd.Entries[0].SharesDst = sdk.NewDec(21)
	app.StakingKeeper.SetRedelegation(ctx, rd)

	resRed, found = app.StakingKeeper.GetRedelegation(ctx, addrDels[0], addrVals[0], addrVals[1])
	require.True(t, found)
	require.Equal(t, rd, resRed)

	redelegations = app.StakingKeeper.GetRedelegationsFromSrcValidator(ctx, addrVals[0])
	require.Equal(t, 1, len(redelegations))
	require.Equal(t, redelegations[0], resRed)

	redelegations = app.StakingKeeper.GetRedelegations(ctx, addrDels[0], 5)
	require.Equal(t, 1, len(redelegations))
	require.Equal(t, redelegations[0], resRed)

	// delete a record
	app.StakingKeeper.RemoveRedelegation(ctx, rd)
	_, found = app.StakingKeeper.GetRedelegation(ctx, addrDels[0], addrVals[0], addrVals[1])
	require.False(t, found)

	redelegations = app.StakingKeeper.GetRedelegations(ctx, addrDels[0], 5)
	require.Equal(t, 0, len(redelegations))

	redelegations = app.StakingKeeper.GetAllRedelegations(ctx, addrDels[0], nil, nil)
	require.Equal(t, 0, len(redelegations))
}

func TestRedelegateToSameValidator(t *testing.T) {
	_, app, ctx := createTestInput(t)

	addrDels := simapp.AddTestAddrsIncremental(app, ctx, 1, sdk.NewInt(0))
	addrVals := simapp.ConvertAddrsToValAddrs(addrDels)

	valTokens := app.StakingKeeper.TokensFromConsensusPower(ctx, 10)
	startCoins := sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), valTokens))

	// add bonded tokens to pool for delegations
	notBondedPool := app.StakingKeeper.GetNotBondedPool(ctx)
	require.NoError(t, testutil.FundModuleAccount(app.BankKeeper, ctx, notBondedPool.GetName(), startCoins))
	app.AccountKeeper.SetModuleAccount(ctx, notBondedPool)

	// create a validator with a self-delegation
	validator := teststaking.NewValidator(t, addrVals[0], PKs[0])
	validator, issuedShares := validator.AddTokensFromDel(valTokens)
	require.Equal(t, valTokens, issuedShares.RoundInt())
	validator = keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validator, true)
	require.True(t, validator.IsBonded())

	val0AccAddr := sdk.AccAddress(addrVals[0].Bytes())
	selfDelegation := types.NewDelegation(val0AccAddr, addrVals[0], issuedShares, false)
	app.StakingKeeper.SetDelegation(ctx, selfDelegation)

	_, err := app.StakingKeeper.BeginRedelegation(ctx, val0AccAddr, addrVals[0], addrVals[0], sdk.NewDec(5))
	require.Error(t, err)
}

func TestRedelegationMaxEntries(t *testing.T) {
	_, app, ctx := createTestInput(t)

	addrDels := simapp.AddTestAddrsIncremental(app, ctx, 2, sdk.NewInt(0))
	addrVals := simapp.ConvertAddrsToValAddrs(addrDels)

	startTokens := app.StakingKeeper.TokensFromConsensusPower(ctx, 20)
	startCoins := sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), startTokens))

	// add bonded tokens to pool for delegations
	notBondedPool := app.StakingKeeper.GetNotBondedPool(ctx)
	require.NoError(t, testutil.FundModuleAccount(app.BankKeeper, ctx, notBondedPool.GetName(), startCoins))
	app.AccountKeeper.SetModuleAccount(ctx, notBondedPool)

	// create a validator with a self-delegation
	validator := teststaking.NewValidator(t, addrVals[0], PKs[0])
	valTokens := app.StakingKeeper.TokensFromConsensusPower(ctx, 10)
	validator, issuedShares := validator.AddTokensFromDel(valTokens)
	require.Equal(t, valTokens, issuedShares.RoundInt())
	_ = keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validator, true)
	val0AccAddr := sdk.AccAddress(addrVals[0].Bytes())
	selfDelegation := types.NewDelegation(val0AccAddr, addrVals[0], issuedShares, false)
	app.StakingKeeper.SetDelegation(ctx, selfDelegation)

	// create a second validator
	validator2 := teststaking.NewValidator(t, addrVals[1], PKs[1])
	validator2, issuedShares = validator2.AddTokensFromDel(valTokens)
	require.Equal(t, valTokens, issuedShares.RoundInt())

	validator2 = keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validator2, true)
	require.Equal(t, sdkstaking.Bonded, validator2.Status)

	maxEntries := app.StakingKeeper.MaxEntries(ctx)

	// redelegations should pass
	var completionTime time.Time
	for i := uint32(0); i < maxEntries; i++ {
		var err error
		completionTime, err = app.StakingKeeper.BeginRedelegation(ctx, val0AccAddr, addrVals[0], addrVals[1], sdk.NewDec(1))
		require.NoError(t, err)
	}

	// an additional redelegation should fail due to max entries
	_, err := app.StakingKeeper.BeginRedelegation(ctx, val0AccAddr, addrVals[0], addrVals[1], sdk.NewDec(1))
	require.Error(t, err)

	// mature redelegations
	ctx = ctx.WithBlockTime(completionTime)
	_, err = app.StakingKeeper.CompleteRedelegation(ctx, val0AccAddr, addrVals[0], addrVals[1])
	require.NoError(t, err)

	// redelegation should work again
	_, err = app.StakingKeeper.BeginRedelegation(ctx, val0AccAddr, addrVals[0], addrVals[1], sdk.NewDec(1))
	require.NoError(t, err)
}

func TestExemptDelegationUndelegate(t *testing.T) {
	_, app, ctx := createTestInput(t)

	addrDels := simapp.AddTestAddrs(app, ctx, 2, app.StakingKeeper.TokensFromConsensusPower(ctx, 10000))
	addrVals := simapp.ConvertAddrsToValAddrs(addrDels)

	startTokens := app.StakingKeeper.TokensFromConsensusPower(ctx, 10)

	bondDenom := app.StakingKeeper.BondDenom(ctx)
	notBondedPool := app.StakingKeeper.GetNotBondedPool(ctx)

	require.NoError(t, testutil.FundModuleAccount(app.BankKeeper, ctx, notBondedPool.GetName(), sdk.NewCoins(sdk.NewCoin(bondDenom, startTokens))))
	app.AccountKeeper.SetModuleAccount(ctx, notBondedPool)

	// create a validator and a delegator to that validator
	validator := teststaking.NewValidator(t, addrVals[0], PKs[0])
	app.StakingKeeper.SetValidator(ctx, validator)

	// set exemption factor
	params := app.StakingKeeper.GetParams(ctx)
	params.ExemptionFactor = sdk.NewDec(1)
	app.StakingKeeper.SetParams(ctx, params)

	// convert to exempt delegation
	msgServer := keeper.NewMsgServerImpl(app.StakingKeeper)

	validator, _ = app.StakingKeeper.GetLiquidValidator(ctx, addrVals[0])
	err := delegateCoinsFromAccount(ctx, app, addrDels[0], startTokens, validator)
	require.NoError(t, err)
	_, err = msgServer.ExemptDelegation(sdk.WrapSDKContext(ctx), &types.MsgExemptDelegation{
		DelegatorAddress: addrDels[0].String(),
		ValidatorAddress: addrVals[0].String(),
	})
	require.NoError(t, err)

	// tokenize share for 2nd account delegation
	validator, _ = app.StakingKeeper.GetLiquidValidator(ctx, addrVals[0])
	err = delegateCoinsFromAccount(ctx, app, addrDels[1], startTokens, validator)
	require.NoError(t, err)
	tokenizeShareResp, err := msgServer.TokenizeShares(sdk.WrapSDKContext(ctx), &types.MsgTokenizeShares{
		DelegatorAddress:    addrDels[1].String(),
		ValidatorAddress:    addrVals[0].String(),
		Amount:              sdk.NewCoin(sdk.DefaultBondDenom, startTokens),
		TokenizedShareOwner: addrDels[0].String(),
	})
	require.NoError(t, err)

	// try undelegating
	_, err = msgServer.Undelegate(ctx, &types.MsgUndelegate{
		DelegatorAddress: addrDels[0].String(),
		ValidatorAddress: addrVals[0].String(),
		Amount:           sdk.NewCoin(sdk.DefaultBondDenom, startTokens),
	})
	require.Error(t, err)

	// redeem full amount on 2nd account and try undelegation
	validator, _ = app.StakingKeeper.GetLiquidValidator(ctx, addrVals[0])
	err = delegateCoinsFromAccount(ctx, app, addrDels[1], startTokens, validator)
	require.NoError(t, err)
	_, err = msgServer.RedeemTokens(sdk.WrapSDKContext(ctx), &types.MsgRedeemTokensforShares{
		DelegatorAddress: addrDels[1].String(),
		Amount:           tokenizeShareResp.Amount,
	})
	require.NoError(t, err)

	// try undelegating
	_, err = msgServer.Undelegate(ctx, &types.MsgUndelegate{
		DelegatorAddress: addrDels[0].String(),
		ValidatorAddress: addrVals[0].String(),
		Amount:           sdk.NewCoin(sdk.DefaultBondDenom, startTokens),
	})
	require.NoError(t, err)

	validator, _ = app.StakingKeeper.GetLiquidValidator(ctx, addrVals[0])
	require.Equal(t, validator.TotalExemptShares, sdk.ZeroDec())
}

func TestExemptDelegationRedelegate(t *testing.T) {
	_, app, ctx := createTestInput(t)

	addrDels := simapp.AddTestAddrs(app, ctx, 2, app.StakingKeeper.TokensFromConsensusPower(ctx, 10000))
	addrVals := simapp.ConvertAddrsToValAddrs(addrDels)

	startTokens := app.StakingKeeper.TokensFromConsensusPower(ctx, 10)

	bondDenom := app.StakingKeeper.BondDenom(ctx)
	notBondedPool := app.StakingKeeper.GetNotBondedPool(ctx)

	require.NoError(t, testutil.FundModuleAccount(app.BankKeeper, ctx, notBondedPool.GetName(), sdk.NewCoins(sdk.NewCoin(bondDenom, startTokens))))
	app.AccountKeeper.SetModuleAccount(ctx, notBondedPool)

	// create a validator and a delegator to that validator
	validator := teststaking.NewValidator(t, addrVals[0], PKs[0])
	app.StakingKeeper.SetValidator(ctx, validator)
	validator2 := teststaking.NewValidator(t, addrVals[1], PKs[1])
	app.StakingKeeper.SetValidator(ctx, validator2)

	// set exemption factor
	params := app.StakingKeeper.GetParams(ctx)
	params.ExemptionFactor = sdk.NewDec(1)
	app.StakingKeeper.SetParams(ctx, params)

	// convert to exempt delegation
	msgServer := keeper.NewMsgServerImpl(app.StakingKeeper)

	validator, _ = app.StakingKeeper.GetLiquidValidator(ctx, addrVals[0])
	err := delegateCoinsFromAccount(ctx, app, addrDels[0], startTokens, validator)
	require.NoError(t, err)
	_, err = msgServer.ExemptDelegation(sdk.WrapSDKContext(ctx), &types.MsgExemptDelegation{
		DelegatorAddress: addrDels[0].String(),
		ValidatorAddress: addrVals[0].String(),
	})
	require.NoError(t, err)

	// tokenize share for 2nd account delegation
	validator, _ = app.StakingKeeper.GetLiquidValidator(ctx, addrVals[0])
	err = delegateCoinsFromAccount(ctx, app, addrDels[1], startTokens, validator)
	require.NoError(t, err)
	tokenizeShareResp, err := msgServer.TokenizeShares(sdk.WrapSDKContext(ctx), &types.MsgTokenizeShares{
		DelegatorAddress:    addrDels[1].String(),
		ValidatorAddress:    addrVals[0].String(),
		Amount:              sdk.NewCoin(sdk.DefaultBondDenom, startTokens),
		TokenizedShareOwner: addrDels[0].String(),
	})
	require.NoError(t, err)

	// try undelegating
	_, err = msgServer.BeginRedelegate(ctx, &types.MsgBeginRedelegate{
		DelegatorAddress:    addrDels[0].String(),
		ValidatorSrcAddress: addrVals[0].String(),
		ValidatorDstAddress: addrVals[1].String(),
		Amount:              sdk.NewCoin(sdk.DefaultBondDenom, startTokens),
	})
	require.Error(t, err)

	// redeem full amount on 2nd account and try undelegation
	validator, _ = app.StakingKeeper.GetLiquidValidator(ctx, addrVals[0])
	err = delegateCoinsFromAccount(ctx, app, addrDels[1], startTokens, validator)
	require.NoError(t, err)
	_, err = msgServer.RedeemTokens(sdk.WrapSDKContext(ctx), &types.MsgRedeemTokensforShares{
		DelegatorAddress: addrDels[1].String(),
		Amount:           tokenizeShareResp.Amount,
	})
	require.NoError(t, err)

	// try undelegating
	_, err = msgServer.BeginRedelegate(ctx, &types.MsgBeginRedelegate{
		DelegatorAddress:    addrDels[0].String(),
		ValidatorSrcAddress: addrVals[0].String(),
		ValidatorDstAddress: addrVals[1].String(),
		Amount:              sdk.NewCoin(sdk.DefaultBondDenom, startTokens),
	})
	require.NoError(t, err)

	validator, _ = app.StakingKeeper.GetLiquidValidator(ctx, addrVals[0])
	require.Equal(t, validator.TotalExemptShares, sdk.ZeroDec())
}
