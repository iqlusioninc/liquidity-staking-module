package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/iqlusioninc/liquidity-staking-module/x/staking/types"
)

// SetTotalLiquidStakedTokens stores the total outstanding tokens owned by a liquid staking provider
func (k Keeper) SetTotalLiquidStakedTokens(ctx sdk.Context, tokens sdk.Int) {
	store := ctx.KVStore(k.storeKey)

	tokensBz, err := tokens.Marshal()
	if err != nil {
		panic(err)
	}

	store.Set(types.TotalLiquidStakedTokensKey, tokensBz)
}

// GetTotalLiquidStakedTokens returns the total outstanding tokens owned by a liquid staking provider
// Returns zero if the total liquid stake amount has not been initialized
func (k Keeper) GetTotalLiquidStakedTokens(ctx sdk.Context) sdk.Int {
	store := ctx.KVStore(k.storeKey)
	tokensBz := store.Get(types.TotalLiquidStakedTokensKey)

	if tokensBz == nil {
		// QUESTION: Should we panic here instead?
		// This is basically protecting against the case where we failed
		// to bootstrap the total liquid staked in the upgrade handler
		return sdk.ZeroInt()
	}

	var tokens sdk.Int
	if err := tokens.Unmarshal(tokensBz); err != nil {
		panic(err)
	}

	return tokens
}

// IncreaseTotalLiquidStakedTokens increments the total liquid staked tokens
func (k Keeper) IncreaseTotalLiquidStakedTokens(ctx sdk.Context, amount sdk.Int) {
	k.SetTotalLiquidStakedTokens(ctx, k.GetTotalLiquidStakedTokens(ctx).Add(amount))
}

// DecreaseTotalLiquidStakedTokens decrements the total liquid staked tokens
func (k Keeper) DecreaseTotalLiquidStakedTokens(ctx sdk.Context, amount sdk.Int) {
	k.SetTotalLiquidStakedTokens(ctx, k.GetTotalLiquidStakedTokens(ctx).Sub(amount))
}

// IncreaseValidatorTotalLiquidShares increments the total liquid shares on a validator
func (k Keeper) IncreaseValidatorTotalLiquidShares(ctx sdk.Context, validator types.Validator, shares sdk.Dec) {
	validator.TotalLiquidShares = validator.TotalLiquidShares.Add(shares)
	k.SetValidator(ctx, validator)
}

// DecreaseValidatorTotalLiquidShares decrements the total liquid shares on a validator
func (k Keeper) DecreaseValidatorTotalLiquidShares(ctx sdk.Context, validator types.Validator, shares sdk.Dec) {
	validator.TotalLiquidShares = validator.TotalLiquidShares.Sub(shares)
	k.SetValidator(ctx, validator)
}

// Check if an account is a owned by a liquid staking provider
// This is determined by checking if the account is a 32-length module account
func (k Keeper) AccountIsLiquidStakingProvider(ctx sdk.Context, address sdk.AccAddress) bool {
	account := k.authKeeper.GetAccount(ctx, address)
	_, isModuleAccount := account.(*authtypes.ModuleAccount)
	return isModuleAccount && len(address) == 32
}

// ExceedsGlobalLiquidStakingCap checks if a liquid delegation would cause the
// global liquid staking cap to be exceeded
// A liquid delegation is defined as either tokenized shares, or a delegation from an ICA Account
// The total stake is determined by the sum of the Bonded and NotBonded pools
// Returns true if the cap is exceeded
func (k Keeper) CheckExceedsGlobalLiquidStakingCap(ctx sdk.Context, tokens sdk.Int) bool {
	liquidStakingCap := k.GlobalLiquidStakingCap(ctx)
	liquidStakedAmount := k.GetTotalLiquidStakedTokens(ctx)

	// Determine the total stake as the sum of the bonded and not-bonded pools
	bondedPoolAddress := k.authKeeper.GetModuleAddress(types.BondedPoolName)
	notBondedPoolAddress := k.authKeeper.GetModuleAddress(types.NotBondedPoolName)

	bondedPoolBalance := k.bankKeeper.GetBalance(ctx, bondedPoolAddress, k.BondDenom(ctx)).Amount
	notBondedPoolBalance := k.bankKeeper.GetBalance(ctx, notBondedPoolAddress, k.BondDenom(ctx)).Amount

	totalStakedAmount := bondedPoolBalance.Add(notBondedPoolBalance)

	// Calculate the percentage of stake that is liquid
	updatedTotalStaked := sdk.NewDecFromInt(totalStakedAmount.Add(tokens))
	updatedLiquidStaked := sdk.NewDecFromInt(liquidStakedAmount.Add(tokens))
	liquidStakePercent := updatedLiquidStaked.Quo(updatedTotalStaked)

	return liquidStakePercent.GT(liquidStakingCap)
}

// ExceedsValidatorBondCap checks if a liquid delegation to a validator would cause
// their liquid shares to exceed their validator bond factor
// A liquid delegation is defined as either tokenized shares, or a delegation from an ICA Account
// Returns true if the cap is exceeded
func (k Keeper) CheckExceedsValidatorBondCap(ctx sdk.Context, validator types.Validator, shares sdk.Dec) bool {
	validatorBondFactor := k.ValidatorBondFactor(ctx)

	// If the validator bond factor is negative, the cap has been disabled
	if validatorBondFactor.IsNegative() {
		return false
	}

	maxValTotalShare := validator.TotalValidatorBondShares.Mul(validatorBondFactor)
	return validator.TotalLiquidShares.Add(shares).GT(maxValTotalShare)
}
