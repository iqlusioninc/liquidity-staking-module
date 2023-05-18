package keeper_test

import (
	"testing"
	"time"

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

// Helper function to clear the Bonded pool balances before a unit test
func clearPoolBalance(t *testing.T, app *simapp.SimApp, ctx sdk.Context) {
	bondDenom := app.StakingKeeper.BondDenom(ctx)
	initialBondedBalance := app.BankKeeper.GetBalance(ctx, app.AccountKeeper.GetModuleAddress(types.BondedPoolName), bondDenom)

	err := app.BankKeeper.SendCoinsFromModuleToModule(ctx, types.BondedPoolName, minttypes.ModuleName, sdk.NewCoins(initialBondedBalance))
	require.NoError(t, err, "no error expected when clearing bonded pool balance")
}

// Helper function to fund the Bonded pool balances before a unit test
func fundPoolBalance(t *testing.T, app *simapp.SimApp, ctx sdk.Context, amount sdk.Int) {
	bondDenom := app.StakingKeeper.BondDenom(ctx)
	bondedPoolCoin := sdk.NewCoin(bondDenom, amount)

	err := app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, sdk.NewCoins(bondedPoolCoin))
	require.NoError(t, err, "no error expected when minting")

	err = app.BankKeeper.SendCoinsFromModuleToModule(ctx, minttypes.ModuleName, types.BondedPoolName, sdk.NewCoins(bondedPoolCoin))
	require.NoError(t, err, "no error expected when sending tokens to bonded pool")
}

