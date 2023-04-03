package keeper

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
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
	if err := json.Unmarshal(sharesBz, &shares); err != nil {
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
