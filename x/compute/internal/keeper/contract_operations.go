package keeper

import (
	"bytes"

	//"encoding/json"
	"fmt"
	"time"

	//"log"

	codedctypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/telemetry"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	sdktx "github.com/cosmos/cosmos-sdk/types/tx"
	sdktxsigning "github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"

	wasmTypes "github.com/trstlabs/trst/go-cosmwasm/types"
	"github.com/trstlabs/trst/x/compute/internal/types"
)

// Create uploads and compiles a WASM contract, returning a short identifier for the contract
func (k Keeper) Create(ctx sdk.Context, creator sdk.AccAddress, wasmCode []byte, source string, builder string, maxDuration time.Duration, title string, description string) (codeID uint64, err error) {

	wasmCode, err = uncompress(wasmCode)
	if err != nil {
		return 0, sdkerrors.Wrap(types.ErrCreateFailed, err.Error())
	}
	ctx.GasMeter().ConsumeGas(types.CompileCost*uint64(len(wasmCode)), "Compiling WASM Bytecode")

	codeHash, err := k.wasmer.Create(wasmCode)
	if err != nil {
		// return 0, sdkerrors.Wrap(err, "cosmwasm create")
		return 0, sdkerrors.Wrap(types.ErrCreateFailed, err.Error())
	}

	//hash = string(codeHash)
	store := ctx.KVStore(k.storeKey)
	codeID = k.autoIncrementID(ctx, types.KeyLastCodeID)
	/*
		if instantiateAccess == nil {
			defaultAccessConfig := k.getInstantiateAccessConfig(ctx).With(creator)
			instantiateAccess = &defaultAccessConfig
		}
	*/
	codeInfo := types.NewCodeInfo(codeHash, creator, source, builder, maxDuration /* , *instantiateAccess */, title, description)
	// 0x01 | codeID (uint64) -> ContractInfo
	store.Set(types.GetCodeKey(codeID), k.cdc.MustMarshal(&codeInfo))

	return codeID, nil
}

func (k Keeper) importCode(ctx sdk.Context, codeID uint64, codeInfo types.CodeInfo, wasmCode []byte) error {
	wasmCode, err := uncompress(wasmCode)
	if err != nil {
		return sdkerrors.Wrap(types.ErrCreateFailed, err.Error())
	}
	newCodeHash, err := k.wasmer.Create(wasmCode)
	if err != nil {
		return sdkerrors.Wrap(types.ErrCreateFailed, err.Error())
	}
	if !bytes.Equal(codeInfo.CodeHash, newCodeHash) {
		return sdkerrors.Wrap(types.ErrInvalid, "code hashes not same")
	}

	store := ctx.KVStore(k.storeKey)
	key := types.GetCodeKey(codeID)
	if store.Has(key) {
		return sdkerrors.Wrapf(types.ErrDuplicate, "duplicate code: %d", codeID)
	}
	// 0x01 | codeID (uint64) -> ContractInfo
	store.Set(key, k.cdc.MustMarshal(&codeInfo))
	return nil
}

