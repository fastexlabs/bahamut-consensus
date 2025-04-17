// todo unit act
package altair_test

import (
	"context"
	"fmt"
	fuzz "github.com/google/gofuzz"
	"github.com/prysmaticlabs/go-bitfield"
	"github.com/prysmaticlabs/prysm/v4/beacon-chain/core/altair"
	"github.com/prysmaticlabs/prysm/v4/beacon-chain/core/helpers"
	"github.com/prysmaticlabs/prysm/v4/beacon-chain/core/signing"
	"github.com/prysmaticlabs/prysm/v4/beacon-chain/core/time"
	"github.com/prysmaticlabs/prysm/v4/beacon-chain/state"
	state_native "github.com/prysmaticlabs/prysm/v4/beacon-chain/state/state-native"
	fieldparams "github.com/prysmaticlabs/prysm/v4/config/fieldparams"
	"github.com/prysmaticlabs/prysm/v4/config/params"
	"github.com/prysmaticlabs/prysm/v4/consensus-types/blocks"
	"github.com/prysmaticlabs/prysm/v4/consensus-types/primitives"
	"github.com/prysmaticlabs/prysm/v4/container/slice"
	"github.com/prysmaticlabs/prysm/v4/crypto/bls"
	"github.com/prysmaticlabs/prysm/v4/math"
	ethpb "github.com/prysmaticlabs/prysm/v4/proto/prysm/v1alpha1"
	"github.com/prysmaticlabs/prysm/v4/proto/prysm/v1alpha1/attestation"
	"github.com/prysmaticlabs/prysm/v4/testing/require"
	"github.com/prysmaticlabs/prysm/v4/testing/util"
	"math/rand"
	"testing"
)

func TestProcessAttestations_InclusionDelayFailure(t *testing.T) {
	attestations := []*ethpb.Attestation{
		util.HydrateAttestation(&ethpb.Attestation{
			Data: &ethpb.AttestationData{
				Target: &ethpb.Checkpoint{Epoch: 0, Root: make([]byte, fieldparams.RootLength)},
				Slot:   5,
			},
		}),
	}
	b := util.NewBeaconBlockAltair()
	b.Block = &ethpb.BeaconBlockAltair{
		Body: &ethpb.BeaconBlockBodyAltair{
			Attestations: attestations,
		},
	}
	beaconState, _ := util.DeterministicGenesisStateAltair(t, 100)

	want := fmt.Sprintf(
		"attestation slot %d + inclusion delay %d > state slot %d",
		attestations[0].Data.Slot,
		params.BeaconConfig().MinAttestationInclusionDelay,
		beaconState.Slot(),
	)
	wsb, err := blocks.NewSignedBeaconBlock(b)
	require.NoError(t, err)
	_, err = altair.ProcessAttestationsNoVerifySignature(context.Background(), beaconState, wsb)
	require.ErrorContains(t, want, err)
}

func TestProcessAttestations_NeitherCurrentNorPrevEpoch(t *testing.T) {
	att := util.HydrateAttestation(&ethpb.Attestation{
		Data: &ethpb.AttestationData{
			Source: &ethpb.Checkpoint{Epoch: 0, Root: []byte("hello-world")},
			Target: &ethpb.Checkpoint{Epoch: 0},
		},
	})

	b := util.NewBeaconBlockAltair()
	b.Block = &ethpb.BeaconBlockAltair{
		Body: &ethpb.BeaconBlockBodyAltair{
			Attestations: []*ethpb.Attestation{att},
		},
	}
	beaconState, _ := util.DeterministicGenesisStateAltair(t, 100)
	err := beaconState.SetSlot(beaconState.Slot() + params.BeaconConfig().SlotsPerEpoch*4 + params.BeaconConfig().MinAttestationInclusionDelay)
	require.NoError(t, err)
	pfc := beaconState.PreviousJustifiedCheckpoint()
	pfc.Root = []byte("hello-world")
	require.NoError(t, beaconState.SetPreviousJustifiedCheckpoint(pfc))

	want := fmt.Sprintf(
		"expected target epoch (%d) to be the previous epoch (%d) or the current epoch (%d)",
		att.Data.Target.Epoch,
		time.PrevEpoch(beaconState),
		time.CurrentEpoch(beaconState),
	)
	wsb, err := blocks.NewSignedBeaconBlock(b)
	require.NoError(t, err)
	_, err = altair.ProcessAttestationsNoVerifySignature(context.Background(), beaconState, wsb)
	require.ErrorContains(t, want, err)
}

func TestProcessAttestations_CurrentEpochFFGDataMismatches(t *testing.T) {
	attestations := []*ethpb.Attestation{
		{
			Data: &ethpb.AttestationData{
				Target: &ethpb.Checkpoint{Epoch: 0, Root: make([]byte, fieldparams.RootLength)},
				Source: &ethpb.Checkpoint{Epoch: 1, Root: make([]byte, fieldparams.RootLength)},
			},
			AggregationBits: bitfield.Bitlist{0x09},
		},
	}
	b := util.NewBeaconBlockAltair()
	b.Block = &ethpb.BeaconBlockAltair{
		Body: &ethpb.BeaconBlockBodyAltair{
			Attestations: attestations,
		},
	}
	beaconState, _ := util.DeterministicGenesisStateAltair(t, 100)
	require.NoError(t, beaconState.SetSlot(beaconState.Slot()+params.BeaconConfig().MinAttestationInclusionDelay))
	cfc := beaconState.CurrentJustifiedCheckpoint()
	cfc.Root = []byte("hello-world")
	require.NoError(t, beaconState.SetCurrentJustifiedCheckpoint(cfc))

	want := "source check point not equal to current justified checkpoint"
	wsb, err := blocks.NewSignedBeaconBlock(b)
	require.NoError(t, err)
	_, err = altair.ProcessAttestationsNoVerifySignature(context.Background(), beaconState, wsb)
	require.ErrorContains(t, want, err)
	b.Block.Body.Attestations[0].Data.Source.Epoch = time.CurrentEpoch(beaconState)
	b.Block.Body.Attestations[0].Data.Source.Root = []byte{}
	wsb, err = blocks.NewSignedBeaconBlock(b)
	require.NoError(t, err)
	_, err = altair.ProcessAttestationsNoVerifySignature(context.Background(), beaconState, wsb)
	require.ErrorContains(t, want, err)
}

