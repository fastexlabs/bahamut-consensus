package state_native_test

import (
	"testing"

	statenative "github.com/prysmaticlabs/prysm/v3/beacon-chain/state/state-native"
	fieldparams "github.com/prysmaticlabs/prysm/v3/config/fieldparams"
	ethpb "github.com/prysmaticlabs/prysm/v3/proto/prysm/v1alpha1"
	"github.com/prysmaticlabs/prysm/v3/testing/assert"
	"github.com/prysmaticlabs/prysm/v3/testing/require"
)

func TestReadOnlyContractsContaier_Contracts(t *testing.T) {
	contract := [fieldparams.ExecutionLayerAddressLength]byte{0xFF}
	contracts := [][]byte{contract[:]}
	rocc, err := statenative.NewContractsContainer(
		&ethpb.ContractsContainer{
			Contracts: contracts,
		},
	)
	require.NoError(t, err)
	assert.Equal(t, contract, rocc.Contracts()[0])
}

func TestReadOnlyContractsContaier_ContractAtIndex(t *testing.T) {
	contract := [fieldparams.ExecutionLayerAddressLength]byte{0xFF}
	contracts := [][]byte{contract[:]}
	rocc, err := statenative.NewContractsContainer(
		&ethpb.ContractsContainer{
			Contracts: contracts,
		},
	)
	require.NoError(t, err)
	c, err := rocc.ContractAtIndex(0)
	require.NoError(t, err)
	assert.Equal(t, contract, c)
	_, err = rocc.ContractAtIndex(1)
	e := statenative.NewContractIndexOutOfRangeError(1)
	require.Equal(t, err.Error(), e.Error())
}
