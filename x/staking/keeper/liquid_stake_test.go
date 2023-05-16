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
			// => Total Liquid Stake: 20+31=51, Total Stake: 220+31=251 => 51/251 = 21% > 20% cap
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

	// Set the global cap such that a small delegation would exceed the cap
	params := app.StakingKeeper.GetParams(ctx)
	params.GlobalLiquidStakingCap = sdk.MustNewDecFromStr("0.0001")
	app.StakingKeeper.SetParams(ctx, params)

	// Attempt to increase the total liquid stake again,it should error since
	// the cap was exceeded
	err := app.StakingKeeper.SafelyIncreaseTotalLiquidStakedTokens(ctx, increaseAmount)
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

	// Decrease the total liquid stake and confirm the total was updated
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
		{
			// Factor of -1 (disabled): Should always return false
			name:                "factor disabled",
			validatorShares:     sdk.NewDec(1),
			validatorBondFactor: sdk.NewDec(-1),
			currentLiquidShares: sdk.NewDec(1),
			newShares:           sdk.NewDec(1_000_000),
			expectedExceeds:     false,
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

// Tests TestCheckExceedsValidatorLiquidStakingCap
func TestCheckExceedsValidatorLiquidStakingCap(t *testing.T) {
	_, app, ctx := createTestInput(t)

	testCases := []struct {
		name                  string
		validatorLiquidCap    sdk.Dec
		validatorLiquidShares sdk.Dec
		validatorTotalShares  sdk.Dec
		newLiquidShares       sdk.Dec
		expectedExceeds       bool
	}{
		{
			// Cap: 10% - Delegation Below Threshold
			// Liquid Shares: 5, Total Shares: 95, New Liquid Shares: 1
			// => Liquid Shares: 5+1=6, Total Shares: 95+1=96 => 6/96 = 6% < 10% cap
			name:                  "10 percent cap _ delegation below cap",
			validatorLiquidCap:    sdk.MustNewDecFromStr("0.1"),
			validatorLiquidShares: sdk.NewDec(5),
			validatorTotalShares:  sdk.NewDec(95),
			newLiquidShares:       sdk.NewDec(1),
			expectedExceeds:       false,
		},
		{
			// Cap: 10% - Delegation At Threshold
			// Liquid Shares: 5, Total Shares: 95, New Liquid Shares: 5
			// => Liquid Shares: 5+5=10, Total Shares: 95+5=100 => 10/100 = 10% == 10% cap
			name:                  "10 percent cap _ delegation equals cap",
			validatorLiquidCap:    sdk.MustNewDecFromStr("0.1"),
			validatorLiquidShares: sdk.NewDec(5),
			validatorTotalShares:  sdk.NewDec(95),
			newLiquidShares:       sdk.NewDec(4),
			expectedExceeds:       false,
		},
		{
			// Cap: 10% - Delegation Exceeds Threshold
			// Liquid Shares: 5, Total Shares: 95, New Liquid Shares: 6
			// => Liquid Shares: 5+6=11, Total Shares: 95+6=101 => 11/101 = 11% > 10% cap
			name:                  "10 percent cap _ delegation exceeds cap",
			validatorLiquidCap:    sdk.MustNewDecFromStr("0.1"),
			validatorLiquidShares: sdk.NewDec(5),
			validatorTotalShares:  sdk.NewDec(95),
			newLiquidShares:       sdk.NewDec(6),
			expectedExceeds:       true,
		},
		{
			// Cap: 20% - Delegation Below Threshold
			// Liquid Shares: 20, Total Shares: 220, New Liquid Shares: 29
			// => Liquid Shares: 20+29=49, Total Shares: 220+29=249 => 49/249 = 19% < 20% cap
			name:                  "20 percent cap _ delegation below cap",
			validatorLiquidCap:    sdk.MustNewDecFromStr("0.2"),
			validatorLiquidShares: sdk.NewDec(20),
			validatorTotalShares:  sdk.NewDec(220),
			newLiquidShares:       sdk.NewDec(29),
			expectedExceeds:       false,
		},
		{
			// Cap: 20% - Delegation At Threshold
			// Liquid Shares: 20, Total Shares: 220, New Liquid Shares: 30
			// => Liquid Shares: 20+30=50, Total Shares: 220+30=250 => 50/250 = 20% == 20% cap
			name:                  "20 percent cap _ delegation equals cap",
			validatorLiquidCap:    sdk.MustNewDecFromStr("0.2"),
			validatorLiquidShares: sdk.NewDec(20),
			validatorTotalShares:  sdk.NewDec(220),
			newLiquidShares:       sdk.NewDec(30),
			expectedExceeds:       false,
		},
		{
			// Cap: 20% - Delegation Exceeds Threshold
			// Liquid Shares: 20, Total Shares: 220, New Liquid Shares: 31
			// => Liquid Shares: 20+31=51, Total Shares: 220+31=251 => 51/251 = 21% > 20% cap
			name:                  "20 percent cap _ delegation exceeds cap",
			validatorLiquidCap:    sdk.MustNewDecFromStr("0.2"),
			validatorLiquidShares: sdk.NewDec(20),
			validatorTotalShares:  sdk.NewDec(220),
			newLiquidShares:       sdk.NewDec(31),
			expectedExceeds:       true,
		},
		{
			// Cap of 0% - everything should exceed
			name:                  "0 percent cap",
			validatorLiquidCap:    sdk.ZeroDec(),
			validatorLiquidShares: sdk.NewDec(0),
			validatorTotalShares:  sdk.NewDec(1_000_000),
			newLiquidShares:       sdk.NewDec(1),
			expectedExceeds:       true,
		},
		{
			// Cap of 100% - nothing should exceed
			name:                  "100 percent cap",
			validatorLiquidCap:    sdk.OneDec(),
			validatorLiquidShares: sdk.NewDec(1),
			validatorTotalShares:  sdk.NewDec(1_000_000),
			newLiquidShares:       sdk.NewDec(1),
			expectedExceeds:       false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Update the validator liquid staking cap
			params := app.StakingKeeper.GetParams(ctx)
			params.ValidatorLiquidStakingCap = tc.validatorLiquidCap
			app.StakingKeeper.SetParams(ctx, params)

			// Create a validator with designated self-bond shares
			validator := types.Validator{
				TotalLiquidShares: tc.validatorLiquidShares,
				DelegatorShares:   tc.validatorTotalShares,
			}

			// Check whether the cap is exceeded
			actualExceeds := app.StakingKeeper.CheckExceedsValidatorLiquidStakingCap(ctx, validator, tc.newLiquidShares)
			require.Equal(t, tc.expectedExceeds, actualExceeds, tc.name)
		})
	}
}

