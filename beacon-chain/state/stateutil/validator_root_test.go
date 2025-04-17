package stateutil_test

import (
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/prysmaticlabs/prysm/v4/beacon-chain/state/stateutil"
	"github.com/prysmaticlabs/prysm/v4/testing/require"
)

func BenchmarkUint64ListRootWithRegistryLimit(b *testing.B) {
	balances := make([]uint64, 100000)
	for i := 0; i < len(balances); i++ {
		balances[i] = uint64(i)
	}
	b.Run("100k balances", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := stateutil.Uint64ListRootWithRegistryLimit(balances)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func TestUint64ListRootWithRegistryLimit(t *testing.T) {
	balances := []uint64{8192, 8192, 8192, 8191, 8190}
	expectedRoot := "0x0dfb483c541f2abd4e42f8fc7620e712fd1a970ab638134f26aaec0b0171b828"
	root, err := stateutil.Uint64ListRootWithRegistryLimit(balances)
	require.NoError(t, err)
	require.Equal(t, expectedRoot, hexutil.Encode(root[:]))
}
