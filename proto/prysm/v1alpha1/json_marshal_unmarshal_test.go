package eth_test

import (
	"encoding/json"
	"testing"

	"github.com/prysmaticlabs/prysm/v4/encoding/bytesutil"
	eth "github.com/prysmaticlabs/prysm/v4/proto/prysm/v1alpha1"
	"github.com/prysmaticlabs/prysm/v4/testing/require"
)

func TestJsonMarshalUnmarshal(t *testing.T) {
	t.Run("activity change", func(t *testing.T) {
		address := bytesutil.PadTo([]byte("contract"), 20)
		delta := uint64(123)
		jsonActivityChange := &eth.ActivityChange{
			ContractAddress: address,
			DeltaActivity:   delta,
		}
		enc, err := json.Marshal(jsonActivityChange)
		require.NoError(t, err)
		activityChange := &eth.ActivityChange{}
		require.NoError(t, json.Unmarshal(enc, activityChange))
		require.DeepEqual(t, address, activityChange.ContractAddress)
		require.DeepEqual(t, delta, activityChange.DeltaActivity)
	})
	t.Run("block activitities", func(t *testing.T) {
		address := bytesutil.PadTo([]byte("contract"), 20)
		delta := uint64(123)
		baseFee := uint64(123123)
		txCount := uint64(1)
		jsonActivityChange := &eth.ActivityChange{
			ContractAddress: address,
			DeltaActivity:   delta,
		}
		jsonBlockActivities := &eth.BlockActivities{
			BaseFee:    baseFee,
			TxCount:    txCount,
			Activities: []*eth.ActivityChange{jsonActivityChange},
		}
		enc, err := json.Marshal(jsonBlockActivities)
		require.NoError(t, err)
		blockActivities := &eth.BlockActivities{}
		require.NoError(t, json.Unmarshal(enc, blockActivities))
		require.DeepEqual(t, baseFee, blockActivities.BaseFee)
		require.DeepEqual(t, txCount, blockActivities.TxCount)
		activityChange := blockActivities.Activities[0]
		require.DeepEqual(t, address, activityChange.ContractAddress)
		require.DeepEqual(t, delta, activityChange.DeltaActivity)
	})
}
