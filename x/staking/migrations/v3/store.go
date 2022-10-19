package v3

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/iqlusioninc/liquidity-staking-module/x/staking/exported"
	v2 "github.com/iqlusioninc/liquidity-staking-module/x/staking/migrations/v2"
	"github.com/iqlusioninc/liquidity-staking-module/x/staking/types"
)

// subspace contains the method needed for migrations of the
// legacy Params subspace
type subspace interface {
	GetParamSet(ctx sdk.Context, ps paramtypes.ParamSet)
	HasKeyTable() bool
	WithKeyTable(paramtypes.KeyTable) paramtypes.Subspace
	Set(ctx sdk.Context, key []byte, value interface{})
}

// MigrateStore performs in-place store migrations from v0.43/v0.44/v0.45 to v0.46.
// The migration includes:
//
// - Setting the MinCommissionRate param in the paramstore
func MigrateStore(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec, paramstore exported.Subspace) error {
	store := ctx.KVStore(storeKey)
	migrateParamsStore(ctx, paramstore.(subspace))
	migrateValidators(store, cdc)
	migrateDelegations(store, cdc)
	migrateHistoricalInfos(store, cdc)

	return nil
}

func migrateParamsStore(ctx sdk.Context, paramstore subspace) {
	if paramstore.HasKeyTable() {
		paramstore.Set(ctx, types.KeyMinCommissionRate, types.DefaultMinCommissionRate)
		paramstore.Set(ctx, types.KeyExemptionFactor, types.DefaultExemptionFactor)
	} else {
		paramstore.WithKeyTable(types.ParamKeyTable())
		paramstore.Set(ctx, types.KeyMinCommissionRate, types.DefaultMinCommissionRate)
		paramstore.Set(ctx, types.KeyExemptionFactor, types.DefaultExemptionFactor)
	}
}

func getMigratedValidator(val v2.Validator) types.Validator {
	return types.Validator{
		OperatorAddress:      val.OperatorAddress,
		ConsensusPubkey:      val.ConsensusPubkey,
		Jailed:               val.Jailed,
		Status:               val.Status,
		Tokens:               val.Tokens,
		DelegatorShares:      val.DelegatorShares,
		Description:          val.Description,
		UnbondingHeight:      val.UnbondingHeight,
		UnbondingTime:        val.UnbondingTime,
		Commission:           val.Commission,
		TotalExemptShares:    sdk.ZeroDec(),
		TotalTokenizedShares: sdk.ZeroDec(),
	}
}
func migrateValidators(store sdk.KVStore, cdc codec.BinaryCodec) {
	oldStore := prefix.NewStore(store, types.ValidatorsKey)

	storeIter := oldStore.Iterator(nil, nil)
	defer storeIter.Close()

	for ; storeIter.Valid(); storeIter.Next() {
		val := v2.Validator{}
		cdc.MustUnmarshal(storeIter.Value(), &val)

		validator := getMigratedValidator(val)
		bz := cdc.MustMarshal(&validator)
		store.Set(storeIter.Key(), bz)
	}
}

func migrateDelegations(store sdk.KVStore, cdc codec.BinaryCodec) {
	oldStore := prefix.NewStore(store, types.DelegationKey)

	storeIter := oldStore.Iterator(nil, nil)
	defer storeIter.Close()

	for ; storeIter.Valid(); storeIter.Next() {
		del := v2.Delegation{}
		cdc.MustUnmarshal(storeIter.Value(), &del)

		delegation := types.Delegation{
			DelegatorAddress: del.DelegatorAddress,
			ValidatorAddress: del.ValidatorAddress,
			Shares:           del.Shares,
			Exempt:           false,
		}

		bz := cdc.MustMarshal(&delegation)
		store.Set(storeIter.Key(), bz)
	}
}

func migrateHistoricalInfos(store sdk.KVStore, cdc codec.BinaryCodec) {
	oldStore := prefix.NewStore(store, types.HistoricalInfoKey)

	storeIter := oldStore.Iterator(nil, nil)
	defer storeIter.Close()

	for ; storeIter.Valid(); storeIter.Next() {
		info := v2.HistoricalInfo{}
		cdc.MustUnmarshal(storeIter.Value(), &info)

		valSet := []types.Validator{}
		for _, val := range info.Valset {
			valSet = append(valSet, getMigratedValidator(val))
		}
		historicalInfo := types.HistoricalInfo{
			Header: info.Header,
			Valset: valSet,
		}

		bz := cdc.MustMarshal(&historicalInfo)
		store.Set(storeIter.Key(), bz)
	}
}
