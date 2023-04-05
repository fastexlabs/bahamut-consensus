package basefeecache

import (
	"context"
	"math/big"
	"testing"

	"github.com/prysmaticlabs/prysm/v3/testing/require"
)

var _ BaseFeeFetcher = (*BaseFeeCache)(nil)

func TestBaseFeeCache_InsertBaseFee(t *testing.T) {
	bc := New()

	err := bc.InsertBaseFee(context.Background(), new(big.Int).SetUint64(1), 413534546)
	require.NoError(t, err)
	require.Equal(t, uint64(413534546), bc.baseFee[1], "Base Fee is not inserted")
}

func TestBaseFeeCache_GetBaseFee(t *testing.T) {
	bc := New()

	err := bc.InsertBaseFee(context.Background(), new(big.Int).SetUint64(1), 413534546)
	require.NoError(t, err)
	require.Equal(t, uint64(413534546), bc.baseFee[1], "Base Fee is not inserted")

	baseFee := bc.GetBaseFee(context.Background(), new(big.Int).SetUint64(2))
	require.Equal(t, uint64(0), baseFee, "Base Fee is not 0")

	baseFee = bc.GetBaseFee(context.Background(), new(big.Int).SetUint64(1))
	require.Equal(t, uint64(413534546), baseFee, "Base Fee returned")
}