func TestProcessAttestations_PrevEpochFFGDataMismatches(t *testing.T) {
	beaconState, _ := util.DeterministicGenesisStateAltair(t, 100)

	aggBits := bitfield.NewBitlist(3)
	aggBits.SetBitAt(0, true)
	attestations := []*ethpb.Attestation{
		{
			Data: &ethpb.AttestationData{
				Source: &ethpb.Checkpoint{Epoch: 1, Root: make([]byte, fieldparams.RootLength)},
				Target: &ethpb.Checkpoint{Epoch: 1, Root: make([]byte, fieldparams.RootLength)},
				Slot:   params.BeaconConfig().SlotsPerEpoch,
			},
			AggregationBits: aggBits,
		},
	}
	b := util.NewBeaconBlockAltair()
	b.Block = &ethpb.BeaconBlockAltair{
		Body: &ethpb.BeaconBlockBodyAltair{
			Attestations: attestations,
		},
	}

	err := beaconState.SetSlot(beaconState.Slot() + 2*params.BeaconConfig().SlotsPerEpoch)
	require.NoError(t, err)
	pfc := beaconState.PreviousJustifiedCheckpoint()
	pfc.Root = []byte("hello-world")
	require.NoError(t, beaconState.SetPreviousJustifiedCheckpoint(pfc))

	want := "source check point not equal to previous justified checkpoint"
	wsb, err := blocks.NewSignedBeaconBlock(b)
	require.NoError(t, err)
	_, err = altair.ProcessAttestationsNoVerifySignature(context.Background(), beaconState, wsb)
	require.ErrorContains(t, want, err)
	b.Block.Body.Attestations[0].Data.Source.Epoch = time.PrevEpoch(beaconState)
	b.Block.Body.Attestations[0].Data.Target.Epoch = time.PrevEpoch(beaconState)
	b.Block.Body.Attestations[0].Data.Source.Root = []byte{}
	wsb, err = blocks.NewSignedBeaconBlock(b)
	require.NoError(t, err)
	_, err = altair.ProcessAttestationsNoVerifySignature(context.Background(), beaconState, wsb)
	require.ErrorContains(t, want, err)
}

func TestProcessAttestations_InvalidAggregationBitsLength(t *testing.T) {
	beaconState, _ := util.DeterministicGenesisStateAltair(t, 100)

	aggBits := bitfield.NewBitlist(4)
	att := &ethpb.Attestation{
		Data: &ethpb.AttestationData{
			Source: &ethpb.Checkpoint{Epoch: 0, Root: []byte("hello-world")},
			Target: &ethpb.Checkpoint{Epoch: 0},
		},
		AggregationBits: aggBits,
	}

	b := util.NewBeaconBlockAltair()
	b.Block = &ethpb.BeaconBlockAltair{
		Body: &ethpb.BeaconBlockBodyAltair{
			Attestations: []*ethpb.Attestation{att},
		},
	}

	err := beaconState.SetSlot(beaconState.Slot() + params.BeaconConfig().MinAttestationInclusionDelay)
	require.NoError(t, err)

	cfc := beaconState.CurrentJustifiedCheckpoint()
	cfc.Root = []byte("hello-world")
	require.NoError(t, beaconState.SetCurrentJustifiedCheckpoint(cfc))

	expected := "failed to verify aggregation bitfield: wanted participants bitfield length 3, got: 4"
	wsb, err := blocks.NewSignedBeaconBlock(b)
	require.NoError(t, err)
	_, err = altair.ProcessAttestationsNoVerifySignature(context.Background(), beaconState, wsb)
	require.ErrorContains(t, expected, err)
}

func TestProcessAttestations_OK(t *testing.T) {
	beaconState, privKeys := util.DeterministicGenesisStateAltair(t, 100)

	aggBits := bitfield.NewBitlist(3)
	aggBits.SetBitAt(0, true)
	var mockRoot [32]byte
	copy(mockRoot[:], "hello-world")
	att := util.HydrateAttestation(&ethpb.Attestation{
		Data: &ethpb.AttestationData{
			Source: &ethpb.Checkpoint{Root: mockRoot[:]},
			Target: &ethpb.Checkpoint{Root: mockRoot[:]},
		},
		AggregationBits: aggBits,
	})

	cfc := beaconState.CurrentJustifiedCheckpoint()
	cfc.Root = mockRoot[:]
	require.NoError(t, beaconState.SetCurrentJustifiedCheckpoint(cfc))

	committee, err := helpers.BeaconCommitteeFromState(context.Background(), beaconState, att.Data.Slot, att.Data.CommitteeIndex)
	require.NoError(t, err)
	attestingIndices, err := attestation.AttestingIndices(att.AggregationBits, committee)
	require.NoError(t, err)
	sigs := make([]bls.Signature, len(attestingIndices))
	for i, indice := range attestingIndices {
		sb, err := signing.ComputeDomainAndSign(beaconState, 0, att.Data, params.BeaconConfig().DomainBeaconAttester, privKeys[indice])
		require.NoError(t, err)
		sig, err := bls.SignatureFromBytes(sb)
		require.NoError(t, err)
		sigs[i] = sig
	}
	att.Signature = bls.AggregateSignatures(sigs).Marshal()

	block := util.NewBeaconBlockAltair()
	block.Block.Body.Attestations = []*ethpb.Attestation{att}

	err = beaconState.SetSlot(beaconState.Slot() + params.BeaconConfig().MinAttestationInclusionDelay)
	require.NoError(t, err)
	wsb, err := blocks.NewSignedBeaconBlock(block)
	require.NoError(t, err)
	_, err = altair.ProcessAttestationsNoVerifySignature(context.Background(), beaconState, wsb)
	require.NoError(t, err)
}

