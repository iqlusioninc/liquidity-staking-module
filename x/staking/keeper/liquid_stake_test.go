package keeper_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	simapp "github.com/iqlusioninc/liquidity-staking-module/app"
	"github.com/iqlusioninc/liquidity-staking-module/x/staking/types"
	"github.com/stretchr/testify/require"
)

// Tests Set/Get TotalLiquidStakedTokens
func TestTotalLiquidStakedTokens(t *testing.T) {
	_, app, ctx := createTestInput(t)

	// Before it's been initialized, it should return zero
	require.Equal(t, sdk.ZeroDec(), app.StakingKeeper.GetTotalLiquidStakedTokens(ctx), "zero")

	// Update the total liquid staked
	total := sdk.NewInt(100)
	app.StakingKeeper.SetTotalLiquidStakedTokens(ctx, total)

	// Confirm it was updated
	require.Equal(t, total, app.StakingKeeper.GetTotalLiquidStakedTokens(ctx), "initial")
}

// Tests Increase/Decrease TotalValidatorTotalLiquidShares
func TestValidatorTotalLiquidShares(t *testing.T) {
	_, app, ctx := createTestInput(t)

	// Create a validator address
	privKey := secp256k1.GenPrivKey()
	pubKey := privKey.PubKey()
	valAddress := sdk.ValAddress(pubKey.Address())

	// Set an initial total
	initial := sdk.NewDec(100)
	validator := types.Validator{
		OperatorAddress:   valAddress.String(),
		TotalLiquidShares: initial,
	}
	app.StakingKeeper.SetValidator(ctx, validator)
}

// Tests AccountIsLiquidStakingProvider
func TestAccountIsLiquidStakingProvider(t *testing.T) {
	_, app, ctx := createTestInput(t)

	// Create base account
	baseAccountAddress := sdk.AccAddress("base-account")
	app.AccountKeeper.SetAccount(ctx, authtypes.NewBaseAccountWithAddress(baseAccountAddress))

	// Create an ICA module account
	icaModuleAccountName := "ica-account"
	icaAccountAddress := address.Module(icaModuleAccountName, []byte("ica-module-account"))
	icaAccount := authtypes.NewModuleAccount(
		authtypes.NewBaseAccountWithAddress(icaAccountAddress),
		icaModuleAccountName,
	)
	app.AccountKeeper.SetAccount(ctx, icaAccount)

	// Only the ICA module account should be considered a liquid staking provider
	require.False(t, app.StakingKeeper.AccountIsLiquidStakingProvider(ctx, baseAccountAddress), "base account")
	require.True(t, app.StakingKeeper.AccountIsLiquidStakingProvider(ctx, icaAccountAddress), "ICA module account")
}

// Helper function to clear the Bonded and NotBonded pool balances before a unit test
func clearPoolBalances(t *testing.T, app *simapp.SimApp, ctx sdk.Context) {
	bondDenom := app.StakingKeeper.BondDenom(ctx)
	initialBondedBalance := app.BankKeeper.GetBalance(ctx, app.AccountKeeper.GetModuleAddress(types.BondedPoolName), bondDenom)
	initialNotBondedBalance := app.BankKeeper.GetBalance(ctx, app.AccountKeeper.GetModuleAddress(types.NotBondedPoolName), bondDenom)

	err := app.BankKeeper.SendCoinsFromModuleToModule(ctx, types.BondedPoolName, minttypes.ModuleName, sdk.NewCoins(initialBondedBalance))
	require.NoError(t, err, "no error expected when clearing bonded pool balance")

	err = app.BankKeeper.SendCoinsFromModuleToModule(ctx, types.NotBondedPoolName, minttypes.ModuleName, sdk.NewCoins(initialNotBondedBalance))
	require.NoError(t, err, "no error expected when clearing notbonded pool balance")
}

