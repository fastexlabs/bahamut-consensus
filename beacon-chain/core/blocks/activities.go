package blocks

import (
	"context"

	"github.com/pkg/errors"
	"github.com/prysmaticlabs/prysm/v3/beacon-chain/core/helpers"
	"github.com/prysmaticlabs/prysm/v3/beacon-chain/core/time"
	"github.com/prysmaticlabs/prysm/v3/beacon-chain/state"
	"github.com/prysmaticlabs/prysm/v3/encoding/bytesutil"
	ethpb "github.com/prysmaticlabs/prysm/v3/proto/prysm/v1alpha1"
	"go.opencensus.io/trace"
)

// ProcessBlockActivities applies processing operations to a block's inner
// activities records.
func ProcessBlockActivities(
	ctx context.Context,
	beaconState state.BeaconState,
	activities []*ethpb.ActivityChange,
) (state.BeaconState, error) {
	var err error
	for _, activity := range activities {
		beaconState, err = ProcessActivity(ctx, beaconState, activity)
		if err != nil {
			return nil, errors.Wrap(err, "could not process activties changes")
		}
	}
	return beaconState, nil
}

// ProcessActivity applies processing operations to a single inner
// activities record.
func ProcessActivity(
	ctx context.Context,
	beaconState state.BeaconState,
	activity *ethpb.ActivityChange,
) (state.BeaconState, error) {
	ctx, span := trace.StartSpan(ctx, "core.ProcessActivtiyNoVerifySignature")
	defer span.End()

	contract := bytesutil.ToBytes20(activity.ContractAddress)
	idx, ok := beaconState.ValidatorIndexByContractAddress(contract)
	if !ok {
		nonStakersGas := beaconState.NonStakersGasPerEpoch()
		if err := beaconState.SetNonStakersGasPerEpoch(nonStakersGas + activity.DeltaActivity); err != nil {
			return nil, err
		}
		return beaconState, nil
	}

	epoch := time.CurrentEpoch(beaconState)
	val, err := beaconState.ValidatorAtIndexReadOnly(idx)
	if err != nil {
		return nil, err
	}

	isActive := helpers.IsActiveValidatorUsingTrie(val, epoch)
	if !isActive {
		nonStakersGas := beaconState.NonStakersGasPerEpoch()
		if err := beaconState.SetNonStakersGasPerEpoch(nonStakersGas + activity.DeltaActivity); err != nil {
			return nil, err
		}
		return beaconState, nil
	}

	valActivity, err := beaconState.ActivityAtIndex(idx)
	if err != nil {
		return nil, err
	}

	valActivity += activity.DeltaActivity
	err = beaconState.UpdateActivitiesAtIndex(idx, valActivity)
	if err != nil {
		return nil, err
	}

	return beaconState, nil
}

// ProcessLatestProcessedBlock applies latest processed block activities number to state.
func ProcessLatestProcessedBlock(
	ctx context.Context,
	beaconState state.BeaconState,
	val uint64,
) (state.BeaconState, error) {
	ctx, span := trace.StartSpan(ctx, "core.ProcessLatestProcessedBlockNoVerifySignature")
	defer span.End()

	err := beaconState.SetLatestProcessedBlockActivities(val)
	if err != nil {
		return nil, err
	}

	return beaconState, nil
}

// ProcessTransactionsCount applies transactions count in latest processed block to state.
func ProcessTransactionsCount(
	ctx context.Context,
	beaconState state.BeaconState,
	val uint64,
) (state.BeaconState, error) {
	ctx, span := trace.StartSpan(ctx, "core.ProcessTransactionsCountNoVerifySignature")
	defer span.End()

	err := beaconState.SetTransactionsPerLatestEpoch(beaconState.TransactionsPerLatestEpoch() + val)
	if err != nil {
		return nil, err
	}

	return beaconState, nil
}
