package stateutil_test

import (
	"testing"

	"github.com/prysmaticlabs/prysm/v4/beacon-chain/state/stateutil"
	ethpb "github.com/prysmaticlabs/prysm/v4/proto/prysm/v1alpha1"
	"github.com/prysmaticlabs/prysm/v4/testing/require"
)

func TestSharedActivityRoot(t *testing.T) {
	sharedActivity := ethpb.SharedActivity{
		TransactionsGasPerPeriod: 1000,
		BaseFeePerPeriod:         102020,
		BaseFeePerEpoch:          90912,
	}

	expectedRoot, err := sharedActivity.HashTreeRoot()
	require.NoError(t, err)
	root, err := stateutil.SharedActivityRoot(&sharedActivity)
	require.NoError(t, err)
	require.DeepEqual(t, expectedRoot, root)
}
