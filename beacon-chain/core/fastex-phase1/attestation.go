package fastexphase1

import (
	"context"
	"fmt"
	"math/big"

	"github.com/pkg/errors"
	"github.com/prysmaticlabs/prysm/v3/beacon-chain/core/altair"
	"github.com/prysmaticlabs/prysm/v3/beacon-chain/core/blocks"
	"github.com/prysmaticlabs/prysm/v3/beacon-chain/core/helpers"
	"github.com/prysmaticlabs/prysm/v3/beacon-chain/core/time"
	"github.com/prysmaticlabs/prysm/v3/beacon-chain/state"
	"github.com/prysmaticlabs/prysm/v3/config/params"
	consensusblocks "github.com/prysmaticlabs/prysm/v3/consensus-types/blocks"
	"github.com/prysmaticlabs/prysm/v3/consensus-types/interfaces"
	"github.com/prysmaticlabs/prysm/v3/consensus-types/primitives"
	ethpb "github.com/prysmaticlabs/prysm/v3/proto/prysm/v1alpha1"
	"github.com/prysmaticlabs/prysm/v3/proto/prysm/v1alpha1/attestation"
	"go.opencensus.io/trace"
)

// ProcessAttestationsNoVerifySignature applies processing operations to a block's inner attestation
// records. The only difference would be that the attestation signature would not be verified.
func ProcessAttestationsNoVerifySignature(
	ctx context.Context,
	beaconState state.BeaconState,
	b interfaces.ReadOnlySignedBeaconBlock,
) (state.BeaconState, error) {
	if err := consensusblocks.BeaconBlockIsNil(b); err != nil {
		return nil, err
	}
	body := b.Block().Body()
	totalBalance, err := helpers.TotalActiveBalance(beaconState)
	if err != nil {
		return nil, err
	}
	for idx, att := range body.Attestations() {
		beaconState, err = ProcessAttestationNoVerifySignature(ctx, beaconState, att, totalBalance)
		if err != nil {
			return nil, errors.Wrapf(err, "could not verify attestation at index %d in block", idx)
		}
	}
	return beaconState, nil
}

// ProcessAttestationNoVerifySignature processes the attestation without verifying the attestation signature. This
// method is used to validate attestations whose signatures have already been verified or will be verified later.
func ProcessAttestationNoVerifySignature(
	ctx context.Context,
	beaconState state.BeaconState,
	att *ethpb.Attestation,
	totalBalance uint64,
) (state.BeaconState, error) {
	ctx, span := trace.StartSpan(ctx, "altair.ProcessAttestationNoVerifySignature")
	defer span.End()

	if err := blocks.VerifyAttestationNoVerifySignature(ctx, beaconState, att); err != nil {
		return nil, err
	}

	delay, err := beaconState.Slot().SafeSubSlot(att.Data.Slot)
	if err != nil {
		return nil, fmt.Errorf("att slot %d can't be greater than state slot %d", att.Data.Slot, beaconState.Slot())
	}
	participatedFlags, err := altair.AttestationParticipationFlagIndices(beaconState, att.Data, delay)
	if err != nil {
		return nil, err
	}
	committee, err := helpers.BeaconCommitteeFromState(ctx, beaconState, att.Data.Slot, att.Data.CommitteeIndex)
	if err != nil {
		return nil, err
	}
	indices, err := attestation.AttestingIndices(att.AggregationBits, committee)
	if err != nil {
		return nil, err
	}

	return SetParticipationAndRewardProposer(ctx, beaconState, att.Data.Target.Epoch, indices, participatedFlags, totalBalance)
}

// SetParticipationAndRewardProposer retrieves and sets the epoch participation bits in state. Based on the epoch participation, it rewards
// the proposer in state.
func SetParticipationAndRewardProposer(
	ctx context.Context,
	beaconState state.BeaconState,
	targetEpoch primitives.Epoch,
	indices []uint64,
	participatedFlags map[uint8]bool,
	totalBalance uint64,
) (state.BeaconState, error) {
	var proposerRewardNumerator uint64
	var proposerRewardDenominator uint64
	currentEpoch := time.CurrentEpoch(beaconState)
	var stateErr error
	if targetEpoch == currentEpoch {
		stateErr = beaconState.ModifyCurrentParticipationBits(func(val []byte) ([]byte, error) {
			propRewardNum, propRewardDenom, epochParticipation, err := EpochParticipation(beaconState, indices, val, participatedFlags, totalBalance)
			if err != nil {
				return nil, err
			}
			proposerRewardNumerator = propRewardNum
			proposerRewardDenominator = propRewardDenom
			return epochParticipation, nil
		})
	} else {
		stateErr = beaconState.ModifyPreviousParticipationBits(func(val []byte) ([]byte, error) {
			propRewardNum, propRewardDenom, epochParticipation, err := EpochParticipation(beaconState, indices, val, participatedFlags, totalBalance)
			if err != nil {
				return nil, err
			}
			proposerRewardNumerator = propRewardNum
			proposerRewardDenominator = propRewardDenom
			return epochParticipation, nil
		})
	}
	if stateErr != nil {
		return nil, stateErr
	}

	if err := RewardProposer(ctx, beaconState, proposerRewardNumerator, proposerRewardDenominator); err != nil {
		return nil, err
	}

	return beaconState, nil
}

