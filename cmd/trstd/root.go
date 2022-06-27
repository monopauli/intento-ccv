package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
	"github.com/cosmos/cosmos-sdk/types/module"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	"github.com/rs/zerolog"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	prefix "github.com/trstlabs/trst/types"
	"github.com/trstlabs/trst/x/compute"

	//"github.com/tendermint/tendermint/libs/cli"
	"github.com/cosmos/cosmos-sdk/snapshots"
	app "github.com/trstlabs/trst/app"

	//"github.com/trstlabs/trst/app"

	"github.com/spf13/cast"
	"github.com/spf13/cobra"
	tmcfg "github.com/tendermint/tendermint/config"
	tmcli "github.com/tendermint/tendermint/libs/cli"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/debug"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/server"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
)

const flagIsBootstrap = "bootstrap"
const cfgFileName = "config-cli.toml"

var bootstrap bool

func bindFlags(cmd *cobra.Command, v *viper.Viper) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		// Environment variables can't have dashes in them, so bind them to their equivalent
		// keys with underscores, e.g. --favorite-color to STING_FAVORITE_COLOR
		_ = v.BindEnv(f.Name, fmt.Sprintf("%s_%s", "TRUSTLESS_HUB", strings.ToUpper(strings.ReplaceAll(f.Name, "-", "_"))))
		_ = v.BindPFlag(f.Name, f)

		// Apply the viper config value to the flag when the flag is not set and viper has a value
		if !f.Changed && v.IsSet(f.Name) {
			val := v.Get(f.Name)
			_ = cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
		}
	})
}

// NewRootCmd creates a new root command for simd. It is called once in the
// main function.
func NewRootCmd() (*cobra.Command, app.EncodingConfig) {
	encodingConfig := app.MakeEncodingConfig()

	initClientCtx := client.Context{}.
		WithCodec(encodingConfig.Marshaler).
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithTxConfig(encodingConfig.TxConfig).
		WithLegacyAmino(encodingConfig.Amino).
		WithInput(os.Stdin).
		WithAccountRetriever(types.AccountRetriever{}).
		WithBroadcastMode(flags.BroadcastBlock).
		WithHomeDir(app.DefaultNodeHome)

	rootCmd := &cobra.Command{
		Use:   app.Name + "d",
		Short: "The Trustless Hub App Daemon (server)",
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {

			trstAppTemplate, trstAppConfig := initAppConfig()

			if err := server.InterceptConfigsPreRunHandler(cmd, trstAppTemplate, trstAppConfig); err != nil {
				return err
			}

			ctx := server.GetServerContextFromCmd(cmd)

			bindFlags(cmd, ctx.Viper)

			return initConfig(&initClientCtx, cmd)
		},
	}

	initRootCmd(rootCmd, encodingConfig)

	//rosetta
	rootCmd.AddCommand(server.RosettaCommand(encodingConfig.InterfaceRegistry, encodingConfig.Marshaler))
	return rootCmd, encodingConfig
}

// Execute executes the root command.
func Execute(rootCmd *cobra.Command) error {
	// Create and set a client.Context on the command's Context. During the pre-run
	// of the root command, a default initialized client.Context is provided to
	// seed child command execution with values such as AccountRetriver, Keyring,
	// and a Tendermint RPC. This requires the use of a pointer reference when
	// getting and setting the client.Context. Ideally, we utilize
	// https://github.com/spf13/cobra/pull/1118.
	ctx := context.Background()
	ctx = context.WithValue(ctx, client.ClientContextKey, &client.Context{})
	ctx = context.WithValue(ctx, server.ServerContextKey, server.NewDefaultContext())

	rootCmd.PersistentFlags().String(flags.FlagLogLevel, zerolog.InfoLevel.String(), "The logging level (trace|debug|info|warn|error|fatal|panic)")
	rootCmd.PersistentFlags().String(flags.FlagLogFormat, tmcfg.LogFormatPlain, "The logging format (json|plain)")
	executor := tmcli.PrepareBaseCmd(rootCmd, "TRUSTLESS_HUB", app.DefaultNodeHome)
	return executor.ExecuteContext(ctx)
}

