package cli

import (
	"fmt"
	"os"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/trstlabs/trst/x/auto-ibc-tx/types"
)

// GetTxCmd creates and returns the auto-ibc-tx tx command
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		getRegisterAccountCmd(),
		getSubmitTxCmd(),
		getSubmitAutoTxCmd(),
		getRegisterAccountAndSubmitAutoTxCmd(),
		getUpdateAutoTxCmd(),
	)

	return cmd
}

func getRegisterAccountCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "register",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// TestVersion defines a reusable interchainaccounts version string for testing purposes
			version := string(icatypes.ModuleCdc.MustMarshalJSON(&icatypes.Metadata{
				Version:                icatypes.Version,
				ControllerConnectionId: viper.GetString(flagConnectionID),
				HostConnectionId:       viper.GetString(flagCounterpartyConnectionID),
				Encoding:               icatypes.EncodingProtobuf,
				TxType:                 icatypes.TxTypeSDKMultiMsg,
			}))

			msg := types.NewMsgRegisterAccount(
				clientCtx.GetFromAddress().String(),
				viper.GetString(flagConnectionID),
				version,
			)

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(flagCounterpartyConnectionID, "", "Counterparty Connection ID, from host chain")
	cmd.Flags().String(flagConnectionID, "", "Connection ID, an IBC ID from this chain to the host chain")
	_ = cmd.MarkFlagRequired(flagConnectionID)

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func getSubmitTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "submit-tx [path/to/sdk_msg.json]",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			cdc := codec.NewProtoCodec(clientCtx.InterfaceRegistry)

			var txMsg sdk.Msg
			if err := cdc.UnmarshalInterfaceJSON([]byte(args[0]), &txMsg); err != nil {

				// check for file path if JSON input is not provided
				contents, err := os.ReadFile(args[0])
				if err != nil {
					return errors.Wrap(err, "neither JSON input nor path to .json file for sdk msg were provided")
				}

				if err := cdc.UnmarshalInterfaceJSON(contents, &txMsg); err != nil {
					return errors.Wrap(err, "error unmarshalling sdk msg")
				}
			}

			msg, err := types.NewMsgSubmitTx(clientCtx.GetFromAddress().String(), txMsg, viper.GetString(flagConnectionID))
			if err != nil {
				return err
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(flagConnectionID, "", "Connection ID, an IBC ID from this chain to the host chain, optional")

	_ = cmd.MarkFlagRequired(flagConnectionID)

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func getSubmitAutoTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "submit-auto-tx [path/to/sdk_msg.json]",
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			cdc := codec.NewProtoCodec(clientCtx.InterfaceRegistry)

			var txMsgs []sdk.Msg
			for _, arg := range args {
				var txMsg sdk.Msg
				if err := cdc.UnmarshalInterfaceJSON([]byte(arg), &txMsg); err != nil {
					// Check for file path if JSON input is not provided
					var msgContents []byte

					// Check if arg is a valid file path
					if _, err := os.Stat(arg); err == nil {
						// arg is a file path, read the file contents
						msgContents, err = os.ReadFile(arg)
						if err != nil {
							return errors.Wrap(err, "failed to read file")
						}
					} else {
						// arg is not a valid file path, assume it is a JSON string
						msgContents = []byte(arg)
					}

					// Parse the JSON content
					if err := cdc.UnmarshalInterfaceJSON(msgContents, &txMsg); err != nil {
						return errors.Wrap(err, "error unmarshalling sdk msg")
					}
					txMsgs = append(txMsgs, txMsg)
				}
			}

			var txIds []uint64
			dependsOnString := viper.GetStringSlice(flagDependsOn)
			for _, id := range dependsOnString {
				txId, err := strconv.ParseUint(id, 10, 64)
				if err != nil {
					return errors.Wrap(err, "invalid id, must be a number")
				}
				txIds = append(txIds, txId)
			}
			funds := sdk.Coins{}
			amount := viper.GetString(flagFeeFunds)
			if amount != "" {
				funds, err = sdk.ParseCoinsNormalized(amount)
				if err != nil {
					return err
				}
			}

			msg, err := types.NewMsgSubmitAutoTx(clientCtx.GetFromAddress().String(), viper.GetString(flagLabel), txMsgs, viper.GetString(flagConnectionID), viper.GetString(flagDuration), viper.GetString(flagInterval), viper.GetUint64(flagStartAt), funds, txIds /*  viper.GetUint64(flagRetries) */)
			if err != nil {
				return err
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(flagLabel, "", "A custom label for the AutoTx e.g. AutoTransfer, UpdateContractParams, optional")
	cmd.Flags().String(flagDuration, "", "A custom duration for the AutoTx e.g. 2h, 6000s, 72h3m0.5s, optional")
	cmd.Flags().String(flagInterval, "", "A custom interval for the AutoTx e.g. 2h, 6000s, 72h3m0.5s, optional")
	cmd.Flags().String(flagFeeFunds, "", "Coins to sent to limit the fees incurred, optional")
	cmd.Flags().Uint64(flagStartAt, 0, "A custom start time in UNIX time, optional")
	cmd.Flags().StringArray(flagDependsOn, []string{}, "array of auto-tx-ids this auto-tx depends on e.g. 5, 6")
	cmd.Flags().String(flagConnectionID, "", "Connection ID, an IBC ID from this chain to the host chain, optional")
	// cmd.Flags().Uint64(flagRetries, 0, "Maximum amount of retries to make the tx succeed, optional")

	//_ = cmd.MarkFlagRequired(flagConnectionID)
	_ = cmd.MarkFlagRequired(flagDuration)

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func getRegisterAccountAndSubmitAutoTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "register-ica-and-submit-auto-tx [path/to/sdk_msg.json]",
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			cdc := codec.NewProtoCodec(clientCtx.InterfaceRegistry)

			var txMsgs []sdk.Msg
			for _, arg := range args {
				var txMsg sdk.Msg
				if err := cdc.UnmarshalInterfaceJSON([]byte(arg), &txMsg); err != nil {
					// Check for file path if JSON input is not provided
					var msgContents []byte

					// Check if arg is a valid file path
					if _, err := os.Stat(arg); err == nil {
						// arg is a file path, read the file contents
						msgContents, err = os.ReadFile(arg)
						if err != nil {
							return errors.Wrap(err, "failed to read file")
						}
					} else {
						// arg is not a valid file path, assume it is a JSON string
						msgContents = []byte(arg)
					}

					// Parse the JSON content
					if err := cdc.UnmarshalInterfaceJSON(msgContents, &txMsg); err != nil {
						return errors.Wrap(err, "error unmarshalling sdk msg")
					}
					txMsgs = append(txMsgs, txMsg)
				}
			}

			var txIds []uint64
			dependsOnString := viper.GetStringSlice(flagDependsOn)
			for _, id := range dependsOnString {
				txId, err := strconv.ParseUint(id, 10, 64)
				if err != nil {
					return errors.Wrap(err, "invalid id, must be a number")
				}
				txIds = append(txIds, txId)
			}
			funds := sdk.Coins{}
			amount := viper.GetString(flagFeeFunds)
			if amount != "" {
				funds, err = sdk.ParseCoinsNormalized(amount)
				if err != nil {
					return err
				}
			}
			// TestVersion defines a reusable interchainaccounts version string for testing purposes
			version := string(icatypes.ModuleCdc.MustMarshalJSON(&icatypes.Metadata{
				Version:                icatypes.Version,
				ControllerConnectionId: viper.GetString(flagConnectionID),
				HostConnectionId:       viper.GetString(flagCounterpartyConnectionID),
				Encoding:               icatypes.EncodingProtobuf,
				TxType:                 icatypes.TxTypeSDKMultiMsg,
			}))

			msg, err := types.NewMsgRegisterAccountAndSubmitAutoTx(clientCtx.GetFromAddress().String(), viper.GetString(flagLabel), txMsgs, viper.GetString(flagConnectionID), viper.GetString(flagDuration), viper.GetString(flagInterval), viper.GetUint64(flagStartAt), funds, txIds /* viper.GetUint64(flagRetries) */, version)
			if err != nil {
				return err
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(flagLabel, "", "A custom label for the AutoTx e.g. AutoTransfer, UpdateContractParams, optional")
	cmd.Flags().String(flagDuration, "", "A custom duration for the AutoTx e.g. 2h, 6000s, 72h3m0.5s, optional")
	cmd.Flags().String(flagInterval, "", "A custom interval for the AutoTx e.g. 2h, 6000s, 72h3m0.5s, optional")
	cmd.Flags().String(flagStartAt, "0", "A custom start time in UNIX time, optional")
	cmd.Flags().String(flagFeeFunds, "", "Coins to sent to limit the fees incurred, optional")
	cmd.Flags().String(flagCounterpartyConnectionID, "", "Counterparty Connection ID, from host chain, optional")
	cmd.Flags().String(flagConnectionID, "", "Connection ID, an IBC ID from this chain to the host chain, optional")

	cmd.Flags().StringArray(flagDependsOn, []string{}, "array of auto-tx-ids this auto-tx depends on e.g. 5, 6")
	// cmd.Flags().Uint64(flagRetries, 0, "Maximum amount of retries to make the tx succeed, optional")

	//_ = cmd.MarkFlagRequired(flagConnectionID)
	_ = cmd.MarkFlagRequired(flagDuration)

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func getUpdateAutoTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "update-auto-tx [AutoTxID] [path/to/sdk_msg.json, optional]",
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			cdc := codec.NewProtoCodec(clientCtx.InterfaceRegistry)

			txID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			var txMsgs []sdk.Msg
			for _, arg := range args {
				if arg == args[0] {
					continue
				}
				var txMsg sdk.Msg
				if err := cdc.UnmarshalInterfaceJSON([]byte(arg), &txMsg); err != nil {
					// Check for file path if JSON input is not provided
					var msgContents []byte

					// Check if arg is a valid file path
					if _, err := os.Stat(arg); err == nil {
						// arg is a file path, read the file contents
						msgContents, err = os.ReadFile(arg)
						if err != nil {
							return errors.Wrap(err, "failed to read file")
						}
					} else {
						// arg is not a valid file path, assume it is a JSON string
						msgContents = []byte(arg)
					}

					// Parse the JSON content
					if err := cdc.UnmarshalInterfaceJSON(msgContents, &txMsg); err != nil {
						return errors.Wrap(err, "error unmarshalling sdk msg")
					}
					txMsgs = append(txMsgs, txMsg)
				}
			}
			var txIds []uint64
			dependsOnString := viper.GetStringSlice(flagDependsOn)
			for _, id := range dependsOnString {
				txId, err := strconv.ParseUint(id, 10, 64)
				if err != nil {
					return errors.Wrap(err, "invalid id, must be a number")
				}
				txIds = append(txIds, txId)
			}
			funds := sdk.Coins{}
			amount := viper.GetString(flagFeeFunds)
			if amount != "" {
				funds, err = sdk.ParseCoinsNormalized(amount)
				if err != nil {
					return err
				}
			}
			msg, err := types.NewMsgUpdateAutoTx(clientCtx.GetFromAddress().String(), txID, viper.GetString(flagLabel), txMsgs, viper.GetString(flagConnectionID), viper.GetUint64(flagEndTime), viper.GetString(flagInterval), viper.GetUint64(flagStartAt), funds, txIds /* viper.GetUint64(flagRetries) */ /* , viper.GetString(flagVersion) */)
			if err != nil {
				return err
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(flagLabel, "", "A custom label for the AutoTx e.g. AutoTransfer, UpdateContractParams, optional")
	cmd.Flags().String(flagEndTime, "", "A custom end time in UNIX time")
	cmd.Flags().String(flagInterval, "", "A custom interval for the AutoTx e.g. 2h, 6000s, 72h3m0.5s, optional")
	cmd.Flags().String(flagStartAt, "0", "A custom start time in UNIX time, optional")
	cmd.Flags().String(flagFeeFunds, "", "Coins to sent to limit the fees incurred, optional")
	cmd.Flags().StringArray(flagDependsOn, []string{}, "array of auto-tx-ids this auto-tx depends on e.g. 5, 6")
	cmd.Flags().String(flagConnectionID, "", "Connection ID, an IBC ID from this chain to the host chain, optional")
	// cmd.Flags().Uint64(flagRetries, 0, "Maximum amount of retries to make the tx succeed, optional")

	//_ = cmd.MarkFlagRequired(flagConnectionID)
	_ = cmd.MarkFlagRequired(flagDuration)

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