func TestProcessAttestationNoVerify_SourceTargetHead(t *testing.T) {
	beaconState, _ := util.DeterministicGenesisStateAltair(t, 64)
	err := beaconState.SetSlot(beaconState.Slot() + params.BeaconConfig().MinAttestationInclusionDelay)
	require.NoError(t, err)

	aggBits := bitfield.NewBitlist(2)
	aggBits.SetBitAt(0, true)
	aggBits.SetBitAt(1, true)
	r, err := helpers.BlockRootAtSlot(beaconState, 0)
	require.NoError(t, err)
	att := &ethpb.Attestation{
		Data: &ethpb.AttestationData{
			BeaconBlockRoot: r,
			Source:          &ethpb.Checkpoint{Epoch: 0, Root: make([]byte, fieldparams.RootLength)},
			Target:          &ethpb.Checkpoint{Epoch: 0, Root: make([]byte, fieldparams.RootLength)},
		},
		AggregationBits: aggBits,
	}
	var zeroSig [96]byte
	att.Signature = zeroSig[:]

	ckp := beaconState.CurrentJustifiedCheckpoint()
	copy(ckp.Root, make([]byte, fieldparams.RootLength))
	require.NoError(t, beaconState.SetCurrentJustifiedCheckpoint(ckp))

	b, err := helpers.TotalActiveBalance(beaconState)
	require.NoError(t, err)
	beaconState, err = altair.ProcessAttestationNoVerifySignature(context.Background(), beaconState, att, b)
	require.NoError(t, err)

	p, err := beaconState.CurrentEpochParticipation()
	require.NoError(t, err)

	committee, err := helpers.BeaconCommitteeFromState(context.Background(), beaconState, att.Data.Slot, att.Data.CommitteeIndex)
	require.NoError(t, err)
	indices, err := attestation.AttestingIndices(att.AggregationBits, committee)
	require.NoError(t, err)
	for _, index := range indices {
		has, err := altair.HasValidatorFlag(p[index], params.BeaconConfig().TimelyHeadFlagIndex)
		require.NoError(t, err)
		require.Equal(t, true, has)
		has, err = altair.HasValidatorFlag(p[index], params.BeaconConfig().TimelySourceFlagIndex)
		require.NoError(t, err)
		require.Equal(t, true, has)
		has, err = altair.HasValidatorFlag(p[index], params.BeaconConfig().TimelyTargetFlagIndex)
		require.NoError(t, err)
		require.Equal(t, true, has)
	}
}

func TestValidatorFlag_Has(t *testing.T) {
	tests := []struct {
		name     string
		set      uint8
		expected []uint8
	}{
		{
			name:     "none",
			set:      0,
			expected: []uint8{},
		},
		{
			name:     "source",
			set:      1,
			expected: []uint8{params.BeaconConfig().TimelySourceFlagIndex},
		},
		{
			name:     "target",
			set:      2,
			expected: []uint8{params.BeaconConfig().TimelyTargetFlagIndex},
		},
		{
			name:     "head",
			set:      4,
			expected: []uint8{params.BeaconConfig().TimelyHeadFlagIndex},
		},
		{
			name:     "source, target",
			set:      3,
			expected: []uint8{params.BeaconConfig().TimelySourceFlagIndex, params.BeaconConfig().TimelyTargetFlagIndex},
		},
		{
			name:     "source, head",
			set:      5,
			expected: []uint8{params.BeaconConfig().TimelySourceFlagIndex, params.BeaconConfig().TimelyHeadFlagIndex},
		},
		{
			name:     "target, head",
			set:      6,
			expected: []uint8{params.BeaconConfig().TimelyTargetFlagIndex, params.BeaconConfig().TimelyTargetFlagIndex},
		},
		{
			name:     "source, target, head",
			set:      7,
			expected: []uint8{params.BeaconConfig().TimelySourceFlagIndex, params.BeaconConfig().TimelyTargetFlagIndex, params.BeaconConfig().TimelyHeadFlagIndex},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, f := range tt.expected {
				has, err := altair.HasValidatorFlag(tt.set, f)
				require.NoError(t, err)
				require.Equal(t, true, has)
			}
		})
	}
}

func TestValidatorFlag_Has_ExceedsLength(t *testing.T) {
	_, err := altair.HasValidatorFlag(0, 8)
	require.ErrorContains(t, "flag position exceeds length", err)
}

func TestValidatorFlag_Add(t *testing.T) {
	tests := []struct {
		name          string
		set           []uint8
		expectedTrue  []uint8
		expectedFalse []uint8
	}{
		{
			name:          "none",
			set:           []uint8{},
			expectedTrue:  []uint8{},
			expectedFalse: []uint8{params.BeaconConfig().TimelySourceFlagIndex, params.BeaconConfig().TimelyTargetFlagIndex, params.BeaconConfig().TimelyHeadFlagIndex},
		},
		{
			name:          "source",
			set:           []uint8{params.BeaconConfig().TimelySourceFlagIndex},
			expectedTrue:  []uint8{params.BeaconConfig().TimelySourceFlagIndex},
			expectedFalse: []uint8{params.BeaconConfig().TimelyTargetFlagIndex, params.BeaconConfig().TimelyHeadFlagIndex},
		},
		{
			name:          "source, target",
			set:           []uint8{params.BeaconConfig().TimelySourceFlagIndex, params.BeaconConfig().TimelyTargetFlagIndex},
			expectedTrue:  []uint8{params.BeaconConfig().TimelySourceFlagIndex, params.BeaconConfig().TimelyTargetFlagIndex},
			expectedFalse: []uint8{params.BeaconConfig().TimelyHeadFlagIndex},
		},
		{
			name:          "source, target, head",
			set:           []uint8{params.BeaconConfig().TimelySourceFlagIndex, params.BeaconConfig().TimelyTargetFlagIndex, params.BeaconConfig().TimelyHeadFlagIndex},
			expectedTrue:  []uint8{params.BeaconConfig().TimelySourceFlagIndex, params.BeaconConfig().TimelyTargetFlagIndex, params.BeaconConfig().TimelyHeadFlagIndex},
			expectedFalse: []uint8{},
		},
	}
	var err error
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := uint8(0)
			for _, f := range tt.set {
				b, err = altair.AddValidatorFlag(b, f)
				require.NoError(t, err)
			}
			for _, f := range tt.expectedFalse {
				has, err := altair.HasValidatorFlag(b, f)
				require.NoError(t, err)
				require.Equal(t, false, has)
			}
			for _, f := range tt.expectedTrue {
				has, err := altair.HasValidatorFlag(b, f)
				require.NoError(t, err)
				require.Equal(t, true, has)
			}
		})
	}
}

