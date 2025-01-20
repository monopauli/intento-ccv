package keeper

import (
	"fmt"
	"strings"
	"time"

	errorsmod "cosmossdk.io/errors"
	storetypes "cosmossdk.io/store/types"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authztypes "github.com/cosmos/cosmos-sdk/x/authz"
	"github.com/cosmos/gogoproto/proto"
	icacontrollerkeeper "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/controller/keeper"
	icacontrollertypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/controller/types"
	icatypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/types"
	"github.com/trstlabs/intento/x/intent/types"
)

func (k Keeper) TriggerAction(ctx sdk.Context, action *types.ActionInfo) (bool, []*cdctypes.Any, error) {
	// local action
	if (action.ICAConfig == nil || action.ICAConfig.ConnectionID == "") && (action.HostedConfig == nil || action.HostedConfig.HostedAddress == "") {
		txMsgs := action.GetTxMsgs(k.cdc)
		msgResponses, err := handleLocalAction(k, ctx, txMsgs, *action)
		return err == nil, msgResponses, errorsmod.Wrap(err, "could execute local action")
	}

	connectionID := action.ICAConfig.ConnectionID
	portID := action.ICAConfig.PortID
	triggerAddress := action.Owner
	//get hosted account from hosted config
	if action.HostedConfig != nil && action.HostedConfig.HostedAddress != "" {
		hostedAccount := k.GetHostedAccount(ctx, action.HostedConfig.HostedAddress)
		connectionID = hostedAccount.ICAConfig.ConnectionID
		portID = hostedAccount.ICAConfig.PortID
		triggerAddress = hostedAccount.HostedAddress
		err := k.SendFeesToHosted(ctx, *action, hostedAccount)
		if err != nil {
			return false, nil, errorsmod.Wrap(err, "could not pay hosted account")
		}

	}

	//check channel is active
	channelID, found := k.icaControllerKeeper.GetActiveChannelID(ctx, connectionID, portID)
	if !found {
		return false, nil, icatypes.ErrActiveChannelNotFound
	}

	//if a message contains "ICA_ADDR" string, the ICA address for the action is retrieved and parsed
	txMsgs, err := k.parseAndSetMsgs(ctx, action, connectionID, portID)
	if err != nil {
		return false, nil, errorsmod.Wrap(err, "could parse and set messages")
	}
	data, err := icatypes.SerializeCosmosTx(k.cdc, txMsgs, icatypes.EncodingProtobuf)
	if err != nil {
		return false, nil, err
	}
	packetData := icatypes.InterchainAccountPacketData{
		Type: icatypes.EXECUTE_TX,
		Data: data,
	}

	relativeTimeoutTimestamp := uint64(time.Minute.Nanoseconds())

	msgServer := icacontrollerkeeper.NewMsgServerImpl(&k.icaControllerKeeper)
	icaMsg := icacontrollertypes.NewMsgSendTx(triggerAddress, connectionID, relativeTimeoutTimestamp, packetData)

	res, err := msgServer.SendTx(ctx, icaMsg)
	if err != nil {
		return false, nil, errorsmod.Wrap(err, "could not send ICA message")
	}

	k.Logger(ctx).Debug("action", "ibc_sequence", res.Sequence)
	k.setTmpActionID(ctx, action.ID, portID, channelID, res.Sequence)
	return false, nil, nil
}

func handleLocalAction(k Keeper, ctx sdk.Context, txMsgs []sdk.Msg, action types.ActionInfo) ([]*cdctypes.Any, error) {
	// CacheContext returns a new context with the multi-store branched into a cached storage object
	// writeCache is called only if all msgs succeed, performing state transitions atomically
	var msgResponses []*cdctypes.Any

	cacheCtx, writeCache := ctx.CacheContext()
	for index, msg := range txMsgs {
		if action.Msgs[index].TypeUrl == "/ibc.applications.transfer.v1.MsgTransfer" {
			transferMsg, err := types.GetTransferMsg(k.cdc, action.Msgs[index])
			if err != nil {
				return nil, err
			}
			_, err = k.transferKeeper.Transfer(ctx, &transferMsg)
			if err != nil {
				return nil, err
			}
			continue
		}

		handler := k.msgRouter.Handler(msg)

		signers, _, err := k.cdc.GetMsgV1Signers(msg)
		if err != nil {
			return nil, err
		}
		for _, acct := range signers {
			if sdk.AccAddress(acct).String() != action.Owner {
				return nil, errorsmod.Wrap(types.ErrUnauthorized, "owner doesn't have permission to send this message")
			}
		}

		res, err := handler(cacheCtx, msg)
		if err != nil {
			return nil, err
		}

		msgResponses = append(msgResponses, res.MsgResponses...)

	}
	writeCache()
	if !action.Configuration.SaveResponses {
		msgResponses = nil
	}
	return msgResponses, nil
}

