package stateutil_test

import (
	"testing"

	"github.com/prysmaticlabs/prysm/v4/beacon-chain/state/stateutil"
	field_params "github.com/prysmaticlabs/prysm/v4/config/fieldparams"
	"github.com/prysmaticlabs/prysm/v4/consensus-types/primitives"
	ethpb "github.com/prysmaticlabs/prysm/v4/proto/prysm/v1alpha1"
	"github.com/prysmaticlabs/prysm/v4/testing/require"
)

func TestContractMapHandler(t *testing.T) {
	vals := []*ethpb.Validator{
		&ethpb.Validator{
			Contract:  []byte("zc"),
			PublicKey: []byte("zp"),
			ExitEpoch: 5, // Validator must be excluded from list cause exit epoch is lower than current.
		},
		&ethpb.Validator{
			Contract:  []byte("ac"),
			PublicKey: []byte("ap"),
			ExitEpoch: 1000,
		},
		&ethpb.Validator{
			Contract:  []byte("bc"),
			PublicKey: []byte("bp"),
			ExitEpoch: 1000,
		},
		&ethpb.Validator{
			Contract:  []byte("cc"),
			PublicKey: []byte("cp"),
			ExitEpoch: 1000,
		},
	}
	ep := primitives.Epoch(10)
	contractMapHandler := stateutil.NewContractMapHandler(vals, ep)

	vIdx, exists := contractMapHandler.Get([field_params.ContractAddressLength]byte{'b', 'c'})
	require.Equal(t, primitives.ValidatorIndex(2), vIdx)
	require.Equal(t, true, exists)

	// Validator was excluded from list, because exitEpoch is lower than current.
	_, exists = contractMapHandler.Get([field_params.ContractAddressLength]byte{'z', 'c'})
	require.Equal(t, false, exists)

	contractAddress := [field_params.ContractAddressLength]byte{'d', 'c'}
	contractMapHandler.Set(contractAddress, primitives.ValidatorIndex(5))
	vIdx, exists = contractMapHandler.Get(contractAddress)
	require.Equal(t, true, exists)
	require.Equal(t, primitives.ValidatorIndex(5), vIdx)

	copyOf := contractMapHandler.Copy()
	cVIdx, exists := copyOf.Get(contractAddress)
	require.Equal(t, true, exists)
	require.Equal(t, vIdx, cVIdx)
}