// Tests CheckExceedsGlobalLiquidStakingCap
func TestCheckExceedsGlobalLiquidStakingCap(t *testing.T) {
	_, app, ctx := createTestInput(t)

	testCases := []struct {
		name             string
		globalLiquidCap  sdk.Dec
		totalLiquidStake sdk.Int
		totalStake       sdk.Int
		newLiquidStake   sdk.Int
		expectedExceeds  bool
	}{
		{
			// Cap: 10% - Delegation Below Threshold
			// Total Liquid Stake: 5, Total Stake: 95, New Liquid Stake: 1
			// => Total Liquid Stake: 5+1=6, Total Stake: 95+1=96 => 6/96 = 6% < 10% cap
			name:             "10 percent cap _ delegation below cap",
			globalLiquidCap:  sdk.MustNewDecFromStr("0.1"),
			totalLiquidStake: sdk.NewInt(5),
			totalStake:       sdk.NewInt(95),
			newLiquidStake:   sdk.NewInt(1),
			expectedExceeds:  false,
		},
		{
			// Cap: 10% - Delegation At Threshold
			// Total Liquid Stake: 5, Total Stake: 95, New Liquid Stake: 5
			// => Total Liquid Stake: 5+5=10, Total Stake: 95+5=100 => 10/100 = 10% == 10% cap
			name:             "10 percent cap _ delegation equals cap",
			globalLiquidCap:  sdk.MustNewDecFromStr("0.1"),
			totalLiquidStake: sdk.NewInt(5),
			totalStake:       sdk.NewInt(95),
			newLiquidStake:   sdk.NewInt(5),
			expectedExceeds:  false,
		},
		{
			// Cap: 10% - Delegation Exceeds Threshold
			// Total Liquid Stake: 5, Total Stake: 95, New Liquid Stake: 6
			// => Total Liquid Stake: 5+6=11, Total Stake: 95+6=101 => 11/101 = 11% > 10% cap
			name:             "10 percent cap _ delegation exceeds cap",
			globalLiquidCap:  sdk.MustNewDecFromStr("0.1"),
			totalLiquidStake: sdk.NewInt(5),
			totalStake:       sdk.NewInt(95),
			newLiquidStake:   sdk.NewInt(6),
			expectedExceeds:  true,
		},
		{
			// Cap: 20% - Delegation Below Threshold
			// Total Liquid Stake: 20, Total Stake: 220, New Liquid Stake: 29
			// => Total Liquid Stake: 20+29=49, Total Stake: 220+29=249 => 49/249 = 19% < 20% cap
			name:             "20 percent cap _ delegation below cap",
			globalLiquidCap:  sdk.MustNewDecFromStr("0.20"),
			totalLiquidStake: sdk.NewInt(20),
			totalStake:       sdk.NewInt(220),
			newLiquidStake:   sdk.NewInt(29),
			expectedExceeds:  false,
		},
		{
			// Cap: 20% - Delegation At Threshold
			// Total Liquid Stake: 20, Total Stake: 220, New Liquid Stake: 30
			// => Total Liquid Stake: 20+30=50, Total Stake: 220+30=250 => 50/250 = 20% == 20% cap
			name:             "20 percent cap _ delegation equals cap",
			globalLiquidCap:  sdk.MustNewDecFromStr("0.20"),
			totalLiquidStake: sdk.NewInt(20),
			totalStake:       sdk.NewInt(220),
			newLiquidStake:   sdk.NewInt(30),
			expectedExceeds:  false,
		},
		{
			// Cap: 20% - Delegation Exceeds Threshold
			// Total Liquid Stake: 20, Total Stake: 220, New Liquid Stake: 31
			// => Total Liquid Stake: 20+31=51, Total Total: 220+31=251 => 51/251 = 21% > 20% cap
			name:             "20 percent cap _ delegation exceeds cap",
			globalLiquidCap:  sdk.MustNewDecFromStr("0.20"),
			totalLiquidStake: sdk.NewInt(20),
			totalStake:       sdk.NewInt(220),
			newLiquidStake:   sdk.NewInt(31),
			expectedExceeds:  true,
		},
		{
			// Cap of 0% - everything should exceed
			name:             "0 percent cap",
			globalLiquidCap:  sdk.ZeroDec(),
			totalLiquidStake: sdk.NewInt(0),
			totalStake:       sdk.NewInt(1_000_000),
			newLiquidStake:   sdk.NewInt(1),
			expectedExceeds:  true,
		},
		{
			// Cap of 100% - nothing should exceed
			name:             "100 percent cap",
			globalLiquidCap:  sdk.OneDec(),
			totalLiquidStake: sdk.NewInt(1),
			totalStake:       sdk.NewInt(1),
			newLiquidStake:   sdk.NewInt(1_000_000),
			expectedExceeds:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Update the global liquid staking cap
			params := app.StakingKeeper.GetParams(ctx)
			params.GlobalLiquidStakingCap = tc.globalLiquidCap
			app.StakingKeeper.SetParams(ctx, params)

			// Update the total liquid tokens
			app.StakingKeeper.SetTotalLiquidStakedTokens(ctx, tc.totalLiquidStake)

			// Fund each pool for the given test case
			clearPoolBalance(t, app, ctx)
			fundPoolBalance(t, app, ctx, tc.totalStake)

			// Check if the new tokens would exceed the global cap
			actualExceeds := app.StakingKeeper.CheckExceedsGlobalLiquidStakingCap(ctx, tc.newLiquidStake)
			require.Equal(t, tc.expectedExceeds, actualExceeds, tc.name)
		})
	}
}

