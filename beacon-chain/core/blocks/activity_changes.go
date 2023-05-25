package blocks

import (
	"context"

	"github.com/pkg/errors"
	"github.com/prysmaticlabs/prysm/v4/beacon-chain/core/helpers"
	"github.com/prysmaticlabs/prysm/v4/beacon-chain/core/time"
	"github.com/prysmaticlabs/prysm/v4/beacon-chain/state"
	"github.com/prysmaticlabs/prysm/v4/config/params"
	"github.com/prysmaticlabs/prysm/v4/encoding/bytesutil"
	ethpb "github.com/prysmaticlabs/prysm/v4/proto/prysm/v1alpha1"
	"go.opencensus.io/trace"
)

// ProcessActivityChanges is one of the operations performed
// on each processed beacon block to update current epoch
// validators' activities.
func ProcessActivityChanges(
	ctx context.Context,
	beaconState state.BeaconState,
	activityChanges []*ethpb.ActivityChange,
) (state.BeaconState, error) {
	var err error
	for _, ac := range activityChanges {
		if ac == nil || ac.ContractAddress == nil {
			return nil, errors.New("got a nil activity change in block")
		}
		beaconState, err = ProcessActivityChange(ctx, beaconState, ac)
		if err != nil {
			return nil, errors.Wrapf(err, "could not process activity change from 0x%x", ac.ContractAddress)
		}
	}
	return beaconState, nil
}

// ProcessActivityChange perform activity updates if
// contract exists in beacon state contract map and
// its owner is active validator.
func ProcessActivityChange(
	ctx context.Context,
	beaconState state.BeaconState,
	activityChange *ethpb.ActivityChange,
) (state.BeaconState, error) {
	ctx, span := trace.StartSpan(ctx, "core.ProcessActivityChanges")
	defer span.End()

	epoch := time.CurrentEpoch(beaconState)
	ownerIdx, exist := beaconState.ValidatorIndexByContract(bytesutil.ToBytes20(activityChange.ContractAddress))
	if !exist {
		return beaconState, nil
	}

	owner, err := beaconState.ValidatorAtIndexReadOnly(ownerIdx)
	if err != nil {
		return nil, err
	}

	if !helpers.IsActiveValidatorUsingTrie(owner, epoch) {
		return beaconState, nil
	}

	activity, err := beaconState.ActivityAtIndex(ownerIdx)
	if err != nil {
		return nil, err
	}

	activity += activityChange.DeltaActivity
	if err := beaconState.UpdateActivityAtIndex(ownerIdx, activity); err != nil {
		return nil, err
	}

	return beaconState, nil
}

// ProcessTransactionsCount perform transactions gas per epoch updates.
func ProcessTransactionsCount(
	ctx context.Context,
	beaconState state.BeaconState,
	transactionsCount uint64,
) (state.BeaconState, error) {
	ctx, span := trace.StartSpan(ctx, "core.ProcessTransactionsCount")
	defer span.End()

	transactionsGas := transactionsCount * params.BeaconConfig().BaseTransactionCost
	sharedActivity := beaconState.SharedActivity()
	if sharedActivity == nil {
		return nil, errors.New("nil shared activity in state")
	}

	sharedActivity.TransactionsGasPerEpoch += transactionsGas
	if err := beaconState.SetSharedActivity(sharedActivity); err != nil {
		return nil, err
	}
	return beaconState, nil
}

// ProcessBaseFee perform base fee per epoch updates.
func ProcessBaseFee(
	ctx context.Context,
	beaconState state.BeaconState,
	baseFee uint64,
) (state.BeaconState, error) {
	ctx, span := trace.StartSpan(ctx, "core.ProcessBaseFee")
	defer span.End()

	sharedActivity := beaconState.SharedActivity()
	if sharedActivity == nil {
		return nil, errors.New("nil shared activity in state")
	}
	sharedActivity.BaseFeePerEpoch += baseFee
	if err := beaconState.SetSharedActivity(sharedActivity); err != nil {
		return nil, err
	}
	return beaconState, nil
}

// ProcessExecutionHeight perform execution height updates.
func ProcessExecutionHeight(
	ctx context.Context,
	beaconState state.BeaconState,
	executionHeight uint64,
) (state.BeaconState, error) {
	ctx, span := trace.StartSpan(ctx, "core.ProcessExecutionHeight")
	defer span.End()

	if err := beaconState.SetExecutionHeight(executionHeight); err != nil {
		return nil, err
	}
	return beaconState, nil
}
