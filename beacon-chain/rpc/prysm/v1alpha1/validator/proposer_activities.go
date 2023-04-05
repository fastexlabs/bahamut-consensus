package validator

import (
	"context"
	"math/big"

	"github.com/prysmaticlabs/prysm/v3/beacon-chain/state"
	ethpb "github.com/prysmaticlabs/prysm/v3/proto/prysm/v1alpha1"
	"go.opencensus.io/trace"
)

// getActivityChanges returns a list of activityChanges that are ready for inclusion in the next beacon block.
func (vs *Server) getActivityChanges(ctx context.Context, beaconState state.BeaconState) ([]*ethpb.ActivityChange, uint64, uint64, error) {
	ctx, span := trace.StartSpan(ctx, "ProposerServer.getActivityChanges")
	defer span.End()

	if !vs.Eth1InfoFetcher.ExecutionClientConnected() {
		log.Warn("not connected to eth1 node, skip activity changes insertion")
		return []*ethpb.ActivityChange{}, 0, 0, nil
	}

	var activityChanges []*ethpb.ActivityChange
	var transactionsCount uint64

	followedBlockHeight, err := vs.FollowedHeightFetcher.FollowedBlockHeight(ctx)
	if err != nil {
		return []*ethpb.ActivityChange{}, 0, 0, err
	}

	latestProcessedBlock := beaconState.LatestProcessedBlockActivities()

	for i := latestProcessedBlock + 1; i <= followedBlockHeight; i++ {
		ac := vs.ActivityChangeFetcher.GetActivityChanges(ctx, big.NewInt(0).SetUint64(i))
		activityChanges = append(activityChanges, ac...)
		transactionsCount += vs.ActivityChangeFetcher.GetTxCount(ctx, big.NewInt(0).SetUint64(i))
	}

	if len(activityChanges) == 0  && transactionsCount == 0 {
		log.Debug("no activity changes for inclusion in block")
		return []*ethpb.ActivityChange{}, 0, followedBlockHeight, nil
	}

	return activityChanges, transactionsCount, followedBlockHeight, nil
}
