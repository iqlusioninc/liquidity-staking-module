package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	gogotypes "github.com/gogo/protobuf/types"

	"github.com/iqlusioninc/liquidity-staking-module/x/staking/types"
)

func (k Keeper) GetLastTokenizeShareRecordId(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bytes := store.Get(types.LastTokenizeShareRecordIdKey)
	if bytes == nil {
		return 0
	}
	return sdk.BigEndianToUint64(bytes)
}

func (k Keeper) SetLastTokenizeShareRecordId(ctx sdk.Context, id uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.LastTokenizeShareRecordIdKey, sdk.Uint64ToBigEndian(id))
}

func (k Keeper) GetTokenizeShareRecord(ctx sdk.Context, id uint64) (tokenizeShareRecord types.TokenizeShareRecord, err error) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.GetTokenizeShareRecordByIndexKey(id))
	if bz == nil {
		return tokenizeShareRecord, sdkerrors.Wrap(types.ErrTokenizeShareRecordNotExists, fmt.Sprintf("tokenizeShareRecord %d does not exist", id))
	}

	k.cdc.MustUnmarshal(bz, &tokenizeShareRecord)
	return tokenizeShareRecord, nil
}

func (k Keeper) GetTokenizeShareRecordsByOwner(ctx sdk.Context, owner sdk.AccAddress) (tokenizeShareRecords []types.TokenizeShareRecord, err error) {
	store := ctx.KVStore(k.storeKey)

	var it sdk.Iterator = sdk.KVStorePrefixIterator(store, types.GetTokenizeShareRecordsByOwnerKey(owner))
	defer it.Close()

	for ; it.Valid(); it.Next() {
		var id gogotypes.UInt64Value
		k.cdc.MustUnmarshal(it.Value(), &id)

		tokenizeShareRecord, err := k.GetTokenizeShareRecord(ctx, id.Value)
		if err != nil {
			continue
		}
		tokenizeShareRecords = append(tokenizeShareRecords, tokenizeShareRecord)
	}
	return
}

func (k Keeper) GetAllTokenizeShareRecords(ctx sdk.Context) (tokenizeShareRecords []types.TokenizeShareRecord) {
	store := ctx.KVStore(k.storeKey)

	var it sdk.Iterator = sdk.KVStorePrefixIterator(store, types.TokenizeShareRecordPrefix)
	defer it.Close()

	for ; it.Valid(); it.Next() {
		var tokenizeShareRecord types.TokenizeShareRecord
		k.cdc.MustUnmarshal(it.Value(), &tokenizeShareRecord)

		tokenizeShareRecords = append(tokenizeShareRecords, tokenizeShareRecord)
	}
	return
}

func (k Keeper) AddTokenizeShareRecord(ctx sdk.Context, tokenizeShareRecord types.TokenizeShareRecord) error {
	if k.hasTokenizeShareRecord(ctx, tokenizeShareRecord.Id) {
		return sdkerrors.Wrapf(types.ErrTokenizeShareRecordAlreadyExists, "TokenizeShareRecord already exists: %d", tokenizeShareRecord.Id)
	}

	k.setTokenizeShareRecord(ctx, tokenizeShareRecord)

	owner, err := sdk.AccAddressFromBech32(tokenizeShareRecord.Owner)
	if err != nil {
		return err
	}

	k.setTokenizeShareRecordWithOwner(ctx, owner, tokenizeShareRecord.Id)

	return nil
}

func (k Keeper) hasTokenizeShareRecord(ctx sdk.Context, id uint64) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.GetTokenizeShareRecordByIndexKey(id))
}

func (k Keeper) setTokenizeShareRecord(ctx sdk.Context, tokenizeShareRecord types.TokenizeShareRecord) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&tokenizeShareRecord)

	store.Set(types.GetTokenizeShareRecordByIndexKey(tokenizeShareRecord.Id), bz)
}

func (k Keeper) setTokenizeShareRecordWithOwner(ctx sdk.Context, owner sdk.AccAddress, id uint64) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&gogotypes.UInt64Value{Value: id})

	store.Set(types.SetTokenizeShareRecordByOwnerKey(owner, id), bz)
}
