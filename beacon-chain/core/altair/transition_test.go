package altair_test

import (
	"context"
	beaconState "github.com/prysmaticlabs/prysm/v4/beacon-chain/state"
	"github.com/prysmaticlabs/prysm/v4/consensus-types/primitives"
	"github.com/prysmaticlabs/prysm/v4/runtime/version"
	"testing"

	"github.com/prysmaticlabs/prysm/v4/beacon-chain/core/altair"
	"github.com/prysmaticlabs/prysm/v4/config/params"
	"github.com/prysmaticlabs/prysm/v4/testing/require"
	"github.com/prysmaticlabs/prysm/v4/testing/util"
)

func TestProcessEpoch_CanProcess(t *testing.T) {
	st, _ := util.DeterministicGenesisStateAltair(t, params.BeaconConfig().MaxValidatorsPerCommittee)
	require.NoError(t, st.SetSlot(10*params.BeaconConfig().SlotsPerEpoch))
	newState, err := altair.ProcessEpoch(context.Background(), st)
	require.NoError(t, err)
	require.Equal(t, uint64(0), newState.Slashings()[2], "Unexpected slashed balance")

	b := st.Balances()
	require.Equal(t, params.BeaconConfig().MaxValidatorsPerCommittee, uint64(len(b)))

	cfg := params.BeaconConfig()

	pDelta, err := penaltyDelta(st, 0)
	require.NoError(t, err)
	require.Equal(t, cfg.MaxEffectiveBalance-pDelta, b[0])

	s, err := st.InactivityScores()
	require.NoError(t, err)
	require.Equal(t, params.BeaconConfig().MaxValidatorsPerCommittee, uint64(len(s)))

	p, err := st.PreviousEpochParticipation()
	require.NoError(t, err)
	require.Equal(t, params.BeaconConfig().MaxValidatorsPerCommittee, uint64(len(p)))

	p, err = st.CurrentEpochParticipation()
	require.NoError(t, err)
	require.Equal(t, params.BeaconConfig().MaxValidatorsPerCommittee, uint64(len(p)))

	sc, err := st.CurrentSyncCommittee()
	require.NoError(t, err)
	require.Equal(t, params.BeaconConfig().SyncCommitteeSize, uint64(len(sc.Pubkeys)))

	sc, err = st.NextSyncCommittee()
	require.NoError(t, err)
	require.Equal(t, params.BeaconConfig().SyncCommitteeSize, uint64(len(sc.Pubkeys)))
}

func TestProcessEpoch_CanProcessBellatrix(t *testing.T) {
	st, _ := util.DeterministicGenesisStateBellatrix(t, params.BeaconConfig().MaxValidatorsPerCommittee)
	require.NoError(t, st.SetSlot(10*params.BeaconConfig().SlotsPerEpoch))
	newState, err := altair.ProcessEpoch(context.Background(), st)
	require.NoError(t, err)
	require.Equal(t, uint64(0), newState.Slashings()[2], "Unexpected slashed balance")

	cfg := params.BeaconConfig()

	pDelta, err := penaltyDelta(st, 0)
	require.NoError(t, err)

	b := st.Balances()
	require.Equal(t, params.BeaconConfig().MaxValidatorsPerCommittee, uint64(len(b)))
	require.Equal(t, cfg.MaxEffectiveBalance-pDelta, b[0])

	s, err := st.InactivityScores()
	require.NoError(t, err)
	require.Equal(t, params.BeaconConfig().MaxValidatorsPerCommittee, uint64(len(s)))

	p, err := st.PreviousEpochParticipation()
	require.NoError(t, err)
	require.Equal(t, params.BeaconConfig().MaxValidatorsPerCommittee, uint64(len(p)))

	p, err = st.CurrentEpochParticipation()
	require.NoError(t, err)
	require.Equal(t, params.BeaconConfig().MaxValidatorsPerCommittee, uint64(len(p)))

	sc, err := st.CurrentSyncCommittee()
	require.NoError(t, err)
	require.Equal(t, params.BeaconConfig().SyncCommitteeSize, uint64(len(sc.Pubkeys)))

	sc, err = st.NextSyncCommittee()
	require.NoError(t, err)
	require.Equal(t, params.BeaconConfig().SyncCommitteeSize, uint64(len(sc.Pubkeys)))
}

func penaltyDelta(st beaconState.ReadOnlyBeaconState, idx primitives.ValidatorIndex) (uint64, error) {
	cfg := params.BeaconConfig()

	reward, err := altair.BaseReward(st, idx)
	if err != nil {
		return 0, err
	}

	val, err := st.ValidatorAtIndexReadOnly(idx)
	if err != nil {
		return 0, err
	}

	var iPQ uint64
	switch st.Version() {
	case version.Phase0:
		iPQ = cfg.InactivityPenaltyQuotient
	case version.Altair:
		iPQ = cfg.InactivityPenaltyQuotientAltair
	case version.Bellatrix, version.Capella, version.Deneb:
		iPQ = cfg.InactivityPenaltyQuotientBellatrix
	}

	sourceWeightDelta := (reward * cfg.TimelySourceWeight) / cfg.WeightDenominator
	targetWeightDelta := (reward * cfg.TimelyTargetWeight) / cfg.WeightDenominator

	inactivityTargetDelta := val.EffectiveBalance() * cfg.InactivityScoreBias / (cfg.InactivityScoreBias * iPQ)
	delta := sourceWeightDelta + targetWeightDelta + inactivityTargetDelta
	return delta, nil
}
