package v3_test

import (
	"testing"
	"time"

	simapp "github.com/iqlusioninc/liquidity-staking-module/app"
	"github.com/stretchr/testify/require"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/mock"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	sdkstaking "github.com/cosmos/cosmos-sdk/x/staking/types"
	v2 "github.com/iqlusioninc/liquidity-staking-module/x/staking/migrations/v2"
	v3 "github.com/iqlusioninc/liquidity-staking-module/x/staking/migrations/v3"
	"github.com/iqlusioninc/liquidity-staking-module/x/staking/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtypes "github.com/tendermint/tendermint/types"
)

func TestStoreMigration(t *testing.T) {
	encCfg := simapp.MakeTestEncodingConfig()
	stakingKey := sdk.NewKVStoreKey("staking")
	tStakingKey := sdk.NewTransientStoreKey("transient_test")
	ctx := testutil.DefaultContext(stakingKey, tStakingKey)
	paramstore := paramtypes.NewSubspace(encCfg.Codec, encCfg.Amino, stakingKey, tStakingKey, "staking")

	// Check no params
	require.False(t, paramstore.Has(ctx, types.KeyMinCommissionRate))

	// add validator
	store := ctx.KVStore(stakingKey)

	privVal := mock.NewPV()
	pubKey, err := privVal.GetPubKey()
	require.NoError(t, err)

	val := tmtypes.NewValidator(pubKey, 1)
	pk, err := cryptocodec.FromTmPubKeyInterface(val.PubKey)
	require.NoError(t, err)
	pkAny, err := codectypes.NewAnyWithValue(pk)
	require.NoError(t, err)
	oldValidator := v2.Validator{
		OperatorAddress:         sdk.ValAddress(val.Address).String(),
		ConsensusPubkey:         pkAny,
		Jailed:                  false,
		Status:                  sdkstaking.Bonded,
		Tokens:                  sdk.NewInt(1000_000),
		DelegatorShares:         sdk.OneDec(),
		Description:             types.Description{},
		UnbondingHeight:         int64(0),
		UnbondingTime:           time.Unix(0, 0).UTC(),
		Commission:              types.NewCommission(sdk.ZeroDec(), sdk.ZeroDec(), sdk.ZeroDec()),
		MinSelfDelegation:       sdk.NewInt(1),
		UnbondingOnHoldRefCount: 1,
		UnbondingIds:            []uint64{1},
	}

	bz := encCfg.Codec.MustMarshal(&oldValidator)
	store.Set(types.GetValidatorKey(sdk.ValAddress(val.Address)), bz)

	// add delegation
	oldDelegation := v2.Delegation{
		DelegatorAddress: val.Address.String(),
		ValidatorAddress: sdk.ValAddress(val.Address).String(),
		Shares:           sdk.NewDec(1),
	}

	bz = encCfg.Codec.MustMarshal(&oldDelegation)
	store.Set(types.GetDelegationKey(sdk.AccAddress(val.Address), sdk.ValAddress(val.Address)), bz)

	// add historical info
	header := tmproto.Header{
		ChainID: "HelloChain",
		Height:  10,
	}
	oldHistoricalInfo := v2.HistoricalInfo{
		Header: header,
		Valset: []v2.Validator{oldValidator},
	}
	bz = encCfg.Codec.MustMarshal(&oldHistoricalInfo)
	store.Set(types.GetHistoricalInfoKey(1), bz)

	// Run migrations.
	err = v3.MigrateStore(ctx, stakingKey, encCfg.Codec, paramstore)
	require.NoError(t, err)

	// Make sure the new params are set.
	require.True(t, paramstore.Has(ctx, types.KeyMinCommissionRate))
	require.True(t, paramstore.Has(ctx, types.KeyExemptionFactor))

	// check validator
	bz = store.Get(types.GetValidatorKey(sdk.ValAddress(val.Address)))
	newValidator := types.Validator{}
	err = encCfg.Codec.Unmarshal(bz, &newValidator)
	require.NoError(t, err)

	//  check delegation
	bz = store.Get(types.GetDelegationKey(sdk.AccAddress(val.Address), sdk.ValAddress(val.Address)))
	newDelegation := types.Delegation{}
	err = encCfg.Codec.Unmarshal(bz, &newDelegation)
	require.NoError(t, err)

	// check historical info
	bz = store.Get(types.GetHistoricalInfoKey(1))
	newHistoricalInfo := types.HistoricalInfo{}
	err = encCfg.Codec.Unmarshal(bz, &newHistoricalInfo)
	require.NoError(t, err)
}
