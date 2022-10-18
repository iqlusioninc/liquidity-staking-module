package v3

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/iqlusioninc/liquidity-staking-module/x/staking/exported"
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
	migrateParamsStore(ctx, paramstore.(subspace))

	// TODO: update validator object
	// - remove string min_self_delegation = 11
	// + add string total_exempt_shares = 11
	// + add string total_tokenized_shares = 12
	// TODO: update delegation object
	// + add bool exempt = 4;
	// TODO: update params for
	// + add string exemption_factor = 7 [
	// TODO: HistoricalInfo for validator updates
	// TODO:  repeated TokenizeShareRecord tokenize_share_records = 9
	// TODO:  uint64 last_tokenize_share_record_id = 10;

	return nil
}

func migrateParamsStore(ctx sdk.Context, paramstore subspace) {
	if paramstore.HasKeyTable() {
		paramstore.Set(ctx, types.KeyMinCommissionRate, types.DefaultMinCommissionRate)
	} else {
		paramstore.WithKeyTable(types.ParamKeyTable())
		paramstore.Set(ctx, types.KeyMinCommissionRate, types.DefaultMinCommissionRate)
	}
}
