package fastexphase1

import (
	"context"

	"github.com/pkg/errors"
	"github.com/prysmaticlabs/prysm/v3/beacon-chain/core/altair"
	e "github.com/prysmaticlabs/prysm/v3/beacon-chain/core/epoch"
	"github.com/prysmaticlabs/prysm/v3/beacon-chain/core/epoch/precompute"
	"github.com/prysmaticlabs/prysm/v3/beacon-chain/state"
	"go.opencensus.io/trace"
)

// ProcessEpoch describes the per epoch operations that are performed on the beacon state.
func ProcessEpoch(ctx context.Context, state state.BeaconState) (state.BeaconState, error) {
	ctx, span := trace.StartSpan(ctx, "fastex-phase1.ProcessEpoch")
	defer span.End()

	if state == nil || state.IsNil() {
		return nil, errors.New("nil state")
	}
	vp, bp, err := altair.InitializePrecomputeValidators(ctx, state)
	if err != nil {
		return nil, err
	}

	// New in Altair.
	vp, bp, err = altair.ProcessEpochParticipation(ctx, state, bp, vp)
	if err != nil {
		return nil, err
	}

	state, err = precompute.ProcessJustificationAndFinalizationPreCompute(state, bp)
	if err != nil {
		return nil, errors.Wrap(err, "could not process justification")
	}

	// New in Altair.
	state, vp, err = altair.ProcessInactivityScores(ctx, state, vp)
	if err != nil {
		return nil, errors.Wrap(err, "could not process inactivity updates")
	}

	// Updated in FastexPhase1.
	state, err = ProcessRewardsAndPenaltiesPrecompute(state, bp, vp)
	if err != nil {
		return nil, errors.Wrap(err, "could not process rewards and penalties")
	}

	state, err = e.ProcessRegistryUpdates(ctx, state)
	if err != nil {
		return nil, errors.Wrap(err, "could not process registry updates")
	}

	// Modified in Altair and Bellatrix.
	proportionalSlashingMultiplier, err := state.ProportionalSlashingMultiplier()
	if err != nil {
		return nil, err
	}
	state, err = e.ProcessSlashings(state, proportionalSlashingMultiplier)
	if err != nil {
		return nil, err
	}
	state, err = e.ProcessEth1DataReset(state)
	if err != nil {
		return nil, err
	}
	state, err = e.ProcessEffectiveBalanceUpdates(state)
	if err != nil {
		return nil, err
	}
	state, err = e.ProcessEffectiveActivityUpdates(state)
	if err != nil {
		return nil, err
	}
	state, err = e.ProcessTransactionsGasPerPeriodUpdate(state)
	if err != nil {
		return nil, err
	}
	state, err = e.ProcessNonStakersGasPerPeriodUpdate(state)
	if err != nil {
		return nil, err
	}
	state, err = e.ProcessActivityReset(state)
	if err != nil {
		return nil, err
	}
	state, err = e.ProcessSlashingsReset(state)
	if err != nil {
		return nil, err
	}
	state, err = e.ProcessRandaoMixesReset(state)
	if err != nil {
		return nil, err
	}
	state, err = e.ProcessHistoricalDataUpdate(state)
	if err != nil {
		return nil, err
	}
	state, err = e.ProcessBaseFeePerPeriodUpdate(state)
	if err != nil {
		return nil, err
	}

	// New in Altair.
	state, err = altair.ProcessParticipationFlagUpdates(state)
	if err != nil {
		return nil, err
	}

	// New in Altair.
	state, err = altair.ProcessSyncCommitteeUpdates(ctx, state)
	if err != nil {
		return nil, err
	}

	return state, nil
}