func TestValidatorFlag_Add_ExceedsLength(t *testing.T) {
	_, err := altair.AddValidatorFlag(0, 8)
	require.ErrorContains(t, "flag position exceeds length", err)
}

func TestFuzzProcessAttestationsNoVerify_10000(t *testing.T) {
	fuzzer := fuzz.NewWithSeed(0)
	st := &ethpb.BeaconStateAltair{}
	b := &ethpb.SignedBeaconBlockAltair{Block: &ethpb.BeaconBlockAltair{}}
	for i := 0; i < 10000; i++ {
		fuzzer.Fuzz(st)
		fuzzer.Fuzz(b)
		if b.Block == nil {
			b.Block = &ethpb.BeaconBlockAltair{}
		}
		s, err := state_native.InitializeFromProtoUnsafeAltair(st)
		require.NoError(t, err)
		if b.Block == nil || b.Block.Body == nil {
			continue
		}
		wsb, err := blocks.NewSignedBeaconBlock(b)
		require.NoError(t, err)
		r, err := altair.ProcessAttestationsNoVerifySignature(context.Background(), s, wsb)
		if err != nil && r != nil {
			t.Fatalf("return value should be nil on err. found: %v on error: %v for state: %v and block: %v", r, err, s, b)
		}
	}
}

// TODO(fastex): figure out how to test SetParticipationAndRewardProposer
func TestSetParticipationAndRewardProposer(t *testing.T) {
	cfg := params.BeaconConfig()
	sourceFlagIndex := cfg.TimelySourceFlagIndex
	targetFlagIndex := cfg.TimelyTargetFlagIndex
	headFlagIndex := cfg.TimelyHeadFlagIndex

	indices := func(size uint64) []uint64 {
		buf := make([]uint64, size)
		for i := uint64(0); i < size; i++ {
			buf[i] = i
		}
		return buf
	}
	validators := params.BeaconConfig().MaxValidatorsPerCommittee

	tests := []struct {
		name                string
		indices             []uint64
		epochParticipation  []byte
		participatedFlags   map[uint8]bool
		epoch               primitives.Epoch
		wantedBalance       uint64
		wantedParticipation []byte
	}{
		{
			name:               "none participated",
			indices:            indices(0),
			epochParticipation: make([]byte, validators),
			// epochParticipation: []byte{0, 0, 0, 0, 0, 0, 0, 0},
			participatedFlags: map[uint8]bool{
				sourceFlagIndex: false,
				targetFlagIndex: false,
				headFlagIndex:   false,
			},
			wantedParticipation: []byte{0, 0, 0, 0, 0, 0, 0, 0},
			wantedBalance:       8192000000000,
		},
		{
			name:               "some participated without flags",
			indices:            indices(validators / uint64(cfg.SlotsPerEpoch)),
			epochParticipation: make([]byte, validators),
			// epochParticipation: []byte{0, 0, 0, 0, 0, 0, 0, 0},
			participatedFlags: map[uint8]bool{
				sourceFlagIndex: false,
				targetFlagIndex: false,
				headFlagIndex:   false,
			},
			wantedParticipation: []byte{0, 0, 0, 0, 0, 0, 0, 0},
			wantedBalance:       8192000000000,
		},
		{
			name:               "some participated with some flags",
			indices:            indices(validators / uint64(cfg.SlotsPerEpoch)),
			epochParticipation: make([]byte, validators),
			// epochParticipation: []byte{0, 0, 0, 0, 0, 0, 0, 0},
			participatedFlags: map[uint8]bool{
				sourceFlagIndex: true,
				targetFlagIndex: true,
				headFlagIndex:   false,
			},
			wantedParticipation: []byte{3, 3, 3, 3, 0, 0, 0, 0},
			wantedBalance:       8192000000000 + 171226317,
		},
		{
			name:               "all participated with some flags (source)",
			indices:            indices(validators / uint64(cfg.SlotsPerEpoch)),
			epochParticipation: make([]byte, validators),
			// epochParticipation: []byte{0, 0, 0, 0, 0, 0, 0, 0},
			participatedFlags: map[uint8]bool{
				sourceFlagIndex: true,
				targetFlagIndex: false,
				headFlagIndex:   false,
			},
			wantedParticipation: []byte{1, 1, 1, 1, 1, 1, 1, 1},
			wantedBalance:       8192000000000 + 59929211,
		},
		{
			name:               "all participated with some flags (source, head)",
			indices:            indices(validators / uint64(cfg.SlotsPerEpoch)),
			epochParticipation: make([]byte, validators),
			// epochParticipation: []byte{0, 0, 0, 0, 0, 0, 0, 0},
			participatedFlags: map[uint8]bool{
				sourceFlagIndex: true,
				targetFlagIndex: false,
				headFlagIndex:   true,
			},
			wantedParticipation: []byte{1, 1, 1, 1, 1, 1, 1, 1},
			wantedBalance:       8192000000000 + 119858422,
		},
		{
			name:               "all participated with all flags",
			indices:            indices(validators / uint64(cfg.SlotsPerEpoch)),
			epochParticipation: make([]byte, validators),
			// epochParticipation: []byte{0, 0, 0, 0, 0, 0, 0, 0},
			participatedFlags: map[uint8]bool{
				sourceFlagIndex: true,
				targetFlagIndex: true,
				headFlagIndex:   true,
			},
			wantedParticipation: []byte{7, 7, 7, 7, 7, 7, 7, 7},
			wantedBalance:       8192000000000 + 231155529,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg := params.BeaconConfig()
			beaconState, _ := util.DeterministicGenesisStateAltair(t, validators)
			require.NoError(t, beaconState.SetSlot(params.BeaconConfig().SlotsPerEpoch))
			require.NoError(t, beaconState.SetSharedActivity(&ethpb.SharedActivity{
				TransactionsGasPerPeriod: 15_981_123 * uint64(cfg.EpochsPerActivityPeriod) * uint64(cfg.SlotsPerEpoch),
				BaseFeePerPeriod:         15 * uint64(cfg.EpochsPerActivityPeriod) * uint64(cfg.SlotsPerEpoch),
			}))

			currentEpoch := time.CurrentEpoch(beaconState)
			if test.epoch == currentEpoch {
				require.NoError(t, beaconState.SetCurrentParticipationBits(test.epochParticipation))
			} else {
				require.NoError(t, beaconState.SetPreviousParticipationBits(test.epochParticipation))
			}

			b, err := helpers.TotalActiveBalance(beaconState)
			require.NoError(t, err)
			committeeBalance, err := helpers.BeaconCommitteeBalanceFromState(context.Background(), beaconState, params.BeaconConfig().SlotsPerEpoch)
			require.NoError(t, err)
			st, err := altair.SetParticipationAndRewardProposer(context.Background(), beaconState, test.epoch, test.indices, test.participatedFlags, b, committeeBalance)
			require.NoError(t, err)

			i, err := helpers.BeaconProposerIndex(context.Background(), st)
			require.NoError(t, err)
			b, err = beaconState.BalanceAtIndex(i)
			require.NoError(t, err)
			require.Equal(t, test.wantedBalance, b)

			if test.epoch == currentEpoch {
				p, err := beaconState.CurrentEpochParticipation()
				require.NoError(t, err)
				require.DeepSSZEqual(t, test.wantedParticipation, p)
			} else {
				p, err := beaconState.PreviousEpochParticipation()
				require.NoError(t, err)
				require.DeepSSZEqual(t, test.wantedParticipation, p)
			}
		})
	}
}

