package state_native

import (
	"github.com/prysmaticlabs/prysm/v3/beacon-chain/state/state-native/types"
	"github.com/prysmaticlabs/prysm/v3/beacon-chain/state/stateutil"
	ethpb "github.com/prysmaticlabs/prysm/v3/proto/prysm/v1alpha1"
	"github.com/prysmaticlabs/prysm/v3/runtime/version"
)

// SetEth1Data for the beacon state.
func (b *BeaconState) SetEth1Data(val *ethpb.Eth1Data) error {
	b.lock.Lock()
	defer b.lock.Unlock()

	b.eth1Data = val
	b.markFieldAsDirty(types.Eth1Data)
	return nil
}

// SetEth1DataVotes for the beacon state. Updates the entire
// list to a new value by overwriting the previous one.
func (b *BeaconState) SetEth1DataVotes(val []*ethpb.Eth1Data) error {
	b.lock.Lock()
	defer b.lock.Unlock()

	b.sharedFieldReferences[types.Eth1DataVotes].MinusRef()
	b.sharedFieldReferences[types.Eth1DataVotes] = stateutil.NewRef(1)

	b.eth1DataVotes = val
	b.markFieldAsDirty(types.Eth1DataVotes)
	b.rebuildTrie[types.Eth1DataVotes] = true
	return nil
}

// SetEth1DepositIndex for the beacon state.
func (b *BeaconState) SetEth1DepositIndex(val uint64) error {
	b.lock.Lock()
	defer b.lock.Unlock()

	b.eth1DepositIndex = val
	b.markFieldAsDirty(types.Eth1DepositIndex)
	return nil
}

// AppendEth1DataVotes for the beacon state. Appends the new value
// to the the end of list.
func (b *BeaconState) AppendEth1DataVotes(val *ethpb.Eth1Data) error {
	b.lock.Lock()
	defer b.lock.Unlock()

	votes := b.eth1DataVotes
	if b.sharedFieldReferences[types.Eth1DataVotes].Refs() > 1 {
		// Copy elements in underlying array by reference.
		votes = make([]*ethpb.Eth1Data, len(b.eth1DataVotes))
		copy(votes, b.eth1DataVotes)
		b.sharedFieldReferences[types.Eth1DataVotes].MinusRef()
		b.sharedFieldReferences[types.Eth1DataVotes] = stateutil.NewRef(1)
	}

	b.eth1DataVotes = append(votes, val)
	b.markFieldAsDirty(types.Eth1DataVotes)
	b.addDirtyIndices(types.Eth1DataVotes, []uint64{uint64(len(b.eth1DataVotes) - 1)})
	return nil
}

// SetLatestProcessedBlockActivities for the beacon state.
func (b *BeaconState) SetLatestProcessedBlockActivities(val uint64) error {
	b.lock.Lock()
	defer b.lock.Unlock()

	b.latestProcessedBlockActivities = val
	b.markFieldAsDirty(types.LatestProcessedBlockActivities)
	return nil
}

// SetTransactionsGasPerPeriod for the beacon state.
func (b *BeaconState) SetTransactionsGasPerPeriod(val uint64) error {
	b.lock.Lock()
	defer b.lock.Unlock()

	b.transactionsGasPerPeriod = val
	b.markFieldAsDirty(types.TransactionsGasPerPeriod)
	return nil
}

// SetTransactionsPerLatestEpoch for the beacon state.
func (b *BeaconState) SetTransactionsPerLatestEpoch(val uint64) error {
	b.lock.Lock()
	defer b.lock.Unlock()

	b.transactionsPerLatestEpoch = val
	b.markFieldAsDirty(types.TransactionsPerLatestEpoch)
	return nil
}

// SetNonStakersGasPerEpoch for the beacon state.
func (b *BeaconState) SetNonStakersGasPerEpoch(val uint64) error {
	b.lock.Lock()
	defer b.lock.Unlock()

	b.nonStakersGasPerEpoch = val
	b.markFieldAsDirty(types.NonStakersGasPerEpoch)
	return nil
}

// SetNonStakersGasPerPeriod for the beacon state.
func (b *BeaconState) SetNonStakersGasPerPeriod(val uint64) error {
	b.lock.Lock()
	defer b.lock.Unlock()

	b.nonStakersGasPerPeriod = val
	b.markFieldAsDirty(types.NonStakersGasPerPeriod)
	return nil
}

func (b *BeaconState) SetBaseFeePerEpoch(val uint64) error {
	b.lock.Lock()
	defer b.lock.Unlock()

	if b.version < version.FastexPhase1 {
		return errNotSupported("SetBaseFeePerEpoch", b.version)
	}

	b.baseFeePerEpoch = val
	b.markFieldAsDirty(types.BaseFeePerEpoch)
	return nil
}

func (b *BeaconState) SetBaseFeePerPeriod(val uint64) error {
	b.lock.Lock()
	defer b.lock.Unlock()

	if b.version < version.FastexPhase1 {
		return errNotSupported("SetBaseFeePerPeriod", b.version)
	}

	b.baseFeePerPeriod = val
	b.markFieldAsDirty(types.BaseFeePerPeriod)
	return nil
}
