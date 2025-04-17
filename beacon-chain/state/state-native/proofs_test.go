package state_native_test

import (
	"context"
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"
	statenative "github.com/prysmaticlabs/prysm/v4/beacon-chain/state/state-native"
	"github.com/prysmaticlabs/prysm/v4/container/trie"
	"github.com/prysmaticlabs/prysm/v4/testing/require"
	"github.com/prysmaticlabs/prysm/v4/testing/util"
)

func TestBeaconStateMerkleProofs_phase0_notsupported(t *testing.T) {
	ctx := context.Background()
	st, _ := util.DeterministicGenesisState(t, 256)
	t.Run("current sync committee", func(t *testing.T) {
		_, err := st.CurrentSyncCommitteeProof(ctx)
		require.ErrorContains(t, "not supported", err)
	})
	t.Run("next sync committee", func(t *testing.T) {
		_, err := st.NextSyncCommitteeProof(ctx)
		require.ErrorContains(t, "not supported", err)
	})
	t.Run("finalized root", func(t *testing.T) {
		_, err := st.FinalizedRootProof(ctx)
		require.ErrorContains(t, "not supported", err)
	})
}
func TestBeaconStateMerkleProofs_altair(t *testing.T) {
	ctx := context.Background()
	altair, err := util.NewBeaconStateAltair()
	require.NoError(t, err)
	htr, err := altair.HashTreeRoot(ctx)
	require.NoError(t, err)
	results := []string{
		"0x173669ae8794c057def63b20372114a628abb029354a2ef50d7a1aaa9a3dab4a",
		"0xf5a5fd42d16a20302798ef6ed309979b43003d2320d9f0e8ea9831a92759fb4b",
		"0xdb56114e00fdd4c1f85c892bf35ac9a89289aaecb1ebd0a96cde606a748b5d71",
		"0xa07848503b66e6c6c699d31177115a1446c76dced4b85476e75707b10ee53189",
		"0x36b127a953720f642c5e1cd6a24081c9e21b80ff5f21b08367f76f8e3e9dd71d",
	}
	t.Run("current sync committee", func(t *testing.T) {
		cscp, err := altair.CurrentSyncCommitteeProof(ctx)
		require.NoError(t, err)
		require.Equal(t, len(cscp), 5)
		for i, bytes := range cscp {
			res := hexutil.Encode(bytes)
			require.Equal(t, results[i], res)
		}
	})
	t.Run("next sync committee", func(t *testing.T) {
		nscp, err := altair.NextSyncCommitteeProof(ctx)
		require.NoError(t, err)
		require.Equal(t, len(nscp), 5)
		for i, bytes := range nscp {
			require.Equal(t, hexutil.Encode(bytes), results[i])
		}
	})
	// TODO: Values are not equal, want: true (bool), got: false (bool).
	t.Run("finalized root", func(t *testing.T) {
		finalizedRoot := altair.FinalizedCheckpoint().Root
		proof, err := altair.FinalizedRootProof(ctx)
		require.NoError(t, err)
		gIndex := statenative.FinalizedRootGeneralizedIndex()
		valid := trie.VerifyMerkleProof(htr[:], finalizedRoot, gIndex, proof)
		require.Equal(t, true, valid)
	})
	// TODO: Values are not equal, want: true (bool), got: false (bool).
	t.Run("recomputes root on dirty fields", func(t *testing.T) {
		currentRoot, err := altair.HashTreeRoot(ctx)
		require.NoError(t, err)
		cpt := altair.FinalizedCheckpoint()
		require.NoError(t, err)

		// Edit the checkpoint.
		cpt.Epoch = 100
		require.NoError(t, altair.SetFinalizedCheckpoint(cpt))

		// Produce a proof for the finalized root.
		proof, err := altair.FinalizedRootProof(ctx)
		require.NoError(t, err)

		// We expect the previous step to have triggered
		// a recomputation of dirty fields in the beacon state, resulting
		// in a new hash tree root as the finalized checkpoint had previously
		// changed and should have been marked as a dirty state field.
		// The proof validity should be false for the old root, but true for the new.
		finalizedRoot := altair.FinalizedCheckpoint().Root
		gIndex := statenative.FinalizedRootGeneralizedIndex()
		valid := trie.VerifyMerkleProof(currentRoot[:], finalizedRoot, gIndex, proof)
		require.Equal(t, false, valid)

		newRoot, err := altair.HashTreeRoot(ctx)
		require.NoError(t, err)

		valid = trie.VerifyMerkleProof(newRoot[:], finalizedRoot, gIndex, proof)
		require.Equal(t, true, valid)
	})
}

