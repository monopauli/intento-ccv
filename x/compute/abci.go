package compute

import (
	//"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/trstlabs/trst/x/compute/internal/keeper"
	"github.com/trstlabs/trst/x/compute/internal/types"
)

// EndBlocker called every block, process inflation, update validator set.
func EndBlocker(ctx sdk.Context, k keeper.Keeper) []abci.ValidatorUpdate {

	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)

	logger := k.Logger(ctx)

	//	addressList := k.GetAllContractAddresses
	incentiveList, contracts := k.GetContractAddressesForBlock(ctx)
	//var rewardCoins sdk.Coins
	if len(incentiveList) > 0 {
		k.SetIncentiveCoins(ctx, incentiveList)

	}
	var gasUsed uint64
	for _, contract := range contracts {
		//for _, addr := range incentiveList {
		if contract.Address.Equals(contract.Address) && contract.AutoMsg != nil {
			// attempt to execute all message
			// Messages may mutate state thus we can use a cached context. If one of
			// the handlers fails, no state mutation is written and the error
			// message is logged.
			cacheCtx, writeCache := ctx.CacheContext()
			gas, err := k.SelfExecute(cacheCtx, contract.Address, contract.ContractInfo.AutoMsg, contract.ContractInfo.CallbackSig)
			if err == nil {
				logger.Info(
					"aut_omsg",
					"gas", gas,
				)

				ctx.EventManager().EmitEvent(
					sdk.NewEvent(
						types.EventTypeAutoMsgContract,
						sdk.NewAttribute(types.AttributeKeyContractAddr, contract.Address.String()),
					),
				)
				writeCache()
			}
			gasUsed = gas

		}

		logger.Info(
			"expired",
			"contract", contract.Address.String(),
		)
		err := k.DeductFeesAndFundCreator(ctx, contract.Address, gasUsed)
		if err != nil {
			logger.Info(
				"contract payout creator",
				"err", err.Error(),
			)
		}
		k.RemoveFromContractQueue(ctx, contract.Address.String(), contract.ContractInfo.EndTime)
		_ = k.Delete(ctx, contract.Address)
		logger.Info(
			"deleted",
			"contract", contract.Address.String(),
		)
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeContractExpired,
				sdk.NewAttribute(types.AttributeKeyContractAddr, contract.Address.String()),
			),
		)
	}

	return []abci.ValidatorUpdate{}
}
