package stateutil

import (
	ethpb "github.com/prysmaticlabs/prysm/v4/proto/prysm/v1alpha1"
)

// SharedActivityRoot computes the HashTreeRoot Merkleization of
// a BeaconBlockState struct according to the eth2
// Simple Serialize specification.
func SharedActivityRoot(sharedActivity *ethpb.SharedActivity) ([32]byte, error) {
	return SharedActivityRootWithHasher(sharedActivity)
}