func (k Keeper) GetSignerInfo(ctx sdk.Context, signer sdk.AccAddress) ([]byte, sdktxsigning.SignMode, []byte, []byte, []byte, error) {
	tx := sdktx.Tx{}
	err := k.cdc.Unmarshal(ctx.TxBytes(), &tx)
	if err != nil {
		return nil, 0, nil, nil, nil, sdkerrors.Wrap(types.ErrSigFailed, fmt.Sprintf("Unable to decode transaction from bytes: %s", err.Error()))
	}

	// for MsgInstantiateContract, there is only one signer which is msg.Sender
	// (https://github.com/enigmampc/SecretNetwork/blob/d7813792fa07b93a10f0885eaa4c5e0a0a698854/x/compute/internal/types/msg.go#L192-L194)
	signerAcc, err := ante.GetSignerAcc(ctx, k.accountKeeper, signer)
	if err != nil {

		return nil, 0, nil, nil, nil, sdkerrors.Wrap(types.ErrSigFailed, fmt.Sprintf("Unable to retrieve account by address: %s", err.Error()))
	}

	txConfig := authtx.NewTxConfig(k.cdc.(*codec.ProtoCodec), authtx.DefaultSignModes)
	modeHandler := txConfig.SignModeHandler()
	signingData := authsigning.SignerData{
		ChainID:       ctx.ChainID(),
		AccountNumber: signerAcc.GetAccountNumber(),
		Sequence:      signerAcc.GetSequence() - 1,
	}

	protobufTx := authtx.WrapTx(&tx).GetTx()

	pubKeys, err := protobufTx.GetPubKeys()
	if err != nil {
		return nil, 0, nil, nil, nil, sdkerrors.Wrap(types.ErrSigFailed, fmt.Sprintf("Unable to get public keys for instantiate: %s", err.Error()))
	}

	pkIndex := -1
	var _signers [][]byte // This is just used for the error message below
	for index, pubKey := range pubKeys {
		thisSigner := pubKey.Address().Bytes()
		_signers = append(_signers, thisSigner)
		if bytes.Equal(thisSigner, signer.Bytes()) {
			pkIndex = index
		}
	}
	if pkIndex == -1 {
		return nil, 0, nil, nil, nil, sdkerrors.Wrap(types.ErrSigFailed, fmt.Sprintf("Message sender: %v is not found in the tx signer set: %v, callback signature not provided", signer, _signers))
	}

	signatures, _ := protobufTx.GetSignaturesV2()
	var signMode sdktxsigning.SignMode
	switch signData := signatures[pkIndex].Data.(type) {
	case *sdktxsigning.SingleSignatureData:
		signMode = signData.SignMode
	case *sdktxsigning.MultiSignatureData:
		signMode = sdktxsigning.SignMode_SIGN_MODE_LEGACY_AMINO_JSON
	}

	signBytes, err := modeHandler.GetSignBytes(signMode, signingData, protobufTx)
	if err != nil {
		return nil, 0, nil, nil, nil, sdkerrors.Wrap(types.ErrSigFailed, fmt.Sprintf("Unable to recreate sign bytes for the tx: %s", err.Error()))
	}

	modeInfoBytes, err := sdktxsigning.SignatureDataToProto(signatures[pkIndex].Data).Marshal()
	if err != nil {
		return nil, 0, nil, nil, nil, sdkerrors.Wrap(types.ErrSigFailed, "couldn't marshal mode info")
	}

	var pkBytes []byte
	pubKey := pubKeys[pkIndex]
	anyPubKey, err := codedctypes.NewAnyWithValue(pubKey)
	if err != nil {
		return nil, 0, nil, nil, nil, sdkerrors.Wrap(types.ErrSigFailed, "couldn't turn public key into Any")
	}

	pkBytes, err = k.cdc.Marshal(anyPubKey)
	if err != nil {
		return nil, 0, nil, nil, nil, sdkerrors.Wrap(types.ErrSigFailed, "couldn't marshal public key")
	}

	return signBytes, signMode, modeInfoBytes, pkBytes, tx.Signatures[pkIndex], nil
}