func initRootCmd(rootCmd *cobra.Command, encodingConfig app.EncodingConfig) {
	//authclient.Codec = e

	rootCmd.AddCommand(
		//genutilcli.InitCmd(app.ModuleBasics, app.DefaultNodeHome),
		//updateTmParamsAndInit(app.ModuleBasics, app.DefaultNodeHome),
		genutilcli.InitCmd(app.ModuleBasics, app.DefaultNodeHome),
		genutilcli.CollectGenTxsCmd(banktypes.GenesisBalancesIterator{}, app.DefaultNodeHome),
		genutilcli.MigrateGenesisCmd(),
		genutilcli.GenTxCmd(app.ModuleBasics, encodingConfig.TxConfig, banktypes.GenesisBalancesIterator{}, app.DefaultNodeHome),
		genutilcli.ValidateGenesisCmd(app.ModuleBasics),
		AddGenesisAccountCmd(app.DefaultNodeHome),
		AddGenesisWasmMsgCmd(app.DefaultNodeHome),
		ImportGenesisAccountsFromSnapshotCmd(app.DefaultNodeHome),
		ExportAirdropSnapshotCmd(),
		PrepareGenesisCmd(app.DefaultNodeHome, app.ModuleBasics),
		tmcli.NewCompletionCmd(rootCmd, true),
		// testnetCmd(app.ModuleBasics, banktypes.GenesisBalancesIterator{}),
		debug.Cmd(),
	)

	server.AddCommands(rootCmd, app.DefaultNodeHome, newApp, exportAppStateAndTMValidators, addModuleInitFlags)

	// add keybase, auxiliary RPC, query, and tx child commands
	rootCmd.AddCommand(
		rpc.StatusCommand(),
		queryCommand(),
		txCommand(),
		InitAttestationCmd(),
		InitBootstrapCmd(),
		ParseCert(),
		ConfigureCredentialsCmd(),
		HealthCheck(),
		ResetEnclave(),
		keys.Commands(app.DefaultNodeHome),
	)

	// This is needed for `newApp` and `exportAppStateAndTMValidators`
	rootCmd.PersistentFlags().BoolVar(&bootstrap, flagIsBootstrap,
		false, "Start the node as the bootstrap node for the network (only used when starting a new network)")
}

func addModuleInitFlags(startCmd *cobra.Command) {
	crisis.AddModuleInitFlags(startCmd)
}

func queryCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "query",
		Aliases:                    []string{"q"},
		Short:                      "Querying subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		authcmd.GetAccountCmd(),
		rpc.ValidatorCommand(),
		rpc.BlockCommand(),
		authcmd.QueryTxsByEventsCmd(),
		authcmd.QueryTxCmd(),
		//S20GetQueryCmd(),
	)

	app.ModuleBasics.AddQueryCommands(cmd)
	cmd.PersistentFlags().String(flags.FlagChainID, "", "The network chain ID")
	cmd.PersistentFlags().String(tmcli.OutputFlag, "text", "Output format (text|json)")

	return cmd
}

func txCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "tx",
		Short:                      "Transactions subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		authcmd.GetSignCommand(),
		authcmd.GetSignBatchCommand(),
		authcmd.GetMultiSignCommand(),
		authcmd.GetValidateSignaturesCommand(),
		flags.LineBreak,
		authcmd.GetBroadcastCommand(),
		authcmd.GetEncodeCommand(),
		authcmd.GetDecodeCommand(),
		flags.LineBreak,
		//vestingcli.GetTxCmd(),
		//S20GetTxCmd(),
	)

	app.ModuleBasics.AddTxCommands(cmd)
	cmd.PersistentFlags().String(flags.FlagChainID, "", "The network chain ID")
	cmd.PersistentFlags().String(tmcli.OutputFlag, "text", "Output format (text|json)")

	return cmd
}

func newApp(logger log.Logger, db dbm.DB, traceStore io.Writer, appOpts servertypes.AppOptions) servertypes.Application {
	var cache sdk.MultiStorePersistentCache

	if cast.ToBool(appOpts.Get(server.FlagInterBlockCache)) {
		cache = store.NewCommitKVStoreCacheManager()
	}

	skipUpgradeHeights := make(map[int64]bool)
	for _, h := range cast.ToIntSlice(appOpts.Get(server.FlagUnsafeSkipUpgrades)) {
		skipUpgradeHeights[int64(h)] = true
	}

	pruningOpts, err := server.GetPruningOptionsFromFlags(appOpts)
	if err != nil {
		panic(err)
	}

	snapshotDir := filepath.Join(cast.ToString(appOpts.Get(flags.FlagHome)), "data", "snapshots")
	snapshotDB, err := sdk.NewLevelDB("metadata", snapshotDir)
	if err != nil {
		panic(err)
	}
	snapshotStore, err := snapshots.NewStore(snapshotDB, snapshotDir)
	if err != nil {
		panic(err)
	}

	bootstrap := cast.ToBool(appOpts.Get("bootstrap"))
	queryGasLimit := viper.GetUint64("query-gas-limit")

	//fmt.Printf("bootstrap: %s", cast.ToString(bootstrap))

	return app.NewTrstApp(logger, db, traceStore, true, skipUpgradeHeights,
		cast.ToString(appOpts.Get(flags.FlagHome)),
		cast.ToUint(appOpts.Get(server.FlagInvCheckPeriod)),
		queryGasLimit,
		bootstrap,
		appOpts,
		compute.GetConfig(appOpts),
		app.GetEnabledProposals(),
		baseapp.SetPruning(pruningOpts),
		baseapp.SetMinGasPrices(cast.ToString(appOpts.Get(server.FlagMinGasPrices))),
		baseapp.SetHaltHeight(cast.ToUint64(appOpts.Get(server.FlagHaltHeight))),
		baseapp.SetHaltTime(cast.ToUint64(appOpts.Get(server.FlagHaltTime))),
		baseapp.SetMinRetainBlocks(cast.ToUint64(appOpts.Get(server.FlagMinRetainBlocks))),
		baseapp.SetInterBlockCache(cache),
		baseapp.SetTrace(cast.ToBool(appOpts.Get(server.FlagTrace))),
		baseapp.SetIndexEvents(cast.ToStringSlice(appOpts.Get(server.FlagIndexEvents))),
		baseapp.SetSnapshotStore(snapshotStore),
		baseapp.SetSnapshotInterval(cast.ToUint64(appOpts.Get(server.FlagStateSyncSnapshotInterval))),
		baseapp.SetSnapshotKeepRecent(cast.ToUint32(appOpts.Get(server.FlagStateSyncSnapshotKeepRecent))),
	)
}