// Helper function to fund the Bonded and NotBonded pool balances before a unit test
func fundPoolBalances(t *testing.T, app *simapp.SimApp, ctx sdk.Context, bondedBalance sdk.Int, notBondedBalance sdk.Int) {
	bondDenom := app.StakingKeeper.BondDenom(ctx)
	bondedPoolCoin := sdk.NewCoin(bondDenom, bondedBalance)
	notbondedPoolCoin := sdk.NewCoin(bondDenom, notBondedBalance)
	totalPoolCoin := sdk.NewCoin(bondDenom, bondedBalance.Add(notBondedBalance))

	err := app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, sdk.NewCoins(totalPoolCoin))
	require.NoError(t, err, "no error expected when minting")

	err = app.BankKeeper.SendCoinsFromModuleToModule(ctx, minttypes.ModuleName, types.BondedPoolName, sdk.NewCoins(bondedPoolCoin))
	require.NoError(t, err, "no error expected when sending tokens to bonded pool")

	err = app.BankKeeper.SendCoinsFromModuleToModule(ctx, minttypes.ModuleName, types.NotBondedPoolName, sdk.NewCoins(notbondedPoolCoin))
	require.NoError(t, err, "no error expected when sending tokens to notbonded pool")
}

// Tests CheckExceedsGlobalLiquidStakingCap
func TestCheckExceedsGlobalLiquidStakingCap(t *testing.T) {
	_, app, ctx := createTestInput(t)

	testCases := []struct {
		name                string
		bondedBalance       sdk.Int
		notBondedBalance    sdk.Int
		globalLiquidCap     sdk.Dec
		currentLiquidTokens sdk.Int
		newDelegation       sdk.Int
		expectedExceeds     bool
	}{
		{
			// Cap: 10% - Delegation Below Threshold
			// Total Liquid: 5   | Bonded Balance: 60, Notbonded Balance: 35 => Total Stake: 95
			// New Delegation: 1 | New Liquid: 5+1=6,  New Total: 95+1=96    => 6/96 = 6% < 10% cap
			name:                "10 percent cap _ delegation below cap",
			globalLiquidCap:     sdk.MustNewDecFromStr("0.1"),
			currentLiquidTokens: sdk.NewInt(5),
			bondedBalance:       sdk.NewInt(60),
			notBondedBalance:    sdk.NewInt(35),
			newDelegation:       sdk.NewInt(1),
			expectedExceeds:     false,
		},
		{
			// Cap: 10% - Delegation At Threshold
			// Total Liquid: 5   | Bonded Balance: 35, Notbonded Balance: 60 => Total Stake: 95
			// New Delegation: 5 | New Liquid: 5+5=10, New Total: 95+5=100   => 10/100 = 10% == 10% cap
			name:                "10 percent cap _ delegation equals cap",
			globalLiquidCap:     sdk.MustNewDecFromStr("0.1"),
			currentLiquidTokens: sdk.NewInt(5),
			bondedBalance:       sdk.NewInt(35),
			notBondedBalance:    sdk.NewInt(60),
			newDelegation:       sdk.NewInt(5),
			expectedExceeds:     false,
		},
		{
			// Cap: 10% - Delegation Exceeds Threshold
			// Total Liquid: 5   | Bonded Balance: 95, Notbonded Balance: 0 => Total Stake: 95
			// New Delegation: 6 | New Liquid: 5+6=11, New Total: 95+6=101  => 11/101 = 11% > 10% cap
			name:                "10 percent cap _ delegation exceeds cap",
			globalLiquidCap:     sdk.MustNewDecFromStr("0.1"),
			currentLiquidTokens: sdk.NewInt(5),
			bondedBalance:       sdk.NewInt(95),
			notBondedBalance:    sdk.NewInt(0),
			newDelegation:       sdk.NewInt(6),
			expectedExceeds:     true,
		},
		{
			// Cap: 20% - Delegation Below Threshold
			// Total Liquid: 20   | Bonded Balance: 0,    Notbonded Balance: 200 => Total Stake: 200
			// New Delegation: 10 | New Liquid: 20+10=30, New Total: 200+10=210  => 30/210 = 14% < 20% cap
			name:                "20 percent cap _ delegation below cap",
			globalLiquidCap:     sdk.MustNewDecFromStr("0.20"),
			currentLiquidTokens: sdk.NewInt(20),
			bondedBalance:       sdk.NewInt(0),
			notBondedBalance:    sdk.NewInt(200),
			newDelegation:       sdk.NewInt(10),
			expectedExceeds:     false,
		},
		{
			// Cap: 20% - Delegation At Threshold
			// Total Liquid: 20   | Bonded Balance: 220,  Notbonded Balance: 0   => Total Stake: 220
			// New Delegation: 30 | New Liquid: 20+30=50, New Total: 220+30=250  => 50/250 = 20% == 20% cap
			name:                "20 percent cap _ delegation equals cap",
			globalLiquidCap:     sdk.MustNewDecFromStr("0.20"),
			currentLiquidTokens: sdk.NewInt(20),
			bondedBalance:       sdk.NewInt(220),
			notBondedBalance:    sdk.NewInt(0),
			newDelegation:       sdk.NewInt(30),
			expectedExceeds:     false,
		},
		{
			// Cap: 20% - Delegation Exceeds Threshold
			// Total Liquid: 20   | Bonded Balance: 220,  Notbonded Balance: 0   => Total Stake: 220
			// New Delegation: 31 | New Liquid: 20+31=51, New Total: 220+31=251  => 51/251 = 21% > 20% cap
			name:                "20 percent cap _ delegation exceeds cap",
			globalLiquidCap:     sdk.MustNewDecFromStr("0.20"),
			currentLiquidTokens: sdk.NewInt(20),
			bondedBalance:       sdk.NewInt(220),
			notBondedBalance:    sdk.NewInt(0),
			newDelegation:       sdk.NewInt(31),
			expectedExceeds:     true,
		},
		{
			// Cap of 0% - everything should exceed
			name:                "0 percent cap",
			globalLiquidCap:     sdk.ZeroDec(),
			currentLiquidTokens: sdk.NewInt(0),
			bondedBalance:       sdk.NewInt(1_000_000),
			notBondedBalance:    sdk.NewInt(1_000_000),
			newDelegation:       sdk.NewInt(1),
			expectedExceeds:     true,
		},
		{
			// Cap of 100% - nothing should exceed
			name:                "100 percent cap",
			globalLiquidCap:     sdk.OneDec(),
			currentLiquidTokens: sdk.NewInt(1),
			bondedBalance:       sdk.NewInt(1),
			notBondedBalance:    sdk.NewInt(1),
			newDelegation:       sdk.NewInt(1_000_000),
			expectedExceeds:     false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Update the global liquid staking cap
			params := app.StakingKeeper.GetParams(ctx)
			params.GlobalLiquidStakingCap = tc.globalLiquidCap
			app.StakingKeeper.SetParams(ctx, params)

			// Update the total liquid tokens
			app.StakingKeeper.SetTotalLiquidStakedTokens(ctx, tc.currentLiquidTokens)

			// Fund each pool for the given test case
			clearPoolBalances(t, app, ctx)
			fundPoolBalances(t, app, ctx, tc.bondedBalance, tc.notBondedBalance)

			// Check if the new tokens would exceed the global cap
			actualExceeds := app.StakingKeeper.CheckExceedsGlobalLiquidStakingCap(ctx, tc.newDelegation)
			require.Equal(t, tc.expectedExceeds, actualExceeds, tc.name)
		})
	}
}

