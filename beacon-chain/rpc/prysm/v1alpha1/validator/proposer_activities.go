package validator

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/prysmaticlabs/prysm/v4/beacon-chain/core/blocks"
	"github.com/prysmaticlabs/prysm/v4/beacon-chain/state"
	"github.com/prysmaticlabs/prysm/v4/consensus-types/interfaces"
)

// Sets the activity changes, transactions count, base fee and execution height for the block.
// Activity changes come from EL client.
func (vs *Server) setActivities(
	ctx context.Context,
	blk interfaces.SignedBeaconBlock,
	beaconState state.BeaconState,
) error {
	mergeComplete, err := blocks.IsMergeTransitionComplete(beaconState)
	if err != nil {
		return err
	}

	if !mergeComplete {
		return nil
	}

	latestExecutionHeader, err := beaconState.LatestExecutionPayloadHeader()
	if err != nil {
		return err
	}

	blockHash := common.BytesToHash(latestExecutionHeader.BlockHash())

	blockActivities, err := vs.ExecutionEngineCaller.GetBlockActivitiesByHash(ctx, blockHash)
	if err != nil {
		return errors.Wrap(err, "could not get block activities from execution layer")
	}

	blk.SetActivityChanges(blockActivities.Activities)
	blk.SetTransactionsCount(blockActivities.TxCount)
	blk.SetBaseFee(blockActivities.BaseFee)
	blk.SetExecutionHeight(latestExecutionHeader.BlockNumber())

	return nil
}
