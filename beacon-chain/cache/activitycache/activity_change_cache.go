package activitycache

import (
	"context"
	"errors"
	"math/big"
	"sync"

	ethpb "github.com/prysmaticlabs/prysm/v3/proto/prysm/v1alpha1"
	"go.opencensus.io/trace"
)

// ActivityChangesFetcher defines a struct which can retreive information about activity changes from a store.
type ActivityChangesFetcher interface {
	GetActivityChanges(ctx context.Context, blkNum *big.Int) []*ethpb.ActivityChange
	GetTxCount(ctx context.Context, blkNum *big.Int) uint64
}

// New instantiates a new ActivityChanges cache
func New() *ActivityChangesCache {
	return &ActivityChangesCache{
		activityChanges: make(map[uint64][]*ethpb.ActivityChange),
		txs:             make(map[uint64]uint64),
	}
}

// ActivityChangesCache stores all in-memory activity change objects.
type ActivityChangesCache struct {
	activityChanges     map[uint64][]*ethpb.ActivityChange
	txs                 map[uint64]uint64
	activityChangesLock sync.RWMutex
}

// InsertActivityChanges into the database. If activity changes is a nil slice
// then this method does nothing.
func (acc *ActivityChangesCache) InsertActivityChanges(ctx context.Context, blkNum *big.Int, activityChanges []*ethpb.ActivityChange) error {
	ctx, span := trace.StartSpan(ctx, "ActivityChangesCache.InsertActivityChanges")
	defer span.End()

	if activityChanges == nil {
		log.Warn("Ignoring nil activity changes insertion")
		return errors.New("nil activity changes inserted into the cache")
	}

	acc.activityChangesLock.Lock()
	defer acc.activityChangesLock.Unlock()

	acc.activityChanges[blkNum.Uint64()] = activityChanges

	return nil
}

// InsertTxCount into the database.
func (acc *ActivityChangesCache) InsertTxCount(ctx context.Context, blkNum *big.Int, txCount uint64) error {
	ctx, span := trace.StartSpan(ctx, "ActivityChangesCache.InsertTxCount")
	defer span.End()

	acc.activityChangesLock.Lock()
	defer acc.activityChangesLock.Unlock()

	acc.txs[blkNum.Uint64()] = txCount

	return nil
}

// GetActivityChanges associated with particular block number.
func (acc *ActivityChangesCache) GetActivityChanges(ctx context.Context, blkNum *big.Int) []*ethpb.ActivityChange {
	ctx, span := trace.StartSpan(ctx, "ActivityChangesCache.GetActivityChanges")
	defer span.End()

	acc.activityChangesLock.Lock()
	defer acc.activityChangesLock.Unlock()

	activityChanges, ok := acc.activityChanges[blkNum.Uint64()]
	if !ok {
		return []*ethpb.ActivityChange{}
	}

	return activityChanges
}

// GetTxCount associated with particular block number.
func (acc *ActivityChangesCache) GetTxCount(ctx context.Context, blkNum *big.Int) uint64 {
	ctx, span := trace.StartSpan(ctx, "ActivityChangesCache.GetTxCount")
	defer span.End()

	acc.activityChangesLock.Lock()
	defer acc.activityChangesLock.Unlock()

	txCount, ok := acc.txs[blkNum.Uint64()]
	if !ok {
		return 0
	}

	return txCount
}