// Tests SafelyIncreaseTotalLiquidStakedTokens
func TestSafelyIncreaseTotalLiquidStakedTokens(t *testing.T) {
	_, app, ctx := createTestInput(t)

	intitialTotalLiquidStaked := sdk.NewInt(100)
	increaseAmount := sdk.NewInt(10)
	poolBalance := sdk.NewInt(100) // for each of bonded and notBonded

	// Set the total liquid staked, Bonded, and NotBonded pool balances
	// which are required components when checking the global cap
	app.StakingKeeper.SetTotalLiquidStakedTokens(ctx, intitialTotalLiquidStaked)
	clearPoolBalances(t, app, ctx)
	fundPoolBalances(t, app, ctx, poolBalance, poolBalance)

	// Set the cap to 100% (meaning it's disabled)
	params := app.StakingKeeper.GetParams(ctx)
	params.GlobalLiquidStakingCap = sdk.OneDec()
	app.StakingKeeper.SetParams(ctx, params)

	// Attempt to increase the total liquid stake, it should not be changed and there
	// should be no error since the cap is disabled
	err := app.StakingKeeper.SafelyIncreaseTotalLiquidStakedTokens(ctx, increaseAmount)
	require.NoError(t, err)
	require.Equal(t, intitialTotalLiquidStaked, app.StakingKeeper.GetTotalLiquidStakedTokens(ctx))

	// Change the cap that it is enabled, but a small delegation would exceed the cap
	params.GlobalLiquidStakingCap = sdk.MustNewDecFromStr("0.0001")
	app.StakingKeeper.SetParams(ctx, params)

	// Attempt to increase the total liquid stake again, this time it should error since
	// the cap was exceeded
	err = app.StakingKeeper.SafelyIncreaseTotalLiquidStakedTokens(ctx, increaseAmount)
	require.ErrorIs(t, err, types.ErrGlobalLiquidStakingCapExceeded)
	require.Equal(t, intitialTotalLiquidStaked, app.StakingKeeper.GetTotalLiquidStakedTokens(ctx))

	// Now relax the cap so that the increase succeeds
	params.GlobalLiquidStakingCap = sdk.MustNewDecFromStr("0.99")
	app.StakingKeeper.SetParams(ctx, params)

	// Confirm the total increased
	err = app.StakingKeeper.SafelyIncreaseTotalLiquidStakedTokens(ctx, increaseAmount)
	require.NoError(t, err)
	require.Equal(t, intitialTotalLiquidStaked.Add(increaseAmount), app.StakingKeeper.GetTotalLiquidStakedTokens(ctx))
}

