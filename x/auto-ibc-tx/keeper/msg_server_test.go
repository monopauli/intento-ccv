package keeper_test

import (
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	//banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	icatypes "github.com/cosmos/ibc-go/v4/modules/apps/27-interchain-accounts/types"
	channeltypes "github.com/cosmos/ibc-go/v4/modules/core/04-channel/types"
	ibctesting "github.com/cosmos/ibc-go/v4/testing"
	"github.com/trstlabs/trst/x/auto-ibc-tx/keeper"
	"github.com/trstlabs/trst/x/auto-ibc-tx/types"
)

func (suite *KeeperTestSuite) TestRegisterInterchainAccount() {
	var (
		owner string
		path  *ibctesting.Path
	)

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"success", func() {}, true,
		},
		{
			"port is already bound",
			func() {
				GetICAApp(suite.chainA.TestChain).GetIBCKeeper().PortKeeper.BindPort(suite.chainA.GetContext(), TestPortID)
			},
			false,
		},
		{
			"fails to generate port-id",
			func() {
				owner = ""
			},
			false,
		},
		{
			"MsgChanOpenInit fails - channel is already active",
			func() {
				portID, err := icatypes.NewControllerPortID(owner)
				suite.Require().NoError(err)

				channel := channeltypes.NewChannel(
					channeltypes.OPEN,
					channeltypes.ORDERED,
					channeltypes.NewCounterparty(path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID),
					[]string{path.EndpointA.ConnectionID},
					path.EndpointA.ChannelConfig.Version,
				)
				GetICAApp(suite.chainA.TestChain).AppKeepers.IbcKeeper.ChannelKeeper.SetChannel(suite.chainA.GetContext(), portID, ibctesting.FirstChannelID, channel)

				GetICAApp(suite.chainA.TestChain).AppKeepers.ICAControllerKeeper.SetActiveChannelID(suite.chainA.GetContext(), ibctesting.FirstConnectionID, portID, ibctesting.FirstChannelID)
			},
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			suite.SetupTest()

			owner = TestOwnerAddress // must be explicitly changed

			path = NewICAPath(suite.chainA, suite.chainB)
			suite.coordinator.SetupConnections(path)

			tc.malleate() // malleate mutates test data

			msgSrv := keeper.NewMsgServerImpl(GetICAKeeper(suite.chainA))
			msg := types.NewMsgRegisterAccount(owner, path.EndpointA.ConnectionID, path.EndpointA.ChannelConfig.Version)

			res, err := msgSrv.RegisterAccount(sdk.WrapSDKContext(suite.chainA.GetContext()), msg)

			// resp, err := GetICAApp(suite.chainA).AppKeepers.AutoIBCTXKeeper.InterchainAccountFromAddress(sdk.WrapSDKContext(suite.chainA.GetContext()), &types.QueryInterchainAccountFromAddressRequest{Owner: owner, ConnectionId: path.EndpointA.ConnectionID})
			// fmt.Printf("%v", resp)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(res)
			} else {
				suite.Require().Error(err)
				suite.Require().Nil(res)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestSubmitTx() {
	var (
		path                      *ibctesting.Path
		registerInterchainAccount bool
		owner                     string
		connectionId              string
		icaMsg                    sdk.Msg
	)

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"success", func() {
				registerInterchainAccount = true
				owner = TestOwnerAddress
				connectionId = path.EndpointA.ConnectionID
			}, true,
		},
		{
			"failure - owner address is empty", func() {
				registerInterchainAccount = true
				owner = ""
				connectionId = path.EndpointA.ConnectionID

			}, false,
		},
		{
			"failure - active channel does not exist for connection ID", func() {
				registerInterchainAccount = true
				owner = TestOwnerAddress
				connectionId = "connection-100"

			}, false,
		},
		{
			"failure - active channel does not exist for port ID", func() {
				registerInterchainAccount = true
				owner = "cosmos153lf4zntqt33a4v0sm5cytrxyqn78q7kz8j8x5"
				connectionId = path.EndpointA.ConnectionID

			}, false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			suite.SetupTest()

			icaAppA := GetICAApp(suite.chainA.TestChain)
			icaAppB := GetICAApp(suite.chainB.TestChain)

			path = NewICAPath(suite.chainA, suite.chainB)
			suite.coordinator.SetupConnections(path)

			tc.malleate() // malleate mutates test data

			if registerInterchainAccount {
				err := SetupICAPath(path, TestOwnerAddress)
				suite.Require().NoError(err)

				portID, err := icatypes.NewControllerPortID(TestOwnerAddress)
				suite.Require().NoError(err)

				// Get the address of the interchain account stored in state during handshake step
				interchainAccountAddr, found := GetICAApp(suite.chainA.TestChain).AppKeepers.ICAControllerKeeper.GetInterchainAccountAddress(suite.chainA.GetContext(), path.EndpointA.ConnectionID, portID)
				suite.Require().True(found)

				icaAddr, err := sdk.AccAddressFromBech32(interchainAccountAddr)
				suite.Require().NoError(err)

				// Check if account is created
				interchainAccount := icaAppB.AppKeepers.AccountKeeper.GetAccount(suite.chainB.GetContext(), icaAddr)
				suite.Require().Equal(interchainAccount.GetAddress().String(), interchainAccountAddr)

				// Create bank transfer message to execute on the host
				icaMsg = &banktypes.MsgSend{
					FromAddress: interchainAccountAddr,
					ToAddress:   suite.chainB.SenderAccount.GetAddress().String(),
					Amount:      sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100))),
				}
			}

			// ownerAddr, err := sdk.AccAddressFromBech32(owner)
			// suite.Require().NoError(err)

			msgSrv := keeper.NewMsgServerImpl(GetICAKeeper2(icaAppA))
			msg, err := types.NewMsgSubmitTx(owner, icaMsg, connectionId)
			suite.Require().NoError(err)

			res, err := msgSrv.SubmitTx(sdk.WrapSDKContext(suite.chainA.GetContext()), msg)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(res)
			} else {
				suite.Require().Error(err)
				suite.Require().Nil(res)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestSubmitAutoTx() {
	var (
		path                      *ibctesting.Path
		registerInterchainAccount bool
		owner                     string
		connectionId              string
		icaMsg                    sdk.Msg
		parseIcaAddress           bool
	)

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"success - IBC autoTx", func() {
				registerInterchainAccount = true
				owner = TestOwnerAddress
				connectionId = path.EndpointA.ConnectionID
			}, true,
		},
		{
			"success - local autoTx", func() {
				registerInterchainAccount = false
				owner = TestOwnerAddress
				connectionId = ""
				icaMsg = &banktypes.MsgSend{
					FromAddress: owner,
					ToAddress:   suite.chainB.SenderAccount.GetAddress().String(),
					Amount:      sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100))),
				}
			}, true,
		},
		{
			"success - parse ICA address", func() {
				registerInterchainAccount = true
				owner = TestOwnerAddress
				connectionId = path.EndpointA.ConnectionID
				parseIcaAddress = true
			}, true,
		},
		{
			"failure - owner address is empty", func() {
				registerInterchainAccount = true
				owner = ""
				connectionId = path.EndpointA.ConnectionID
				parseIcaAddress = false
			}, false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			suite.SetupTest()

			icaAppA := GetICAApp(suite.chainA.TestChain)
			icaAppB := GetICAApp(suite.chainB.TestChain)

			path = NewICAPath(suite.chainA, suite.chainB)
			suite.coordinator.SetupConnections(path)

			tc.malleate() // malleate mutates test data

			if registerInterchainAccount {
				err := SetupICAPath(path, TestOwnerAddress)
				suite.Require().NoError(err)

				portID, err := icatypes.NewControllerPortID(TestOwnerAddress)
				suite.Require().NoError(err)

				// Get the address of the interchain account stored in state during handshake step
				interchainAccountAddr, found := GetICAApp(suite.chainA.TestChain).AppKeepers.ICAControllerKeeper.GetInterchainAccountAddress(suite.chainA.GetContext(), path.EndpointA.ConnectionID, portID)
				suite.Require().True(found)

				icaAddr, err := sdk.AccAddressFromBech32(interchainAccountAddr)
				suite.Require().NoError(err)

				// Check if account is created
				interchainAccount := icaAppB.AppKeepers.AccountKeeper.GetAccount(suite.chainB.GetContext(), icaAddr)
				suite.Require().Equal(interchainAccount.GetAddress().String(), interchainAccountAddr)
				if parseIcaAddress {
					interchainAccountAddr = "ICA_ADDR"
				}
				// Create bank transfer message to execute on the host
				icaMsg = &banktypes.MsgSend{
					FromAddress: interchainAccountAddr,
					ToAddress:   suite.chainB.SenderAccount.GetAddress().String(),
					Amount:      sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100))),
				}
			}

			GetICAApp(suite.chainA.TestChain).AppKeepers.AutoIBCTXKeeper.SetParams(suite.chainA.GetContext(), types.Params{
				AutoTxFundsCommission:      2,
				AutoTxConstantFee:          1_000_000,
				AutoTxFlexFeeMul:           100,
				RecurringAutoTxConstantFee: 1_000_000,
				MaxAutoTxDuration:          time.Hour * 24 * 366 * 10,
				MinAutoTxDuration:          time.Second * 60,
				MinAutoTxInterval:          time.Second * 20,
			})

			msg, err := types.NewMsgSubmitAutoTx(owner, "label", []sdk.Msg{icaMsg}, connectionId, "200s", "100s", 0, sdk.Coins{}, []uint64{})
			suite.Require().NoError(err)
			wrappedCtx := sdk.WrapSDKContext(suite.chainA.GetContext())
			msgSrv := keeper.NewMsgServerImpl(GetICAKeeper2(icaAppA))
			res, err := msgSrv.SubmitAutoTx(wrappedCtx, msg)

			if parseIcaAddress {
				icaApp := GetICAApp(suite.chainA.TestChain)
				err := icaMsg.ValidateBasic()
				suite.Require().Contains(err.Error(), "bech32")
				err = msg.ValidateBasic()
				suite.Require().NoError(err)
				suite.chainA.NextBlock()
				unwrappedCtx := sdk.UnwrapSDKContext(wrappedCtx)
				autoTx := icaApp.AppKeepers.AutoIBCTXKeeper.GetAutoTxInfo(unwrappedCtx, 1)
				suite.Require().NotEqual(autoTx, types.AutoTxInfo{})

				err = icaApp.AppKeepers.AutoIBCTXKeeper.SendAutoTx(unwrappedCtx, autoTx)
				suite.Require().NoError(err)
				icaAddrToParse := "ICA_ADDR"
				autoTx = icaApp.AppKeepers.AutoIBCTXKeeper.GetAutoTxInfo(unwrappedCtx, 1)
				//txMsgs := autoTx.GetTxMsgs()

				var txMsg sdk.Msg
				err = icaApp.AppCodec().UnpackAny(autoTx.Msgs[0], &txMsg)
				suite.Require().NoError(err)
				msgJSON, err := icaApp.AppCodec().MarshalInterfaceJSON(txMsg)
				suite.Require().NoError(err)

				msgJSONString := string(msgJSON)
				index := strings.Index(msgJSONString, icaAddrToParse)
				suite.Require().Equal(-1, index)

			}

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(res)
			} else {
				suite.Require().Error(err)
				suite.Require().Nil(res)
			}

		})
	}
}

