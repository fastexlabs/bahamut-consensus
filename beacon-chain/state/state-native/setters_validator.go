package state_native

import (
	"github.com/pkg/errors"
	"github.com/prysmaticlabs/prysm/v3/beacon-chain/state/state-native/types"
	"github.com/prysmaticlabs/prysm/v3/beacon-chain/state/stateutil"
	"github.com/prysmaticlabs/prysm/v3/consensus-types/primitives"
	"github.com/prysmaticlabs/prysm/v3/encoding/bytesutil"
	ethpb "github.com/prysmaticlabs/prysm/v3/proto/prysm/v1alpha1"
	"github.com/prysmaticlabs/prysm/v3/runtime/version"
)

// SetValidators for the beacon state. Updates the entire
// to a new value by overwriting the previous one.
func (b *BeaconState) SetValidators(val []*ethpb.Validator) error {
	b.lock.Lock()
	defer b.lock.Unlock()

	b.validators = val
	b.sharedFieldReferences[types.Validators].MinusRef()
	b.sharedFieldReferences[types.Validators] = stateutil.NewRef(1)
	b.markFieldAsDirty(types.Validators)
	b.rebuildTrie[types.Validators] = true
	b.valMapHandler = stateutil.NewValMapHandler(b.validators)
	return nil
}

// ApplyToEveryValidator applies the provided callback function to each validator in the
// validator registry.
func (b *BeaconState) ApplyToEveryValidator(f func(idx int, val *ethpb.Validator) (bool, *ethpb.Validator, error)) error {
	b.lock.Lock()
	v := b.validators
	if ref := b.sharedFieldReferences[types.Validators]; ref.Refs() > 1 {
		v = b.validatorsReferences()
		ref.MinusRef()
		b.sharedFieldReferences[types.Validators] = stateutil.NewRef(1)
	}
	b.lock.Unlock()
	var changedVals []uint64
	for i, val := range v {
		changed, newVal, err := f(i, val)
		if err != nil {
			return err
		}
		if changed {
			changedVals = append(changedVals, uint64(i))
			v[i] = newVal
		}
	}

	b.lock.Lock()
	defer b.lock.Unlock()

	b.validators = v
	b.markFieldAsDirty(types.Validators)
	b.addDirtyIndices(types.Validators, changedVals)

	return nil
}

// UpdateValidatorAtIndex for the beacon state. Updates the validator
// at a specific index to a new value.
func (b *BeaconState) UpdateValidatorAtIndex(idx primitives.ValidatorIndex, val *ethpb.Validator) error {
	if uint64(len(b.validators)) <= uint64(idx) {
		return errors.Errorf("invalid index provided %d", idx)
	}
	b.lock.Lock()
	defer b.lock.Unlock()

	v := b.validators
	if ref := b.sharedFieldReferences[types.Validators]; ref.Refs() > 1 {
		v = b.validatorsReferences()
		ref.MinusRef()
		b.sharedFieldReferences[types.Validators] = stateutil.NewRef(1)
	}

	v[idx] = val
	b.validators = v
	b.markFieldAsDirty(types.Validators)
	b.addDirtyIndices(types.Validators, []uint64{uint64(idx)})

	return nil
}

// SetBalances for the beacon state. Updates the entire
// list to a new value by overwriting the previous one.
func (b *BeaconState) SetBalances(val []uint64) error {
	b.lock.Lock()
	defer b.lock.Unlock()

	b.sharedFieldReferences[types.Balances].MinusRef()
	b.sharedFieldReferences[types.Balances] = stateutil.NewRef(1)

	b.balances = val
	b.markFieldAsDirty(types.Balances)
	b.rebuildTrie[types.Balances] = true
	return nil
}

// UpdateBalancesAtIndex for the beacon state. This method updates the balance
// at a specific index to a new value.
func (b *BeaconState) UpdateBalancesAtIndex(idx primitives.ValidatorIndex, val uint64) error {
	if uint64(len(b.balances)) <= uint64(idx) {
		return errors.Errorf("invalid index provided %d", idx)
	}
	b.lock.Lock()
	defer b.lock.Unlock()

	bals := b.balances
	if b.sharedFieldReferences[types.Balances].Refs() > 1 {
		bals = b.balancesVal()
		b.sharedFieldReferences[types.Balances].MinusRef()
		b.sharedFieldReferences[types.Balances] = stateutil.NewRef(1)
	}

	bals[idx] = val
	b.balances = bals
	b.markFieldAsDirty(types.Balances)
	b.addDirtyIndices(types.Balances, []uint64{uint64(idx)})
	return nil
}

