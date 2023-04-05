package state_native

import (
	"errors"
	"fmt"

	"github.com/prysmaticlabs/prysm/v3/beacon-chain/state"
	fieldparams "github.com/prysmaticlabs/prysm/v3/config/fieldparams"
	ethpb "github.com/prysmaticlabs/prysm/v3/proto/prysm/v1alpha1"
)

// ErrNilWrappedContractsContainer returns when caller attepts to wrap a nil pointer contractsContainer
var ErrNilWrappedContractsContainer = errors.New("nil contractsContainer cannot be wrapped as readonly")

// ContractsIndexOutOfRangeError represents an error scenario where a validator does not exist
// at a given index in the validator's array.
type ContractIndexOutOfRangeError struct {
	message string
}

// NewContractIndexOutOfRangeError creates a new error instance.
func NewContractIndexOutOfRangeError(index int) ContractIndexOutOfRangeError {
	return ContractIndexOutOfRangeError{
		message: fmt.Sprintf("index %d out of range", index),
	}
}

// Error returns the underlying error message.
func (e *ContractIndexOutOfRangeError) Error() string {
	return e.message
}

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
func (cc readOnlyContractsContainer) ContractAtIndex(idx int) ([fieldparams.ExecutionLayerAddressLength]byte, error) {
	var contract [fieldparams.ExecutionLayerAddressLength]byte
	if len(cc.contractsContainer.Contracts) <= idx {
		e := NewContractIndexOutOfRangeError(idx)
		return contract, &e
	}
	copy(contract[:], cc.contractsContainer.Contracts[idx])
	return contract, nil
}

// IsNil returns true if the contracts container is nil.
func (cc readOnlyContractsContainer) IsNil() bool {
	return cc.contractsContainer == nil
}
