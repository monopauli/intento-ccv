package compute

import (
	"fmt"
	"time"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/trstlabs/trst/x/compute/internal/types"
)

// NewHandler returns a handler for "bank" type messages.
func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {

		case *MsgStoreCode:
			return handleStoreCode(ctx, k, msg)
		case *MsgInstantiateContract:
			return handleInstantiate(ctx, k, msg)
		case *MsgExecuteContract:
			return handleExecute(ctx, k, msg)
			/*
				case MsgMigrateContract:
					return handleMigration(ctx, k, &msg)
				case MsgUpdateAdmin:
					return handleUpdateContractAdmin(ctx, k, &msg)
				case MsgClearAdmin:
					return handleClearContractAdmin(ctx, k, &msg)
			*/
		default:
			errMsg := fmt.Sprintf("unrecognized wasm message type: %T", msg)
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, errMsg)
		}
	}
}

// filteredMessageEvents returns the same events with all of type == EventTypeMessage removed.
// this is so only our top-level message event comes through
func filteredMessageEvents(manager *sdk.EventManager) []abci.Event {
	events := manager.ABCIEvents()
	res := make([]abci.Event, 0, len(events)+1)
	for _, e := range events {
		if e.Type != sdk.EventTypeMessage {
			res = append(res, e)
		}
	}
	return res
}

func handleStoreCode(ctx sdk.Context, k Keeper, msg *MsgStoreCode) (*sdk.Result, error) {
	err := msg.ValidateBasic()
	if err != nil {
		return nil, err
	}
	activePeriod := k.GetParams(ctx).MaxActivePeriod
	//submitTime := ctx.BlockHeader().Time

	endTime := time.Hour * time.Duration(msg.ContractPeriod)
	maxEndTime := activePeriod

	if msg.ContractPeriod == 0 {
		endTime = maxEndTime
	}
	if endTime > maxEndTime {
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, "Time period invalid for this contract code")
	}
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}
	codeID, err := k.Create(ctx, sender, msg.WASMByteCode, msg.Source, msg.Builder, endTime, msg.Title, msg.Description)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, ModuleName),
			sdk.NewAttribute(types.AttributeKeySigner, msg.Sender),
			sdk.NewAttribute(types.AttributeKeyCodeID, fmt.Sprintf("%d", codeID)),
		),
	})

	return &sdk.Result{
		Data:   []byte(fmt.Sprintf("%d", codeID)),
		Events: ctx.EventManager().ABCIEvents(),
	}, nil
}

func handleInstantiate(ctx sdk.Context, k Keeper, msg *MsgInstantiateContract) (*sdk.Result, error) {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}
	contractAddr, err := k.Instantiate(ctx, msg.CodeID, sender, msg.InitMsg, msg.ContractId, msg.InitFunds, nil)
	if err != nil {
		return nil, err
	}

	events := filteredMessageEvents(ctx.EventManager())
	custom := sdk.Events{sdk.NewEvent(
		sdk.EventTypeMessage,
		sdk.NewAttribute(sdk.AttributeKeyModule, ModuleName),
		sdk.NewAttribute(types.AttributeKeySigner, msg.Sender),
		sdk.NewAttribute(types.AttributeKeyCodeID, fmt.Sprintf("%d", msg.CodeID)),
		sdk.NewAttribute(types.AttributeKeyContract, contractAddr.String()),
	)}
	events = append(events, custom.ToABCIEvents()...)

	/*activePeriod := k.GetParams(ctx).MaxActivePeriod
	submitTime := ctx.BlockHeader().Time
	endTime := submitTime.Add(activePeriod)
	var info *types.CodeInfo
	info = k.GetCodeInfo(ctx, msg.CodeID)
	if info.EndTime <= endTime {
		endTime = info.EndTime
	} else {
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, "Time period invalid for this contract code")

	}*/
	//var info *types.CodeInfo
	info := k.GetCodeInfo(ctx, msg.CodeID)
	submitTime := ctx.BlockHeader().Time
	endTime := submitTime.Add(info.EndTime)
	k.InsertContractQueue(ctx, contractAddr.String(), endTime)
	return &sdk.Result{
		Data:   contractAddr,
		Events: events,
	}, nil
}