// Tests DecreaseTotalLiquidStakedTokens
func TestDecreaseTotalLiquidStakedTokens(t *testing.T) {
	_, app, ctx := createTestInput(t)

	intitialTotalLiquidStaked := sdk.NewInt(100)
	decreaseAmount := sdk.NewInt(10)

	// Set the total liquid staked to an arbitrary value
	app.StakingKeeper.SetTotalLiquidStakedTokens(ctx, intitialTotalLiquidStaked)

	// Set the cap to 100% (meaning it's disabled)
	params := app.StakingKeeper.GetParams(ctx)
	params.GlobalLiquidStakingCap = sdk.OneDec()
	app.StakingKeeper.SetParams(ctx, params)

	// Attempt to decrease the total liquid stake, it should not be changed since the cap is disabled
	app.StakingKeeper.DecreaseTotalLiquidStakedTokens(ctx, decreaseAmount)
	require.Equal(t, intitialTotalLiquidStaked, app.StakingKeeper.GetTotalLiquidStakedTokens(ctx))

	// Now relax the cap to anything other than 1 so that the cap is enabled
	params.GlobalLiquidStakingCap = sdk.MustNewDecFromStr("0.50")
	app.StakingKeeper.SetParams(ctx, params)

	// Attempt to decrease the total liquid stake again, now it should succeed
	app.StakingKeeper.DecreaseTotalLiquidStakedTokens(ctx, decreaseAmount)
	require.Equal(t, intitialTotalLiquidStaked.Sub(decreaseAmount), app.StakingKeeper.GetTotalLiquidStakedTokens(ctx))
}