func exportAppStateAndTMValidators(
	logger log.Logger, db dbm.DB, traceStore io.Writer, height int64, forZeroHeight bool, jailWhiteList []string, appOpts servertypes.AppOptions,
) (servertypes.ExportedApp, error) {

	bootstrap := viper.GetBool("bootstrap")
	queryGasLimit := viper.GetUint64("query-gas-limit")

	encCfg := app.MakeEncodingConfig()
	encCfg.Marshaler = codec.NewProtoCodec(encCfg.InterfaceRegistry)
	var wasmApp *app.App
	if height != -1 {
		wasmApp = app.NewTrstApp(logger, db, traceStore, false, map[int64]bool{}, "", uint(1), queryGasLimit, bootstrap, appOpts, compute.DefaultWasmConfig(), app.GetEnabledProposals())

		if err := wasmApp.LoadHeight(height); err != nil {
			return servertypes.ExportedApp{}, err
		}
	} else {
		wasmApp = app.NewTrstApp(logger, db, traceStore, true, map[int64]bool{}, "", uint(1), queryGasLimit, bootstrap, appOpts, compute.DefaultWasmConfig(), app.GetEnabledProposals())
	}
	return wasmApp.ExportAppStateAndValidators(forZeroHeight, jailWhiteList)
}

// writeParamsAndConfigCmd patches the write-params cmd to additionally update the app pruning config.
func updateTmParamsAndInit(mbm module.BasicManager, defaultNodeHome string) *cobra.Command {
	cmd := genutilcli.InitCmd(mbm, defaultNodeHome)
	originalFunc := cmd.RunE

	wrappedFunc := func(cmd *cobra.Command, args []string) error {
		ctx := server.GetServerContextFromCmd(cmd)

		// time is in NS
		ctx.Config.Consensus.TimeoutPrecommit = 2_000_000_000

		appConfigFilePath := filepath.Join(defaultNodeHome, "config/app.toml")
		appConf, _ := serverconfig.ParseConfig(viper.GetViper())
		appConf.MinGasPrices = "0.00025utrst"

		serverconfig.WriteConfigFile(appConfigFilePath, appConf)

		if err := originalFunc(cmd, args); err != nil {
			return err
		}

		return nil
	}

	cmd.RunE = wrappedFunc
	return cmd
}

func initConfig(ctx *client.Context, cmd *cobra.Command) error {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(prefix.Bech32PrefixAccAddr, prefix.Bech32PrefixAccPub)
	config.SetBech32PrefixForValidator(prefix.Bech32PrefixValAddr, prefix.Bech32PrefixValPub)
	config.SetBech32PrefixForConsensusNode(prefix.Bech32PrefixConsAddr, prefix.Bech32PrefixConsPub)

	cfgFilePath := filepath.Join(app.DefaultNodeHome, "config", cfgFileName)
	if _, err := os.Stat(cfgFilePath); err == nil {
		viper.SetConfigFile(cfgFilePath)

		if err := viper.ReadInConfig(); err != nil {
			return err
		}
	}

	cfgFlags := []string{flags.FlagChainID, flags.FlagKeyringBackend}
	for _, flag := range cfgFlags {
		err := setFlagFromConfig(cmd, flag)
		if err != nil {
			return err
		}
	}

	return client.SetCmdClientContextHandler(*ctx, cmd)
}

func setFlagFromConfig(cmd *cobra.Command, flag string) error {
	if viper.GetString(flag) != "" && cmd.Flags().Lookup(flag) != nil {
		err := cmd.Flags().Set(flag, viper.GetString(flag))
		if err != nil {
			return err
		}
	}
	return nil
}
