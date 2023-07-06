package blocks_test

import (
	"context"
	"testing"

	"github.com/prysmaticlabs/prysm/v4/beacon-chain/core/blocks"
	"github.com/prysmaticlabs/prysm/v4/beacon-chain/state"
	state_native "github.com/prysmaticlabs/prysm/v4/beacon-chain/state/state-native"
	"github.com/prysmaticlabs/prysm/v4/config/params"
	"github.com/prysmaticlabs/prysm/v4/consensus-types/primitives"
	ethpb "github.com/prysmaticlabs/prysm/v4/proto/prysm/v1alpha1"
	"github.com/prysmaticlabs/prysm/v4/testing/assert"
	"github.com/prysmaticlabs/prysm/v4/testing/require"
	"github.com/prysmaticlabs/prysm/v4/testing/util"
)

func TestProcessActivityChange_NoAcvitveValidators(t *testing.T) {
	registry := []*ethpb.Validator{
		{
			PublicKey: []byte{1}, Contract: []byte{1, 1, 1},
			ActivationEpoch: 42,
			ExitEpoch:       params.BeaconConfig().FarFutureEpoch,
		},
		{
			PublicKey:       []byte{2},
			Contract:        []byte{2, 2, 2},
			ActivationEpoch: 42,
			ExitEpoch:       params.BeaconConfig().FarFutureEpoch,
		},
	}
	activityChanges := []*ethpb.ActivityChange{
		{
			ContractAddress: []byte{1, 1, 1},
			DeltaActivity:   42,
		},
		{
			ContractAddress: []byte{2, 2, 2},
			DeltaActivity:   4242,
		},
		{
			ContractAddress: []byte{1, 1, 1},
			DeltaActivity:   4200,
		},
		{
			ContractAddress: []byte{3, 3, 3},
			DeltaActivity:   80000,
		},
	}
	activities := []uint64{0, 0}
	st, err := state_native.InitializeFromProtoPhase0(&ethpb.BeaconState{
		Validators: registry,
		Activities: activities,
		Fork: &ethpb.Fork{
			CurrentVersion:  params.BeaconConfig().GenesisForkVersion,
			PreviousVersion: params.BeaconConfig().GenesisForkVersion,
		},
		Slot: 0,
	})
	require.NoError(t, err)
	b := util.NewBeaconBlock()
	b.Block = &ethpb.BeaconBlock{
		Body: &ethpb.BeaconBlockBody{
			ActivityChanges: activityChanges,
		},
	}
	newSt, err := blocks.ProcessActivityChanges(context.Background(), st, b.GetBlock().Body.GetActivityChanges())
	require.NoError(t, err)
	var activity uint64
	activity, err = newSt.ActivityAtIndex(primitives.ValidatorIndex(0))
	require.NoError(t, err)
	assert.Equal(t, uint64(0), activity)
	activity, err = newSt.ActivityAtIndex(primitives.ValidatorIndex(1))
	require.NoError(t, err)
	assert.Equal(t, uint64(0), activity)
}

