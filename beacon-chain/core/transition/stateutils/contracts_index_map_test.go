package stateutils_test

import (
	"testing"

	"github.com/prysmaticlabs/prysm/v3/beacon-chain/core/transition/stateutils"
	state_native "github.com/prysmaticlabs/prysm/v3/beacon-chain/state/state-native"
	fieldparams "github.com/prysmaticlabs/prysm/v3/config/fieldparams"
	types "github.com/prysmaticlabs/prysm/v3/consensus-types/primitives"
	"github.com/prysmaticlabs/prysm/v3/encoding/bytesutil"
	ethpb "github.com/prysmaticlabs/prysm/v3/proto/prysm/v1alpha1"
	"github.com/prysmaticlabs/prysm/v3/testing/assert"
	"github.com/prysmaticlabs/prysm/v3/testing/require"
)

func TestContractsIndexMap_OK(t *testing.T) {
	base := &ethpb.BeaconState{
		Contracts: []*ethpb.ContractsContainer{
			{
				Contracts: [][]byte{{1}, {2}, {3}},
			},
			{
				Contracts: [][]byte{{4}, {5}},
			},
		},
	}
	state, err := state_native.InitializeFromProtoPhase0(base)
	require.NoError(t, err)

	tests := []struct {
		key [fieldparams.ExecutionLayerAddressLength]byte
		val types.ValidatorIndex
		ok  bool
	}{
		{
			key: bytesutil.ToBytes20([]byte{1}),
			val: 0,
			ok:  true,
		}, {
			key: bytesutil.ToBytes20([]byte{2}),
			val: 0,
			ok:  true,
		}, {
			key: bytesutil.ToBytes20([]byte{3}),
			val: 0,
			ok:  true,
		}, {
			key: bytesutil.ToBytes20([]byte{4}),
			val: 1,
			ok:  true,
		}, {
			key: bytesutil.ToBytes20([]byte{5}),
			val: 1,
			ok:  true,
		}, {
			key: bytesutil.ToBytes20([]byte("no")),
			val: 0,
			ok:  false,
		},
	}

	m := stateutils.ContractsIndexMap(state.Contracts())
	for _, tt := range tests {
		result, ok := m[tt.key]
		assert.Equal(t, tt.val, result)
		assert.Equal(t, tt.ok, ok)
	}
}