// TODO(fastex): figure out how to test EpochParticipation
func TestEpochParticipation(t *testing.T) {
	beaconState, _ := util.DeterministicGenesisStateAltair(t, params.BeaconConfig().MaxValidatorsPerCommittee)
	cfg := params.BeaconConfig()
	sourceFlagIndex := cfg.TimelySourceFlagIndex
	targetFlagIndex := cfg.TimelyTargetFlagIndex
	headFlagIndex := cfg.TimelyHeadFlagIndex
	tests := []struct {
		name                     string
		indices                  []uint64
		epochParticipation       []byte
		participatedFlags        map[uint8]bool
		wantedNumerator          uint64
		wantedEpochParticipation []byte
	}{
		{
			name:    "none participated",
			indices: []uint64{}, epochParticipation: []byte{0, 0, 0, 0, 0, 0, 0, 0}, participatedFlags: map[uint8]bool{
				sourceFlagIndex: false,
				targetFlagIndex: false,
				headFlagIndex:   false,
			},
			wantedNumerator:          0,
			wantedEpochParticipation: []byte{0, 0, 0, 0, 0, 0, 0, 0},
		},
		{
			name:    "some participated without flags",
			indices: []uint64{0, 1, 2, 3}, epochParticipation: []byte{0, 0, 0, 0, 0, 0, 0, 0}, participatedFlags: map[uint8]bool{
				sourceFlagIndex: false,
				targetFlagIndex: false,
				headFlagIndex:   false,
			},
			wantedNumerator:          0,
			wantedEpochParticipation: []byte{0, 0, 0, 0, 0, 0, 0, 0},
		},
		{
			name:    "some participated with some flags",
			indices: []uint64{0, 1, 2, 3}, epochParticipation: []byte{0, 0, 0, 0, 0, 0, 0, 0}, participatedFlags: map[uint8]bool{
				sourceFlagIndex: true,
				targetFlagIndex: true,
				headFlagIndex:   false,
			},
			wantedNumerator:          40473600,
			wantedEpochParticipation: []byte{3, 3, 3, 3, 0, 0, 0, 0},
		},
		{
			name:    "all participated with some flags",
			indices: []uint64{0, 1, 2, 3, 4, 5, 6, 7}, epochParticipation: []byte{0, 0, 0, 0, 0, 0, 0, 0}, participatedFlags: map[uint8]bool{
				sourceFlagIndex: true,
				targetFlagIndex: false,
				headFlagIndex:   false,
			},
			wantedNumerator:          28331520,
			wantedEpochParticipation: []byte{1, 1, 1, 1, 1, 1, 1, 1},
		},
		{
			name:    "all participated with all flags",
			indices: []uint64{0, 1, 2, 3, 4, 5, 6, 7}, epochParticipation: []byte{0, 0, 0, 0, 0, 0, 0, 0}, participatedFlags: map[uint8]bool{
				sourceFlagIndex: true,
				targetFlagIndex: true,
				headFlagIndex:   true,
			},
			wantedNumerator:          109278720,
			wantedEpochParticipation: []byte{7, 7, 7, 7, 7, 7, 7, 7},
		},
	}
	for _, test := range tests {
		b, err := helpers.TotalActiveBalance(beaconState)
		require.NoError(t, err)
		n, p, err := altair.EpochParticipation(beaconState, test.indices, test.epochParticipation, test.participatedFlags, b)
		require.NoError(t, err)
		require.Equal(t, test.wantedNumerator, n)
		require.DeepSSZEqual(t, test.wantedEpochParticipation, p)
	}
}

