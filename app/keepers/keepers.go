package keepers

import (
	"path/filepath"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	crisiskeeper "github.com/cosmos/cosmos-sdk/x/crisis/keeper"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	evidencekeeper "github.com/cosmos/cosmos-sdk/x/evidence/keeper"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	feegrantkeeper "github.com/cosmos/cosmos-sdk/x/feegrant/keeper"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	paramproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
	upgradekeeper "github.com/cosmos/cosmos-sdk/x/upgrade/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	icacontroller "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/controller"
	icacontrollerkeeper "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/controller/keeper"
	icacontrollertypes "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/controller/types"
	icahost "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/host"
	icahostkeeper "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/host/keeper"
	icahosttypes "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/host/types"
	"github.com/cosmos/ibc-go/v3/modules/apps/transfer"
	ibctransferkeeper "github.com/cosmos/ibc-go/v3/modules/apps/transfer/keeper"
	ibctransfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	ibcclient "github.com/cosmos/ibc-go/v3/modules/core/02-client"
	ibcclienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	porttypes "github.com/cosmos/ibc-go/v3/modules/core/05-port/types"
	ibchost "github.com/cosmos/ibc-go/v3/modules/core/24-host"
	ibckeeper "github.com/cosmos/ibc-go/v3/modules/core/keeper"
	allockeeper "github.com/trstlabs/trst/x/alloc/keeper"
	alloctypes "github.com/trstlabs/trst/x/alloc/types"
	icaauth "github.com/trstlabs/trst/x/auto-ibc-tx"
	icaauthkeeper "github.com/trstlabs/trst/x/auto-ibc-tx/keeper"
	icaauthtypes "github.com/trstlabs/trst/x/auto-ibc-tx/types"
	"github.com/trstlabs/trst/x/compute"
	mintkeeper "github.com/trstlabs/trst/x/mint/keeper"
	minttypes "github.com/trstlabs/trst/x/mint/types"
	reg "github.com/trstlabs/trst/x/registration"

	claimkeeper "github.com/trstlabs/trst/x/claim/keeper"
	claimtypes "github.com/trstlabs/trst/x/claim/types"
)

type TrstAppKeepers struct {
	// keepers
	AccountKeeper       *authkeeper.AccountKeeper
	AuthzKeeper         *authzkeeper.Keeper
	BankKeeper          *bankkeeper.BaseKeeper
	CapabilityKeeper    *capabilitykeeper.Keeper
	StakingKeeper       *stakingkeeper.Keeper
	SlashingKeeper      *slashingkeeper.Keeper
	MintKeeper          *mintkeeper.Keeper
	DistrKeeper         *distrkeeper.Keeper
	GovKeeper           *govkeeper.Keeper
	CrisisKeeper        *crisiskeeper.Keeper
	UpgradeKeeper       *upgradekeeper.Keeper
	ParamsKeeper        *paramskeeper.Keeper
	EvidenceKeeper      *evidencekeeper.Keeper
	FeegrantKeeper      *feegrantkeeper.Keeper
	ComputeKeeper       *compute.Keeper
	RegKeeper           *reg.Keeper
	IbcKeeper           *ibckeeper.Keeper // IBC Keeper must be a pointer in the app, so we can SetRouter on it correctly
	TransferKeeper      *ibctransferkeeper.Keeper
	ClaimKeeper         *claimkeeper.Keeper
	AllocKeeper         *allockeeper.Keeper
	ICAControllerKeeper *icacontrollerkeeper.Keeper
	ICAHostKeeper       *icahostkeeper.Keeper
	ICAAuthKeeper       *icaauthkeeper.Keeper

	// make scoped keepers public for test purposes
	ScopedIBCKeeper      capabilitykeeper.ScopedKeeper
	ScopedTransferKeeper capabilitykeeper.ScopedKeeper

	ScopedICAControllerKeeper capabilitykeeper.ScopedKeeper
	ScopedICAHostKeeper       capabilitykeeper.ScopedKeeper
	ScopedICAAuthKeeper       capabilitykeeper.ScopedKeeper

	ScopedComputeKeeper capabilitykeeper.ScopedKeeper

	//

	// keys to access the substores
	keys    map[string]*sdk.KVStoreKey
	tKeys   map[string]*sdk.TransientStoreKey
	memKeys map[string]*sdk.MemoryStoreKey
}

