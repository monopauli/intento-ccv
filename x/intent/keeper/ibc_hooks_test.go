package keeper_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/trstlabs/intento/x/intent/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"

	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
)

func (suite *KeeperTestSuite) TestOnRecvTransferPacket() {
	var (
		trace    transfertypes.DenomTrace
		amount   math.Int
		receiver string
	)

	suite.SetupTest()

	path := NewTransferPath(suite.IntentoChain, suite.HostChain)
	suite.Coordinator.Setup(path)
	receiver = suite.HostChain.SenderAccount.GetAddress().String() // must be explicitly changed

	amount = math.NewInt(100) // must be explicitly changed in malleate
	seq := uint64(1)

	trace = transfertypes.ParseDenomTrace(sdk.DefaultBondDenom)

	// send coin from IntentoChain to HostChain
	transferMsg := transfertypes.NewMsgTransfer(path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID, sdk.NewCoin(trace.IBCDenom(), amount), suite.IntentoChain.SenderAccount.GetAddress().String(), receiver, clienttypes.NewHeight(1, 110), 0, "")
	_, err := suite.IntentoChain.SendMsgs(transferMsg)
	suite.Require().NoError(err) // message committed

	data := transfertypes.NewFungibleTokenPacketData(trace.GetFullDenomPath(), amount.String(), suite.IntentoChain.SenderAccount.GetAddress().String(), receiver, "")
	packet := channeltypes.NewPacket(data.GetBytes(), seq, path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID, path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID, clienttypes.NewHeight(1, 100), 0)

	//a little hack as this check would be on HostChain OnRecvPacket
	ack := GetICAApp(suite.IntentoChain).TransferStack.OnRecvPacket(suite.IntentoChain.GetContext(), packet, suite.IntentoChain.SenderAccount.GetAddress())

	suite.Require().True(ack.Success())

}

func (suite *KeeperTestSuite) TestOnRecvTransferPacketWithAction() {
	suite.SetupTest()

	params := types.DefaultParams()
	params.GasFeeCoins = sdk.NewCoins(sdk.NewCoin("stake", math.OneInt()))
	params.ActionFlexFeeMul = 1
	GetICAApp(suite.IntentoChain).IntentKeeper.SetParams(suite.IntentoChain.GetContext(), params)

	addr := suite.IntentoChain.SenderAccount.GetAddress().String()
	addrTo := suite.TestAccs[0].String()
	msg := fmt.Sprintf(`{
		"@type":"/cosmos.bank.v1beta1.MsgSend",
		"amount": [{
			"amount": "70",
			"denom": "stake"
		}],
		"from_address": "%s",
		"to_address": "%s"
	}`, addr, addrTo)

	ackBytes := suite.receiveTransferPacket(addr, fmt.Sprintf(`{"action": {"owner": "%s","label": "my_trigger", "msgs": [%s], "duration": "500s", "interval": "60s", "start_at": "0"} }`, addr, msg))

	var ack map[string]string // This can't be unmarshalled to Acknowledgement because it's fetched from the events
	err := json.Unmarshal(ackBytes, &ack)
	suite.Require().NoError(err)
	suite.Require().NotContains(ack, "error")

	action := GetICAApp(suite.IntentoChain).IntentKeeper.GetActionInfo(suite.IntentoChain.GetContext(), 1)

	suite.Require().Equal(action.Owner, addr)
	suite.Require().Equal(action.Label, "my_trigger")
	suite.Require().Equal(action.ICAConfig.PortID, "")
	suite.Require().Equal(action.Interval, time.Second*60)

	var txMsgAny codectypes.Any
	cdc := codec.NewProtoCodec(GetICAApp(suite.IntentoChain).InterfaceRegistry())

	err = cdc.UnmarshalJSON([]byte(msg), &txMsgAny)
	suite.Require().NoError(err)
	suite.True(action.Msgs[0].Equal(txMsgAny))
}

