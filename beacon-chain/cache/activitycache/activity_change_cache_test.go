package activitycache

import (
	"context"
	"math/big"
	"testing"

	ethpb "github.com/prysmaticlabs/prysm/v3/proto/prysm/v1alpha1"
	"github.com/prysmaticlabs/prysm/v3/testing/assert"
	"github.com/prysmaticlabs/prysm/v3/testing/require"
	logTest "github.com/sirupsen/logrus/hooks/test"
)

const nilActivityChangesErr = "Ignoring nil activity changes insertion"

var _ ActivityChangesFetcher = (*ActivityChangesCache)(nil)

func TestInsertActivityChanges_LogsOnNilActivityChangesInsertion(t *testing.T) {
    hook := logTest.NewGlobal()
    acc := New()

    assert.ErrorContains(t, "nil activity changes inserted into the cache", acc.InsertActivityChanges(context.Background(), new(big.Int).SetUint64(1), nil))

    require.Equal(t, 0, len(acc.activityChanges), "Number of activity changes is not changed")
    assert.Equal(t, nilActivityChangesErr, hook.LastEntry().Message)
}

func TestGetActivityChanges(t *testing.T) {
	var activityChanges []*ethpb.ActivityChange
	for i := 0; i < 100; i++ {
		activityChanges = append(activityChanges, &ethpb.ActivityChange{
			ContractAddress: []byte{byte(i)},
			DeltaActivity:   uint64(i),
		})
	}

	acc := New()

	err := acc.InsertActivityChanges(context.Background(), new(big.Int).SetUint64(1), activityChanges)
	require.NoError(t, err)
	require.Equal(t, 100, len(acc.activityChanges[1]), "All deposits inserted")

	receivedChanges := acc.GetActivityChanges(context.Background(), new(big.Int).SetUint64(1))
	require.Equal(t, 1, len(acc.activityChanges), "Number of activity changes changed")
	require.Equal(t, len(activityChanges), len(receivedChanges), "Number of activity changes received")
}