// HandleResponseAndSetActionResult sets the result of the last executed ID set at SendAction.
func (k Keeper) HandleResponseAndSetActionResult(ctx sdk.Context, portID string, channelID string, relayer sdk.AccAddress, seq uint64, msgResponses []*cdctypes.Any) error {
	id := k.getTmpActionID(ctx, portID, channelID, seq)
	if id <= 0 {
		return nil
	}
	action := k.GetActionInfo(ctx, id)

	actionHistoryEntry, newErr := k.GetLatestActionHistoryEntry(ctx, id)
	if newErr != nil {
		actionHistoryEntry.Errors = append(actionHistoryEntry.Errors, newErr.Error())
	}

	msgResponses, msgClass, err := k.HandleDeepResponses(ctx, msgResponses, relayer, action, len(actionHistoryEntry.MsgResponses))
	if err != nil {
		return err
	}

	k.UpdateActionIbcUsage(ctx, action)
	owner, err := sdk.AccAddressFromBech32(action.Owner)
	if err != nil {
		return err
	}
	// reward hooks
	if msgClass == 3 {
		k.hooks.AfterActionLocal(ctx, owner)
	} else if msgClass == 1 {
		k.hooks.AfterActionICA(ctx, owner)
	}

	actionHistoryEntry.Executed = true

	if action.Configuration.SaveResponses {
		actionHistoryEntry.MsgResponses = append(actionHistoryEntry.MsgResponses, msgResponses...)
	}

	// Refactor to handle multiple FeedbackLoops
	if len(action.Conditions.FeedbackLoops) != 0 {
		for _, feedbackLoop := range action.Conditions.FeedbackLoops {
			// Validate MsgsIndex and ActionID
			if feedbackLoop.MsgsIndex == 0 || feedbackLoop.ActionID != 0 {
				continue // Skip invalid FeedbackLoops or if ActionID is set to non-default
			}

			// Ensure MsgsIndex is within bounds
			if len(actionHistoryEntry.MsgResponses)-1 < int(feedbackLoop.MsgsIndex) {
				continue // Skip if MsgsIndex exceeds available responses
			}

			// Trigger remaining execution for the valid FeedbackLoop
			tmpAction := action
			tmpAction.Msgs = action.Msgs[feedbackLoop.MsgsIndex:]

			if err := triggerRemainingMsgs(k, ctx, tmpAction, actionHistoryEntry); err != nil {
				return err // Return on the first encountered error
			}
		}
	}

	k.SetCurrentActionHistoryEntry(ctx, action.ID, actionHistoryEntry)
	return nil

}

func triggerRemainingMsgs(k Keeper, ctx sdk.Context, action types.ActionInfo, actionHistoryEntry *types.ActionHistoryEntry) error {
	var errorString = ""

	allowed, err := k.allowedToExecute(ctx, action)
	if !allowed {
		k.recordActionNotAllowed(ctx, &action, ctx.BlockTime(), err)

	}

	actionCtx := ctx.WithGasMeter(storetypes.NewGasMeter(types.MaxGas))
	cacheCtx, writeCtx := actionCtx.CacheContext()
	k.Logger(ctx).Debug("continuing msg execution", "id", action.ID)

	feeAddr, feeDenom, err := k.GetFeeAccountForMinFees(cacheCtx, action, types.MaxGas)
	if err != nil {
		errorString = appendError(errorString, err.Error())
	} else if feeAddr == nil || feeDenom == "" {
		errorString = appendError(errorString, (types.ErrBalanceTooLow + feeDenom))
	}

	err = k.RunFeedbackLoops(cacheCtx, action.ID, &action.Msgs, action.Conditions)
	if err != nil {
		return errorsmod.Wrap(err, fmt.Sprintf(types.ErrSettingActionResult, err))
	}

	k.Logger(ctx).Debug("triggering msgs", "id", action.ID, "msgs", len(action.Msgs))
	_, _, err = k.TriggerAction(cacheCtx, &action)
	if err != nil {
		errorString = appendError(errorString, fmt.Sprintf(types.ErrActionMsgHandling, err.Error()))
	}
	fee, err := k.DistributeCoins(cacheCtx, action, feeAddr, feeDenom, ctx.BlockHeader().ProposerAddress)
	if err != nil {
		errorString = appendError(errorString, fmt.Sprintf(types.ErrActionFeeDistribution, err.Error()))
	}
	actionHistoryEntry.ExecFee = actionHistoryEntry.ExecFee.Add(fee)

	if errorString != "" {
		actionHistoryEntry.Executed = false
		actionHistoryEntry.Errors = append(actionHistoryEntry.Errors, types.ErrActionMsgHandling+err.Error())

	}
	k.SetCurrentActionHistoryEntry(cacheCtx, action.ID, actionHistoryEntry)
	writeCtx()
	return nil
}

