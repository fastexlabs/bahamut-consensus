package fastexphase1

import (
	"github.com/prysmaticlabs/prysm/v3/beacon-chain/core/time"
	"github.com/prysmaticlabs/prysm/v3/beacon-chain/state"
	state_native "github.com/prysmaticlabs/prysm/v3/beacon-chain/state/state-native"
	"github.com/prysmaticlabs/prysm/v3/config/params"
	enginev1 "github.com/prysmaticlabs/prysm/v3/proto/engine/v1"
	ethpb "github.com/prysmaticlabs/prysm/v3/proto/prysm/v1alpha1"
)

// UpgradeToFastexPhase1 updates a generic state to return the version FastexPhase1 state.
func UpgradeToFastexPhase1(state state.BeaconState) (state.BeaconState, error) {
	epoch := time.CurrentEpoch(state)

	currentSyncCommittee, err := state.CurrentSyncCommittee()
	if err != nil {
		return nil, err
	}
	nextSyncCommittee, err := state.NextSyncCommittee()
	if err != nil {
		return nil, err
	}
	prevEpochParticipation, err := state.PreviousEpochParticipation()
	if err != nil {
		return nil, err
	}
	currentEpochParticipation, err := state.CurrentEpochParticipation()
	if err != nil {
		return nil, err
	}
	inactivityScores, err := state.InactivityScores()
	if err != nil {
		return nil, err
	}
	payloadHeader, err := state.LatestExecutionPayloadHeader()
	if err != nil {
		return nil, err
	}
	txRoot, err := payloadHeader.TransactionsRoot()
	if err != nil {
		return nil, err
	}
	hrs, err := state.HistoricalRoots()
	if err != nil {
		return nil, err
	}

	s := &ethpb.BeaconStateFastexPhase1{
		GenesisTime:           state.GenesisTime(),
		GenesisValidatorsRoot: state.GenesisValidatorsRoot(),
		Slot:                  state.Slot(),
		Fork: &ethpb.Fork{
			PreviousVersion: state.Fork().CurrentVersion,
			CurrentVersion:  params.BeaconConfig().FastexPhase1ForkVersion,
			Epoch:           epoch,
		},
		LatestBlockHeader:              state.LatestBlockHeader(),
		BlockRoots:                     state.BlockRoots(),
		StateRoots:                     state.StateRoots(),
		HistoricalRoots:                hrs,
		Eth1Data:                       state.Eth1Data(),
		Eth1DataVotes:                  state.Eth1DataVotes(),
		Eth1DepositIndex:               state.Eth1DepositIndex(),
		LatestProcessedBlockActivities: state.LatestProcessedBlockActivities(),
		TransactionsGasPerPeriod:       state.TransactionsGasPerPeriod(),
		TransactionsPerLatestEpoch:     state.TransactionsPerLatestEpoch(),
		NonStakersGasPerPeriod:         state.NonStakersGasPerPeriod(),
		NonStakersGasPerEpoch:          state.NonStakersGasPerEpoch(),
		Validators:                     state.Validators(),
		Balances:                       state.Balances(),
		Contracts:                      state.Contracts(),
		Activities:                     state.Activities(),
		RandaoMixes:                    state.RandaoMixes(),
		Slashings:                      state.Slashings(),
		PreviousEpochParticipation:     prevEpochParticipation,
		CurrentEpochParticipation:      currentEpochParticipation,
		JustificationBits:              state.JustificationBits(),
		PreviousJustifiedCheckpoint:    state.PreviousJustifiedCheckpoint(),
		CurrentJustifiedCheckpoint:     state.CurrentJustifiedCheckpoint(),
		FinalizedCheckpoint:            state.FinalizedCheckpoint(),
		InactivityScores:               inactivityScores,
		CurrentSyncCommittee:           currentSyncCommittee,
		NextSyncCommittee:              nextSyncCommittee,
		LatestExecutionPayloadHeader: &enginev1.ExecutionPayloadHeader{
			ParentHash:       payloadHeader.ParentHash(),
			FeeRecipient:     payloadHeader.FeeRecipient(),
			StateRoot:        payloadHeader.StateRoot(),
			ReceiptsRoot:     payloadHeader.ReceiptsRoot(),
			LogsBloom:        payloadHeader.LogsBloom(),
			PrevRandao:       payloadHeader.PrevRandao(),
			BlockNumber:      payloadHeader.BlockNumber(),
			GasLimit:         payloadHeader.GasLimit(),
			GasUsed:          payloadHeader.GasUsed(),
			Timestamp:        payloadHeader.Timestamp(),
			ExtraData:        payloadHeader.ExtraData(),
			BaseFeePerGas:    payloadHeader.BaseFeePerGas(),
			BlockHash:        payloadHeader.BlockHash(),
			TransactionsRoot: txRoot,
		},
		BaseFeePerEpoch:  0,
		BaseFeePerPeriod: 0,
	}

	return state_native.InitializeFromProtoUnsafeFastexPhase1(s)
}
