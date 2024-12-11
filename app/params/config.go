package appparams

import (
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
)

// CustomAppConfig defines the configuration for the Nois app.
type CustomAppConfig struct {
	serverconfig.Config
	Wasm wasmtypes.WasmConfig `mapstructure:"wasm" json:"wasm"`
}

func CustomconfigTemplate(config wasmtypes.WasmConfig) string {
	return serverconfig.DefaultConfigTemplate + wasmtypes.ConfigTemplate(config)
}

func DefaultConfig() (string, interface{}) {
	serverConfig := serverconfig.DefaultConfig()
	serverConfig.MinGasPrices = "0uinto"

	wasmConfig := wasmtypes.DefaultWasmConfig()
	simulationLimit := uint64(50_000_000)

	wasmConfig.SimulationGasLimit = &simulationLimit
	wasmConfig.SmartQueryGasLimit = 25_000_000
	wasmConfig.MemoryCacheSize = 512
	wasmConfig.ContractDebugMode = false

	customConfig := CustomAppConfig{
		Config: *serverConfig,
		Wasm:   wasmConfig,
	}

	return CustomconfigTemplate(wasmConfig), customConfig
}
