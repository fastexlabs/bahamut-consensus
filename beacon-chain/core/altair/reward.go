package altair

import (
	stdmath "math"
	"math/big"

	"github.com/pkg/errors"
	"github.com/prysmaticlabs/prysm/v4/beacon-chain/core/helpers"
	"github.com/prysmaticlabs/prysm/v4/beacon-chain/state"
	"github.com/prysmaticlabs/prysm/v4/config/params"
	"github.com/prysmaticlabs/prysm/v4/consensus-types/primitives"
	"github.com/prysmaticlabs/prysm/v4/math"
)

// BaseReward takes state and validator index and calculate
// individual validator's base reward.
//
// Spec code:
//
//	def get_base_reward(state: BeaconState, index: ValidatorIndex) -> Gwei:
//	  """
//	  Return the base reward for the validator defined by ``index`` with respect to the current ``state``.
//
//	  Note: An optimally performing validator can earn one base reward per epoch over a long time horizon.
//	  This takes into account both per-epoch (e.g. attestation) and intermittent duties (e.g. block proposal
//	  and sync committees).
//	  """
//	  increments = state.validators[index].effective_balance // EFFECTIVE_BALANCE_INCREMENT
//	  return Gwei(increments * get_base_reward_per_increment(state))
func BaseReward(s state.ReadOnlyBeaconState, index primitives.ValidatorIndex) (uint64, error) {
	totalBalance, err := helpers.TotalActiveBalance(s)
	if err != nil {
		return 0, errors.Wrap(err, "could not calculate active balance")
	}
	return BaseRewardWithTotalBalance(s, index, totalBalance)
}

// BaseRewardWithTotalBalance calculates the base reward with the provided total balance.
func BaseRewardWithTotalBalance(s state.ReadOnlyBeaconState, index primitives.ValidatorIndex, totalBalance uint64) (uint64, error) {
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

// BaseRewardPerIncrement of the beacon state
//
// Spec code:
// def get_base_reward_per_increment(state: BeaconState) -> Gwei:
//
//	return Gwei(EFFECTIVE_BALANCE_INCREMENT * BASE_REWARD_FACTOR // integer_squareroot(get_total_active_balance(state)))
func BaseRewardPerIncrement(activeBalance uint64) (uint64, error) {
	if activeBalance == 0 {
		return 0, errors.New("active balance can't be 0")
	}
	cfg := params.BeaconConfig()
	return cfg.EffectiveBalanceIncrement * cfg.BaseRewardFactor / math.CachedSquareRoot(activeBalance), nil
}

// BaseProposerReward of the beacon state.
func BaseProposerReward(s state.ReadOnlyBeaconState, totalPower, totalEffectivePower uint64) (uint64, error) {
	activity, err := helpers.TotalEffectiveActivity(s)
	if err != nil {
		return 0, errors.Wrap(err, "could not calculate total effective activity")
	}

	sharedActivity := s.SharedActivity()
	if sharedActivity == nil {
		return 0, errors.New("nil shared activity in state")
	}

	period := uint64(params.BeaconConfig().EpochsPerActivityPeriod)
	slotsPerEpoch := uint64(params.BeaconConfig().SlotsPerEpoch)
	denominator := period * period * slotsPerEpoch * slotsPerEpoch
	transactionsGas := sharedActivity.TransactionsGasPerPeriod
	baseFee := sharedActivity.BaseFeePerPeriod
	reward := baseFee * (activity + transactionsGas) / denominator

	return baseProposerReward(reward, totalEffectivePower, totalPower), nil
}

func baseProposerReward(reward, totalEffectivePower, totalPower uint64) uint64 {
	if totalPower == 0 {
		return reward
	}

	if stdmath.MaxUint64/totalEffectivePower > reward {
		return reward * totalEffectivePower / totalPower
	}

	var (
		rewardBig              = big.NewInt(0).SetUint64(reward)
		totalEffectivePowerBig = big.NewInt(0).SetUint64(totalEffectivePower)
		totalPowerBig          = big.NewInt(0).SetUint64(totalPower)
		ret                    = big.NewInt(0)
	)

	ret.Mul(rewardBig, totalEffectivePowerBig)
	ret.Div(ret, totalPowerBig)

	return ret.Uint64()
}