// SetSlashings for the beacon state. Updates the entire
// list to a new value by overwriting the previous one.
func (b *BeaconState) SetSlashings(val []uint64) error {
	b.lock.Lock()
	defer b.lock.Unlock()

	b.sharedFieldReferences[types.Slashings].MinusRef()
	b.sharedFieldReferences[types.Slashings] = stateutil.NewRef(1)

	b.slashings = val
	b.markFieldAsDirty(types.Slashings)
	return nil
}

// UpdateSlashingsAtIndex for the beacon state. Updates the slashings
// at a specific index to a new value.
func (b *BeaconState) UpdateSlashingsAtIndex(idx, val uint64) error {
	if uint64(len(b.slashings)) <= idx {
		return errors.Errorf("invalid index provided %d", idx)
	}
	b.lock.Lock()
	defer b.lock.Unlock()

	s := b.slashings
	if b.sharedFieldReferences[types.Slashings].Refs() > 1 {
		s = b.slashingsVal()
		b.sharedFieldReferences[types.Slashings].MinusRef()
		b.sharedFieldReferences[types.Slashings] = stateutil.NewRef(1)
	}

	s[idx] = val

	b.slashings = s

	b.markFieldAsDirty(types.Slashings)
	return nil
}

// AppendValidator for the beacon state. Appends the new value
// to the the end of list.
func (b *BeaconState) AppendValidator(val *ethpb.Validator) error {
	b.lock.Lock()
	defer b.lock.Unlock()

	vals := b.validators
	if b.sharedFieldReferences[types.Validators].Refs() > 1 {
		vals = b.validatorsReferences()
		b.sharedFieldReferences[types.Validators].MinusRef()
		b.sharedFieldReferences[types.Validators] = stateutil.NewRef(1)
	}

	// append validator to slice
	b.validators = append(vals, val)
	valIdx := primitives.ValidatorIndex(len(b.validators) - 1)

	b.valMapHandler.Set(bytesutil.ToBytes48(val.PublicKey), valIdx)

	b.markFieldAsDirty(types.Validators)
	b.addDirtyIndices(types.Validators, []uint64{uint64(valIdx)})
	return nil
}

// AppendBalance for the beacon state. Appends the new value
// to the the end of list.
func (b *BeaconState) AppendBalance(bal uint64) error {
	b.lock.Lock()
	defer b.lock.Unlock()

	bals := b.balances
	if b.sharedFieldReferences[types.Balances].Refs() > 1 {
		bals = b.balancesVal()
		b.sharedFieldReferences[types.Balances].MinusRef()
		b.sharedFieldReferences[types.Balances] = stateutil.NewRef(1)
	}

	b.balances = append(bals, bal)
	balIdx := len(b.balances) - 1
	b.markFieldAsDirty(types.Balances)
	b.addDirtyIndices(types.Balances, []uint64{uint64(balIdx)})
	return nil
}

// AppendInactivityScore for the beacon state.
func (b *BeaconState) AppendInactivityScore(s uint64) error {
	b.lock.Lock()
	defer b.lock.Unlock()

	if b.version == version.Phase0 {
		return errNotSupported("AppendInactivityScore", b.version)
	}

	scores := b.inactivityScores
	if b.sharedFieldReferences[types.InactivityScores].Refs() > 1 {
		scores = b.inactivityScoresVal()
		b.sharedFieldReferences[types.InactivityScores].MinusRef()
		b.sharedFieldReferences[types.InactivityScores] = stateutil.NewRef(1)
	}

	b.inactivityScores = append(scores, s)
	b.markFieldAsDirty(types.InactivityScores)
	return nil
}

// SetInactivityScores for the beacon state. Updates the entire
// list to a new value by overwriting the previous one.
func (b *BeaconState) SetInactivityScores(val []uint64) error {
	b.lock.Lock()
	defer b.lock.Unlock()

	if b.version == version.Phase0 {
		return errNotSupported("SetInactivityScores", b.version)
	}

	b.sharedFieldReferences[types.InactivityScores].MinusRef()
	b.sharedFieldReferences[types.InactivityScores] = stateutil.NewRef(1)

	b.inactivityScores = val
	b.markFieldAsDirty(types.InactivityScores)
	return nil
}

