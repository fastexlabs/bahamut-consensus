package fastexphase1

import (
	"github.com/pkg/errors"
	"github.com/prysmaticlabs/prysm/v3/beacon-chain/core/epoch/precompute"
	"github.com/prysmaticlabs/prysm/v3/beacon-chain/core/helpers"
	"github.com/prysmaticlabs/prysm/v3/beacon-chain/core/time"
	"github.com/prysmaticlabs/prysm/v3/beacon-chain/state"
	"github.com/prysmaticlabs/prysm/v3/config/params"
	"github.com/prysmaticlabs/prysm/v3/math"
)

// ProcessRewardsAndPenaltiesPrecompute processes the rewards and penalties of individual validator.
// This is an optimized version by passing in precomputed validator attesting records and and total epoch balances.
func ProcessRewardsAndPenaltiesPrecompute(
	beaconState state.BeaconState,
	bal *precompute.Balance,
	vals []*precompute.Validator,
) (state.BeaconState, error) {
	// Don't process rewards and penalties in genesis epoch.
	cfg := params.BeaconConfig()
	if time.CurrentEpoch(beaconState) == cfg.GenesisEpoch {
		return beaconState, nil
	}

	numOfVals := beaconState.NumValidators()
	// Guard against an out-of-bounds using validator balance precompute.
	if len(vals) != numOfVals || len(vals) != beaconState.BalancesLength() {
		return beaconState, errors.New("validator registries not the same length as state's validator registries")
	}

	attsRewards, attsPenalties, err := AttestationsDelta(beaconState, bal, vals)
	if err != nil {
		return nil, errors.Wrap(err, "could not get attestation delta")
	}

	balances := beaconState.Balances()
	for i := 0; i < numOfVals; i++ {
		vals[i].BeforeEpochTransitionBalance = balances[i]

		// Compute the post balance of the validator after accounting for the
		// attester and proposer rewards and penalties.
		balances[i], err = helpers.IncreaseBalanceWithVal(balances[i], attsRewards[i])
		if err != nil {
			return nil, err
		}
		balances[i] = helpers.DecreaseBalanceWithVal(balances[i], attsPenalties[i])

		vals[i].AfterEpochTransitionBalance = balances[i]
	}

	if err := beaconState.SetBalances(balances); err != nil {
		return nil, errors.Wrap(err, "could not set validator balances")
	}

	return beaconState, nil
}

// AttestationsDelta computes and returns the rewards and penalties differences for individual validators based on the
// voting records.
func AttestationsDelta(beaconState state.BeaconState, bal *precompute.Balance, vals []*precompute.Validator) (rewards, penalties []uint64, err error) {
	numOfVals := beaconState.NumValidators()
	rewards = make([]uint64, numOfVals)
	penalties = make([]uint64, numOfVals)

	cfg := params.BeaconConfig()
	prevEpoch := time.PrevEpoch(beaconState)
	finalizedEpoch := beaconState.FinalizedCheckpointEpoch()
	increment := cfg.EffectiveBalanceIncrement
	factor := cfg.BaseRewardFactorFTN
	baseRewardMultiplier := increment * factor / math.IntegerSquareRoot(bal.ActiveCurrentEpoch)
	leak := helpers.IsInInactivityLeak(prevEpoch, finalizedEpoch)

	// Modified in Altair and Bellatrix.
	bias := cfg.InactivityScoreBias
	inactivityPenaltyQuotient, err := beaconState.InactivityPenaltyQuotient()
	if err != nil {
		return nil, nil, err
	}
	inactivityDenominator := bias * inactivityPenaltyQuotient

	for i, v := range vals {
		rewards[i], penalties[i], err = attestationDelta(bal, v, baseRewardMultiplier, inactivityDenominator, leak)
		if err != nil {
			return nil, nil, err
		}
	}

	return rewards, penalties, nil
}

func attestationDelta(
	bal *precompute.Balance,
	val *precompute.Validator,
	baseRewardMultiplier, inactivityDenominator uint64,
	inactivityLeak bool) (reward, penalty uint64, err error) {
	eligible := val.IsActivePrevEpoch || (val.IsSlashed && !val.IsWithdrawableCurrentEpoch)
	// Per spec `ActiveCurrentEpoch` can't be 0 to process attestation delta.
	if !eligible || bal.ActiveCurrentEpoch == 0 {
		return 0, 0, nil
	}

	cfg := params.BeaconConfig()
	increment := cfg.EffectiveBalanceIncrement
	effectiveBalance := val.CurrentEpochEffectiveBalance
	baseReward := (effectiveBalance / increment) * baseRewardMultiplier
	activeIncrement := bal.ActiveCurrentEpoch / increment

	weightDenominator := cfg.WeightDenominatorFTN
	srcWeight := cfg.TimelySourceWeight
	tgtWeight := cfg.TimelyTargetWeight
	headWeight := cfg.TimelyHeadWeight
	reward, penalty = uint64(0), uint64(0)
	// Process source reward / penalty
	if val.IsPrevEpochSourceAttester && !val.IsSlashed {
		if !inactivityLeak {
			n := baseReward * srcWeight * (bal.PrevEpochAttested / increment)
			reward += n / (activeIncrement * weightDenominator)
		}
	} else {
		penalty += baseReward * srcWeight / weightDenominator
	}

	// Process target reward / penalty
	if val.IsPrevEpochTargetAttester && !val.IsSlashed {
		if !inactivityLeak {
			n := baseReward * tgtWeight * (bal.PrevEpochTargetAttested / increment)
			reward += n / (activeIncrement * weightDenominator)
		}
	} else {
		penalty += baseReward * tgtWeight / weightDenominator
	}

	// Process head reward / penalty
	if val.IsPrevEpochHeadAttester && !val.IsSlashed {
		if !inactivityLeak {
			n := baseReward * headWeight * (bal.PrevEpochHeadAttested / increment)
			reward += n / (activeIncrement * weightDenominator)
		}
	}

	// Process finality delay penalty
	// Apply an additional penalty to validators that did not vote on the correct target or slashed
	if !val.IsPrevEpochTargetAttester || val.IsSlashed {
		n, err := math.Mul64(effectiveBalance, val.InactivityScore)
		if err != nil {
			return 0, 0, err
		}
		penalty += n / inactivityDenominator
	}

	return reward, penalty, nil
}