// TODO(fastex): figure out how to test RewardProposer
func TestRewardProposer(t *testing.T) {
	beaconState, _ := util.DeterministicGenesisStateAltair(t, params.BeaconConfig().MaxValidatorsPerCommittee)
	require.NoError(t, beaconState.SetSlot(1))
	tests := []struct {
		rewardNumerator uint64
		want            uint64
	}{
		{rewardNumerator: 1, want: 32000000000},
		{rewardNumerator: 10000, want: 32000000022},
		{rewardNumerator: 1000000, want: 32000002254},
		{rewardNumerator: 1000000000, want: 32002234396},
		{rewardNumerator: 1000000000000, want: 34234377253},
	}
	for _, test := range tests {
		require.NoError(t, altair.RewardProposer(context.Background(), beaconState, test.rewardNumerator, test.rewardNumerator))
		i, err := helpers.BeaconProposerIndex(context.Background(), beaconState)
		require.NoError(t, err)
		b, err := beaconState.BalanceAtIndex(i)
		require.NoError(t, err)
		require.Equal(t, test.want, b)
	}
}

func TestAttestationParticipationFlagIndices(t *testing.T) {
	beaconState, _ := util.DeterministicGenesisStateAltair(t, params.BeaconConfig().MaxValidatorsPerCommittee)
	require.NoError(t, beaconState.SetSlot(1))
	cfg := params.BeaconConfig()
	sourceFlagIndex := cfg.TimelySourceFlagIndex
	targetFlagIndex := cfg.TimelyTargetFlagIndex
	headFlagIndex := cfg.TimelyHeadFlagIndex

	denebState, _ := util.DeterministicGenesisStateDeneb(t, params.BeaconConfig().MaxValidatorsPerCommittee)
	require.NoError(t, denebState.SetSlot(1))

	tests := []struct {
		name                 string
		inputState           state.BeaconState
		inputData            *ethpb.AttestationData
		inputDelay           primitives.Slot
		participationIndices map[uint8]bool
	}{
		{
			name: "none",
			inputState: func() state.BeaconState {
				return beaconState
			}(),
			inputData: &ethpb.AttestationData{
				Source: &ethpb.Checkpoint{Root: params.BeaconConfig().ZeroHash[:]},
				Target: &ethpb.Checkpoint{},
			},
			inputDelay:           params.BeaconConfig().SlotsPerEpoch,
			participationIndices: map[uint8]bool{},
		},
		{
			name: "participated source",
			inputState: func() state.BeaconState {
				return beaconState
			}(),
			inputData: &ethpb.AttestationData{
				Source: &ethpb.Checkpoint{Root: params.BeaconConfig().ZeroHash[:]},
				Target: &ethpb.Checkpoint{},
			},
			inputDelay: primitives.Slot(math.IntegerSquareRoot(uint64(cfg.SlotsPerEpoch)) - 1),
			participationIndices: map[uint8]bool{
				sourceFlagIndex: true,
			},
		},
		{
			name: "participated source and target",
			inputState: func() state.BeaconState {
				return beaconState
			}(),
			inputData: &ethpb.AttestationData{
				Source: &ethpb.Checkpoint{Root: params.BeaconConfig().ZeroHash[:]},
				Target: &ethpb.Checkpoint{Root: params.BeaconConfig().ZeroHash[:]},
			},
			inputDelay: primitives.Slot(math.IntegerSquareRoot(uint64(cfg.SlotsPerEpoch)) - 1),
			participationIndices: map[uint8]bool{
				sourceFlagIndex: true,
				targetFlagIndex: true,
			},
		},
		{
			name: "participated source and target with delay",
			inputState: func() state.BeaconState {
				return beaconState
			}(),
			inputData: &ethpb.AttestationData{
				Source: &ethpb.Checkpoint{Root: params.BeaconConfig().ZeroHash[:]},
				Target: &ethpb.Checkpoint{Root: params.BeaconConfig().ZeroHash[:]},
			},
			inputDelay: params.BeaconConfig().SlotsPerEpoch + 1,
			participationIndices: map[uint8]bool{
				targetFlagIndex: true,
			},
		},
		{
			name: "participated source and target with delay in deneb",
			inputState: func() state.BeaconState {
				return denebState
			}(),
			inputData: &ethpb.AttestationData{
				Source: &ethpb.Checkpoint{Root: params.BeaconConfig().ZeroHash[:]},
				Target: &ethpb.Checkpoint{Root: params.BeaconConfig().ZeroHash[:]},
			},
			inputDelay: params.BeaconConfig().SlotsPerEpoch + 1,
			participationIndices: map[uint8]bool{
				targetFlagIndex: true,
			},
		},
		{
			name: "participated source and target and head",
			inputState: func() state.BeaconState {
				return beaconState
			}(),
			inputData: &ethpb.AttestationData{
				BeaconBlockRoot: params.BeaconConfig().ZeroHash[:],
				Source:          &ethpb.Checkpoint{Root: params.BeaconConfig().ZeroHash[:]},
				Target:          &ethpb.Checkpoint{Root: params.BeaconConfig().ZeroHash[:]},
			},
			inputDelay: 1,
			participationIndices: map[uint8]bool{
				sourceFlagIndex: true,
				targetFlagIndex: true,
				headFlagIndex:   true,
			},
		},
	}
	for _, test := range tests {
		flagIndices, err := altair.AttestationParticipationFlagIndices(test.inputState, test.inputData, test.inputDelay)
		require.NoError(t, err)
		require.DeepEqual(t, test.participationIndices, flagIndices)
	}
}