// Instantiate creates an instance of a WASM contract
func (k Keeper) Instantiate(ctx sdk.Context, codeID uint64, creator /* , admin */ sdk.AccAddress, initMsg []byte, autoMsg []byte, id string, deposit sdk.Coins, callbackSig []byte, customDuration time.Duration) (sdk.AccAddress, error) {

	ctx.GasMeter().ConsumeGas(types.InstanceCost, "Loading CosmWasm module: init")

	signBytes := []byte{}
	signMode := sdktxsigning.SignMode_SIGN_MODE_UNSPECIFIED
	modeInfoBytes := []byte{}
	pkBytes := []byte{}
	signerSig := []byte{}
	var err error

	fmt.Printf("callbackSig.. %s \n", callbackSig)

	// If no callback signature - we should send the actual msg sender sign bytes and signature
	if callbackSig == nil {
		signBytes, signMode, modeInfoBytes, pkBytes, signerSig, err = k.GetSignerInfo(ctx, creator)
		if err != nil {
			return nil, err
		}
	}

	verificationInfo := types.NewVerificationInfo(signBytes, signMode, modeInfoBytes, pkBytes, signerSig, callbackSig)

	store := ctx.KVStore(k.storeKey)
	existingAddress := store.Get(types.GetContractIdPrefix(id))

	if existingAddress != nil {
		return nil, sdkerrors.Wrap(types.ErrAccountExists, id)
	}

	contractAddress := k.generateContractAddress(ctx, codeID)
	existingAcct := k.accountKeeper.GetAccount(ctx, contractAddress)
	if existingAcct != nil {
		return nil, sdkerrors.Wrap(types.ErrAccountExists, existingAcct.GetAddress().String())
	}

	// deposit initial contract funds
	if !deposit.IsZero() {
		if k.bankKeeper.BlockedAddr(creator) {
			return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "blocked address can not be used")
		}
		sdkerr := k.bankKeeper.SendCoins(ctx, creator, contractAddress, deposit)
		if sdkerr != nil {
			return nil, sdkerr
		}
	} else {
		// create an empty account (so we don't have issues later)
		contractAccount := k.accountKeeper.NewAccountWithAddress(ctx, contractAddress)
		k.accountKeeper.SetAccount(ctx, contractAccount)
	}

	// get contact info

	bz := store.Get(types.GetCodeKey(codeID))
	if bz == nil {
		return nil, sdkerrors.Wrap(types.ErrNotFound, "code")
	}

	var codeInfo types.CodeInfo
	k.cdc.MustUnmarshal(bz, &codeInfo)

	// prepare params for contract instantiate call
	params := types.NewEnv(ctx, creator, deposit, contractAddress, nil)

	// create prefixed data store
	// 0x03 | contractAddress (sdk.AccAddress)
	prefixStoreKey := types.GetContractStorePrefixKey(contractAddress)
	prefixStore := prefix.NewStore(ctx.KVStore(k.storeKey), prefixStoreKey)

	// prepare querier
	querier := QueryHandler{
		Ctx:     ctx,
		Plugins: k.queryPlugins,
	}
	if autoMsg == nil {
		autoMsg = make([]byte, 256)
	}
	// instantiate wasm contract
	gas := gasForContract(ctx)
	res, key, callbackSig, gasUsed, err := k.wasmer.Instantiate(codeInfo.CodeHash, params, initMsg, autoMsg, prefixStore, cosmwasmAPI, querier, ctx.GasMeter(), gas, verificationInfo)
	//fmt.Printf("ress")
	consumeGas(ctx, gasUsed)
	if err != nil {
		return contractAddress, sdkerrors.Wrap(types.ErrInstantiateFailed, err.Error())
	}

	// emit all events from this contract itself
	events := types.ParseEvents(res.Log, contractAddress)
	ctx.EventManager().EmitEvents(events)

	// persist instance
	createdAt := types.NewAbsoluteTxPosition(ctx)
	//fmt.Printf("created")
	endTime := ctx.BlockHeader().Time.Add(codeInfo.Duration)
	if customDuration != 0 {
		endTime = ctx.BlockHeader().Time.Add(customDuration)
	}
	if autoMsg != nil {
		instance := types.NewContractInfo(codeID, creator /* admin, */, id, createdAt, endTime, autoMsg, callbackSig)
		store.Set(types.GetContractAddressKey(contractAddress), k.cdc.MustMarshal(&instance))
	} else {
		instance := types.NewContractInfo(codeID, creator /* admin, */, id, createdAt, endTime, nil, nil)
		store.Set(types.GetContractAddressKey(contractAddress), k.cdc.MustMarshal(&instance))
	}

	codeInfo.Instances = codeInfo.Instances + 1
	store.Set(types.GetCodeKey(codeID), k.cdc.MustMarshal(&codeInfo))
	store.Set(types.GetContractEnclaveKey(contractAddress), key)
	store.Set(types.GetContractIdPrefix(id), contractAddress)

	k.InsertContractQueue(ctx, contractAddress.String(), endTime)

	k.hooks.AfterComputeInstantiated(ctx, creator)

	err = k.dispatchMessages(ctx, contractAddress, res.Messages)
	if err != nil {
		return nil, err
	}

	return contractAddress, nil
}

