package state_native_test

import (
	"testing"

	state_native "github.com/prysmaticlabs/prysm/v4/beacon-chain/state/state-native"
	"github.com/prysmaticlabs/prysm/v4/consensus-types/primitives"
	ethpb "github.com/prysmaticlabs/prysm/v4/proto/prysm/v1alpha1"
	"github.com/prysmaticlabs/prysm/v4/testing/require"
)

func BenchmarkAppendBalance(b *testing.B) {
	st, err := state_native.InitializeFromProtoPhase0(&ethpb.BeaconState{})
	require.NoError(b, err)

	max := uint64(16777216)
	for i := uint64(0); i < max-2; i++ {
		require.NoError(b, st.AppendBalance(i))
	}

	ref := st.Copy()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		require.NoError(b, ref.AppendBalance(uint64(i)))
		ref = st.Copy()
	}
}

func BenchmarkAppendInactivityScore(b *testing.B) {
	st, err := state_native.InitializeFromProtoCapella(&ethpb.BeaconStateCapella{})
	require.NoError(b, err)

	max := uint64(16777216)
	for i := uint64(0); i < max-2; i++ {
		require.NoError(b, st.AppendInactivityScore(i))
	}

	ref := st.Copy()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		require.NoError(b, ref.AppendInactivityScore(uint64(i)))
		ref = st.Copy()
	}
}

func TestActivities(t *testing.T) {
	t.Run("activities check", func(t *testing.T) {
		st, err := state_native.InitializeFromProtoCapella(&ethpb.BeaconStateCapella{})
		require.NoError(t, err)

		expectedActivities := []uint64{1, 3, 5}
		require.NoError(t, st.SetActivities(expectedActivities))

		require.DeepEqual(t, expectedActivities, st.Activities())

		require.ErrorContains(t, "invalid index provided", st.UpdateActivityAtIndex(primitives.ValidatorIndex(5), 15))

		require.NoError(t, st.UpdateActivityAtIndex(primitives.ValidatorIndex(1), 15))

		_, err = st.ActivityAtIndex(primitives.ValidatorIndex(15))
		require.ErrorContains(t, "activity index 15 does not exist", err)

		act, err := st.ActivityAtIndex(primitives.ValidatorIndex(1))
		require.NoError(t, err)
		require.Equal(t, uint64(15), act)
		require.Equal(t, 3, st.ActivitiesLength())

		require.NoError(t, st.AppendActivity(78))
		expectedActivities = append(expectedActivities, 78)
		require.Equal(t, 4, st.ActivitiesLength())
		require.DeepEqual(t, expectedActivities, st.Activities())
	})
}