func (ak *TrstAppKeepers) GetKeys() map[string]*sdk.KVStoreKey {
	return ak.keys
}

func (ak *TrstAppKeepers) GetTransientStoreKeys() map[string]*sdk.TransientStoreKey {
	return ak.tKeys
}

func (ak *TrstAppKeepers) GetMemoryStoreKeys() map[string]*sdk.MemoryStoreKey {
	return ak.memKeys
}

func (ak *TrstAppKeepers) GetKey(key string) *sdk.KVStoreKey {
	return ak.keys[key]
}

// getSubspace returns a param subspace for a given module name.
func (ak *TrstAppKeepers) GetSubspace(moduleName string) paramstypes.Subspace {
	subspace, _ := ak.ParamsKeeper.GetSubspace(moduleName)
	return subspace
}

func (ak *TrstAppKeepers) InitSdkKeepers(
	appCodec codec.Codec,
	legacyAmino *codec.LegacyAmino,
	app *baseapp.BaseApp,
	maccPerms map[string][]string,
	blockedAddresses map[string]bool,
	invCheckPeriod uint,
	skipUpgradeHeights map[int64]bool,
	homePath string,
) {
	paramsKeeper := initParamsKeeper(appCodec, legacyAmino, ak.keys[paramstypes.StoreKey], ak.tKeys[paramstypes.TStoreKey])
	ak.ParamsKeeper = &paramsKeeper

	// set the BaseApp's parameter store
	app.SetParamStore(ak.ParamsKeeper.Subspace(baseapp.Paramspace).WithKeyTable(paramskeeper.ConsensusParamsKeyTable()))

	// add keepers
	accountKeeper := authkeeper.NewAccountKeeper(
		appCodec, ak.keys[authtypes.StoreKey], ak.GetSubspace(authtypes.ModuleName), authtypes.ProtoBaseAccount, maccPerms,
	)
	ak.AccountKeeper = &accountKeeper

	bankKeeper := bankkeeper.NewBaseKeeper(
		appCodec, ak.keys[banktypes.StoreKey], ak.AccountKeeper, ak.GetSubspace(banktypes.ModuleName), blockedAddresses,
	)
	ak.BankKeeper = &bankKeeper

	stakingKeeper := stakingkeeper.NewKeeper(
		appCodec, ak.keys[stakingtypes.StoreKey], ak.AccountKeeper, *ak.BankKeeper, ak.GetSubspace(stakingtypes.ModuleName),
	)
	ak.StakingKeeper = &stakingKeeper

	distrKeeper := distrkeeper.NewKeeper(
		appCodec, ak.keys[distrtypes.StoreKey], ak.GetSubspace(distrtypes.ModuleName), ak.AccountKeeper, *ak.BankKeeper,
		ak.StakingKeeper, authtypes.FeeCollectorName, blockedAddresses,
	)
	ak.DistrKeeper = &distrKeeper

	slashkingKeeper := slashingkeeper.NewKeeper(
		appCodec, ak.keys[slashingtypes.StoreKey], ak.StakingKeeper, ak.GetSubspace(slashingtypes.ModuleName),
	)
	ak.SlashingKeeper = &slashkingKeeper

	crisisKeeper := crisiskeeper.NewKeeper(
		ak.GetSubspace(crisistypes.ModuleName), invCheckPeriod, *ak.BankKeeper, authtypes.FeeCollectorName,
	)
	ak.CrisisKeeper = &crisisKeeper

	feegrantKeeper := feegrantkeeper.NewKeeper(appCodec, ak.keys[feegrant.StoreKey], ak.AccountKeeper)
	ak.FeegrantKeeper = &feegrantKeeper

	authzKeeper := authzkeeper.NewKeeper(ak.keys[authzkeeper.StoreKey], appCodec, app.MsgServiceRouter())
	ak.AuthzKeeper = &authzKeeper

	upgradeKeeper := upgradekeeper.NewKeeper(skipUpgradeHeights, ak.keys[upgradetypes.StoreKey], appCodec, homePath, app)
	ak.UpgradeKeeper = &upgradeKeeper

	// add capability keeper and ScopeToModule for ibc module
	ak.CapabilityKeeper = capabilitykeeper.NewKeeper(appCodec, ak.keys[capabilitytypes.StoreKey], ak.memKeys[capabilitytypes.MemStoreKey])
	ak.CreateScopedKeepers()

	// Create IBC Keeper
	ak.IbcKeeper = ibckeeper.NewKeeper(
		appCodec, ak.keys[ibchost.StoreKey], ak.GetSubspace(ibchost.ModuleName), ak.StakingKeeper, ak.UpgradeKeeper, ak.ScopedIBCKeeper,
	)

	// Create evidence keeper with router
	ak.EvidenceKeeper = evidencekeeper.NewKeeper(
		appCodec, ak.keys[evidencetypes.StoreKey], ak.StakingKeeper, ak.SlashingKeeper,
	)

}

