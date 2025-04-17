// todo unit act
package stateutil

import (
	"bytes"
	"encoding/binary"

	"github.com/pkg/errors"
	"github.com/prysmaticlabs/prysm/v4/encoding/ssz"
	ethpb "github.com/prysmaticlabs/prysm/v4/proto/prysm/v1alpha1"
)

func ActivityChangesRoot(activities []*ethpb.ActivityChange) ([32]byte, error) {
	roots := make([][32]byte, 0, len(activities))
	for _, activity := range activities {
		root, err := activity.HashTreeRoot()
		if err != nil {
			return [32]byte{}, errors.Wrap(err, "could not compute activity change merkleization")
		}
		roots = append(roots, root)
	}
	activitiesRoot, err := ssz.BitwiseMerkleize(roots, uint64(len(roots)), uint64(len(roots)))
	if err != nil {
		return [32]byte{}, errors.Wrap(err, "could not compute activity changes merkleization")
	}
	activitiesLenBuf := new(bytes.Buffer)
	if err := binary.Write(activitiesLenBuf, binary.LittleEndian, uint64(len(activities))); err != nil {
		return [32]byte{}, errors.Wrap(err, "could not marshal activity changes length")
	}
	activitiesLenRoot := make([]byte, 32)
	copy(activitiesLenRoot, activitiesLenBuf.Bytes())
	res := ssz.MixInLength(activitiesRoot, activitiesLenRoot)
	return res, nil
}