// EpochParticipation sets and returns the proposer reward numerator and epoch participation.
func EpochParticipation(
	beaconState state.BeaconState,
	indices []uint64,
	epochParticipation []byte,
	participatedFlags map[uint8]bool,
	totalBalance uint64,
) (uint64, uint64, []byte, error) {
	cfg := params.BeaconConfig()
	sourceFlagIndex := cfg.TimelySourceFlagIndex
	targetFlagIndex := cfg.TimelyTargetFlagIndex
	headFlagIndex := cfg.TimelyHeadFlagIndex
	proposerRewardNumerator := uint64(0)
	proposerRewardDenominator := uint64(0)
	for _, index := range indices {
		if index >= uint64(len(epochParticipation)) {
			return 0, 0, nil, fmt.Errorf("index %d exceeds participation length %d", index, len(epochParticipation))
		}
		br, err := BaseRewardWithTotalBalance(beaconState, primitives.ValidatorIndex(index), totalBalance)
		if err != nil {
			return 0, 0, nil, err
		}
		proposerRewardDenominator += br * (cfg.TimelyHeadWeight + cfg.TimelyTargetWeight + cfg.TimelySourceWeight)
		has, err := altair.HasValidatorFlag(epochParticipation[index], sourceFlagIndex)
		if err != nil {
			return 0, 0, nil, err
		}
		if participatedFlags[sourceFlagIndex] && !has {
			epochParticipation[index], err = altair.AddValidatorFlag(epochParticipation[index], sourceFlagIndex)
			if err != nil {
				return 0, 0, nil, err
			}
			proposerRewardNumerator += br * cfg.TimelySourceWeight
		}
		has, err = altair.HasValidatorFlag(epochParticipation[index], targetFlagIndex)
		if err != nil {
			return 0, 0, nil, err
		}
		if participatedFlags[targetFlagIndex] && !has {
			epochParticipation[index], err = altair.AddValidatorFlag(epochParticipation[index], targetFlagIndex)
			if err != nil {
				return 0, 0, nil, err
			}
			proposerRewardNumerator += br * cfg.TimelyTargetWeight
		}
		has, err = altair.HasValidatorFlag(epochParticipation[index], headFlagIndex)
		if err != nil {
			return 0, 0, nil, err
		}
		if participatedFlags[headFlagIndex] && !has {
			epochParticipation[index], err = altair.AddValidatorFlag(epochParticipation[index], headFlagIndex)
			if err != nil {
				return 0, 0, nil, err
			}
			proposerRewardNumerator += br * cfg.TimelyHeadWeight
		}
	}

	return proposerRewardNumerator, proposerRewardDenominator, epochParticipation, nil
}

// RewardProposer rewards proposer by increasing proposer's balance with input reward numerator and calculated reward denominator.
func RewardProposer(ctx context.Context, beaconState state.BeaconState, proposerRewardNumerator, proposerRewardDenominator uint64) error {
	cfg := params.BeaconConfig()
	epoch := time.CurrentEpoch(beaconState)
	validatorsCount, err := helpers.ActiveValidatorCount(ctx, beaconState, epoch)
	if err != nil {
		return err
	}
	baseProposerReward, err := BaseProposerReward(beaconState, validatorsCount)
	if err != nil {
		return err
	}

	var bigProposerReward *big.Int
	bigBaseProposerReward := new(big.Int).SetUint64(baseProposerReward)
	bigProposerRewardNumerator := new(big.Int).SetUint64(proposerRewardNumerator)
	bigProposerRewardDenominator := new(big.Int).SetUint64(proposerRewardDenominator)
	bigWeightDenominatorFTN := new(big.Int).SetUint64(cfg.WeightDenominatorFTN)
	// WeightDenominatorFTN - SyncRewardWeight == TimelySourceWeight + TimelyHeadWeight + TimelyTargetWeigth
	bigWeightNumerator := new(big.Int).SetUint64(cfg.TimelySourceWeight + cfg.TimelyHeadWeight + cfg.TimelyTargetWeight)
	// bigWeightNumerator := new(big.Int).SetUint64(cfg.WeightDenominatorFTN - cfg.SyncRewardWeight)

	bigProposerReward = new(big.Int).Mul(bigBaseProposerReward, bigProposerRewardNumerator)
	bigProposerReward = bigProposerReward.Mul(bigProposerReward, bigWeightNumerator)
	bigProposerReward = bigProposerReward.Div(bigProposerReward, bigProposerRewardDenominator)
	bigProposerReward = bigProposerReward.Div(bigProposerReward, bigWeightDenominatorFTN)

	i, err := helpers.BeaconProposerIndex(ctx, beaconState)
	if err != nil {
		return err
	}

	return helpers.IncreaseBalance(beaconState, i, bigProposerReward.Uint64())
}