func (ak *TrstAppKeepers) CreateScopedKeepers() {
	ak.ScopedIBCKeeper = ak.CapabilityKeeper.ScopeToModule(ibchost.ModuleName)
	ak.ScopedTransferKeeper = ak.CapabilityKeeper.ScopeToModule(ibctransfertypes.ModuleName)
	ak.ScopedICAControllerKeeper = ak.CapabilityKeeper.ScopeToModule(icacontrollertypes.SubModuleName)
	ak.ScopedICAHostKeeper = ak.CapabilityKeeper.ScopeToModule(icahosttypes.SubModuleName)
	ak.ScopedICAAuthKeeper = ak.CapabilityKeeper.ScopeToModule(icaauthtypes.ModuleName)
	ak.ScopedComputeKeeper = ak.CapabilityKeeper.ScopeToModule(compute.ModuleName)
	// Applications that wish to enforce statically created ScopedKeepers should call `Seal` after creating
	// their scoped modules in `NewApp` with `ScopeToModule`
	ak.CapabilityKeeper.Seal()
}

func (ak *TrstAppKeepers) InitCustomKeepers(
	appCodec codec.Codec,
	legacyAmino *codec.LegacyAmino,
	app *baseapp.BaseApp,
	bootstrap bool,
	homePath string,
	computeConfig *compute.WasmConfig,
	enabledProposals []compute.ProposalType,
) {

	mintKeeper := mintkeeper.NewKeeper(
		appCodec, ak.keys[minttypes.StoreKey], ak.GetSubspace(minttypes.ModuleName),
		ak.AccountKeeper, ak.BankKeeper, authtypes.FeeCollectorName,
	)

	ak.MintKeeper = &mintKeeper
	claimKeeper := claimkeeper.NewKeeper(
		appCodec,
		ak.keys[claimtypes.StoreKey],
		ak.AccountKeeper,
		ak.BankKeeper,
		ak.StakingKeeper,
		ak.DistrKeeper,
	)
	ak.ClaimKeeper = &claimKeeper

	// Register the staking hooks
	// NOTE: StakingKeeper above is passed by reference, so that it will contain these hooks
	ak.StakingKeeper.SetHooks(
		stakingtypes.NewMultiStakingHooks(
			ak.DistrKeeper.Hooks(),
			ak.SlashingKeeper.Hooks(),
			ak.ClaimKeeper.Hooks()),
	)

	ak.AllocKeeper = allockeeper.NewKeeper(
		appCodec,
		ak.keys[alloctypes.StoreKey],
		ak.AccountKeeper,
		ak.BankKeeper,
		ak.StakingKeeper,
		ak.DistrKeeper,
		ak.GetSubspace(alloctypes.ModuleName),
	)
	// Just re-use the full router - do we want to limit this more?
	regRouter := app.Router()

	// Replace with bootstrap flag when we figure out how to test properly and everything works
	regKeeper := reg.NewKeeper(appCodec, ak.keys[reg.StoreKey], regRouter, reg.EnclaveApi{}, homePath, bootstrap)
	ak.RegKeeper = &regKeeper

	icaHostKeeper := icahostkeeper.NewKeeper(
		appCodec, ak.keys[icahosttypes.StoreKey], ak.GetSubspace(icahosttypes.SubModuleName),
		ak.IbcKeeper.ChannelKeeper, &ak.IbcKeeper.PortKeeper,
		ak.AccountKeeper, ak.ScopedICAHostKeeper, app.MsgServiceRouter(),
	)
	ak.ICAHostKeeper = &icaHostKeeper

	icaControllerKeeper := icacontrollerkeeper.NewKeeper(
		appCodec, ak.keys[icacontrollertypes.StoreKey], ak.GetSubspace(icacontrollertypes.SubModuleName),
		ak.IbcKeeper.ChannelKeeper, ak.IbcKeeper.ChannelKeeper, &ak.IbcKeeper.PortKeeper,
		ak.ScopedICAControllerKeeper, app.MsgServiceRouter(),
	)
	ak.ICAControllerKeeper = &icaControllerKeeper

	icaAuthKeeper := icaauthkeeper.NewKeeper(appCodec, ak.keys[icaauthtypes.StoreKey], *ak.ICAControllerKeeper, ak.ScopedICAAuthKeeper, ak.BankKeeper, *ak.DistrKeeper, *ak.StakingKeeper, *ak.AccountKeeper, ak.GetSubspace(icaauthtypes.ModuleName))
	ak.ICAAuthKeeper = &icaAuthKeeper

	icaAuthIBCModule := icaauth.NewIBCModule(*ak.ICAAuthKeeper)

	// Create Transfer Keepers
	transferKeeper := ibctransferkeeper.NewKeeper(
		appCodec, ak.keys[ibctransfertypes.StoreKey], ak.GetSubspace(ibctransfertypes.ModuleName),
		ak.IbcKeeper.ChannelKeeper, ak.IbcKeeper.ChannelKeeper, &ak.IbcKeeper.PortKeeper,
		ak.AccountKeeper, ak.BankKeeper, ak.ScopedTransferKeeper,
	)
	ak.TransferKeeper = &transferKeeper

	// Create static IBC router, add ibc-tranfer module route, then set and seal it
	ibcRouter := porttypes.NewRouter()

	icaControllerIBCModule := icacontroller.NewIBCModule(*ak.ICAControllerKeeper, icaAuthIBCModule)
	icaHostIBCModule := icahost.NewIBCModule(*ak.ICAHostKeeper)

	computeDir := filepath.Join(homePath, ".compute")
	// The last arguments can contain custom message handlers, and custom query handlers,
	// if we want to allow any custom callbacks
	supportedFeatures := "staking,stargate,ibc3" //"staking,stargate,ibc3"

	computeKeeper := compute.NewKeeper(
		appCodec,
		*legacyAmino,
		ak.keys[compute.StoreKey],
		*ak.AccountKeeper,
		ak.BankKeeper,
		*ak.DistrKeeper,
		mintKeeper,
		*ak.StakingKeeper,
		ak.ScopedComputeKeeper,
		&ak.IbcKeeper.PortKeeper,
		ak.TransferKeeper,
		ak.IbcKeeper.ChannelKeeper,
		app.Router(),
		app.MsgServiceRouter(),
		app.GRPCQueryRouter(),
		computeDir,
		computeConfig,
		supportedFeatures,
		nil,
		nil, ak.GetSubspace(compute.ModuleName), compute.NewMultiComputeHooks(ak.ClaimKeeper.Hooks()),
	)
	ak.ComputeKeeper = &computeKeeper

	// register the proposal types
	govRouter := govtypes.NewRouter()
	// The gov proposal types can be individually enabled
	if len(enabledProposals) != 0 {
		govRouter.AddRoute(compute.RouterKey, compute.NewWasmProposalHandler(*ak.ComputeKeeper, enabledProposals))
	}

	govRouter.AddRoute(govtypes.RouterKey, govtypes.ProposalHandler).
		AddRoute(paramproposal.RouterKey, params.NewParamChangeProposalHandler(*ak.ParamsKeeper)).
		AddRoute(distrtypes.RouterKey, distr.NewCommunityPoolSpendProposalHandler(*ak.DistrKeeper)).
		AddRoute(upgradetypes.RouterKey, upgrade.NewSoftwareUpgradeProposalHandler(*ak.UpgradeKeeper)).
		AddRoute(ibcclienttypes.RouterKey, ibcclient.NewClientProposalHandler(ak.IbcKeeper.ClientKeeper))

	govKeeper := govkeeper.NewKeeper(
		appCodec, ak.keys[govtypes.StoreKey], ak.GetSubspace(govtypes.ModuleName), ak.AccountKeeper, ak.BankKeeper,
		ak.StakingKeeper, govRouter,
	)
	ak.GovKeeper = &govKeeper

	ibcRouter.AddRoute(compute.ModuleName, compute.NewIBCHandler(ak.ComputeKeeper, ak.IbcKeeper.ChannelKeeper)).
		AddRoute(ibctransfertypes.ModuleName, transfer.NewIBCModule(*ak.TransferKeeper)).
		AddRoute(icacontrollertypes.SubModuleName, icaControllerIBCModule).
		AddRoute(icahosttypes.SubModuleName, icaHostIBCModule).
		AddRoute(icaauthtypes.ModuleName, icaControllerIBCModule)

	// Setting Router will finalize all routes by sealing router
	// No more routes can be added
	ak.IbcKeeper.SetRouter(ibcRouter)

}

