package stateutil

import (
	"bytes"
	"encoding/binary"

	"github.com/pkg/errors"
	fieldparams "github.com/prysmaticlabs/prysm/v3/config/fieldparams"
	"github.com/prysmaticlabs/prysm/v3/crypto/hash"
	"github.com/prysmaticlabs/prysm/v3/encoding/ssz"
	ethpb "github.com/prysmaticlabs/prysm/v3/proto/prysm/v1alpha1"
)

const (
	contractsTrieDepth = 0
)

// ContractsRoot computes the HashTreeRoot Merkleization of
// a list of contract container structs according to the Ethereum
// Simple Serialize specification.
func ContractsRoot(contracts []*ethpb.ContractsContainer) ([32]byte, error) {
	return contractsRoot(contracts)
}

func contractsRoot(contracts []*ethpb.ContractsContainer) ([32]byte, error) {
	hasher := hash.CustomSHA256Hasher()

	roots, err := contractsRoots(hasher, contracts)	
	if err != nil {
		return [32]byte{}, err
	}

	contractsRootsRoot, err := ssz.BitwiseMerkleize(hasher, roots, uint64(len(roots)), fieldparams.ValidatorRegistryLimit)
	if err != nil {
		return [32]byte{}, errors.Wrap(err, "could not compute contracts merkleization")
	}

	contractsRootsBuf := new(bytes.Buffer)
	if err := binary.Write(contractsRootsBuf, binary.LittleEndian, uint64(len(contracts))); err != nil {
		return [32]byte{}, errors.Wrap(err, "could not marshal contracts length")
	}
	var contractsRootsBufRoot [32]byte
	copy(contractsRootsBufRoot[:], contractsRootsBuf.Bytes())
	res := ssz.MixInLength(contractsRootsRoot, contractsRootsBufRoot[:])

	return res, nil
}

func contractsRoots(hasher ssz.HashFn, contracts []*ethpb.ContractsContainer) ([][32]byte, error) {
	roots := make([][32]byte, len(contracts))
	for i := 0; i < len(contracts); i++ {
		contract, err := contractRoot(hasher, contracts[i])	
		if err != nil {
			return [][32]byte{}, err
		}
		roots[i] = contract
	}
	return roots, nil
}

func contractRoot(hasher ssz.HashFn, contract *ethpb.ContractsContainer) ([32]byte, error) {
	if contract == nil {
		return [32]byte{}, errors.New("nil contracts container")
	}
	return ContractsRootWithHasher(hasher, contract)
}