func (suite *KeeperTestSuite) TestOnRecvTransferPacketAndMultippleActions() {
	suite.SetupTest()

	params := types.DefaultParams()
	params.GasFeeCoins = sdk.NewCoins(sdk.NewCoin("stake", math.OneInt()))
	params.ActionFlexFeeMul = 1
	GetICAApp(suite.IntentoChain).IntentKeeper.SetParams(suite.IntentoChain.GetContext(), params)

	addr := suite.IntentoChain.SenderAccount.GetAddress()
	msg := `{
		"@type":"/cosmos.bank.v1beta1.MsgSend",
		"amount": [{
			"amount": "70",
			"denom": "stake"
		}],
		"from_address": "into12gxmzpucje8aflw2vz45rv8x4nyaaj3rp8vjh03dulehkdl5fu6s93ewkp",
		"to_address": "into1ykql5ktedxkpjszj5trzu8f5dxajvgv95nuwjx"
	}`

	path := NewICAPath(suite.IntentoChain, suite.HostChain)
	suite.Coordinator.SetupConnections(path)
	err := suite.SetupICAPath(path, addr.String())
	suite.Require().NoError(err)

	//HostChain sends packet to IntentoChain. connectionID to execute on HostChain is on IntentoChains config
	ackBytes := suite.receiveTransferPacket(addr.String(), fmt.Sprintf(`{"action": {"owner": "%s","label": "my_trigger", "cid":"%s", "host_cid":"%s","msgs": [%s, %s], "duration": "500s", "interval": "60s", "start_at": "0", "fallback": "true" } }`, addr.String(), path.EndpointA.ConnectionID, path.EndpointB.ConnectionID, msg, msg))

	var ack map[string]string // This can't be unmarshalled to Acknowledgement because it's fetched from the events
	err = json.Unmarshal(ackBytes, &ack)
	suite.Require().NoError(err)
	suite.Require().NotContains(ack, "error")

	action := GetICAApp(suite.IntentoChain).IntentKeeper.GetActionInfo(suite.IntentoChain.GetContext(), 1)

	suite.Require().Equal(action.Owner, addr.String())
	suite.Require().Equal(action.Label, "my_trigger")
	suite.Require().Equal(action.Configuration.FallbackToOwnerBalance, true)
	suite.Require().Equal(action.ICAConfig.PortID, "icacontroller-"+addr.String())
	suite.Require().Equal(action.ICAConfig.ConnectionID, path.EndpointA.ConnectionID)

	suite.Require().Equal(action.Interval, time.Second*60)

	_, found := GetICAApp(suite.IntentoChain).ICAControllerKeeper.GetInterchainAccountAddress(suite.IntentoChain.GetContext(), action.ICAConfig.ConnectionID, action.ICAConfig.PortID)
	suite.Require().True(found)

	var txMsgAny codectypes.Any
	cdc := codec.NewProtoCodec(GetICAApp(suite.IntentoChain).InterfaceRegistry())

	err = cdc.UnmarshalJSON([]byte(msg), &txMsgAny)
	suite.Require().NoError(err)
	suite.True(action.Msgs[0].Equal(txMsgAny))
}

func (suite *KeeperTestSuite) TestOnRecvTransferPacketSubmitTxAndAddressParsing() {
	suite.SetupTest()

	params := types.DefaultParams()
	params.GasFeeCoins = sdk.NewCoins(sdk.NewCoin("stake", math.OneInt()))
	params.ActionFlexFeeMul = 1
	GetICAApp(suite.IntentoChain).IntentKeeper.SetParams(suite.IntentoChain.GetContext(), params)

	addr := suite.IntentoChain.SenderAccount.GetAddress()
	msg := `{
		"@type":"/cosmos.bank.v1beta1.MsgSend",
		"amount": [{
			"amount": "70",
			"denom": "stake"
		}],
		"from_address": "ICA_ADDR",
		"to_address": "into1ykql5ktedxkpjszj5trzu8f5dxajvgv95nuwjx"
	}`

	path := NewICAPath(suite.IntentoChain, suite.HostChain)
	suite.Coordinator.SetupConnections(path)
	err := suite.SetupICAPath(path, addr.String())
	suite.Require().NoError(err)

	ackBytes := suite.receiveTransferPacket(addr.String(), fmt.Sprintf(`{"action": {"owner": "%s","label": "my trigger", "cid":"%s","host_cid":"%s","msgs": [%s, %s], "duration": "120s", "interval": "60s", "start_at": "0", "fallback":"true" }}`, addr.String(), path.EndpointA.ConnectionID, path.EndpointB.ConnectionID, msg, msg))
	var ack map[string]string // This can't be unmarshalled to Acknowledgement because it's fetched from the events
	err = json.Unmarshal(ackBytes, &ack)
	suite.Require().NoError(err)
	suite.Require().NotContains(ack, "error")

	actionKeeper := GetICAApp(suite.IntentoChain).IntentKeeper
	action := actionKeeper.GetActionInfo(suite.IntentoChain.GetContext(), 1)
	unpacker := suite.IntentoChain.Codec
	unpackedMsgs := action.GetTxMsgs(unpacker)
	suite.Require().True(strings.Contains(unpackedMsgs[0].String(), types.ParseICAValue))

	suite.IntentoChain.CurrentHeader.Time = suite.IntentoChain.CurrentHeader.Time.Add(time.Minute)
	actionKeeper.HandleAction(suite.IntentoChain.GetContext(), actionKeeper.Logger(suite.IntentoChain.GetContext()), action, suite.IntentoChain.GetContext().BlockTime(), nil)

	action = actionKeeper.GetActionInfo(suite.IntentoChain.GetContext(), 1)
	actionHistory, _ := actionKeeper.GetActionHistory(suite.IntentoChain.GetContext(), action.ID)
	suite.Require().NotNil(actionHistory)
	suite.Require().Empty(actionHistory[0].Errors)
	suite.Require().Equal(action.Owner, addr.String())
	suite.Require().Equal(action.Label, "my trigger")
	suite.Require().Equal(action.ICAConfig.PortID, "icacontroller-"+addr.String())
	suite.Require().Equal(action.ICAConfig.ConnectionID, path.EndpointA.ConnectionID)

	unpackedMsgs = action.GetTxMsgs(unpacker)
	suite.Require().False(strings.Contains(unpackedMsgs[0].String(), types.ParseICAValue))
	suite.Require().Equal(action.Interval, time.Second*60)
}

