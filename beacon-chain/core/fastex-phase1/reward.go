package fastexphase1

import (
	"math/big"

	"github.com/pkg/errors"
	"github.com/prysmaticlabs/prysm/v3/beacon-chain/core/helpers"
	"github.com/prysmaticlabs/prysm/v3/beacon-chain/core/time"
	"github.com/prysmaticlabs/prysm/v3/beacon-chain/state"
	"github.com/prysmaticlabs/prysm/v3/config/params"
	types "github.com/prysmaticlabs/prysm/v3/consensus-types/primitives"
	"github.com/prysmaticlabs/prysm/v3/math"
)

// BaseReward takes state and validator index and calculate
// individual validator's base reward.
func BaseReward(s state.ReadOnlyBeaconState, index types.ValidatorIndex) (uint64, error) {
	totalBalance, err := helpers.TotalActiveBalance(s)
	if err != nil {
		return 0, errors.Wrap(err, "could not calculate active balance")
	}
	return BaseRewardWithTotalBalance(s, index, totalBalance)
}

// BaseRewardWithTotalBalance calculates the base reward with the provided total balance.
func BaseRewardWithTotalBalance(s state.ReadOnlyBeaconState, index types.ValidatorIndex, totalBalance uint64) (uint64, error) {
	val, err := s.ValidatorAtIndexReadOnly(index)
	if err != nil {
		return 0, err
	}
	cfg := params.BeaconConfig()
	increments := val.EffectiveBalance() / cfg.EffectiveBalanceIncrement
	baseRewardPerInc, err := BaseRewardPerIncrement(totalBalance)
	if err != nil {
		return 0, err
	}
	return increments * baseRewardPerInc, nil
}

// BaseRewardPerIncrement of the beacon state.
func BaseRewardPerIncrement(activeBalance uint64) (uint64, error) {
	if activeBalance == 0 {
		return 0, errors.New("active balance can't be 0")
	}
	cfg := params.BeaconConfig()
	return cfg.EffectiveBalanceIncrement * cfg.BaseRewardFactorFTN / math.IntegerSquareRoot(activeBalance), nil
}

// BaseProposerReward of the post-FastexPhase1 beacon state.
func BaseProposerReward(s state.ReadOnlyBeaconState, validatorsCount uint64) (uint64, error) {
	epoch := time.CurrentEpoch(s)
	activity, err := helpers.TotalEffectiveActivity(s, epoch)
	if err != nil {
		return 0, err
	}

	txGas := s.TransactionsGasPerPeriod()
	baseFee, err := s.BaseFeePerPeriod()
	if err != nil {
		return 0, err
	}
	cfg := params.BeaconConfig()

	var bigReward *big.Int
	bigActivity := new(big.Int).SetUint64(activity)
	bigTxGas := new(big.Int).SetUint64(txGas)
	bigBaseFee := new(big.Int).SetUint64(baseFee)
	bigValidatorsCount := new(big.Int).SetUint64(validatorsCount)
	bigWeiPerGwei := new(big.Int).SetUint64(cfg.WeiPerGwei)
	bigActivityPeriod := new(big.Int).SetUint64(uint64(cfg.EpochsPerActivityPeriod))

	bigReward = new(big.Int).Add(bigActivity, bigTxGas)
	bigReward = bigReward.Div(bigReward, bigValidatorsCount)
	bigReward = bigReward.Mul(bigReward, bigBaseFee)
	bigReward = bigReward.Div(bigReward, bigWeiPerGwei)
	bigReward = bigReward.Div(bigReward, bigActivityPeriod)

	return bigReward.Uint64(), nil
}
