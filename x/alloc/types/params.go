package types

import (
	"errors"
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter store keys
var (
	KeyDistributionProportions  = []byte("DistributionProportions")
	KeyDeveloperRewardsReceiver = []byte("DeveloperRewardsReceiver")
	KeySupplementAmount         = []byte("SupplementAmount")
)

// ParamTable for module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

func NewParams(
	distrProportions DistributionProportions,
	weightedDevRewardsReceivers []WeightedAddress,
) Params {
	return Params{
		DistributionProportions:           distrProportions,
		WeightedDeveloperRewardsReceivers: weightedDevRewardsReceivers,
	}
}

// default module parameters
func DefaultParams() Params {
	return Params{
		DistributionProportions: DistributionProportions{
			RelayerIncentives: math.LegacyNewDecWithPrec(20, 2), // 20%
			DeveloperRewards:  math.LegacyNewDecWithPrec(5, 2),  // 15%
			CommunityPool:     math.LegacyNewDecWithPrec(15, 2), // 5%
		},
		WeightedDeveloperRewardsReceivers: []WeightedAddress{},
		SupplementAmount:                  sdk.NewCoins(),
	}
}

// validate params
func (p Params) Validate() error {
	if err := validateDistributionProportions(p.DistributionProportions); err != nil {
		return err
	}
	return validateWeightedRewardsReceivers(p.WeightedDeveloperRewardsReceivers)
}

// Implements params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyDistributionProportions, &p.DistributionProportions, validateDistributionProportions),
		paramtypes.NewParamSetPair(
			KeyDeveloperRewardsReceiver, &p.WeightedDeveloperRewardsReceivers, validateWeightedRewardsReceivers),
		paramtypes.NewParamSetPair(
			KeySupplementAmount, &p.SupplementAmount, validateSupplementAmount),
	}
}

func validateSupplementAmount(i interface{}) error {
	v, ok := i.(sdk.Coins)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if len(v) == 0 {
		return nil
	}
	return v.Validate()
}

func validateDistributionProportions(i interface{}) error {
	v, ok := i.(DistributionProportions)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.RelayerIncentives.IsNegative() {
		return errors.New("NFT incentives distribution ratio should not be negative")
	}

	if v.DeveloperRewards.IsNegative() {
		return errors.New("developer rewards distribution ratio should not be negative")
	}

	if v.CommunityPool.IsNegative() {
		return errors.New("community pool ratio should not be negative")
	}

	totalProportions := v.RelayerIncentives.Add(v.DeveloperRewards).Add(v.CommunityPool)

	if totalProportions.GT(math.LegacyOneDec()) {
		return errors.New("total distributions can not be higher than 100%")
	}

	return nil
}

func validateWeightedRewardsReceivers(i interface{}) error {
	v, ok := i.([]WeightedAddress)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	// fund community pool when rewards address is empty
	if len(v) == 0 {
		return nil
	}

	weightSum := math.LegacyNewDec(0)
	for i, w := range v {
		// we allow address to be "" to go to community pool
		if w.Address != "" {
			_, err := sdk.AccAddressFromBech32(w.Address)
			if err != nil {
				return fmt.Errorf("invalid address at %dth", i)
			}
		}
		if !w.Weight.IsPositive() {
			return fmt.Errorf("non-positive weight at %dth", i)
		}
		if w.Weight.GT(math.LegacyNewDec(1)) {
			return fmt.Errorf("more than 1 weight at %dth", i)
		}
		weightSum = weightSum.Add(w.Weight)
	}

	if !weightSum.Equal(math.LegacyNewDec(1)) {
		return fmt.Errorf("invalid weight sum: %s", weightSum.String())
	}

	return nil
}