func TestMatchingStatus(t *testing.T) {
	beaconState, _ := util.DeterministicGenesisStateAltair(t, params.BeaconConfig().MaxValidatorsPerCommittee)
	require.NoError(t, beaconState.SetSlot(1))
	tests := []struct {
		name          string
		inputState    state.BeaconState
		inputData     *ethpb.AttestationData
		inputCheckpt  *ethpb.Checkpoint
		matchedSource bool
		matchedTarget bool
		matchedHead   bool
	}{
		{
			name:       "non matched",
			inputState: beaconState,
			inputData: &ethpb.AttestationData{
				Source: &ethpb.Checkpoint{Epoch: 1},
				Target: &ethpb.Checkpoint{},
			},
			inputCheckpt: &ethpb.Checkpoint{},
		},
		{
			name:       "source matched",
			inputState: beaconState,
			inputData: &ethpb.AttestationData{
				Source: &ethpb.Checkpoint{},
				Target: &ethpb.Checkpoint{},
			},
			inputCheckpt:  &ethpb.Checkpoint{},
			matchedSource: true,
		},
		{
			name:       "target matched",
			inputState: beaconState,
			inputData: &ethpb.AttestationData{
				Source: &ethpb.Checkpoint{Epoch: 1},
				Target: &ethpb.Checkpoint{Root: params.BeaconConfig().ZeroHash[:]},
			},
			inputCheckpt:  &ethpb.Checkpoint{},
			matchedTarget: true,
		},
		{
			name:       "head matched",
			inputState: beaconState,
			inputData: &ethpb.AttestationData{
				Source:          &ethpb.Checkpoint{Epoch: 1},
				Target:          &ethpb.Checkpoint{},
				BeaconBlockRoot: params.BeaconConfig().ZeroHash[:],
			},
			inputCheckpt: &ethpb.Checkpoint{},
			matchedHead:  true,
		},
		{
			name:       "everything matched",
			inputState: beaconState,
			inputData: &ethpb.AttestationData{
				Source:          &ethpb.Checkpoint{},
				Target:          &ethpb.Checkpoint{Root: params.BeaconConfig().ZeroHash[:]},
				BeaconBlockRoot: params.BeaconConfig().ZeroHash[:],
			},
			inputCheckpt:  &ethpb.Checkpoint{},
			matchedSource: true,
			matchedTarget: true,
			matchedHead:   true,
		},
	}

	for _, test := range tests {
		src, tgt, head, err := altair.MatchingStatus(test.inputState, test.inputData, test.inputCheckpt)
		require.NoError(t, err)
		require.Equal(t, test.matchedSource, src)
		require.Equal(t, test.matchedTarget, tgt)
		require.Equal(t, test.matchedHead, head)
	}
}

