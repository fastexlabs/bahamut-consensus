package state_native

import (
	"errors"

	"github.com/prysmaticlabs/prysm/v3/beacon-chain/state"
	fieldparams "github.com/prysmaticlabs/prysm/v3/config/fieldparams"
	ethpb "github.com/prysmaticlabs/prysm/v3/proto/prysm/v1alpha1"
)

// ErrNilWrappedContractsContainer returns when caller attepts to wrap a nil pointer contractsContainer
var ErrNilWrappedContractsContainer = errors.New("nil contractsContainer cannot be wrapped as readonly")

// readOnlyContractsContainer returns a wrapper that only allows to read contracts containers.
type readOnlyContractsContainer struct {
	contractsContainer *ethpb.ContractsContainer
}

// NewContractsContainer initializes the read only wrapper for contracts container.
func NewContractsContainer(cc *ethpb.ContractsContainer) (state.ReadOnlyContractsContainer, error) {
	rocc := readOnlyContractsContainer{
		contractsContainer: cc,
	}
	if rocc.IsNil() {
		return nil, ErrNilWrappedContractsContainer
	}
	return rocc, nil
}

// Contracts returns a slice of contract addresses.
func (cc readOnlyContractsContainer) Contracts() [][fieldparams.ExecutionLayerAddressLength]byte {
	contracts := make([][fieldparams.ExecutionLayerAddressLength]byte, len(cc.contractsContainer.Contracts))
	for i, contract := range cc.contractsContainer.Contracts {
		copy(contracts[i][:], contract)
	}
	return contracts
}

// Contracts returns a contract address in the container by its index.
func (cc readOnlyContractsContainer) ContractAtIndex(idx int) [fieldparams.ExecutionLayerAddressLength]byte {
	var contract [fieldparams.ExecutionLayerAddressLength]byte
	copy(contract[:], cc.contractsContainer.Contracts[idx])
	return contract
}

// IsNil returns true if the contracts container is nil.
func (cc readOnlyContractsContainer) IsNil() bool {
	return cc.contractsContainer == nil
}
