package validator

import (
	"context"
	"math/big"

	"github.com/prysmaticlabs/prysm/v3/beacon-chain/state"
	"go.opencensus.io/trace"
)

func (vs *Server) getBaseFee(ctx context.Context, beaconState state.BeaconState) (uint64, error) {
	ctx, span := trace.StartSpan(ctx, "ProposerServer.getBaseFee")
	defer span.End()

	if !vs.Eth1InfoFetcher.ExecutionClientConnected() {
		log.Warn("not connected to eth1 node, skip base fee insertion")
		return  0, nil
	}

	var baseFee uint64

	followedBlockHeight, err := vs.FollowedHeightFetcher.FollowedBlockHeight(ctx)
	if err != nil {
		return 0, err
	}

	latestProcessedBlock := beaconState.LatestProcessedBlockActivities()

	for i := latestProcessedBlock + 1; i <= followedBlockHeight; i++ {
		baseFee += vs.BaseFeeFetcher.GetBaseFee(ctx, big.NewInt(0).SetUint64(i))
	}

	return baseFee, nil
}
