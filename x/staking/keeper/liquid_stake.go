package keeper

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/iqlusioninc/liquidity-staking-module/x/staking/types"
)

// SetTotalLiquidStaked stores the total outstanding shares owned by a liquid staking provider
func (k Keeper) SetTotalLiquidStaked(ctx sdk.Context, shares sdk.Dec) {
	store := ctx.KVStore(k.storeKey)

	sharesBz, err := shares.Marshal()
	if err != nil {
		panic(err)
	}

	store.Set(types.TotalLiquidStakedSharesKey, sharesBz)
}

// GetTotalLiquidStaked returns the total outstanding shares owned by a liquid staking provider
// Returns zero if the total liquid stake amount has not been initialized
func (k Keeper) GetTotalLiquidStaked(ctx sdk.Context) sdk.Dec {
	store := ctx.KVStore(k.storeKey)
	sharesBz := store.Get(types.TotalLiquidStakedSharesKey)

	if sharesBz == nil {
		return sdk.ZeroDec()
	}

	var shares sdk.Dec
	if err := shares.Unmarshal(sharesBz); err != nil {
		panic(err)
	}

	return shares
}

// IncreaseTotalLiquidStaked increments the total liquid staked shares
func (k Keeper) IncreaseTotalLiquidStaked(ctx sdk.Context, amount sdk.Dec) {
	k.SetTotalLiquidStaked(ctx, k.GetTotalLiquidStaked(ctx).Add(amount))
}

// DecreaseTotalLiquidStaked decrements the total liquid staked shares
func (k Keeper) DecreaseTotalLiquidStaked(ctx sdk.Context, amount sdk.Dec) {
	k.SetTotalLiquidStaked(ctx, k.GetTotalLiquidStaked(ctx).Sub(amount))
}

// Check if an account is a owned by a liquid staking provider
// This is determined by checking if the account is a module account
// that's not owned by the LSM module
// TODO: Verify the best way to check if an address belongs to a module account
func (k Keeper) AccountIsLiquidStakingProvider(ctx sdk.Context, address sdk.AccAddress) bool {
	// Check if the account is a module account
	// Module accounts are stored with 32-length addresses
	isModuleAccount := len(address) == 32
	if !isModuleAccount {
		return false
	}

	// If the address is a module account, check whether it is used by the staking module
	account := k.authKeeper.GetAccount(ctx, address)
	moduleAccount, isModuleAccount := account.(*authtypes.ModuleAccount)
	if !isModuleAccount {
		return false
	}
	isTokenizedSharesCustodian := strings.HasPrefix(moduleAccount.Name, types.TokenizeShareModuleAccountPrefix)

	return !isTokenizedSharesCustodian
}

// ExceedsGlobalLiquidStakingCap checks if a liquid delegation would cause the
// global liquid staking cap to be exceeded
// A liquid delegation is defined as either tokenized shares, or a delegation from an ICA Account
// The total stake is determined by the sum of the Bonded and NotBonded pools
// Returns true if the cap is exceeded
func (k Keeper) CheckExceedsGlobalLiquidStakingCap(ctx sdk.Context, shares sdk.Dec) bool {
	// TODO
	return false
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
