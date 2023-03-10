package execution

import (
	"context"
	"fmt"
	"math/big"

	ethpb "github.com/prysmaticlabs/prysm/v3/proto/prysm/v1alpha1"
	"github.com/sirupsen/logrus"
	"go.opencensus.io/trace"
)

var GetBlockActivitiesMethod string = "eth_getBlockActivities"

type BlockActivitiesResult struct {
	TxCount         uint64                  `json:"txCount"`
	ActivityChanges []*ethpb.ActivityChange `json:"activities"`
}

// GetBlockActivities from execution layer from particular block.
func (s *Service) GetBlockActivities(ctx context.Context, blkNum *big.Int) ([]*ethpb.ActivityChange, uint64, error) {
	ctx, span := trace.StartSpan(ctx, "powchain.engine-api-client.GetBlockActivities")
	defer span.End()
	var result BlockActivitiesResult

	if err := s.rpcClient.CallContext(
		ctx,
		&result,
		GetBlockActivitiesMethod,
		fmt.Sprintf("0x%x", blkNum.Uint64()),
	); err != nil {
		return nil, 0, handleRPCError(err)
	}

	return result.ActivityChanges, result.TxCount, nil
}

// ProcessBlockActivities insert activity change data into cache.
func (s *Service) ProcessBlockActivities(ctx context.Context, blkNum *big.Int) error {
	activityChanges, txCount, err := s.GetBlockActivities(ctx, blkNum)
	if err != nil {
		return err
	}

	if err := s.cfg.activityChangesCache.InsertTxCount(ctx, blkNum, txCount); err != nil {
		return err
	}

	if activityChanges != nil {
		if err := s.cfg.activityChangesCache.InsertActivityChanges(ctx, blkNum, activityChanges); err != nil {
			return err
		}
	}

	log.WithFields(logrus.Fields{
		"block":          blkNum.Uint64(),
		"activiyChanges": len(activityChanges),
		"txCount":        txCount,
	}).Info("Processing activity changes from execution layer")

	return nil
}
