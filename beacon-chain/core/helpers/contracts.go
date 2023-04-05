package helpers

import (
	"github.com/prysmaticlabs/prysm/v3/beacon-chain/state"
	types "github.com/prysmaticlabs/prysm/v3/consensus-types/primitives"
	ethpb "github.com/prysmaticlabs/prysm/v3/proto/prysm/v1alpha1"
)

// AppendValidatorContracts updates list of validator's contracts at index.
func AppendValidatorContracts(state state.BeaconState, idx types.ValidatorIndex, contract []byte) error {
	ccAtIndex, err := state.ContractsAtIndex(idx)
	if err != nil {
		return err
	}

	newContractsContainer := appendValidatorContractsWithVal(ccAtIndex, contract)

	return state.UpdateContractsAtIndex(idx, newContractsContainer)
}

func appendValidatorContractsWithVal(cc *ethpb.ContractsContainer, contract []byte) *ethpb.ContractsContainer {
	contracts := cc.Contracts
	contracts = append(contracts, contract)
	cc.Contracts = contracts
	return cc
}
