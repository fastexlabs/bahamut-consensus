package basefeecache

import (
	"context"
	"math/big"
	"sync"

	"go.opencensus.io/trace"
)

type BaseFeeFetcher interface {
	GetBaseFee(ctx context.Context, blkNum *big.Int) uint64
}

type BaseFeeCache struct {
	baseFee map[uint64]uint64
	baseFeeLock sync.RWMutex
}

func New() *BaseFeeCache {
	return &BaseFeeCache{
		baseFee:     make(map[uint64]uint64),
	}
}

func (bc *BaseFeeCache) InsertBaseFee(ctx context.Context, blkNum *big.Int, baseFee uint64) error {
	ctx, span := trace.StartSpan(ctx, "BaseFeeCache.InsertBaseFee")
	defer span.End()

	bc.baseFeeLock.Lock()
	defer bc.baseFeeLock.Unlock()

	bc.baseFee[blkNum.Uint64()] = baseFee

	return nil
}

func (bc *BaseFeeCache) GetBaseFee(ctx context.Context, blkNum *big.Int) uint64 {
	ctx, span := trace.StartSpan(ctx, "BaseFeeCache.GetBaseFee")
	defer span.End()

	bc.baseFeeLock.Lock()
	defer bc.baseFeeLock.Unlock()

	baseFee, ok := bc.baseFee[blkNum.Uint64()]
	if !ok {
		return 0
	}

	return baseFee
}