func TestBeaconStateMerkleProofs_bellatrix(t *testing.T) {
	ctx := context.Background()
	bellatrix, err := util.NewBeaconStateBellatrix()
	require.NoError(t, err)
	htr, err := bellatrix.HashTreeRoot(ctx)
	require.NoError(t, err)
	results := []string{
		"0x173669ae8794c057def63b20372114a628abb029354a2ef50d7a1aaa9a3dab4a",
		"0xb68b2f519878bcdc8fce2bba633a841e89e757c901224e731ec16a2397fdca74",
		"0xdb56114e00fdd4c1f85c892bf35ac9a89289aaecb1ebd0a96cde606a748b5d71",
		"0xa07848503b66e6c6c699d31177115a1446c76dced4b85476e75707b10ee53189",
		"0x36b127a953720f642c5e1cd6a24081c9e21b80ff5f21b08367f76f8e3e9dd71d",
	}
	t.Run("current sync committee", func(t *testing.T) {
		cscp, err := bellatrix.CurrentSyncCommitteeProof(ctx)
		require.NoError(t, err)
		require.Equal(t, len(cscp), 5)

		for i, bytes := range cscp {
			require.Equal(t, hexutil.Encode(bytes), results[i])
		}
	})
	t.Run("next sync committee", func(t *testing.T) {
		nscp, err := bellatrix.NextSyncCommitteeProof(ctx)
		require.NoError(t, err)
		require.Equal(t, len(nscp), 5)
		for i, bytes := range nscp {
			require.Equal(t, hexutil.Encode(bytes), results[i])
		}
	})
	// TODO: Values are not equal, want: true (bool), got: false (bool).
	t.Run("finalized root", func(t *testing.T) {
		finalizedRoot := bellatrix.FinalizedCheckpoint().Root
		proof, err := bellatrix.FinalizedRootProof(ctx)
		require.NoError(t, err)
		gIndex := statenative.FinalizedRootGeneralizedIndex()
		valid := trie.VerifyMerkleProof(htr[:], finalizedRoot, gIndex, proof)
		require.Equal(t, true, valid)
	})
	// TODO: Values are not equal, want: true (bool), got: false (bool).
	t.Run("recomputes root on dirty fields", func(t *testing.T) {
		currentRoot, err := bellatrix.HashTreeRoot(ctx)
		require.NoError(t, err)
		cpt := bellatrix.FinalizedCheckpoint()
		require.NoError(t, err)

		// Edit the checkpoint.
		cpt.Epoch = 100
		require.NoError(t, bellatrix.SetFinalizedCheckpoint(cpt))

		// Produce a proof for the finalized root.
		proof, err := bellatrix.FinalizedRootProof(ctx)
		require.NoError(t, err)

		// We expect the previous step to have triggered
		// a recomputation of dirty fields in the beacon state, resulting
		// in a new hash tree root as the finalized checkpoint had previously
		// changed and should have been marked as a dirty state field.
		// The proof validity should be false for the old root, but true for the new.
		finalizedRoot := bellatrix.FinalizedCheckpoint().Root
		gIndex := statenative.FinalizedRootGeneralizedIndex()
		valid := trie.VerifyMerkleProof(currentRoot[:], finalizedRoot, gIndex, proof)
		require.Equal(t, false, valid)

		newRoot, err := bellatrix.HashTreeRoot(ctx)
		require.NoError(t, err)

		valid = trie.VerifyMerkleProof(newRoot[:], finalizedRoot, gIndex, proof)
		require.Equal(t, true, valid)
	})
}