func TestProcessAttestations_ProposerReward(t *testing.T) {
	prevConfig := params.BeaconConfig().Copy()
	defer params.SetActive(prevConfig)
	config := params.BeaconConfig().Copy()
	config.TargetCommitteeSize = 2
	params.SetActive(config)

	var (
		validators     uint64          = 64 * 2 // 2 validators per sub committee (2 sub committee per slot)
		lookAheadEpoch uint64          = uint64(params.BeaconConfig().EpochsPerHistoricalVector-params.BeaconConfig().MinSeedLookahead) - 1
		randaoMix      [32]byte        = [32]byte{42}
		slot           primitives.Slot = 3
		zeroSig        [96]byte
	)

	beaconState, _ := util.DeterministicGenesisStateAltair(t, validators)
	require.NoError(t, beaconState.SetSlot(slot+params.BeaconConfig().MinAttestationInclusionDelay))
	require.NoError(t, beaconState.SetSharedActivity(&ethpb.SharedActivity{
		TransactionsGasPerPeriod: 15_000_000 * uint64(params.BeaconConfig().SlotsPerEpoch) * uint64(params.BeaconConfig().EpochsPerActivityPeriod),
		BaseFeePerPeriod:         15 * uint64(params.BeaconConfig().SlotsPerEpoch) * uint64(params.BeaconConfig().EpochsPerActivityPeriod),
	}))

	require.NoError(t, beaconState.UpdateRandaoMixesAtIndex(lookAheadEpoch, randaoMix))

	src := rand.NewSource(42)
	rnd := rand.New(src)
	require.NoError(t, beaconState.ApplyToEveryValidator(func(idx int, val *ethpb.Validator) (bool, *ethpb.Validator, error) {
		balance := (4096 + uint64(rnd.Int63n(4097))) * params.BeaconConfig().EffectiveBalanceIncrement
		val.EffectiveBalance = balance
		return true, val, nil
	}))

	for i := primitives.Slot(0); i <= slot; i++ {
		require.NoError(t, beaconState.UpdateBlockRootAtIndex(uint64(i.ModSlot(params.BeaconConfig().SlotsPerHistoricalRoot)), [32]byte{byte(i)}))
	}

	var (
		aggBits bitfield.Bitlist
		r       []byte
		err     error
	)

	atts := make([]*ethpb.Attestation, 0, helpers.SlotCommitteeCount(validators)+2)

	// Previous attestation
	// TimelyHead = false
	// TimelySource = true
	// TimelyTarget = true
	aggBits = bitfield.NewBitlist(2)
	aggBits.SetBitAt(1, true)
	r, err = helpers.BlockRootAtSlot(beaconState, slot-1)
	require.NoError(t, err)
	atts = append(atts, &ethpb.Attestation{
		AggregationBits: aggBits,
		Data: &ethpb.AttestationData{
			Slot:            slot - 1,
			CommitteeIndex:  1,
			BeaconBlockRoot: r,
			Source:          &ethpb.Checkpoint{Epoch: 0, Root: make([]byte, fieldparams.RootLength)},
			Target:          &ethpb.Checkpoint{Epoch: 0, Root: make([]byte, fieldparams.RootLength)},
		},
		Signature: zeroSig[:],
	})

	// Current attestations (committee 0, full)
	// TimelyHead = true
	// TimelySource = true
	// TimelyTarget = true
	aggBits = bitfield.NewBitlist(2)
	aggBits.SetBitAt(0, true)
	aggBits.SetBitAt(1, true)
	r, err = helpers.BlockRootAtSlot(beaconState, slot)
	atts = append(atts, &ethpb.Attestation{
		AggregationBits: aggBits,
		Data: &ethpb.AttestationData{
			Slot:            slot,
			CommitteeIndex:  0,
			BeaconBlockRoot: r,
			Source:          &ethpb.Checkpoint{Epoch: 0, Root: make([]byte, fieldparams.RootLength)},
			Target:          &ethpb.Checkpoint{Epoch: 0, Root: make([]byte, fieldparams.RootLength)},
		},
		Signature: zeroSig[:],
	})

	// Current attestations (committee 1, partial)
	// TimelyHead = true
	// TimelySource = true
	// TimelyTarget = true
	aggBits = bitfield.NewBitlist(2)
	aggBits.SetBitAt(1, true)
	r, err = helpers.BlockRootAtSlot(beaconState, slot)
	atts = append(atts, &ethpb.Attestation{
		AggregationBits: aggBits,
		Data: &ethpb.AttestationData{
			Slot:            slot,
			CommitteeIndex:  1,
			BeaconBlockRoot: r,
			Source:          &ethpb.Checkpoint{Epoch: 0, Root: make([]byte, fieldparams.RootLength)},
			Target:          &ethpb.Checkpoint{Epoch: 0, Root: make([]byte, fieldparams.RootLength)},
		},
		Signature: zeroSig[:],
	})

	// Current attestations (committee 1, partial)
	// TimelyHead = false
	// TimelySource = true
	// TimelyTarget = true
	aggBits = bitfield.NewBitlist(2)
	aggBits.SetBitAt(0, true)
	atts = append(atts, &ethpb.Attestation{
		AggregationBits: aggBits,
		Data: &ethpb.AttestationData{
			Slot:           slot,
			CommitteeIndex: 1,
			Source:         &ethpb.Checkpoint{Epoch: 0, Root: make([]byte, fieldparams.RootLength)},
			Target:         &ethpb.Checkpoint{Epoch: 0, Root: make([]byte, fieldparams.RootLength)},
		},
		Signature: zeroSig[:],
	})

	ckp := beaconState.CurrentJustifiedCheckpoint()
	copy(ckp.Root, make([]byte, fieldparams.RootLength))
	require.NoError(t, beaconState.SetCurrentJustifiedCheckpoint(ckp))

	activeIndices, err := helpers.ActiveValidatorIndices(context.Background(), beaconState, 0)
	require.NoError(t, err)
	shuffledList := make([]primitives.ValidatorIndex, validators)
	copy(shuffledList, activeIndices)
	seed, err := helpers.Seed(beaconState, 0, params.BeaconConfig().DomainBeaconAttester)
	require.NoError(t, err)
	shuffledIndices, err := helpers.UnshuffleList(shuffledList, seed)
	require.NoError(t, err)

	totalBalance, err := helpers.TotalActiveBalance(beaconState)
	require.NoError(t, err)
	baseRewardPerIncerement, err := altair.BaseRewardPerIncrement(totalBalance)
	require.NoError(t, err)
	attestationsCoefficient := params.BeaconConfig().TimelyHeadWeight + params.BeaconConfig().TimelySourceWeight + params.BeaconConfig().TimelyTargetWeight

	var start, end, reward uint64
	denominators := make(map[primitives.Slot]uint64)

	start = slice.SplitOffset(validators, uint64(params.BeaconConfig().SlotsPerEpoch), uint64(slot-1))
	end = slice.SplitOffset(validators, uint64(params.BeaconConfig().SlotsPerEpoch), uint64(slot))
	reward = helpers.TotalBalance(beaconState, shuffledIndices[start:end]) / params.BeaconConfig().EffectiveBalanceIncrement * baseRewardPerIncerement
	denominators[slot-1] = reward * attestationsCoefficient

	start = slice.SplitOffset(validators, uint64(params.BeaconConfig().SlotsPerEpoch), uint64(slot))
	end = slice.SplitOffset(validators, uint64(params.BeaconConfig().SlotsPerEpoch), uint64(slot+1))
	reward = helpers.TotalBalance(beaconState, shuffledIndices[start:end]) / params.BeaconConfig().EffectiveBalanceIncrement * baseRewardPerIncerement
	denominators[slot] = reward * attestationsCoefficient

	totalPower, totalEffectivePower, err := helpers.Powers(context.Background(), beaconState)
	require.NoError(t, err)
	baseProposerReward, err := altair.BaseProposerReward(beaconState, totalPower, totalEffectivePower)
	baseProposerReward = baseProposerReward * (attestationsCoefficient) / params.BeaconConfig().WeightDenominator

	proposerIndex, err := helpers.BeaconProposerIndex(context.Background(), beaconState)
	require.NoError(t, err)
	proposerBalanceBefore, err := beaconState.BalanceAtIndex(proposerIndex)
	require.NoError(t, err)

	var earnedProposerReward uint64
	for _, att := range atts {
		b, err := helpers.TotalActiveBalance(beaconState)
		require.NoError(t, err)
		beaconState, err = altair.ProcessAttestationNoVerifySignature(context.Background(), beaconState, att, b)
		require.NoError(t, err)

		p, err := beaconState.CurrentEpochParticipation()
		require.NoError(t, err)
		committee, err := helpers.BeaconCommitteeFromState(context.Background(), beaconState, att.Data.Slot, att.Data.CommitteeIndex)
		require.NoError(t, err)
		indices, err := attestation.AttestingIndices(att.AggregationBits, committee)
		require.NoError(t, err)
		var numerator uint64
		for _, index := range indices {
			validator, err := beaconState.ValidatorAtIndexReadOnly(primitives.ValidatorIndex(index))
			require.NoError(t, err)
			increments := validator.EffectiveBalance() / params.BeaconConfig().EffectiveBalanceIncrement
			has, err := altair.HasValidatorFlag(p[index], params.BeaconConfig().TimelyHeadFlagIndex)
			require.NoError(t, err)
			if has {
				numerator += increments * baseRewardPerIncerement * params.BeaconConfig().TimelyHeadWeight
			}
			has, err = altair.HasValidatorFlag(p[index], params.BeaconConfig().TimelySourceFlagIndex)
			require.NoError(t, err)
			if has {
				numerator += increments * baseRewardPerIncerement * params.BeaconConfig().TimelySourceWeight
			}
			has, err = altair.HasValidatorFlag(p[index], params.BeaconConfig().TimelyTargetFlagIndex)
			require.NoError(t, err)
			if has {
				numerator += increments * baseRewardPerIncerement * params.BeaconConfig().TimelyTargetWeight
			}
		}
		earnedProposerReward += baseProposerReward * numerator / denominators[att.Data.Slot]
	}

	proposerBalanceAfter, err := beaconState.BalanceAtIndex(proposerIndex)
	require.NoError(t, err)

	require.Equal(t, proposerBalanceBefore+earnedProposerReward, proposerBalanceAfter)
}
