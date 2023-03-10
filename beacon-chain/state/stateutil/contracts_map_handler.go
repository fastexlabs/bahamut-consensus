package stateutil

import (
	"sync"

	coreutils "github.com/prysmaticlabs/prysm/v3/beacon-chain/core/transition/stateutils"
	fieldparams "github.com/prysmaticlabs/prysm/v3/config/fieldparams"
	types "github.com/prysmaticlabs/prysm/v3/consensus-types/primitives"
	ethpb "github.com/prysmaticlabs/prysm/v3/proto/prysm/v1alpha1"
)

// ContractsMapHandler is a container to hold the map and a reference tracker for how many
// states shared this.
type ContractsMapHandler struct {
	valIdxMap map[[fieldparams.ExecutionLayerAddressLength]byte]types.ValidatorIndex
	mapRef *Reference
	*sync.RWMutex
}

// NewContractsMapHandler returns a new contracts map handler.
func NewContractsMapHandler(cc []*ethpb.ContractsContainer) *ContractsMapHandler {
	return &ContractsMapHandler{
		valIdxMap: coreutils.ContractsIndexMap(cc),
		mapRef: &Reference{refs: 1},
		RWMutex: new(sync.RWMutex),
	}
}

// AddRef copies the whole map and returns a map handler with the copied map.
func (c *ContractsMapHandler) AddRef() {
	c.mapRef.AddRef()
}

// IsNil returns true if the underlying validator index map is nil.
func (c *ContractsMapHandler) IsNil() bool {
	return c.mapRef == nil || c.valIdxMap == nil
}

// Copy the whole map and returns a map handler with the copied map.
func (c *ContractsMapHandler) Copy() *ContractsMapHandler {
	if c == nil || c.valIdxMap == nil {
		return &ContractsMapHandler{valIdxMap: map[[fieldparams.ExecutionLayerAddressLength]byte]types.ValidatorIndex{}, mapRef: new(Reference), RWMutex: new(sync.RWMutex)}
	}
	c.RLock()
	defer c.RUnlock()
	m := make(map[[fieldparams.ExecutionLayerAddressLength]byte]types.ValidatorIndex, len(c.valIdxMap))
	for k, v := range c.valIdxMap {
		m[k] = v
	}

	return &ContractsMapHandler{
		valIdxMap: m,
		mapRef: &Reference{refs: 1},
		RWMutex: new(sync.RWMutex),
	}
}

// Get the validator index using the corresponding contract address.
func (c *ContractsMapHandler) Get(key [fieldparams.ExecutionLayerAddressLength]byte) (types.ValidatorIndex, bool) {
	c.RLock()
	defer c.RUnlock()
	idx, ok := c.valIdxMap[key]
	if !ok {
		return 0, false
	}
	return idx, true
}

// Set the validator index using the corresponding contract address.
func (c *ContractsMapHandler) Set(key [fieldparams.ExecutionLayerAddressLength]byte, index types.ValidatorIndex) {
	c.Lock()
	defer c.Unlock()
	c.valIdxMap[key] = index
}
