// Package stateutils contains useful tools for faster computation
// of state transitions using maps to represent validators instead
// of slices.
package stateutils

import (
	fieldparams "github.com/prysmaticlabs/prysm/v4/config/fieldparams"
	"github.com/prysmaticlabs/prysm/v4/config/params"
	"github.com/prysmaticlabs/prysm/v4/consensus-types/primitives"
	"github.com/prysmaticlabs/prysm/v4/encoding/bytesutil"
	ethpb "github.com/prysmaticlabs/prysm/v4/proto/prysm/v1alpha1"
)

// ContractIndexMap builds a lookup map for quickly determining the index of
// a validator by their deployed contract address.
func ContractIndexMap(validators []*ethpb.Validator, epoch primitives.Epoch) map[[fieldparams.ContractAddressLength]byte]primitives.ValidatorIndex {
	m := make(map[[fieldparams.ContractAddressLength]byte]primitives.ValidatorIndex, len(validators))
	if validators == nil {
		return m
	}
	for idx, record := range validators {
		if record == nil || record.ExitEpoch < epoch {
			continue
		}
		key := bytesutil.ToBytes20(record.Contract)
		if key != params.BeaconConfig().ZeroContract {
			m[key] = primitives.ValidatorIndex(idx)
		}
	}
	return m
}
