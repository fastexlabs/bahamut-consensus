package stateutils_test

import (
	"testing"

	"github.com/prysmaticlabs/prysm/v4/beacon-chain/core/transition/stateutils"
	state_native "github.com/prysmaticlabs/prysm/v4/beacon-chain/state/state-native"
	fieldparams "github.com/prysmaticlabs/prysm/v4/config/fieldparams"
	"github.com/prysmaticlabs/prysm/v4/consensus-types/primitives"
	"github.com/prysmaticlabs/prysm/v4/encoding/bytesutil"
	ethpb "github.com/prysmaticlabs/prysm/v4/proto/prysm/v1alpha1"
	"github.com/prysmaticlabs/prysm/v4/testing/assert"
	"github.com/prysmaticlabs/prysm/v4/testing/require"
)

func TestContractIndexMap_OK(t *testing.T) {
	base := &ethpb.BeaconState{
		Validators: []*ethpb.Validator{
			{
				Contract:  []byte("contract_zero"),
				ExitEpoch: 10,
			},
			{
				Contract:  []byte("contract_one"),
				ExitEpoch: 10,
			},
		},
	}
	state, err := state_native.InitializeFromProtoPhase0(base)
	require.NoError(t, err)

	tests := []struct {
		key [fieldparams.ContractAddressLength]byte
		val primitives.ValidatorIndex
		ok  bool
	}{
		{
			key: bytesutil.ToBytes20([]byte("contract_zero")),
			val: 0,
			ok:  true,
		}, {
			key: bytesutil.ToBytes20([]byte("contract_one")),
			val: 1,
			ok:  true,
		}, {
			key: bytesutil.ToBytes20([]byte("no")),
			val: 0,
			ok:  false,
		},
	}

	m := stateutils.ContractIndexMap(state.Validators(), 1)
	for _, tt := range tests {
		result, ok := m[tt.key]
		assert.Equal(t, tt.val, result)
		assert.Equal(t, tt.ok, ok)
	}
}

func TestContractIndexMap_ExitedValidator(t *testing.T) {
	base := &ethpb.BeaconState{
		Validators: []*ethpb.Validator{
			{
				Contract:  []byte("contract_zero"),
				ExitEpoch: 0,
			},
			{
				Contract:  []byte("contract_one"),
				ExitEpoch: 10,
			},
			{
				Contract:  []byte("contract_two"),
				ExitEpoch: 10,
			},
			{
				Contract:  []byte("contract_two"),
				ExitEpoch: 0,
			},
		},
	}
	state, err := state_native.InitializeFromProtoPhase0(base)
	require.NoError(t, err)

	tests := []struct {
		key [fieldparams.ContractAddressLength]byte
		val primitives.ValidatorIndex
		ok  bool
	}{
		{
			key: bytesutil.ToBytes20([]byte("contract_zero")),
			val: 0,
			ok:  false,
		}, {
			key: bytesutil.ToBytes20([]byte("contract_one")),
			val: 1,
			ok:  true,
		}, {
			key: bytesutil.ToBytes20([]byte("no")),
			val: 0,
			ok:  false,
		}, {
			key: bytesutil.ToBytes20([]byte("contract_two")),
			val: 2,
			ok:  true,
		},
	}

	m := stateutils.ContractIndexMap(state.Validators(), 1)
	for _, tt := range tests {
		result, ok := m[tt.key]
		assert.Equal(t, tt.val, result)
		assert.Equal(t, tt.ok, ok)
	}
}
