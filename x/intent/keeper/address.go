package keeper

import (
	"encoding/binary"
	"fmt"
	"strings"

	errorsmod "cosmossdk.io/errors"
	"github.com/cometbft/cometbft/crypto"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"

	sdkaddress "github.com/cosmos/cosmos-sdk/types/address"
	"github.com/cosmos/gogoproto/proto"
	"github.com/trstlabs/intento/x/intent/types"
)

func (k Keeper) parseAndSetMsgs(ctx sdk.Context, flow *types.FlowInfo, connectionID, portID string) (protoMsgs []proto.Message, err error) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	if store.Has(types.GetFlowHistoryKey(flow.ID)) {
		txMsgs := flow.GetTxMsgs(k.cdc)

		protoMsgs = append(protoMsgs, txMsgs...)

		return protoMsgs, nil
	}

	var txMsgs []sdk.Msg
	var parsedIcaAddr bool

	for _, msg := range flow.Msgs {
		var txMsg sdk.Msg
		err := k.cdc.UnpackAny(msg, &txMsg)
		if err != nil {
			return nil, err
		}
		// Marshal the message into a JSON string
		msgJSON, err := k.cdc.MarshalInterfaceJSON(txMsg)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal %s message", msg)
		}
		msgJSONString := string(msgJSON)

		index := strings.Index(msgJSONString, types.ParseICAValue)
		if index == -1 {
			protoMsgs = append(protoMsgs, txMsg)
			txMsgs = append(txMsgs, txMsg)
			continue
		}

		ica, found := k.icaControllerKeeper.GetInterchainAccountAddress(ctx, connectionID, portID)
		if !found {
			return nil, errorsmod.Wrapf(types.ErrNotFound, "ICA address not found")
		}

		// Replace the text "ICA_ADDR" in the JSON string
		msgJSONString = strings.ReplaceAll(msgJSONString, types.ParseICAValue, ica)
		// Unmarshal the modified JSON string back into a proto message
		var updatedMsg sdk.Msg
		err = k.cdc.UnmarshalInterfaceJSON([]byte(msgJSONString), &updatedMsg)
		if err != nil {
			return nil, err
		}
		protoMsgs = append(protoMsgs, updatedMsg)

		txMsgs = append(txMsgs, updatedMsg)
		parsedIcaAddr = true

	}

	if parsedIcaAddr {
		anys, err := types.PackTxMsgAnys(txMsgs)
		if err != nil {
			return nil, err
		}
		flow.Msgs = anys
	}

	return protoMsgs, nil
}

func (k Keeper) createFeeAccount(ctx sdk.Context, id uint64, owner sdk.AccAddress, feeFunds sdk.Coins) (sdk.AccAddress, error) {
	flowAddress := k.generateFlowFeeAddress(ctx, id)
	existingAcct := k.accountKeeper.GetAccount(ctx, flowAddress)
	if existingAcct != nil {
		return nil, errorsmod.Wrap(types.ErrAccountExists, existingAcct.GetAddress().String())
	}

	// deposit initial flow funds
	if !feeFunds.IsZero() && !feeFunds[0].Amount.IsZero() {
		if k.bankKeeper.BlockedAddr(owner) {
			return nil, errorsmod.Wrap(types.ErrInvalidAddress, "blocked address can not be used")
		}
		sdkerr := k.bankKeeper.SendCoins(ctx, owner, flowAddress, feeFunds)
		if sdkerr != nil {
			return nil, sdkerr
		}
	} else {
		// create an empty account (so we don't have issues later)
		flowAccount := k.accountKeeper.NewAccountWithAddress(ctx, flowAddress)
		k.accountKeeper.NewAccount(ctx, flowAccount)
	}
	return flowAddress, nil
}

// generates a flow address from id + instanceID
func (k Keeper) generateFlowFeeAddress(ctx sdk.Context, id uint64) sdk.AccAddress {
	instanceID := k.autoIncrementID(ctx, types.KeyLastTxAddrID)
	return flowAddress(id, instanceID)
}

func flowAddress(id, instanceID uint64) sdk.AccAddress {
	// NOTE: It is possible to get a duplicate address if either id or instanceID
	// overflow 32 bits. This is highly improbable, but something that could be refactored.
	flowID := id<<32 + instanceID
	return addrFromUint64(flowID)

}

func (k Keeper) autoIncrementID(ctx sdk.Context, lastIDKey []byte) uint64 {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	bz := store.Get(lastIDKey)
	id := uint64(1)
	if bz != nil {
		id = binary.BigEndian.Uint64(bz)
	}
	bz = sdk.Uint64ToBigEndian(id + 1)
	store.Set(lastIDKey, bz)
	return id
}

func addrFromUint64(id uint64) sdk.AccAddress {
	addr := make([]byte, 20)
	addr[0] = 'C'
	binary.PutUvarint(addr[1:], id)
	return sdk.AccAddress(crypto.AddressHash(addr))
}

// simplied from https://github.com/cosmos/ibc-go/blob/main/modules/apps/27-interchain-accounts/types/account.go#L46
// to diferentiate between hosted icas
func DeriveHostedAddress(addressString string, connectionID string) (sdk.AccAddress, error) {
	addr, err := sdk.AccAddressFromBech32(addressString)
	if err != nil {
		return nil, err
	}
	buf := []byte(connectionID)
	return sdkaddress.Derive(addr, buf), nil

}
