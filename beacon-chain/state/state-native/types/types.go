package types

import (
	"github.com/pkg/errors"
)

// FieldIndex represents the relevant field position in the
// state struct for a field.
type FieldIndex int

// String returns the name of the field index.
func (f FieldIndex) String(_ int) string {
	switch f {
	case GenesisTime:
		return "genesisTime"
	case GenesisValidatorsRoot:
		return "genesisValidatorsRoot"
	case Slot:
		return "slot"
	case Fork:
		return "fork"
	case LatestBlockHeader:
		return "latestBlockHeader"
	case BlockRoots:
		return "blockRoots"
	case StateRoots:
		return "stateRoots"
	case HistoricalRoots:
		return "historicalRoots"
	case Eth1Data:
		return "eth1Data"
	case Eth1DataVotes:
		return "eth1DataVotes"
	case Eth1DepositIndex:
		return "eth1DepositIndex"
	case LatestProcessedBlockActivities:
		return "latestProcessedBlockActivities"
	case TransactionsGasPerPeriod:
		return "transactionsGasPerPeriod"
	case TransactionsPerLatestEpoch:
		return "transactionsPerLatestEpoch"
	case NonStakersGasPerEpoch:
		return "nonStakersGasPerEpoch"
	case NonStakersGasPerPeriod:
		return "nonStakersGasPerPeriod"
	case Validators:
		return "validators"
	case Balances:
		return "balances"
	case Contracts:
		return "contracts"
	case Activities:
		return "activities"
	case RandaoMixes:
		return "randaoMixes"
	case Slashings:
		return "slashings"
	case PreviousEpochAttestations:
		return "previousEpochAttestations"
	case CurrentEpochAttestations:
		return "currentEpochAttestations"
	case PreviousEpochParticipationBits:
		return "previousEpochParticipationBits"
	case CurrentEpochParticipationBits:
		return "currentEpochParticipationBits"
	case JustificationBits:
		return "justificationBits"
	case PreviousJustifiedCheckpoint:
		return "previousJustifiedCheckpoint"
	case CurrentJustifiedCheckpoint:
		return "currentJustifiedCheckpoint"
	case FinalizedCheckpoint:
		return "finalizedCheckpoint"
	case InactivityScores:
		return "inactivityScores"
	case CurrentSyncCommittee:
		return "currentSyncCommittee"
	case NextSyncCommittee:
		return "nextSyncCommittee"
	case LatestExecutionPayloadHeader:
		return "latestExecutionPayloadHeader"
	default:
		return ""
	}
}

// RealPosition denotes the position of the field in the beacon state.
// The value might differ for different state versions.
func (f FieldIndex) RealPosition() int {
	switch f {
	case GenesisTime:
		return 0
	case GenesisValidatorsRoot:
		return 1
	case Slot:
		return 2
	case Fork:
		return 3
	case LatestBlockHeader:
		return 4
	case BlockRoots:
		return 5
	case StateRoots:
		return 6
	case HistoricalRoots:
		return 7
	case Eth1Data:
		return 8
	case Eth1DataVotes:
		return 9
	case Eth1DepositIndex:
		return 10
	case LatestProcessedBlockActivities:
		return 11
	case TransactionsGasPerPeriod:
		return 12
	case TransactionsPerLatestEpoch:
		return 13
	case NonStakersGasPerEpoch:
		return 14
	case NonStakersGasPerPeriod:
		return 15
	case Validators:
		return 16
	case Balances:
		return 17
	case Contracts:
		return 18
	case Activities:
		return 19
	case RandaoMixes:
		return 20
	case Slashings:
		return 21
	case PreviousEpochAttestations, PreviousEpochParticipationBits:
		return 22
	case CurrentEpochAttestations, CurrentEpochParticipationBits:
		return 23
	case JustificationBits:
		return 24
	case PreviousJustifiedCheckpoint:
		return 25
	case CurrentJustifiedCheckpoint:
		return 26
	case FinalizedCheckpoint:
		return 27
	case InactivityScores:
		return 28
	case CurrentSyncCommittee:
		return 29
	case NextSyncCommittee:
		return 30
	case LatestExecutionPayloadHeader:
		return 31
	default:
		return -1
	}
}

// ElemsInChunk returns the number of elements in the chunk (number of
// elements that are able to be packed).
func (f FieldIndex) ElemsInChunk() (uint64, error) {
	switch f {
	case Balances, Activities:
		return 4, nil
	default:
		return 0, errors.Errorf("field %d doesn't support element compression", f)
	}
}

func (FieldIndex) Native() bool {
	return true
}

// Below we define a set of useful enum values for the field
// indices of the beacon state. For example, genesisTime is the
// 0th field of the beacon state. This is helpful when we are
// updating the Merkle branches up the trie representation
// of the beacon state. The below field indexes correspond
// to the state.
const (
	GenesisTime FieldIndex = iota
	GenesisValidatorsRoot
	Slot
	Fork
	LatestBlockHeader
	BlockRoots
	StateRoots
	HistoricalRoots
	Eth1Data
	Eth1DataVotes
	Eth1DepositIndex
	LatestProcessedBlockActivities
	TransactionsGasPerPeriod
	TransactionsPerLatestEpoch
	NonStakersGasPerEpoch
	NonStakersGasPerPeriod
	Validators
	Balances
	Contracts
	Activities
	RandaoMixes
	Slashings
	PreviousEpochAttestations
	CurrentEpochAttestations
	PreviousEpochParticipationBits
	CurrentEpochParticipationBits
	JustificationBits
	PreviousJustifiedCheckpoint
	CurrentJustifiedCheckpoint
	FinalizedCheckpoint
	InactivityScores
	CurrentSyncCommittee
	NextSyncCommittee
	LatestExecutionPayloadHeader
)
