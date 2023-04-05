package state_native

import (
	ethpb "github.com/prysmaticlabs/prysm/v3/proto/prysm/v1alpha1"
	"github.com/prysmaticlabs/prysm/v3/runtime/version"
)

// Eth1Data corresponding to the proof-of-work chain information stored in the beacon state.
func (b *BeaconState) Eth1Data() *ethpb.Eth1Data {
	if b.eth1Data == nil {
		return nil
	}

	b.lock.RLock()
	defer b.lock.RUnlock()

	return b.eth1DataVal()
}

// eth1DataVal corresponding to the proof-of-work chain information stored in the beacon state.
// This assumes that a lock is already held on BeaconState.
func (b *BeaconState) eth1DataVal() *ethpb.Eth1Data {
	if b.eth1Data == nil {
		return nil
	}

	return ethpb.CopyETH1Data(b.eth1Data)
}

// Eth1DataVotes corresponds to votes from Ethereum on the canonical proof-of-work chain
// data retrieved from eth1.
func (b *BeaconState) Eth1DataVotes() []*ethpb.Eth1Data {
	if b.eth1DataVotes == nil {
		return nil
	}

	b.lock.RLock()
	defer b.lock.RUnlock()

	return b.eth1DataVotesVal()
}

// eth1DataVotesVal corresponds to votes from Ethereum on the canonical proof-of-work chain
// data retrieved from eth1.
// This assumes that a lock is already held on BeaconState.
func (b *BeaconState) eth1DataVotesVal() []*ethpb.Eth1Data {
	if b.eth1DataVotes == nil {
		return nil
	}

	res := make([]*ethpb.Eth1Data, len(b.eth1DataVotes))
	for i := 0; i < len(res); i++ {
		res[i] = ethpb.CopyETH1Data(b.eth1DataVotes[i])
	}
	return res
}

// Eth1DepositIndex corresponds to the index of the deposit made to the
// validator deposit contract at the time of this state's eth1 data.
func (b *BeaconState) Eth1DepositIndex() uint64 {
	b.lock.RLock()
	defer b.lock.RUnlock()

	return b.eth1DepositIndex
}

// LatestProcessedBlockActivities corresponds to the number of the block
// from which latest activity changes was retrieved
func (b *BeaconState) LatestProcessedBlockActivities() uint64 {
	b.lock.RLock()
	defer b.lock.RUnlock()

	return b.latestProcessedBlockActivities
}

// TransactionsGasPerPeriod corresponds to the amount of gas used by
// transactions during activity period.
func (b *BeaconState) TransactionsGasPerPeriod() uint64 {
	b.lock.RLock()
	defer b.lock.RUnlock()

	return b.transactionsGasPerPeriod
}

// TransactionsPerLatestEpoch corresponds to the number of transactions
// executed during latest epoch.
func (b *BeaconState) TransactionsPerLatestEpoch() uint64 {
	b.lock.RLock()
	defer b.lock.RUnlock()

	return b.transactionsPerLatestEpoch
}

// NonStakersGasPerEpoch corresponds to the amount of gas used by
// contracts that are not in validators' contracts list during latest epoch.
func (b *BeaconState) NonStakersGasPerEpoch() uint64 {
	b.lock.RLock()
	defer b.lock.RUnlock()

	return b.nonStakersGasPerEpoch
}

// NonStakersGasPerEpoch corresponds to the amount of gas used by
// contracts that are not in validators' contracts list during acitivity period.
func (b *BeaconState) NonStakersGasPerPeriod() uint64 {
	b.lock.RLock()
	defer b.lock.RUnlock()

	return b.nonStakersGasPerPeriod
}

// BaseFeePerEpoch corresponds to the sum of base fee per gas from blocks
// processed in during latest epoch.
func (b *BeaconState) BaseFeePerEpoch() (uint64, error) {
	if b.version < version.FastexPhase1 {
		return 0, errNotSupported("BaseFeePerEpoch", b.version)
	}

	b.lock.RLock()
	defer b.lock.RUnlock()

	return b.baseFeePerEpoch, nil
}

// BaseFeePerPeriod corresponds to the average base fee per gas
// during activity period.
func (b *BeaconState) BaseFeePerPeriod() (uint64, error) {
	if b.version < version.FastexPhase1 {
		return 0, errNotSupported("BaseFeePerPeriod", b.version)
	}

	b.lock.RLock()
	defer b.lock.RUnlock()

	return b.baseFeePerPeriod, nil
}