// Tests SafelyIncreaseTotalLiquidStakedTokens
func TestSafelyIncreaseTotalLiquidStakedTokens(t *testing.T) {
	_, app, ctx := createTestInput(t)

	intitialTotalLiquidStaked := sdk.NewInt(100)
	increaseAmount := sdk.NewInt(10)
	poolBalance := sdk.NewInt(200)

	// Set the total staked and total liquid staked amounts
	// which are required components when checking the global cap
	// Total stake is calculated from the pool balance
	clearPoolBalance(t, app, ctx)
	fundPoolBalance(t, app, ctx, poolBalance)
	app.StakingKeeper.SetTotalLiquidStakedTokens(ctx, intitialTotalLiquidStaked)

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

// Tests Add/Remove/SetTokenizeSharesLock and IsTokenizeSharesDisabled
func TestTokenizeSharesLock(t *testing.T) {
	_, app, ctx := createTestInput(t)

	addresses := simapp.AddTestAddrs(app, ctx, 2, sdk.NewInt(1))
	addressA, addressB := addresses[0], addresses[1]

	// Confirm both accounts start unlocked
	disabled, _ := app.StakingKeeper.IsTokenizeSharesDisabled(ctx, addressA)
	require.False(t, disabled, "addressA unlocked at start")

	disabled, _ = app.StakingKeeper.IsTokenizeSharesDisabled(ctx, addressB)
	require.False(t, disabled, "addressB unlocked at start")

	// Lock the first account
	app.StakingKeeper.AddTokenizeSharesLock(ctx, addressA)

	// The first account should now have tokenize shares disabled
	// and the unlock time should be the zero time
	disabled, actualUnlockTime := app.StakingKeeper.IsTokenizeSharesDisabled(ctx, addressA)
	require.True(t, disabled, "addressA locked")
	require.True(t, actualUnlockTime.IsZero(), "addressA unlock time uninititalized")

	disabled, _ = app.StakingKeeper.IsTokenizeSharesDisabled(ctx, addressB)
	require.False(t, disabled, "addressB still unlocked")

	// Update the lock time and confirm it was set
	expectedUnlockTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	app.StakingKeeper.SetTokenizeShareUnlockTime(ctx, addressA, expectedUnlockTime)

	disabled, actualUnlockTime = app.StakingKeeper.IsTokenizeSharesDisabled(ctx, addressA)
	require.True(t, disabled, "addressA still locked")
	require.Equal(t, expectedUnlockTime, actualUnlockTime, "addressA unlock time")

	// Confirm B is still unlocked
	disabled, _ = app.StakingKeeper.IsTokenizeSharesDisabled(ctx, addressB)
	require.False(t, disabled, "addressB still unlocked")

	// Remove the lock
	app.StakingKeeper.RemoveTokenizeSharesLock(ctx, addressA)
	disabled, _ = app.StakingKeeper.IsTokenizeSharesDisabled(ctx, addressA)
	require.False(t, disabled, "addressA unlocked at end")

	disabled, _ = app.StakingKeeper.IsTokenizeSharesDisabled(ctx, addressB)
	require.False(t, disabled, "addressB unlocked at end")
}

// Test Get/SetPendingTokenizeShareAuthorizations
func TestPendingTokenizeShareAuthorizations(t *testing.T) {
	_, app, ctx := createTestInput(t)

	// Create dummy accounts and completion times
	addresses := simapp.AddTestAddrs(app, ctx, 3, sdk.NewInt(1))
	addressStrings := []string{}
	for _, address := range addresses {
		addressStrings = append(addressStrings, address.String())
	}

	timeA := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	timeB := timeA.Add(time.Hour)

	// There should be no addresses returned originally
	authorizationsA := app.StakingKeeper.GetPendingTokenizeShareAuthorizations(ctx, timeA)
	require.Empty(t, authorizationsA.Addresses, "no addresses at timeA expected")

	authorizationsB := app.StakingKeeper.GetPendingTokenizeShareAuthorizations(ctx, timeB)
	require.Empty(t, authorizationsB.Addresses, "no addresses at timeB expected")

	// Store addresses for timeB
	app.StakingKeeper.SetPendingTokenizeShareAuthorizations(ctx, timeB, types.PendingTokenizeShareAuthorizations{
		Addresses: addressStrings,
	})

	// Check addresses
	authorizationsA = app.StakingKeeper.GetPendingTokenizeShareAuthorizations(ctx, timeA)
	require.Empty(t, authorizationsA.Addresses, "no addresses at timeA expected at end")

	authorizationsB = app.StakingKeeper.GetPendingTokenizeShareAuthorizations(ctx, timeB)
	require.Equal(t, addressStrings, authorizationsB.Addresses, "address length")
}

// Test QueueTokenizeSharesAuthorization and RemoveExpiredTokenizeShareLocks
func TestTokenizeShareAuthorizationQueue(t *testing.T) {
	_, app, ctx := createTestInput(t)

	// We'll start by adding the following addresses to the queue
	//   Time 0: [address0]
	//   Time 1: []
	//   Time 2: [address1, address2, address3]
	//   Time 3: [address4, address5]
	//   Time 4: [address6]
	addresses := simapp.AddTestAddrs(app, ctx, 7, sdk.NewInt(1))
	addressesByTime := map[int][]sdk.AccAddress{
		0: {addresses[0]},
		1: {},
		2: {addresses[1], addresses[2], addresses[3]},
		3: {addresses[4], addresses[5]},
		4: {addresses[6]},
	}

	// Set the unbonding time to 1 day
	unbondingPeriod := time.Hour * 24
	params := app.StakingKeeper.GetParams(ctx)
	params.UnbondingTime = unbondingPeriod
	app.StakingKeeper.SetParams(ctx, params)

	// Add each address to the queue and then increment the block time
	// such that the times line up as follows
	//   Time 0: 2023-01-01 00:00:00
	//   Time 1: 2023-01-01 00:01:00
	//   Time 2: 2023-01-01 00:02:00
	//   Time 3: 2023-01-01 00:03:00
	startTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	ctx = ctx.WithBlockTime(startTime)
	blockTimeIncrement := time.Hour

	for timeIndex := 0; timeIndex <= 4; timeIndex++ {
		for _, address := range addressesByTime[timeIndex] {
			app.StakingKeeper.QueueTokenizeSharesAuthorization(ctx, address)
		}
		ctx = ctx.WithBlockTime(ctx.BlockTime().Add(blockTimeIncrement))
	}

	// We'll unlock the tokens using the following progression
	// The "alias'"/keys for these times assume a starting point of the Time 0
	// from above, plus the Unbonding Time
	//   Time -1  (2023-01-01 23:59:99): []
	//   Time  0  (2023-01-02 00:00:00): [address0]
	//   Time  1  (2023-01-02 00:01:00): []
	//   Time 2.5 (2023-01-02 00:02:30): [address1, address2, address3]
	//   Time 10  (2023-01-02 00:10:00): [address4, address5, address6]
	unlockBlockTimes := map[string]time.Time{
		"-1":  startTime.Add(unbondingPeriod).Add(-time.Second),
		"0":   startTime.Add(unbondingPeriod),
		"1":   startTime.Add(unbondingPeriod).Add(blockTimeIncrement),
		"2.5": startTime.Add(unbondingPeriod).Add(2 * blockTimeIncrement).Add(blockTimeIncrement / 2),
		"10":  startTime.Add(unbondingPeriod).Add(10 * blockTimeIncrement),
	}
	expectedUnlockedAddresses := map[string][]string{
		"-1":  {},
		"0":   {addresses[0].String()},
		"1":   {},
		"2.5": {addresses[1].String(), addresses[2].String(), addresses[3].String()},
		"10":  {addresses[4].String(), addresses[5].String(), addresses[6].String()},
	}

	// Now we'll remove items from the queue sequentially
	// First check with a block time before the first expiration - it should remove no addresses
	actualAddresses := app.StakingKeeper.RemoveExpiredTokenizeShareLocks(ctx, unlockBlockTimes["-1"])
	require.Equal(t, expectedUnlockedAddresses["-1"], actualAddresses, "no addresses unlocked from time -1")

	// Then pass in (time 0 + unbonding time) - it should remove the first address
	actualAddresses = app.StakingKeeper.RemoveExpiredTokenizeShareLocks(ctx, unlockBlockTimes["0"])
	require.Equal(t, expectedUnlockedAddresses["0"], actualAddresses, "one address unlocked from time 0")

	// Now pass in (time 1 + unbonding time) - it should remove no addresses since
	// the address at time 0 was already removed
	actualAddresses = app.StakingKeeper.RemoveExpiredTokenizeShareLocks(ctx, unlockBlockTimes["1"])
	require.Equal(t, expectedUnlockedAddresses["1"], actualAddresses, "no addresses unlocked from time 1")

	// Now pass in (time 2.5 + unbonding time) - it should remove the three addresses from time 2
	actualAddresses = app.StakingKeeper.RemoveExpiredTokenizeShareLocks(ctx, unlockBlockTimes["2.5"])
	require.Equal(t, expectedUnlockedAddresses["2.5"], actualAddresses, "addresses unlocked from time 2.5")

	// Finally pass in a block time far in the future, which should remove all the remaining locks
	actualAddresses = app.StakingKeeper.RemoveExpiredTokenizeShareLocks(ctx, unlockBlockTimes["10"])
	require.Equal(t, expectedUnlockedAddresses["10"], actualAddresses, "addresses unlocked from time 10")
}
