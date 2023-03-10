package types

import (
	"testing"

	"github.com/prysmaticlabs/prysm/v3/testing/assert"
)

func TestStateFieldIndexes(t *testing.T) {
	assert.Equal(t, FieldIndex(0), GenesisTime)
	assert.Equal(t, FieldIndex(1), GenesisValidatorsRoot)
	assert.Equal(t, FieldIndex(2), Slot)
	assert.Equal(t, FieldIndex(3), Fork)
	assert.Equal(t, FieldIndex(4), LatestBlockHeader)
	assert.Equal(t, FieldIndex(5), BlockRoots)
	assert.Equal(t, FieldIndex(6), StateRoots)
	assert.Equal(t, FieldIndex(7), HistoricalRoots)
	assert.Equal(t, FieldIndex(8), Eth1Data)
	assert.Equal(t, FieldIndex(9), Eth1DataVotes)
	assert.Equal(t, FieldIndex(10), Eth1DepositIndex)
	assert.Equal(t, FieldIndex(11), LatestProcessedBlockActivities)
	assert.Equal(t, FieldIndex(12), Validators)
	assert.Equal(t, FieldIndex(13), Balances)
	assert.Equal(t, FieldIndex(14), Contracts)
	assert.Equal(t, FieldIndex(15), Activities)
	assert.Equal(t, FieldIndex(16), RandaoMixes)
	assert.Equal(t, FieldIndex(17), Slashings)
	assert.Equal(t, FieldIndex(18), PreviousEpochAttestations)
	assert.Equal(t, FieldIndex(18), PreviousEpochParticipationBits)
	assert.Equal(t, FieldIndex(19), CurrentEpochAttestations)
	assert.Equal(t, FieldIndex(19), CurrentEpochParticipationBits)
	assert.Equal(t, FieldIndex(20), JustificationBits)
	assert.Equal(t, FieldIndex(21), PreviousJustifiedCheckpoint)
	assert.Equal(t, FieldIndex(22), CurrentJustifiedCheckpoint)
	assert.Equal(t, FieldIndex(23), FinalizedCheckpoint)
	assert.Equal(t, FieldIndex(24), InactivityScores)
	assert.Equal(t, FieldIndex(25), CurrentSyncCommittee)
	assert.Equal(t, FieldIndex(26), NextSyncCommittee)
}