// Execute executes the contract instance
func (k Keeper) Execute(ctx sdk.Context, contractAddress sdk.AccAddress, caller sdk.AccAddress, msg []byte, coins sdk.Coins, callbackSig []byte) (*sdk.Result, error) {
	ctx.GasMeter().ConsumeGas(types.InstanceCost, "Loading compute module: execute")
	//fmt.Printf("executing caller %s ", caller)
	signBytes := []byte{}
	signMode := sdktxsigning.SignMode_SIGN_MODE_UNSPECIFIED
	modeInfoBytes := []byte{}
	pkBytes := []byte{}
	signerSig := []byte{}
	var err error

	// If no callback signature - we should send the actual msg sender sign bytes and signature
	if callbackSig == nil {
		signBytes, signMode, modeInfoBytes, pkBytes, signerSig, err = k.GetSignerInfo(ctx, caller)
		if err != nil {

			return nil, err
		}
		//fmt.Printf("executing sig %s \n", signMode)
	}

	verificationInfo := types.NewVerificationInfo(signBytes, signMode, modeInfoBytes, pkBytes, signerSig, callbackSig)

	codeInfo, prefixStore, err := k.contractInstance(ctx, contractAddress)
	if err != nil {
		return nil, err
	}

	store := ctx.KVStore(k.storeKey)

	// add funds
	if !coins.IsZero() {
		if k.bankKeeper.BlockedAddr(caller) {
			return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "blocked address can not be used")
		}

		sdkerr := k.bankKeeper.SendCoins(ctx, caller, contractAddress, coins)
		if sdkerr != nil {
			return nil, sdkerr
		}
	}

	contractKey := store.Get(types.GetContractEnclaveKey(contractAddress))
	if contractKey == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "contract key not found")
	}
	//fmt.Printf("Contract Execute: Got contract Key for contract %s: %s\n", contractAddress, base64.StdEncoding.EncodeToString(contractKey))
	params := types.NewEnv(ctx, caller, coins, contractAddress, contractKey)

	// prepare querier
	querier := QueryHandler{
		Ctx:     ctx,
		Plugins: k.queryPlugins,
	}

	gas := gasForContract(ctx)
	//	fmt.Printf("Execute message before wasm is %s \n", base64.StdEncoding.EncodeToString(msg))
	res, gasUsed, execErr := k.wasmer.Execute(codeInfo.CodeHash, params, msg, prefixStore, cosmwasmAPI, querier, gasMeter(ctx), gas, verificationInfo)
	consumeGas(ctx, gasUsed)

	if execErr != nil {
		return nil, sdkerrors.Wrap(types.ErrExecuteFailed, execErr.Error())
	}

	// emit all events from this contract itself
	events := types.ParseEvents(res.Log, contractAddress)
	ctx.EventManager().EmitEvents(events)

	// TODO: capture events here as well
	err = k.dispatchMessages(ctx, contractAddress, res.Messages)
	if err != nil {
		return nil, err
	}

	k.hooks.AfterComputeExecuted(ctx, caller)

	if res.Log[0].Value == "verifiable" && res.Log[0].Key == "output" {
		k.SetContractPublicState(ctx, contractAddress, res.Log)
		return &sdk.Result{Log: res.Log[1].Value}, nil
	} else {
		return &sdk.Result{}, nil
	}
}

// Delete deletes the contract instance
func (k Keeper) Delete(ctx sdk.Context, contractAddress sdk.AccAddress) error {

	_, prefixStore, err := k.contractInstance(ctx, contractAddress)
	if err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)

	contractBz := store.Get(types.GetContractAddressKey(contractAddress))
	if contractBz == nil {
		return sdkerrors.Wrap(types.ErrNotFound, "contract")
	}
	var contract types.ContractInfo
	k.cdc.MustUnmarshal(contractBz, &contract)

	prefixStoreKey := types.GetContractStorePrefixKey(contractAddress)
	//prefixStore := prefix.NewStore(ctx.KVStore(k.storeKey), prefixStoreKey)
	prefixStore.Delete(prefixStoreKey)

	//prefixStore.Delete()
	store.Delete(types.GetContractEnclaveKey(contractAddress))
	store.Delete(types.GetContractIdPrefix(contract.ContractId))
	store.Delete(types.GetContractAddressKey(contractAddress))

	return nil
}

// QuerySmart queries the smart contract itself.
func (k Keeper) QuerySmart(ctx sdk.Context, contractAddr sdk.AccAddress, req []byte, useDefaultGasLimit bool) ([]byte, error) {
	return k.querySmartImpl(ctx, contractAddr, req, useDefaultGasLimit, false)
}