func (suite *KeeperTestSuite) TestUpdateAutoTx() {
	var (
		path                      *ibctesting.Path
		registerInterchainAccount bool
		owner                     string
		connectionId              string
		icaMsg                    sdk.Msg
		newEndTime                uint64
		newStartAt                uint64
		newInterval               string
		withExecHistory           bool
	)

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"success - update autoTx", func() {
				registerInterchainAccount = true
				owner = TestOwnerAddress
				connectionId = path.EndpointA.ConnectionID
				newStartAt = uint64(time.Now().Unix())
				newEndTime = uint64(time.Now().Add(time.Hour).Unix())
				newInterval = "8m20s"
			}, true,
		},
		{
			"success - update local autoTx", func() {
				registerInterchainAccount = false
				owner = TestOwnerAddress
				connectionId = ""
			}, true,
		},
		{
			"failure - interval shorter than min duration", func() {
				registerInterchainAccount = true
				owner = TestOwnerAddress
				connectionId = path.EndpointA.ConnectionID
				newInterval = "1s"
			}, false,
		},
		{
			"failure - start time can not be changed after execution entry", func() {
				registerInterchainAccount = true
				owner = TestOwnerAddress
				connectionId = path.EndpointA.ConnectionID
				newStartAt = uint64(time.Date(2020, time.April, 22, 12, 0, 0, 0, time.UTC).Unix())
				newInterval = "8m20s"
				withExecHistory = true
			}, false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			suite.SetupTest()

			icaAppA := GetICAApp(suite.chainA.TestChain)
			icaAppB := GetICAApp(suite.chainB.TestChain)

			path = NewICAPath(suite.chainA, suite.chainB)
			suite.coordinator.SetupConnections(path)

			tc.malleate() // malleate mutates test data

			if registerInterchainAccount {
				err := SetupICAPath(path, TestOwnerAddress)
				suite.Require().NoError(err)

				portID, err := icatypes.NewControllerPortID(TestOwnerAddress)
				suite.Require().NoError(err)

				// Get the address of the interchain account stored in state during handshake step
				interchainAccountAddr, found := GetICAApp(suite.chainA.TestChain).AppKeepers.ICAControllerKeeper.GetInterchainAccountAddress(suite.chainA.GetContext(), path.EndpointA.ConnectionID, portID)
				suite.Require().True(found)

				icaAddr, err := sdk.AccAddressFromBech32(interchainAccountAddr)
				suite.Require().NoError(err)

				// Check if account is created
				interchainAccount := icaAppB.AppKeepers.AccountKeeper.GetAccount(suite.chainB.GetContext(), icaAddr)
				suite.Require().Equal(interchainAccount.GetAddress().String(), interchainAccountAddr)

				// Create bank transfer message to execute on the host
				icaMsg = &banktypes.MsgSend{
					FromAddress: interchainAccountAddr,
					ToAddress:   suite.chainB.SenderAccount.GetAddress().String(),
					Amount:      sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100))),
				}
			}

			icaAppA.AppKeepers.AutoIBCTXKeeper.SetParams(suite.chainA.GetContext(), types.Params{
				AutoTxFundsCommission:      2,
				AutoTxConstantFee:          1_000_000,
				AutoTxFlexFeeMul:           100,
				RecurringAutoTxConstantFee: 1_000_000,
				MaxAutoTxDuration:          time.Hour * 24 * 366 * 10,
				MinAutoTxDuration:          time.Second * 60,
				MinAutoTxInterval:          time.Second * 20,
			})

			msg, err := types.NewMsgSubmitAutoTx(owner, "label", []sdk.Msg{icaMsg}, connectionId, "200s", "100s", uint64(time.Now().Unix()), sdk.Coins{}, []uint64{})
			suite.Require().NoError(err)
			wrappedCtx := sdk.WrapSDKContext(suite.chainA.GetContext())
			msgSrv := keeper.NewMsgServerImpl(GetICAKeeper2(icaAppA))
			_, err = msgSrv.SubmitAutoTx(wrappedCtx, msg)
			suite.Require().NoError(err)

			if withExecHistory {
				autoTx := icaAppA.AppKeepers.AutoIBCTXKeeper.GetAutoTxInfo(sdk.UnwrapSDKContext(wrappedCtx), 1)
				suite.Require().NotNil(autoTx)
				fakeEntry := types.AutoTxHistoryEntry{ScheduledExecTime: autoTx.ExecTime}
				autoTx.AutoTxHistory = []*types.AutoTxHistoryEntry{&fakeEntry}
				icaAppA.AppKeepers.AutoIBCTXKeeper.SetAutoTxInfo(sdk.UnwrapSDKContext(wrappedCtx), &autoTx)
				suite.chainA.NextBlock()
			}
			updateMsg, err := types.NewMsgUpdateAutoTx(owner, 1, "new_label", []sdk.Msg{icaMsg}, connectionId, newEndTime, newInterval, newStartAt, sdk.Coins{}, []uint64{})
			suite.Require().NoError(err)
			suite.chainA.Coordinator.IncrementTime()
			suite.Require().NotEqual(suite.chainA.GetContext(), wrappedCtx)
			wrappedCtx = sdk.WrapSDKContext(suite.chainA.GetContext())

			msgSrv = keeper.NewMsgServerImpl(GetICAKeeper2(icaAppA))
			res, err := msgSrv.UpdateAutoTx(wrappedCtx, updateMsg)

			if !tc.expPass {
				suite.Require().Nil(res)
				suite.Require().Error(err)

			} else {
				suite.chainA.NextBlock()
				suite.Require().NoError(err)
				suite.Require().NotNil(res)
				autoTx := icaAppA.AppKeepers.AutoIBCTXKeeper.GetAutoTxInfo(sdk.UnwrapSDKContext(wrappedCtx), 1)
				suite.Require().Equal(autoTx.Label, updateMsg.Label)
				suite.Require().Equal(autoTx.StartTime.Unix(), int64(updateMsg.StartAt))
				suite.Require().Equal(autoTx.EndTime.Unix(), int64(updateMsg.EndTime))
				suite.Require().Equal((autoTx.Interval.String()), updateMsg.Interval)

			}

		})
	}
}
