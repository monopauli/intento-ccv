package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/trstlabs/intento/x/intent/types"
)

// IterateActionsQueue iterates over the items in the inactive action queue
// and performs a callback function
func (k Keeper) IterateActionQueue(ctx sdk.Context, execTime time.Time, cb func(action types.ActionInfo) (stop bool)) {
	iterator := k.ActionQueueIterator(ctx, execTime)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {

		actionID, _ := types.SplitActionQueueKey(iterator.Key())
		//fmt.Printf("action id is:  %v \n", actionID)
		action := k.GetActionInfo(ctx, actionID)

		if cb(action) {
			break
		}
	}
}

// GetActionsForBlock returns all expiring actions for a block
func (k Keeper) GetActionsForBlock(ctx sdk.Context) (actions []types.ActionInfo) {
	k.IterateActionQueue(ctx, ctx.BlockHeader().Time, func(action types.ActionInfo) bool {
		actions = append(actions, action)
		return false
	})
	return
}

// ActionQueueIterator returns an sdk.Iterator for all the items in the Inactive Queue that expire by execTime
func (k Keeper) ActionQueueIterator(ctx sdk.Context, execTime time.Time) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return store.Iterator(types.ActionQueuePrefix, sdk.PrefixEndBytes(types.ActionByTimeKey(execTime))) //we check the end of the bites array for the execution time
}

// InsertActionQueue Inserts a action into the auto tx queue
func (k Keeper) InsertActionQueue(ctx sdk.Context, actionID uint64, execTime time.Time) {
	store := ctx.KVStore(k.storeKey)
	bz := types.GetBytesForUint(actionID)

	//here the key is time+action appended (as bytes) and value is action in bytes
	store.Set(types.ActionQueueKey(actionID, execTime), bz)
}

// RemoveFromActionQueue removes a action from the Inactive Item Queue
func (k Keeper) RemoveFromActionQueue(ctx sdk.Context, action types.ActionInfo) {

	store := ctx.KVStore(k.storeKey)
	store.Delete(types.ActionQueueKey(action.ID, action.ExecTime))
}
