package stateutil

import (
	"encoding/binary"

	"github.com/pkg/errors"
	"github.com/prysmaticlabs/prysm/v4/encoding/bytesutil"
	"github.com/prysmaticlabs/prysm/v4/encoding/ssz"
	ethpb "github.com/prysmaticlabs/prysm/v4/proto/prysm/v1alpha1"
)

// SharedActivityRootWithHaher returns the hash tree root of input `sharedActivity`
func SharedActivityRootWithHaher(sharedActivity *ethpb.SharedActivity) ([32]byte, error) {
	if sharedActivity == nil {
		return [32]byte{}, errors.New("nil shared activity")
	}

	fieldRoots := make([][32]byte, 4)
	for i := 0; i < len(fieldRoots); i++ {
		fieldRoots[i] = [32]byte{}
	}
	transactionsCountPerPeriodBuf := make([]byte, 8)
	binary.LittleEndian.PutUint64(transactionsCountPerPeriodBuf, sharedActivity.TransactionsGasPerPeriod)
	fieldRoots[0] = bytesutil.ToBytes32(transactionsCountPerPeriodBuf)
	transactionsCountPerEpochBuf := make([]byte, 8)
	binary.LittleEndian.PutUint64(transactionsCountPerEpochBuf, sharedActivity.TransactionsGasPerEpoch)
	fieldRoots[1] = bytesutil.ToBytes32(transactionsCountPerEpochBuf)
	baseFeePerPeriodBuf := make([]byte, 8)
	binary.LittleEndian.PutUint64(baseFeePerPeriodBuf, sharedActivity.BaseFeePerPeriod)
	fieldRoots[2] = bytesutil.ToBytes32(baseFeePerPeriodBuf)
	baseFeePerEpochBuf := make([]byte, 8)
	binary.LittleEndian.PutUint64(baseFeePerEpochBuf, sharedActivity.BaseFeePerEpoch)
	fieldRoots[3] = bytesutil.ToBytes32(baseFeePerEpochBuf)

	root, err := ssz.BitwiseMerkleize(fieldRoots, uint64(len(fieldRoots)), uint64(len(fieldRoots)))
	if err != nil {
		return [32]byte{}, err
	}
	return root, nil
}
