package validator

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/prysmaticlabs/prysm/v3/beacon-chain/state"
	consensusblocks "github.com/prysmaticlabs/prysm/v3/consensus-types/blocks"
	"github.com/prysmaticlabs/prysm/v3/consensus-types/interfaces"
	enginev1 "github.com/prysmaticlabs/prysm/v3/proto/engine/v1"
	ethpb "github.com/prysmaticlabs/prysm/v3/proto/prysm/v1alpha1"
	"github.com/prysmaticlabs/prysm/v3/runtime/version"
	"github.com/sirupsen/logrus"
)

// Sets the base fee for the block.
func (vs *Server) setBaseFee(ctx context.Context, blk interfaces.SignedBeaconBlock, headState state.BeaconState) error {
	if blk.Version() < version.FastexPhase1 {
		return nil
	}

	baseFee, err := vs.getBaseFee(ctx, headState)
	if err != nil {
		return err
	}

	if err := blk.SetBaseFee(baseFee); err != nil {
		return err
	}

	return nil
}

// This function retrieves the full payload block using the input blind block. This input must be versioned as
// fastex-phase1 blind block. The output block will contain the full payload. The original header block
// will be returned the block builder is not configured.
func (vs *Server) unblindBuilderBlockFastexPhase1(ctx context.Context, b interfaces.ReadOnlySignedBeaconBlock) (interfaces.ReadOnlySignedBeaconBlock, error) {
	if err := consensusblocks.BeaconBlockIsNil(b); err != nil {
		return nil, err
	}

	// No-op if the input block is not version blind and bellatrix.
	if b.Version() != version.FastexPhase1 || !b.IsBlinded() {
		return b, nil
	}
	// No-op nothing if the builder has not been configured.
	if !vs.BlockBuilder.Configured() {
		return b, nil
	}

	agg, err := b.Block().Body().SyncAggregate()
	if err != nil {
		return nil, err
	}
	h, err := b.Block().Body().Execution()
	if err != nil {
		return nil, err
	}
	header, ok := h.Proto().(*enginev1.ExecutionPayloadHeader)
	if !ok {
		return nil, errors.New("execution data must be execution payload header")
	}
	baseFee, err := b.Block().Body().BaseFee()
	if err != nil {
		return nil, err
	}

	parentRoot := b.Block().ParentRoot()
	stateRoot := b.Block().StateRoot()
	randaoReveal := b.Block().Body().RandaoReveal()
	graffiti := b.Block().Body().Graffiti()
	sig := b.Signature()
	psb := &ethpb.SignedBlindedBeaconBlockFastexPhase1{
		Block: &ethpb.BlindedBeaconBlockFastexPhase1{
			Slot:          b.Block().Slot(),
			ProposerIndex: b.Block().ProposerIndex(),
			ParentRoot:    parentRoot[:],
			StateRoot:     stateRoot[:],
			Body: &ethpb.BlindedBeaconBlockBodyFastexPhase1{
				RandaoReveal:           randaoReveal[:],
				Eth1Data:               b.Block().Body().Eth1Data(),
				Graffiti:               graffiti[:],
				ProposerSlashings:      b.Block().Body().ProposerSlashings(),
				AttesterSlashings:      b.Block().Body().AttesterSlashings(),
				Attestations:           b.Block().Body().Attestations(),
				Deposits:               b.Block().Body().Deposits(),
				ActivityChanges:        b.Block().Body().ActivityChanges(),
				LatestProcessedBlock:   b.Block().Body().LatestProcessedBlock(),
				TransactionsCount:      b.Block().Body().TransactionsCount(),
				VoluntaryExits:         b.Block().Body().VoluntaryExits(),
				SyncAggregate:          agg,
				ExecutionPayloadHeader: header,
				BaseFee:                baseFee,
			},
		},
		Signature: sig[:],
	}

	sb, err := consensusblocks.NewSignedBeaconBlock(psb)
	if err != nil {
		return nil, errors.Wrap(err, "could not create signed block")
	}
	payload, err := vs.BlockBuilder.SubmitBlindedBlock(ctx, sb)
	if err != nil {
		return nil, err
	}
	headerRoot, err := header.HashTreeRoot()
	if err != nil {
		return nil, err
	}

	payloadRoot, err := payload.HashTreeRoot()
	if err != nil {
		return nil, err
	}
	if headerRoot != payloadRoot {
		return nil, fmt.Errorf("header and payload root do not match, consider disconnect from relay to avoid further issues, "+
			"%#x != %#x", headerRoot, payloadRoot)
	}

	pbPayload, err := payload.PbBellatrix()
	if err != nil {
		return nil, errors.Wrap(err, "could not get payload")
	}
	bb := &ethpb.SignedBeaconBlockFastexPhase1{
		Block: &ethpb.BeaconBlockFastexPhase1{
			Slot:          psb.Block.Slot,
			ProposerIndex: psb.Block.ProposerIndex,
			ParentRoot:    psb.Block.ParentRoot,
			StateRoot:     psb.Block.StateRoot,
			Body: &ethpb.BeaconBlockBodyFastexPhase1{
				RandaoReveal:         psb.Block.Body.RandaoReveal,
				Eth1Data:             psb.Block.Body.Eth1Data,
				Graffiti:             psb.Block.Body.Graffiti,
				ProposerSlashings:    psb.Block.Body.ProposerSlashings,
				AttesterSlashings:    psb.Block.Body.AttesterSlashings,
				Attestations:         psb.Block.Body.Attestations,
				Deposits:             psb.Block.Body.Deposits,
				VoluntaryExits:       psb.Block.Body.VoluntaryExits,
				ActivityChanges:      psb.Block.Body.ActivityChanges,
				LatestProcessedBlock: psb.Block.Body.LatestProcessedBlock,
				TransactionsCount:    psb.Block.Body.TransactionsCount,
				SyncAggregate:        agg,
				ExecutionPayload:     pbPayload,
				BaseFee:              psb.Block.Body.BaseFee,
			},
		},
		Signature: psb.Signature,
	}
	wb, err := consensusblocks.NewSignedBeaconBlock(bb)
	if err != nil {
		return nil, err
	}

	txs, err := payload.Transactions()
	if err != nil {
		return nil, errors.Wrap(err, "could not get transactions from payload")
	}
	log.WithFields(logrus.Fields{
		"blockHash":    fmt.Sprintf("%#x", h.BlockHash()),
		"feeRecipient": fmt.Sprintf("%#x", h.FeeRecipient()),
		"gasUsed":      h.GasUsed,
		"slot":         b.Block().Slot(),
		"txs":          len(txs),
	}).Info("Retrieved full payload from builder")

	return wb, nil
}