// SetContracts for the beacon state. Updates the entire
// to a new value by overwriting the previous one.
func (b *BeaconState) SetContracts(cc []*ethpb.ContractsContainer) error {
	b.lock.Lock()
	defer b.lock.Unlock()

	b.contracts = cc
	b.sharedFieldReferences[types.Contracts].MinusRef()
	b.sharedFieldReferences[types.Contracts] = stateutil.NewRef(1)
	b.markFieldAsDirty(types.Contracts)
	b.rebuildTrie[types.Contracts] = true
	b.contractsMapHandler = stateutil.NewContractsMapHandler(b.contracts)
	return nil
}

// UpdateContractsAtIndex for the beacon state. Updates the contracts container
// at a specific index to a new value.
func (b *BeaconState) UpdateContractsAtIndex(idx primitives.ValidatorIndex, cc *ethpb.ContractsContainer) error {
	if uint64(len(b.contracts)) <= uint64(idx) {
		return errors.Errorf("invalid index provided %d", idx)
	}
	b.lock.Lock()
	defer b.lock.Unlock()

	c := b.contracts
	if ref := b.sharedFieldReferences[types.Contracts]; ref.Refs() > 1 {
		c = b.contractsReferences()
		ref.MinusRef()
		b.sharedFieldReferences[types.Contracts] = stateutil.NewRef(1)
	}

	c[idx] = cc
	b.contracts = c
	b.markFieldAsDirty(types.Contracts)
	b.addDirtyIndices(types.Contracts, []uint64{uint64(idx)})

	for _, contract := range cc.Contracts {
		b.contractsMapHandler.Set(bytesutil.ToBytes20(contract), idx)
	}

	return nil
}

// AppendContracts for the beacon state. Appends the new value
// to the the end of list.
func (b *BeaconState) AppendContracts(cc *ethpb.ContractsContainer) error {
	b.lock.Lock()
	defer b.lock.Unlock()

	contracts := b.contracts
	if b.sharedFieldReferences[types.Contracts].Refs() > 1 {
		contracts = b.contractsReferences()
		b.sharedFieldReferences[types.Contracts].MinusRef()
		b.sharedFieldReferences[types.Contracts] = stateutil.NewRef(1)
	}

	// append contracts container to slice
	b.contracts = append(contracts, cc)
	valIdx := primitives.ValidatorIndex(len(b.contracts) - 1)

	for _, contract := range cc.Contracts {
		b.contractsMapHandler.Set(bytesutil.ToBytes20(contract), valIdx)
	}

	b.markFieldAsDirty(types.Contracts)
	b.addDirtyIndices(types.Contracts, []uint64{uint64(valIdx)})
	return nil
}

// SetActivities for the beacon state. Updates the entire
// to a new value by overwriting the previous one.
func (b *BeaconState) SetActivities(activities []uint64) error {
	b.lock.Lock()
	defer b.lock.Unlock()

	b.sharedFieldReferences[types.Activities].MinusRef()
	b.sharedFieldReferences[types.Activities] = stateutil.NewRef(1)

	b.activities = activities
	b.markFieldAsDirty(types.Activities)
	b.rebuildTrie[types.Activities] = true
	return nil
}

// UpdateActivitiesAtIndex for the beacon state. This method updates the activity
// at a specific index to a new value.
func (b *BeaconState) UpdateActivitiesAtIndex(idx primitives.ValidatorIndex, activity uint64) error {
	if uint64(len(b.activities)) <= uint64(idx) {
		return errors.Errorf("invalid index provided %d", idx)
	}
	b.lock.Lock()
	defer b.lock.Unlock()

	activities := b.activities
	if b.sharedFieldReferences[types.Activities].Refs() > 1 {
		activities = b.activitiesVal()
		b.sharedFieldReferences[types.Activities].MinusRef()
		b.sharedFieldReferences[types.Activities] = stateutil.NewRef(1)
	}

	activities[idx] = activity
	b.activities = activities
	b.markFieldAsDirty(types.Activities)
	b.addDirtyIndices(types.Activities, []uint64{uint64(idx)})
	return nil
}

// AppendActivtity for the beacon state. Appends the new value
// to the the end of list.
func (b *BeaconState) AppendActivity(activity uint64) error {
	b.lock.Lock()
	defer b.lock.Unlock()

	activities := b.activities
	if b.sharedFieldReferences[types.Activities].Refs() > 1 {
		activities = b.activitiesVal()
		b.sharedFieldReferences[types.Activities].MinusRef()
		b.sharedFieldReferences[types.Activities] = stateutil.NewRef(1)
	}

	b.activities = append(activities, activity)
	activityIdx := len(b.activities) - 1
	b.markFieldAsDirty(types.Activities)
	b.addDirtyIndices(types.Activities, []uint64{uint64(activityIdx)})
	return nil
}