func (k Keeper) HandleDeepResponses(ctx sdk.Context, msgResponses []*cdctypes.Any, relayer sdk.AccAddress, action types.ActionInfo, previousMsgsExecuted int) ([]*cdctypes.Any, int, error) {
	var msgClass int

	for index, anyResp := range msgResponses {
		k.Logger(ctx).Debug("msg response in ICS-27 packet", "response", anyResp.GoString(), "typeURL", anyResp.GetTypeUrl())

		rewardClass := getMsgRewardType(anyResp.GetTypeUrl())
		if index == 0 && rewardClass > 0 {
			msgClass = rewardClass
			k.HandleRelayerReward(ctx, relayer, msgClass)
		}
		if anyResp.GetTypeUrl() == "/cosmos.authz.v1beta1.MsgExecResponse" {

			msgExecResponse := authztypes.MsgExecResponse{}
			err := proto.Unmarshal(anyResp.GetValue(), &msgExecResponse)
			if err != nil {
				k.Logger(ctx).Debug("handling deep action response unmarshalling", "err", err)
				return nil, 0, err
			}

			actionIndex := index + previousMsgsExecuted
			if actionIndex >= len(action.Msgs) {
				return nil, 0, errorsmod.Wrapf(types.ErrMsgResponsesHandling, "expected more message responses")
			}
			msgExec := &authztypes.MsgExec{}
			if err := proto.Unmarshal(action.Msgs[actionIndex].Value, msgExec); err != nil {
				return nil, 0, err
			}

			msgResponses = []*cdctypes.Any{}

			for _, resultBytes := range msgExecResponse.Results {
				var msgResponse = cdctypes.Any{}
				if err := proto.Unmarshal(resultBytes, &msgResponse); err == nil {
					typeUrl := msgResponse.GetTypeUrl()

					if typeUrl != "" && strings.Contains(typeUrl, "Msg") {
						// _, err := k.interfaceRegistry.Resolve(typeUrl)
						// if err == nil {
						k.Logger(ctx).Debug("parsing response authz v0.52+", "msgResponse", msgResponse)
						msgResponses = append(msgResponses, &msgResponse)
						continue
						//}
					}

				}
				// in v0.50.8 we were writing msgResponse.Data in that [][]byte and no marshalled anys
				// https://github.com/cosmos/cosmos-sdk/blob/v0.50.8/x/authz/keeper/keeper.go#L166-L186
				//	k.Logger(ctx).Debug("action result", "resultBytes", resultBytes)

				//as we do not get typeURL (until cosmos 0.52 and is not possible in 51) we have to rely in this, MsgExec is the only regisrered message that should return results
				msgRespProto, _, err := handleMsgData(&sdk.MsgData{Data: resultBytes, MsgType: msgExec.Msgs[0].TypeUrl})
				if err != nil {
					return nil, 0, err
				}
				respAny, err := cdctypes.NewAnyWithValue(msgRespProto)
				if err != nil {
					return nil, 0, err
				}

				msgResponses = append(msgResponses, respAny)

			}
		}
	}
	return msgResponses, msgClass, nil
}

// SetActionOnTimeout sets the action timeout result to the action

func (k Keeper) SetActionOnTimeout(ctx sdk.Context, sourcePort string, channelID string, seq uint64) error {
	id := k.getTmpActionID(ctx, sourcePort, channelID, seq)
	if id <= 0 {
		return nil
	}
	action := k.GetActionInfo(ctx, id)
	if action.Configuration.ReregisterICAAfterTimeout {
		action := k.GetActionInfo(ctx, id)
		metadataString := icatypes.NewDefaultMetadataString(action.ICAConfig.ConnectionID, action.ICAConfig.HostConnectionID)
		err := k.RegisterInterchainAccount(ctx, action.ICAConfig.ConnectionID, action.Owner, metadataString)
		if err != nil {
			return err
		}
	} else {
		k.RemoveFromActionQueue(ctx, action)
	}
	k.Logger(ctx).Debug("action packet timed out", "action_id", id)

	actionHistoryEntry, err := k.GetLatestActionHistoryEntry(ctx, id)
	if err != nil {
		return err
	}

	actionHistoryEntry.TimedOut = true
	k.SetCurrentActionHistoryEntry(ctx, id, actionHistoryEntry)

	return nil
}

// SetActionOnTimeout sets the action timeout result to the action
func (k Keeper) SetActionError(ctx sdk.Context, sourcePort string, channelID string, seq uint64, err string) {
	id := k.getTmpActionID(ctx, sourcePort, channelID, seq)
	if id <= 0 {
		return
	}

	k.Logger(ctx).Debug("action", "id", id, "error", err)

	actionHistoryEntry, newErr := k.GetLatestActionHistoryEntry(ctx, id)
	if newErr != nil {
		actionHistoryEntry.Errors = append(actionHistoryEntry.Errors, newErr.Error())
	}

	actionHistoryEntry.Errors = append(actionHistoryEntry.Errors, err)
	k.SetCurrentActionHistoryEntry(ctx, id, actionHistoryEntry)
}