// Tests CheckExceedsValidatorBondCap
func TestCheckExceedsValidatorBondCap(t *testing.T) {
	_, app, ctx := createTestInput(t)

	testCases := []struct {
		name                string
		validatorShares     sdk.Dec
		validatorBondFactor sdk.Dec
		currentLiquidShares sdk.Dec
		newShares           sdk.Dec
		expectedExceeds     bool
	}{
		{
			// Validator Shares: 100, Factor: 1, Current Shares: 90 => 100 Max Shares, Capacity: 10
			// New Shares: 5 - below cap
			name:                "factor 1 - below cap",
			validatorShares:     sdk.NewDec(100),
			validatorBondFactor: sdk.NewDec(1),
			currentLiquidShares: sdk.NewDec(90),
			newShares:           sdk.NewDec(5),
			expectedExceeds:     false,
		},
		{
			// Validator Shares: 100, Factor: 1, Current Shares: 90 => 100 Max Shares, Capacity: 10
			// New Shares: 10 - at cap
			name:                "factor 1 - at cap",
			validatorShares:     sdk.NewDec(100),
			validatorBondFactor: sdk.NewDec(1),
			currentLiquidShares: sdk.NewDec(90),
			newShares:           sdk.NewDec(10),
			expectedExceeds:     false,
		},
		{
			// Validator Shares: 100, Factor: 1, Current Shares: 90 => 100 Max Shares, Capacity: 10
			// New Shares: 15 - above cap
			name:                "factor 1 - above cap",
			validatorShares:     sdk.NewDec(100),
			validatorBondFactor: sdk.NewDec(1),
			currentLiquidShares: sdk.NewDec(90),
			newShares:           sdk.NewDec(15),
			expectedExceeds:     true,
		},
		{
			// Validator Shares: 100, Factor: 2, Current Shares: 90 => 200 Max Shares, Capacity: 110
			// New Shares: 5 - below cap
			name:                "factor 2 - well below cap",
			validatorShares:     sdk.NewDec(100),
			validatorBondFactor: sdk.NewDec(2),
			currentLiquidShares: sdk.NewDec(90),
			newShares:           sdk.NewDec(5),
			expectedExceeds:     false,
		},
		{
			// Validator Shares: 100, Factor: 2, Current Shares: 90 => 200 Max Shares, Capacity: 110
			// New Shares: 100 - below cap
			name:                "factor 2 - below cap",
			validatorShares:     sdk.NewDec(100),
			validatorBondFactor: sdk.NewDec(2),
			currentLiquidShares: sdk.NewDec(90),
			newShares:           sdk.NewDec(100),
			expectedExceeds:     false,
		},
		{
			// Validator Shares: 100, Factor: 2, Current Shares: 90 => 200 Max Shares, Capacity: 110
			// New Shares: 110 - below cap
			name:                "factor 2 - at cap",
			validatorShares:     sdk.NewDec(100),
			validatorBondFactor: sdk.NewDec(2),
			currentLiquidShares: sdk.NewDec(90),
			newShares:           sdk.NewDec(110),
			expectedExceeds:     false,
		},
		{
			// Validator Shares: 100, Factor: 2, Current Shares: 90 => 200 Max Shares, Capacity: 110
			// New Shares: 111 - above cap
			name:                "factor 2 - above cap",
			validatorShares:     sdk.NewDec(100),
			validatorBondFactor: sdk.NewDec(2),
			currentLiquidShares: sdk.NewDec(90),
			newShares:           sdk.NewDec(111),
			expectedExceeds:     true,
		},
		{
			// Validator Shares: 100, Factor: 100, Current Shares: 90 => 10000 Max Shares, Capacity: 9910
			// New Shares: 100 - below cap
			name:                "factor 100 - below cap",
			validatorShares:     sdk.NewDec(100),
			validatorBondFactor: sdk.NewDec(100),
			currentLiquidShares: sdk.NewDec(90),
			newShares:           sdk.NewDec(100),
			expectedExceeds:     false,
		},
		{
			// Validator Shares: 100, Factor: 100, Current Shares: 90 => 10000 Max Shares, Capacity: 9910
			// New Shares: 9910 - at cap
			name:                "factor 100 - at cap",
			validatorShares:     sdk.NewDec(100),
			validatorBondFactor: sdk.NewDec(100),
			currentLiquidShares: sdk.NewDec(90),
			newShares:           sdk.NewDec(9910),
			expectedExceeds:     false,
		},
		{
			// Validator Shares: 100, Factor: 100, Current Shares: 90 => 10000 Max Shares, Capacity: 9910
			// New Shares: 9911 - above cap
			name:                "factor 100 - above cap",
			validatorShares:     sdk.NewDec(100),
			validatorBondFactor: sdk.NewDec(100),
			currentLiquidShares: sdk.NewDec(90),
			newShares:           sdk.NewDec(9911),
			expectedExceeds:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Update the validator bond factor
			params := app.StakingKeeper.GetParams(ctx)
			params.ValidatorBondFactor = tc.validatorBondFactor
			app.StakingKeeper.SetParams(ctx, params)

			// Create a validator with designated self-bond shares
			validator := types.Validator{
				TotalLiquidShares:        tc.currentLiquidShares,
				TotalValidatorBondShares: tc.validatorShares,
			}

			// Check whether the cap is exceeded
			actualExceeds := app.StakingKeeper.CheckExceedsValidatorBondCap(ctx, validator, tc.newShares)
			require.Equal(t, tc.expectedExceeds, actualExceeds, tc.name)
		})
	}
}

