package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	epochkeeper "github.com/cosmos/cosmos-sdk/x/epoching/keeper"
)

// GetNewActionID returns ID to be used for next message queue item
func (k Keeper) GetNewActionID(ctx sdk.Context) uint64 {
	id := uint64(1)

	store := ctx.KVStore(k.storeKey)
	bz := store.Get(epochkeeper.NextEpochActionID)
	if bz != nil {
		id = sdk.BigEndianToUint64(bz)
	}

	// increment next action ID
	store.Set(epochkeeper.NextEpochActionID, sdk.Uint64ToBigEndian(id+1))
	return id
}

// QueueMsgForEpoch save the actions that need to be executed on next epoch
func (k Keeper) QueueMsgForEpoch(ctx sdk.Context, epochNumber int64, action sdk.Msg) {
	store := ctx.KVStore(k.storeKey)

	// reference from TestMarshalAny(t *testing.T)
	bz, err := k.cdc.MarshalInterface(action)
	if err != nil {
		panic(err)
	}

	actionID := k.GetNewActionID(ctx)
	store.Set(epochkeeper.ActionStoreKey(epochNumber, actionID), bz)
}

// GetNextEpochHeight returns next epoch block height
func (k Keeper) GetNextEpochHeight(ctx sdk.Context) int64 {
	return k.Keeper.GetNextEpochHeight(ctx, k.EpochInterval(ctx))
}

// GetNextEpochTime returns estimated next epoch time
func (k Keeper) GetNextEpochTime(ctx sdk.Context) time.Time {
	return k.Keeper.GetNextEpochTime(ctx, k.EpochInterval(ctx))
}
