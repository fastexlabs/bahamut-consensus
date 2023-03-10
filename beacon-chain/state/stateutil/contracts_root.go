package stateutil

import (
	"github.com/prysmaticlabs/prysm/v3/config/features"
	"github.com/prysmaticlabs/prysm/v3/crypto/hash/htr"
	"github.com/prysmaticlabs/prysm/v3/encoding/ssz"
	ethpb "github.com/prysmaticlabs/prysm/v3/proto/prysm/v1alpha1"
)

// ContractsRootWithHasher describes a method from which the hash tree root
// of a contracts container is returned.
func ContractsRootWithHasher(hasher ssz.HashFn, cc *ethpb.ContractsContainer) ([32]byte, error) {
	fieldRoots, err := ContractsFieldRoot(hasher, cc)
	if err != nil {
		return [32]byte{}, nil
	}
	return ssz.BitwiseMerkleize(hasher, fieldRoots, uint64(len(fieldRoots)), uint64(len(fieldRoots)))
}

// ContractsFieldRoot describes a method from which the hash tree root
// of a contracts container is returned.
func ContractsFieldRoot(hasher ssz.HashFn, cc *ethpb.ContractsContainer) ([][32]byte, error) {
	var fieldRoots [][32]byte
	if cc.Contracts != nil {
		contractsRoot, err := merkleizeContracts(hasher, cc.Contracts)
		if err != nil {
			return [][32]byte{}, nil
		}
		fieldRoots = [][32]byte{contractsRoot}
	}
	return fieldRoots, nil
}

func merkleizeContracts(hasher ssz.HashFn, contracts [][]byte) ([32]byte, error) {
	chunks, err := ssz.PackByChunk(contracts)
	if err != nil {
		return [32]byte{}, err
	}
	var contractsRoot [32]byte
	if features.Get().EnableVectorizedHTR {
		if len(chunks) % 2 == 1 {
			chunks = append(chunks, [32]byte{})
		}
		outputChunk := make([][32]byte, 1)
		htr.VectorizedSha256(chunks, outputChunk)
		contractsRoot = outputChunk[0]
	} else {
		contractsRoot, err = ssz.BitwiseMerkleize(hasher, chunks, uint64(len(chunks)), uint64(len(chunks)))
		if err != nil {
			return [32]byte{}, err
		}

	}
	return contractsRoot, nil
}