// Tests SafelyIncreaseValidatorTotalLiquidShares
func TestSafelyIncreaseValidatorTotalLiquidShares(t *testing.T) {
	_, app, ctx := createTestInput(t)

	initialLiquidShares := sdk.NewDec(0)
	validatorBondShares := sdk.NewDec(10)
	increaseAmount := sdk.NewDec(20)

	// Create a validator with designated self-bond shares
	privKey := secp256k1.GenPrivKey()
	pubKey := privKey.PubKey()
	valAddress := sdk.ValAddress(pubKey.Address())

	initialValidator := types.Validator{
		OperatorAddress:          valAddress.String(),
		TotalLiquidShares:        initialLiquidShares,
		TotalValidatorBondShares: validatorBondShares,
	}
	app.StakingKeeper.SetValidator(ctx, initialValidator)

	// Set validator bond factor so that the check is disabled
	params := app.StakingKeeper.GetParams(ctx)
	params.ValidatorBondFactor = sdk.NewDec(-1)
	app.StakingKeeper.SetParams(ctx, params)

	// Attempt to increase the validator liquid shares
	// it should do nothing since the check is disabled
	err := app.StakingKeeper.SafelyIncreaseValidatorTotalLiquidShares(ctx, initialValidator, increaseAmount)
	require.NoError(t, err)

	actualValidator, found := app.StakingKeeper.GetLiquidValidator(ctx, valAddress)
	require.True(t, found)
	require.Equal(t, initialLiquidShares, actualValidator.TotalLiquidShares, "shares with cap disabled")

	// Change validator bond factor so that it is enabled, but the delegation will fail
	params.ValidatorBondFactor = sdk.NewDec(1)
	app.StakingKeeper.SetParams(ctx, params)

	// Attempt to increase the validator liquid shares again, this time it should throw an
	// error that the cap was exceeded
	err = app.StakingKeeper.SafelyIncreaseValidatorTotalLiquidShares(ctx, initialValidator, increaseAmount)
	require.ErrorIs(t, err, types.ErrInsufficientValidatorBondShares)
	require.Equal(t, initialLiquidShares, actualValidator.TotalLiquidShares, "shares with strict cap")

	// Change validator bond factor one more time, so that the increase succeeds
	params.ValidatorBondFactor = sdk.NewDec(10)
	app.StakingKeeper.SetParams(ctx, params)

	// Finally, try the increase again and check that it succeeded
	err = app.StakingKeeper.SafelyIncreaseValidatorTotalLiquidShares(ctx, initialValidator, increaseAmount)
	require.NoError(t, err)

	actualValidator, found = app.StakingKeeper.GetLiquidValidator(ctx, valAddress)
	require.True(t, found)
	require.Equal(t, initialLiquidShares.Add(increaseAmount), actualValidator.TotalLiquidShares, "shares with loose cap")
}

// Tests DecreaseValidatorTotalLiquidShares
func TestDecreaseValidatorTotalLiquidShares(t *testing.T) {
	_, app, ctx := createTestInput(t)

	initialLiquidShares := sdk.NewDec(0)
	decreaseAmount := sdk.NewDec(10)

	// Create a validator with designated self-bond shares
	privKey := secp256k1.GenPrivKey()
	pubKey := privKey.PubKey()
	valAddress := sdk.ValAddress(pubKey.Address())

	initialValidator := types.Validator{
		OperatorAddress:   valAddress.String(),
		TotalLiquidShares: initialLiquidShares,
	}
	app.StakingKeeper.SetValidator(ctx, initialValidator)

	// Set validator bond factor so that the check is disabled
	params := app.StakingKeeper.GetParams(ctx)
	params.ValidatorBondFactor = sdk.NewDec(-1)
	app.StakingKeeper.SetParams(ctx, params)

	// Attempt to decrease the validator liquid shares, it should not be changed since the cap is disabled
	app.StakingKeeper.DecreaseValidatorTotalLiquidShares(ctx, initialValidator, decreaseAmount)
	actualValidator, found := app.StakingKeeper.GetLiquidValidator(ctx, valAddress)
	require.True(t, found)
	require.Equal(t, initialLiquidShares, actualValidator.TotalLiquidShares, "shares with cap disabled")

	// Now relax the cap to any positive number so that it is enabled
	params.ValidatorBondFactor = sdk.NewDec(10)
	app.StakingKeeper.SetParams(ctx, params)

	// Attempt to decrease the validator liquid shares again, now it should succeed
	app.StakingKeeper.DecreaseValidatorTotalLiquidShares(ctx, initialValidator, decreaseAmount)
	actualValidator, found = app.StakingKeeper.GetLiquidValidator(ctx, valAddress)
	require.True(t, found)
	require.Equal(t, initialLiquidShares.Sub(decreaseAmount), actualValidator.TotalLiquidShares, "shares with cap disabled")
}