// Tests SafelyIncreaseValidatorTotalLiquidShares
func TestSafelyIncreaseValidatorTotalLiquidShares(t *testing.T) {
	_, app, ctx := createTestInput(t)

	// Generate a test validator address
	privKey := secp256k1.GenPrivKey()
	pubKey := privKey.PubKey()
	valAddress := sdk.ValAddress(pubKey.Address())

	// Helper function to check the validator's liquid shares
	checkValidatorLiquidShares := func(expected sdk.Dec, description string) {
		actualValidator, found := app.StakingKeeper.GetLiquidValidator(ctx, valAddress)
		require.True(t, found)
		require.Equal(t, expected.TruncateInt64(), actualValidator.TotalLiquidShares.TruncateInt64(), description)
	}

	// Start with the following:
	//   Initial Liquid Shares: 0
	//   Validator Bond Shares: 10
	//   Validator TotalShares: 75
	//
	// Initial Caps:
	//   ValidatorBondFactor: 1 (Cap applied at 10 shares)
	//   ValidatorLiquidStakingCap: 25% (Cap applied at 25 shares)
	//
	// Cap Increases:
	//   ValidatorBondFactor: 10 (Cap applied at 100 shares)
	//   ValidatorLiquidStakingCap: 40% (Cap applied at 50 shares)
	initialLiquidShares := sdk.NewDec(0)
	validatorBondShares := sdk.NewDec(10)
	validatorTotalShares := sdk.NewDec(75)

	firstIncreaseAmount := sdk.NewDec(20)
	secondIncreaseAmount := sdk.NewDec(40)

	initialBondFactor := sdk.NewDec(1)
	finalBondFactor := sdk.NewDec(10)
	initialLiquidStakingCap := sdk.MustNewDecFromStr("0.25")
	finalLiquidStakingCap := sdk.MustNewDecFromStr("0.4")

	// Create a validator with designated self-bond shares
	initialValidator := types.Validator{
		OperatorAddress:          valAddress.String(),
		TotalLiquidShares:        initialLiquidShares,
		TotalValidatorBondShares: validatorBondShares,
		DelegatorShares:          validatorTotalShares,
	}
	app.StakingKeeper.SetValidator(ctx, initialValidator)

	// Set validator bond factor to a small number such that any delegation would fail,
	// and set the liquid staking cap such that the first stake would succeed, but the second
	// would fail
	params := app.StakingKeeper.GetParams(ctx)
	params.ValidatorBondFactor = initialBondFactor
	params.ValidatorLiquidStakingCap = initialLiquidStakingCap
	app.StakingKeeper.SetParams(ctx, params)

	// Attempt to increase the validator liquid shares, it should throw an
	// error that the validator bond cap was exceeded
	err := app.StakingKeeper.SafelyIncreaseValidatorTotalLiquidShares(ctx, initialValidator, firstIncreaseAmount)
	require.ErrorIs(t, err, types.ErrInsufficientValidatorBondShares)
	checkValidatorLiquidShares(initialLiquidShares, "shares after low bond factor")

	// Change validator bond factor to a more conservative number, so that the increase succeeds
	params.ValidatorBondFactor = finalBondFactor
	app.StakingKeeper.SetParams(ctx, params)

	// Try the increase again and check that it succeeded
	expectedLiquidSharesAfterFirstStake := initialLiquidShares.Add(firstIncreaseAmount)
	err = app.StakingKeeper.SafelyIncreaseValidatorTotalLiquidShares(ctx, initialValidator, firstIncreaseAmount)
	require.NoError(t, err)
	checkValidatorLiquidShares(expectedLiquidSharesAfterFirstStake, "shares with cap loose bond cap")

	// Attempt another increase, it should fail from the liquid staking cap
	err = app.StakingKeeper.SafelyIncreaseValidatorTotalLiquidShares(ctx, initialValidator, secondIncreaseAmount)
	require.ErrorIs(t, err, types.ErrValidatorLiquidStakingCapExceeded)
	checkValidatorLiquidShares(expectedLiquidSharesAfterFirstStake, "shares after liquid staking cap hit")

	// Raise the liquid staking cap so the new increment succeeds
	params.ValidatorLiquidStakingCap = finalLiquidStakingCap
	app.StakingKeeper.SetParams(ctx, params)

	// Finally confirm that the increase succeeded this time
	expectedLiquidSharesAfterSecondStake := initialLiquidShares.Add(secondIncreaseAmount)
	err = app.StakingKeeper.SafelyIncreaseValidatorTotalLiquidShares(ctx, initialValidator, secondIncreaseAmount)
	require.NoError(t, err, "no error expected after increasing liquid staking cap")
	checkValidatorLiquidShares(expectedLiquidSharesAfterSecondStake, "shares after loose liquid stake cap")
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

	// Decrease the validator liquid shares, and confirm the new share amount has been updated
	app.StakingKeeper.DecreaseValidatorTotalLiquidShares(ctx, initialValidator, decreaseAmount)
	actualValidator, found := app.StakingKeeper.GetLiquidValidator(ctx, valAddress)
	require.True(t, found)
	require.Equal(t, initialLiquidShares.Sub(decreaseAmount), actualValidator.TotalLiquidShares, "shares with cap disabled")
}
