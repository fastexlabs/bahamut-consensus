package stateutil

import (
	"testing"

	ethpb "github.com/prysmaticlabs/prysm/v4/proto/prysm/v1alpha1"
	"github.com/prysmaticlabs/prysm/v4/testing/require"
)

func TestSharedActivityRootWithHasher(t *testing.T) {
	sharedActivity := &ethpb.SharedActivity{
		TransactionsGasPerPeriod: 10000000,
		TransactionsGasPerEpoch:  10000000000,
		BaseFeePerPeriod:         100000000000,
		BaseFeePerEpoch:          100000000,
	}

	expectedRoot, err := sharedActivity.HashTreeRoot()
	require.NoError(t, err)

	t.Run("empty shared activity", func(tt *testing.T) {
		_, err := SharedActivityRootWithHasher(nil)
		require.ErrorContains(tt, "empty shared activity", err)
	})

	t.Run("root check with expected", func(tt *testing.T) {
		root, err := SharedActivityRootWithHasher(sharedActivity)
		require.NoError(tt, err)
		require.Equal(tt, expectedRoot, root)
	})
}
