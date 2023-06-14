package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/trstlabs/trst/x/alloc/types"
)

type (
	Keeper struct {
		cdc           codec.BinaryCodec
		storeKey      sdk.StoreKey
		accountKeeper types.AccountKeeper
		bankKeeper    types.BankKeeper
		stakingKeeper types.StakingKeeper
		distrKeeper   types.DistrKeeper
		paramstore    paramtypes.Subspace
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey sdk.StoreKey,
	accountKeeper types.AccountKeeper, bankKeeper types.BankKeeper, stakingKeeper types.StakingKeeper, distrKeeper types.DistrKeeper,
	ps paramtypes.Subspace,
) *Keeper {

	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		cdc:           cdc,
		storeKey:      storeKey,
		accountKeeper: accountKeeper, bankKeeper: bankKeeper, stakingKeeper: stakingKeeper, distrKeeper: distrKeeper, //computeKeeper: ck,
		paramstore: ps,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetBalance gets balance
func (k Keeper) GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	return k.bankKeeper.GetBalance(ctx, addr, denom)
}

// DistributeInflation distributes module-specific inflation
func (k Keeper) DistributeInflation(ctx sdk.Context) error {
	blockInflationAddr := k.accountKeeper.GetModuleAccount(ctx, authtypes.FeeCollectorName).GetAddress()
	blockInflation := k.bankKeeper.GetBalance(ctx, blockInflationAddr, k.stakingKeeper.BondDenom(ctx))
	//blockInflationDec := sdk.NewDecFromInt(blockInflation.Amount)

	params := k.GetParams(ctx)
	proportions := params.DistributionProportions

	// contractIncentiveCoins := sdk.NewCoins(k.GetProportions(ctx, blockInflation, proportions.TrustlessContractIncentives))
	// err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, authtypes.FeeCollectorName, "compute", contractIncentiveCoins)
	// if err != nil {
	// 	return err
	// }
	// k.Logger(ctx).Debug("funded contract module", "amount", contractIncentiveCoins.String(), "from", blockInflationAddr)

	relayerIncentiveCoins := sdk.NewCoins(k.GetProportions(ctx, blockInflation, proportions.RelayerIncentives))
	err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, authtypes.FeeCollectorName, "autoibctx", relayerIncentiveCoins)
	if err != nil {
		return err
	}
	k.Logger(ctx).Debug("funded autoibctx module", "amount", relayerIncentiveCoins.String(), "from", blockInflationAddr)

	/*itemIncentiveCoins := sdk.NewCoins(k.GetProportions(ctx, blockInflation, proportions.ItemIncentives))
	err = k.bankKeeper.SendCoinsFromModuleToModule(ctx, authtypes.FeeCollectorName, "item_incentives", itemIncentiveCoins)
	if err != nil {
		return err
	}*/

	//staking incentives stay in the fee collector account and are to be moved to on next begin blocker
	stakingIncentivesCoins := sdk.NewCoins(k.GetProportions(ctx, blockInflation, proportions.Staking))

	contributorCoin := k.GetProportions(ctx, blockInflation, proportions.ContributorRewards)
	contributorCoins := sdk.NewCoins(contributorCoin)

	for _, w := range params.WeightedContributorRewardsReceivers {
		contributorPortionCoins := sdk.NewCoins(k.GetProportions(ctx, contributorCoin, w.Weight))
		if w.Address == "" {
			err := k.distrKeeper.FundCommunityPool(ctx, contributorPortionCoins, blockInflationAddr)
			if err != nil {
				return err
			}
		} else {
			contributorRewardsAddr, err := sdk.AccAddressFromBech32(w.Address)
			if err != nil {
				return err
			}
			err = k.bankKeeper.SendCoins(ctx, blockInflationAddr, contributorRewardsAddr, contributorPortionCoins)
			if err != nil {
				return err
			}
			k.Logger(ctx).Debug("sent coins to contributor", "amount", contributorPortionCoins.String(), "from", blockInflationAddr)
		}
	}

	// subtract from original provision to ensure no coins left over after the allocations
	communityPoolCoins := sdk.NewCoins(blockInflation).Sub(stakingIncentivesCoins). /*.Sub(itemIncentiveCoins) Sub(contractIncentiveCoins).*/ Sub(relayerIncentiveCoins).Sub(contributorCoins)

	err = k.distrKeeper.FundCommunityPool(ctx, communityPoolCoins, blockInflationAddr)
	if err != nil {
		return err
	}

	return nil
}

// GetProportions gets the balance of the `MintedDenom` from minted coins
// and returns coins according to the `AllocationRatio`
func (k Keeper) GetProportions(ctx sdk.Context, mintedCoin sdk.Coin, ratio sdk.Dec) sdk.Coin {
	return sdk.NewCoin(mintedCoin.Denom, mintedCoin.Amount.ToDec().Mul(ratio).TruncateInt())
}
