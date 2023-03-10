package helpers

import (
	"github.com/prysmaticlabs/prysm/v3/beacon-chain/state"
	types "github.com/prysmaticlabs/prysm/v3/consensus-types/primitives"
	"github.com/prysmaticlabs/prysm/v3/encoding/bytesutil"
	ethpb "github.com/prysmaticlabs/prysm/v3/proto/prysm/v1alpha1"
)

func AppendValidatorContracts(state state.BeaconState, idx types.ValidatorIndex, contract []byte) error {
	index, ok := state.ValidatorIndexByContractAddress(bytesutil.ToBytes20(contract))
	if ok {
		if index == idx {
			return nil
		}
		// TODO: Uncomment this error
		// return fmt.Errorf("contract %x is used by another validator", contract)
		return nil
	}
	ccAtIndex, err := state.ContractsAtIndex(idx)
	if err != nil {
		return err
	}

	newContractsContainer := AppendValidatorContractsWithVal(ccAtIndex, contract)

	return state.UpdateContractsAtIndex(idx, newContractsContainer)
}

func AppendValidatorContractsWithVal(cc *ethpb.ContractsContainer, contract []byte) *ethpb.ContractsContainer {
	contracts := cc.Contracts
	contracts = append(contracts, contract)
	cc.Contracts = contracts
	return cc
}