func (ak *TrstAppKeepers) InitKeys() {
	ak.keys = sdk.NewKVStoreKeys(
		authtypes.StoreKey,
		banktypes.StoreKey,
		stakingtypes.StoreKey,
		minttypes.StoreKey,
		distrtypes.StoreKey,
		slashingtypes.StoreKey,
		govtypes.StoreKey,
		paramstypes.StoreKey,
		ibchost.StoreKey,
		upgradetypes.StoreKey,
		evidencetypes.StoreKey,
		ibctransfertypes.StoreKey,
		capabilitytypes.StoreKey,
		compute.StoreKey,
		reg.StoreKey,
		feegrant.StoreKey,
		authzkeeper.StoreKey,
		icahosttypes.StoreKey,
		claimtypes.StoreKey,
		alloctypes.StoreKey,
		icaauthtypes.StoreKey,
		icacontrollertypes.StoreKey,
	)

	ak.tKeys = sdk.NewTransientStoreKeys(paramstypes.TStoreKey)
	ak.memKeys = sdk.NewMemoryStoreKeys(capabilitytypes.MemStoreKey)
}

// initParamsKeeper init params keeper and its subspaces
func initParamsKeeper(appCodec codec.BinaryCodec, legacyAmino *codec.LegacyAmino, key, tkey sdk.StoreKey) paramskeeper.Keeper {
	paramsKeeper := paramskeeper.NewKeeper(appCodec, legacyAmino, key, tkey)

	paramsKeeper.Subspace(authtypes.ModuleName)
	paramsKeeper.Subspace(banktypes.ModuleName)
	paramsKeeper.Subspace(stakingtypes.ModuleName)
	paramsKeeper.Subspace(minttypes.ModuleName)
	paramsKeeper.Subspace(distrtypes.ModuleName)
	paramsKeeper.Subspace(slashingtypes.ModuleName)
	paramsKeeper.Subspace(ibctransfertypes.ModuleName)
	paramsKeeper.Subspace(ibchost.ModuleName)
	paramsKeeper.Subspace(icahosttypes.SubModuleName)
	paramsKeeper.Subspace(icacontrollertypes.SubModuleName)
	paramsKeeper.Subspace(govtypes.ModuleName).WithKeyTable(govtypes.ParamKeyTable())
	paramsKeeper.Subspace(crisistypes.ModuleName)
	paramsKeeper.Subspace(compute.ModuleName)
	paramsKeeper.Subspace(reg.ModuleName)
	paramsKeeper.Subspace(claimtypes.ModuleName)
	paramsKeeper.Subspace(alloctypes.ModuleName)
	paramsKeeper.Subspace(icaauthtypes.ModuleName)

	return paramsKeeper
}