func (suite *KeeperTestSuite) TestOnRecvTransferPacketSubmitTxWithSentDenomInParams() {
	suite.SetupTest()

	addr := suite.IntentoChain.SenderAccount.GetAddress()
	msg := `{
		"@type":"/cosmos.bank.v1beta1.MsgSend",
		"amount": [{
			"amount": "70",
			"denom": "stake"
		}],
		"from_address": "ICA_ADDR",
		"to_address": "into1ykql5ktedxkpjszj5trzu8f5dxajvgv95nuwjx"
	}`

	path := NewICAPath(suite.IntentoChain, suite.HostChain)
	suite.Coordinator.SetupConnections(path)
	err := suite.SetupICAPath(path, addr.String())
	suite.Require().NoError(err)

	ackBytes := suite.receiveTransferPacket(addr.String(), fmt.Sprintf(`{"action": {"owner": "%s","label": "my trigger", "cid":"%s","host_cid":"%s","msgs": [%s, %s], "duration": "120s", "interval": "60s", "start_at": "0", "fallback": "true" }}`, addr.String(), path.EndpointA.ConnectionID, path.EndpointB.ConnectionID, msg, msg))
	var ack map[string]string // This can't be unmarshalled to Acknowledgement because it's fetched from the events
	err = json.Unmarshal(ackBytes, &ack)
	suite.Require().NoError(err)
	suite.Require().NotContains(ack, "error")

	actionKeeper := GetICAApp(suite.IntentoChain).IntentKeeper
	action := actionKeeper.GetActionInfo(suite.IntentoChain.GetContext(), 1)
	feeAddr, _ := sdk.AccAddressFromBech32(action.FeeAddress)
	bDenom := GetICAApp(suite.IntentoChain).BankKeeper.GetAllBalances(suite.IntentoChain.GetContext(), feeAddr)[0].Denom
	params := types.DefaultParams()
	params.GasFeeCoins = sdk.NewCoins(sdk.NewCoin(bDenom, math.NewInt(2)), sdk.NewCoin("stake", math.OneInt()))
	params.ActionFlexFeeMul = 1
	GetICAApp(suite.IntentoChain).IntentKeeper.SetParams(suite.IntentoChain.GetContext(), params)

	unpacker := suite.IntentoChain.Codec
	unpackedMsgs := action.GetTxMsgs(unpacker)
	suite.Require().True(strings.Contains(unpackedMsgs[0].String(), types.ParseICAValue))

	suite.IntentoChain.CurrentHeader.Time = suite.IntentoChain.CurrentHeader.Time.Add(time.Minute)
	actionKeeper.HandleAction(suite.IntentoChain.GetContext(), actionKeeper.Logger(suite.IntentoChain.GetContext()), action, suite.IntentoChain.GetContext().BlockTime(), nil)

	action = actionKeeper.GetActionInfo(suite.IntentoChain.GetContext(), 1)
	actionHistory, _ := actionKeeper.GetActionHistory(suite.IntentoChain.GetContext(), action.ID)
	suite.Require().NotNil(actionHistory)
	suite.Require().Empty(actionHistory[0].Errors)
}