/*
// QuerySmartRecursive queries the smart contract itself. This should only be called when running inside another query recursively.
func (k Keeper) querySmartRecursive(ctx sdk.Context, contractAddr sdk.AccAddress, req []byte, useDefaultGasLimit bool) ([]byte, error) {
	return k.querySmartImpl(ctx, contractAddr, req, useDefaultGasLimit, true)
}
*/
func (k Keeper) querySmartImpl(ctx sdk.Context, contractAddr sdk.AccAddress, req []byte, useDefaultGasLimit bool, recursive bool) ([]byte, error) {
	defer telemetry.MeasureSince(time.Now(), "compute", "keeper", "query")

	if useDefaultGasLimit {
		ctx = ctx.WithGasMeter(sdk.NewGasMeter(k.queryGasLimit))
	}
	ctx.GasMeter().ConsumeGas(types.InstanceCost, "Loading CosmWasm module: query")

	codeInfo, prefixStore, err := k.contractInstance(ctx, contractAddr)
	if err != nil {
		return nil, err
	}

	// prepare querier
	querier := QueryHandler{
		Ctx:     ctx,
		Plugins: k.queryPlugins,
	}

	store := ctx.KVStore(k.storeKey)
	// 0x01 | codeID (uint64) -> ContractInfo
	contractKey := store.Get(types.GetContractEnclaveKey(contractAddr))

	params := types.NewEnv(
		ctx,
		sdk.AccAddress{}, /* empty because it's unused in queries */
		[]sdk.Coin{},     /* empty because it's unused in queries */
		contractAddr,
		contractKey,
	)
	params.Recursive = recursive

	queryResult, gasUsed, qErr := k.wasmer.Query(codeInfo.CodeHash, params, req, prefixStore, cosmwasmAPI, querier, gasMeter(ctx), gasForContract(ctx))
	consumeGas(ctx, gasUsed)

	if qErr != nil {
		return nil, sdkerrors.Wrap(types.ErrQueryFailed, qErr.Error())
	}
	return queryResult, nil
}

// We don't use this function since we have an encrypted state. It's here for upstream compatibility
// QueryRaw returns the contract's state for give key. For a `nil` key a empty slice result is returned.
func (k Keeper) QueryRaw(ctx sdk.Context, contractAddress sdk.AccAddress, key []byte) []types.Model {
	result := make([]types.Model, 0)
	if key == nil {
		return result
	}
	prefixStoreKey := types.GetContractStorePrefixKey(contractAddress)
	prefixStore := prefix.NewStore(ctx.KVStore(k.storeKey), prefixStoreKey)

	if val := prefixStore.Get(key); val != nil {
		return append(result, types.Model{
			Key:   key,
			Value: val,
		})
	}
	return result
}

func (k Keeper) contractInstance(ctx sdk.Context, contractAddress sdk.AccAddress) (types.CodeInfo, prefix.Store, error) {
	store := ctx.KVStore(k.storeKey)

	contractBz := store.Get(types.GetContractAddressKey(contractAddress))
	if contractBz == nil {
		return types.CodeInfo{}, prefix.Store{}, sdkerrors.Wrap(types.ErrNotFound, "contract")
	}
	var contract types.ContractInfo
	k.cdc.MustUnmarshal(contractBz, &contract)

	contractInfoBz := store.Get(types.GetCodeKey(contract.CodeID))
	if contractInfoBz == nil {
		return types.CodeInfo{}, prefix.Store{}, sdkerrors.Wrap(types.ErrNotFound, "contract info")
	}
	var codeInfo types.CodeInfo
	k.cdc.MustUnmarshal(contractInfoBz, &codeInfo)
	prefixStoreKey := types.GetContractStorePrefixKey(contractAddress)
	prefixStore := prefix.NewStore(ctx.KVStore(k.storeKey), prefixStoreKey)
	return codeInfo, prefixStore, nil
}

func (k Keeper) dispatchMessages(ctx sdk.Context, contractAddr sdk.AccAddress, msgs []wasmTypes.CosmosMsg) error {
	for _, msg := range msgs {

		var err error

		if _, _, err = k.Dispatch(ctx, contractAddr, msg); err != nil {
			fmt.Printf("error dispatching messages \n")
			return err
		}
	}
	return nil
}
