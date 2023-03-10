package stateutils

import (
	fieldparams "github.com/prysmaticlabs/prysm/v3/config/fieldparams"
	types "github.com/prysmaticlabs/prysm/v3/consensus-types/primitives"
	"github.com/prysmaticlabs/prysm/v3/encoding/bytesutil"
	ethpb "github.com/prysmaticlabs/prysm/v3/proto/prysm/v1alpha1"
)

// ContractsIndexMap builds a lookup map for quickly determing the index of
// a validator by their deployed contract.
func ContractsIndexMap(cc []*ethpb.ContractsContainer) map[[fieldparams.ExecutionLayerAddressLength]byte]types.ValidatorIndex {
	var length int
	for _, c := range cc {
		if c == nil {
			continue
		}
		length += len(c.Contracts)	
	}
	m := make(map[[fieldparams.ExecutionLayerAddressLength]byte]types.ValidatorIndex, length)
	for idx, record := range cc {
		if record == nil {
			continue
		}
		for _, contract := range record.Contracts {
			key := bytesutil.ToBytes20(contract)
			m[key] = types.ValidatorIndex(idx)
		}
	}
	return m
}