func TestProcessActivityChanges(t *testing.T) {
	registry := []*ethpb.Validator{
		{
			PublicKey:       []byte{1},
			Contract:        []byte{1, 1, 1},
			ActivationEpoch: 0,
			ExitEpoch:       params.BeaconConfig().FarFutureEpoch,
		},
		{
			PublicKey:       []byte{2},
			Contract:        []byte{2, 2, 2},
			ActivationEpoch: 0,
			ExitEpoch:       params.BeaconConfig().FarFutureEpoch,
		},
	}
	activities := []uint64{0, 0}
	st, err := state_native.InitializeFromProtoPhase0(&ethpb.BeaconState{
		Validators: registry,
		Activities: activities,
		Fork: &ethpb.Fork{
			CurrentVersion:  params.BeaconConfig().GenesisForkVersion,
			PreviousVersion: params.BeaconConfig().GenesisForkVersion,
		},
	})
	require.NoError(t, err)

	reset := func(state state.BeaconState) error {
		if err := state.SetActivities([]uint64{0, 0}); err != nil {
			return err
		}
		return nil
	}

	tests := []struct {
		name    string
		setup   func() []*ethpb.ActivityChange
		check   func(*testing.T, state.BeaconState)
		wantErr string
	}{
		{
			name: "Ok",
			setup: func() []*ethpb.ActivityChange {
				return []*ethpb.ActivityChange{
					{
						ContractAddress: []byte{1, 1, 1},
						DeltaActivity:   42,
					},
					{
						ContractAddress: []byte{2, 2, 2},
						DeltaActivity:   4242,
					},
					{
						ContractAddress: []byte{1, 1, 1},
						DeltaActivity:   4200,
					},
					{
						ContractAddress: []byte{3, 3, 3},
						DeltaActivity:   80000,
					},
				}
			},
			check: func(t *testing.T, state state.BeaconState) {
				var activity uint64
				var err error
				activity, err = state.ActivityAtIndex(primitives.ValidatorIndex(0))
				require.NoError(t, err)
				assert.Equal(t, uint64(4242), activity)
				activity, err = state.ActivityAtIndex(primitives.ValidatorIndex(1))
				require.NoError(t, err)
				assert.Equal(t, uint64(4242), activity)
			},
			wantErr: "",
		},
		{
			name: "Nil activity change",
			setup: func() []*ethpb.ActivityChange {
				return []*ethpb.ActivityChange{
					{
						ContractAddress: []byte{1, 1, 1},
						DeltaActivity:   42,
					},
					{
						ContractAddress: []byte{2, 2, 2},
						DeltaActivity:   4242,
					},
					nil,
					{
						ContractAddress: []byte{1, 1, 1},
						DeltaActivity:   4200,
					},
					{
						ContractAddress: []byte{3, 3, 3},
						DeltaActivity:   80000,
					},
				}
			},
			wantErr: "got a nil activity change in block",
		},
		{
			name: "No changes",
			setup: func() []*ethpb.ActivityChange {
				return []*ethpb.ActivityChange{
					{
						ContractAddress: []byte{3, 3, 3},
						DeltaActivity:   4200,
					},
				}
			},
			check: func(t *testing.T, state state.BeaconState) {
				var activity uint64
				var err error
				activity, err = state.ActivityAtIndex(primitives.ValidatorIndex(0))
				require.NoError(t, err)
				assert.Equal(t, uint64(0), activity)
				activity, err = state.ActivityAtIndex(primitives.ValidatorIndex(1))
				require.NoError(t, err)
				assert.Equal(t, uint64(0), activity)
			},
			wantErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := reset(st)
			require.NoError(t, err)
			b := util.NewBeaconBlock()
			b.Block = &ethpb.BeaconBlock{
				Body: &ethpb.BeaconBlockBody{
					ActivityChanges: tt.setup(),
				},
			}
			newSt, err := blocks.ProcessActivityChanges(context.Background(), st, b.GetBlock().Body.GetActivityChanges())
			if tt.wantErr == "" {
				require.NoError(t, err)
				tt.check(t, newSt)
			} else {
				require.ErrorContains(t, tt.wantErr, err)
			}
		})
	}
}

func TestProcessTransactionsCount(t *testing.T) {
	st, err := state_native.InitializeFromProtoPhase0(&ethpb.BeaconState{
		Fork: &ethpb.Fork{
			CurrentVersion:  params.BeaconConfig().GenesisForkVersion,
			PreviousVersion: params.BeaconConfig().GenesisForkVersion,
		},
		SharedActivity: &ethpb.SharedActivity{},
	})
	require.NoError(t, err)

	for i := primitives.Slot(0); i < params.BeaconConfig().SlotsPerEpoch; i++ {
		st, err = blocks.ProcessTransactionsCount(context.Background(), st, 10)
		assert.Equal(t, 10*uint64(i+1)*21000, st.SharedActivity().TransactionsGasPerEpoch)
		require.NoError(t, err)
	}
	assert.Equal(t, 10*uint64(params.BeaconConfig().SlotsPerEpoch)*21000, st.SharedActivity().TransactionsGasPerEpoch)
}

func TestProcessBaseFee(t *testing.T) {
	st, err := state_native.InitializeFromProtoPhase0(&ethpb.BeaconState{
		Fork: &ethpb.Fork{
			CurrentVersion:  params.BeaconConfig().GenesisForkVersion,
			PreviousVersion: params.BeaconConfig().GenesisForkVersion,
		},
		SharedActivity: &ethpb.SharedActivity{},
	})
	require.NoError(t, err)

	for i := primitives.Slot(0); i < params.BeaconConfig().SlotsPerEpoch; i++ {
		st, err = blocks.ProcessBaseFee(context.Background(), st, 10)
		assert.Equal(t, 10*uint64(i+1), st.SharedActivity().BaseFeePerEpoch)
		require.NoError(t, err)
	}
	assert.Equal(t, 10*uint64(params.BeaconConfig().SlotsPerEpoch), st.SharedActivity().BaseFeePerEpoch)
}