func handleExecute(ctx sdk.Context, k Keeper, msg *MsgExecuteContract) (*sdk.Result, error) {
	fmt.Print("handling..")
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}
	fmt.Print("executing..")
	res, err := k.Execute(
		ctx,
		msg.Contract,
		sender,
		msg.Msg,
		msg.SentFunds,
		nil,
	)
	if err != nil {
		return nil, err
	}
	fmt.Print("setting..")
	k.SetContractResult(ctx, msg.Contract, res)

	events := filteredMessageEvents(ctx.EventManager())
	custom := sdk.Events{sdk.NewEvent(
		sdk.EventTypeMessage,
		sdk.NewAttribute(sdk.AttributeKeyModule, ModuleName),
		sdk.NewAttribute(types.AttributeKeySigner, msg.Sender),
		sdk.NewAttribute(types.AttributeKeyContract, msg.Contract.String()),
	)}
	events = append(events, custom.ToABCIEvents()...)
	fmt.Print("events Execute handled")
	res.Events = events

	return res, nil
}

/*
func handleDeleteContract(ctx sdk.Context, k Keeper, msg *MsgDeleteContract) (*sdk.Result, error) {
	res, err := k.DeleteContract(ctx, msg.Contract, msg.Sender, msg.CodeID, msg.DeleteContractMsg) // for MsgMigrateContract, there is only one signer which is msg.Sender (https://github.com/trstlabs/trst/blob/d7813792fa07b93a10f0885eaa4c5e0a0a698854/x/compute/internal/types/msg.go#L228-L230)
	if err != nil {
		return nil, err
	}

	events := filteredMessageEvents(ctx.EventManager())
	ourEvent := sdk.Events{sdk.NewEvent(
		sdk.EventTypeMessage,
		sdk.NewAttribute(sdk.AttributeKeyModule, ModuleName),
		sdk.NewAttribute(types.AttributeKeySigner, msg.Sender.String()),
		sdk.NewAttribute(types.AttributeKeyContract, msg.Contract.String()),
	)}
	res.Events = append(events, ourEvent.ToABCIEvents()...)
	return res, nil
}
*/
/*
func handleMigration(ctx sdk.Context, k Keeper, msg *MsgMigrateContract) (*sdk.Result, error) {
	res, err := k.Migrate(ctx, msg.Contract, msg.Sender, msg.CodeID, msg.MigrateMsg) // for MsgMigrateContract, there is only one signer which is msg.Sender (https://github.com/trstlabs/trst/blob/d7813792fa07b93a10f0885eaa4c5e0a0a698854/x/compute/internal/types/msg.go#L228-L230)
	if err != nil {
		return nil, err
	}

	events := filteredMessageEvents(ctx.EventManager())
	ourEvent := sdk.NewEvent(
		sdk.EventTypeMessage,
		sdk.NewAttribute(sdk.AttributeKeyModule, ModuleName),
		sdk.NewAttribute(types.AttributeKeySigner, msg.Sender.String()),
		sdk.NewAttribute(types.AttributeKeyContract, msg.Contract.String()),
	)
	res.Events = append(events, ourEvent)
	return res, nil
}

func handleUpdateContractAdmin(ctx sdk.Context, k Keeper, msg *MsgUpdateAdmin) (*sdk.Result, error) {
	if err := k.UpdateContractAdmin(ctx, msg.Contract, msg.Sender, msg.NewAdmin); err != nil {
		return nil, err
	}
	events := ctx.EventManager().Events()
	ourEvent := sdk.NewEvent(
		sdk.EventTypeMessage,
		sdk.NewAttribute(sdk.AttributeKeyModule, ModuleName),
		sdk.NewAttribute(types.AttributeKeySigner, msg.Sender.String()),
		sdk.NewAttribute(types.AttributeKeyContract, msg.Contract.String()),
	)
	return &sdk.Result{
		Events: append(events, ourEvent),
	}, nil
}

func handleClearContractAdmin(ctx sdk.Context, k Keeper, msg *MsgClearAdmin) (*sdk.Result, error) {
	if err := k.ClearContractAdmin(ctx, msg.Contract, msg.Sender); err != nil {
		return nil, err
	}
	events := ctx.EventManager().Events()
	ourEvent := sdk.NewEvent(
		sdk.EventTypeMessage,
		sdk.NewAttribute(sdk.AttributeKeyModule, ModuleName),
		sdk.NewAttribute(types.AttributeKeySigner, msg.Sender.String()),
		sdk.NewAttribute(types.AttributeKeyContract, msg.Contract.String()),
	)
	return &sdk.Result{
		Events: append(events, ourEvent),
	}, nil
}
*/
