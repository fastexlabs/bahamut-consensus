package stateutil

import (
	"sync"

	coreutils "github.com/prysmaticlabs/prysm/v4/beacon-chain/core/transition/stateutils"
	fieldparams "github.com/prysmaticlabs/prysm/v4/config/fieldparams"
	"github.com/prysmaticlabs/prysm/v4/consensus-types/primitives"
	ethpb "github.com/prysmaticlabs/prysm/v4/proto/prysm/v1alpha1"
)

// ContractMapHandler is a container to hold the map and a reference tracker for how many
// states shared this.
type ContractMapHandler struct {
	valIdxMap map[[fieldparams.ContractAddressLength]byte]primitives.ValidatorIndex
	mapRef    *Reference
	*sync.RWMutex
}

// NewContractMapHandler returns a new contract map handler.
func NewContractMapHandler(vals []*ethpb.Validator) *ContractMapHandler {
	return &ContractMapHandler{
		valIdxMap: coreutils.ContractIndexMap(vals),
		mapRef:    &Reference{refs: 1},
		RWMutex:   new(sync.RWMutex),
	}
}

// AddRef copies the whole map and returns a map handler with the copied map.
func (c *ContractMapHandler) AddRef() {
	c.mapRef.AddRef()
}

// IsNil returns true if the underlying validator index map is nil.
func (c *ContractMapHandler) IsNil() bool {
	return c.mapRef == nil || c.valIdxMap == nil
}

// Copy the whole map and returns a map handler with the copied map.
func (c *ContractMapHandler) Copy() *ContractMapHandler {
	if c == nil || c.valIdxMap == nil {
		return &ContractMapHandler{valIdxMap: map[[fieldparams.ContractAddressLength]byte]primitives.ValidatorIndex{}, mapRef: new(Reference), RWMutex: new(sync.RWMutex)}
	}
	c.RLock()
	defer c.RUnlock()
	m := make(map[[fieldparams.ContractAddressLength]byte]primitives.ValidatorIndex, len(c.valIdxMap))
	for k, v := range c.valIdxMap {
		m[k] = v
	}
	return &ContractMapHandler{
		valIdxMap: m,
		mapRef:    &Reference{refs: 1},
		RWMutex:   new(sync.RWMutex),
	}
}

// Get the validator index using the corresponding contract address.
func (c *ContractMapHandler) Get(key [fieldparams.ContractAddressLength]byte) (primitives.ValidatorIndex, bool) {
	c.RLock()
	defer c.RUnlock()
	idx, ok := c.valIdxMap[key]
	if !ok {
		return 0, false
	}
	return idx, true
}

// Set the validator index using the corresponding contract address.
func (c *ContractMapHandler) Set(key [fieldparams.ContractAddressLength]byte, index primitives.ValidatorIndex) {
	c.Lock()
	defer c.Unlock()
	c.valIdxMap[key] = index
}
