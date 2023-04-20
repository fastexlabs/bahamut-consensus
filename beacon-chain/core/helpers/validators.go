package helpers

import (
	"bytes"
	"context"
	"math"
	"math/big"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prysmaticlabs/prysm/v3/beacon-chain/cache"
	"github.com/prysmaticlabs/prysm/v3/beacon-chain/core/time"
	"github.com/prysmaticlabs/prysm/v3/beacon-chain/state"
	"github.com/prysmaticlabs/prysm/v3/config/params"
	"github.com/prysmaticlabs/prysm/v3/consensus-types/primitives"
	"github.com/prysmaticlabs/prysm/v3/crypto/hash"
	"github.com/prysmaticlabs/prysm/v3/encoding/bytesutil"
	ethpb "github.com/prysmaticlabs/prysm/v3/proto/prysm/v1alpha1"
	"github.com/prysmaticlabs/prysm/v3/runtime/version"
	"github.com/prysmaticlabs/prysm/v3/time/slots"
	log "github.com/sirupsen/logrus"
)

var CommitteeCacheInProgressHit = promauto.NewCounter(prometheus.CounterOpts{
	Name: "committee_cache_in_progress_hit",
	Help: "The number of committee requests that are present in the cache.",
})

// IsActiveValidator returns the boolean value on whether the validator
// is active or not.
//
// Spec pseudocode definition:
//
//	def is_active_validator(validator: Validator, epoch: Epoch) -> bool:
//	  """
//	  Check if ``validator`` is active.
//	  """
//	  return validator.activation_epoch <= epoch < validator.exit_epoch
func IsActiveValidator(validator *ethpb.Validator, epoch primitives.Epoch) bool {
	return checkValidatorActiveStatus(validator.ActivationEpoch, validator.ExitEpoch, epoch)
}

// IsActiveValidatorUsingTrie checks if a read only validator is active.
func IsActiveValidatorUsingTrie(validator state.ReadOnlyValidator, epoch primitives.Epoch) bool {
	return checkValidatorActiveStatus(validator.ActivationEpoch(), validator.ExitEpoch(), epoch)
}

// IsActiveNonSlashedValidatorUsingTrie checks if a read only validator is active and not slashed
func IsActiveNonSlashedValidatorUsingTrie(validator state.ReadOnlyValidator, epoch primitives.Epoch) bool {
	active := checkValidatorActiveStatus(validator.ActivationEpoch(), validator.ExitEpoch(), epoch)
	return active && !validator.Slashed()
}

func checkValidatorActiveStatus(activationEpoch, exitEpoch, epoch primitives.Epoch) bool {
	return activationEpoch <= epoch && epoch < exitEpoch
}

// IsSlashableValidator returns the boolean value on whether the validator
// is slashable or not.
//
// Spec pseudocode definition:
//
//	def is_slashable_validator(validator: Validator, epoch: Epoch) -> bool:
//	"""
//	Check if ``validator`` is slashable.
//	"""
//	return (not validator.slashed) and (validator.activation_epoch <= epoch < validator.withdrawable_epoch)
func IsSlashableValidator(activationEpoch, withdrawableEpoch primitives.Epoch, slashed bool, epoch primitives.Epoch) bool {
	return checkValidatorSlashable(activationEpoch, withdrawableEpoch, slashed, epoch)
}

// IsSlashableValidatorUsingTrie checks if a read only validator is slashable.
func IsSlashableValidatorUsingTrie(val state.ReadOnlyValidator, epoch primitives.Epoch) bool {
	return checkValidatorSlashable(val.ActivationEpoch(), val.WithdrawableEpoch(), val.Slashed(), epoch)
}

func checkValidatorSlashable(activationEpoch, withdrawableEpoch primitives.Epoch, slashed bool, epoch primitives.Epoch) bool {
	active := activationEpoch <= epoch
	beforeWithdrawable := epoch < withdrawableEpoch
	return beforeWithdrawable && active && !slashed
}

// ActiveValidatorIndices filters out active validators based on validator status
// and returns their indices in a list.
//
// WARNING: This method allocates a new copy of the validator index set and is
// considered to be very memory expensive. Avoid using this unless you really
// need the active validator indices for some specific reason.
//
// Spec pseudocode definition:
//
//	def get_active_validator_indices(state: BeaconState, epoch: Epoch) -> Sequence[ValidatorIndex]:
//	  """
//	  Return the sequence of active validator indices at ``epoch``.
//	  """
//	  return [ValidatorIndex(i) for i, v in enumerate(state.validators) if is_active_validator(v, epoch)]
func ActiveValidatorIndices(ctx context.Context, s state.ReadOnlyBeaconState, epoch primitives.Epoch) ([]primitives.ValidatorIndex, error) {
	seed, err := Seed(s, epoch, params.BeaconConfig().DomainBeaconAttester)
	if err != nil {
		return nil, errors.Wrap(err, "could not get seed")
	}
	activeIndices, err := committeeCache.ActiveIndices(ctx, seed)
	if err != nil {
		return nil, errors.Wrap(err, "could not interface with committee cache")
	}
	if activeIndices != nil {
		return activeIndices, nil
	}

	if err := committeeCache.MarkInProgress(seed); err != nil {
		if errors.Is(err, cache.ErrAlreadyInProgress) {
			activeIndices, err := committeeCache.ActiveIndices(ctx, seed)
			if err != nil {
				return nil, err
			}
			if activeIndices == nil {
				return nil, errors.New("nil active indices")
			}
			CommitteeCacheInProgressHit.Inc()
			return activeIndices, nil
		}
		return nil, errors.Wrap(err, "could not mark committee cache as in progress")
	}
	defer func() {
		if err := committeeCache.MarkNotInProgress(seed); err != nil {
			log.WithError(err).Error("Could not mark cache not in progress")
		}
	}()

	var indices []primitives.ValidatorIndex
	if err := s.ReadFromEveryValidator(func(idx int, val state.ReadOnlyValidator) error {
		if IsActiveValidatorUsingTrie(val, epoch) {
			indices = append(indices, primitives.ValidatorIndex(idx))
		}
		return nil
	}); err != nil {
		return nil, err
	}

	if err := UpdateCommitteeCache(ctx, s, epoch); err != nil {
		return nil, errors.Wrap(err, "could not update committee cache")
	}

	return indices, nil
}

// ActiveValidatorCount returns the number of active validators in the state
// at the given epoch.
func ActiveValidatorCount(ctx context.Context, s state.ReadOnlyBeaconState, epoch primitives.Epoch) (uint64, error) {
	seed, err := Seed(s, epoch, params.BeaconConfig().DomainBeaconAttester)
	if err != nil {
		return 0, errors.Wrap(err, "could not get seed")
	}
	activeCount, err := committeeCache.ActiveIndicesCount(ctx, seed)
	if err != nil {
		return 0, errors.Wrap(err, "could not interface with committee cache")
	}
	if activeCount != 0 && s.Slot() != 0 {
		return uint64(activeCount), nil
	}

	if err := committeeCache.MarkInProgress(seed); err != nil {
		if errors.Is(err, cache.ErrAlreadyInProgress) {
			activeCount, err := committeeCache.ActiveIndicesCount(ctx, seed)
			if err != nil {
				return 0, err
			}
			CommitteeCacheInProgressHit.Inc()
			return uint64(activeCount), nil
		}
		return 0, errors.Wrap(err, "could not mark committee cache as in progress")
	}
	defer func() {
		if err := committeeCache.MarkNotInProgress(seed); err != nil {
			log.WithError(err).Error("Could not mark cache not in progress")
		}
	}()

	count := uint64(0)
	if err := s.ReadFromEveryValidator(func(idx int, val state.ReadOnlyValidator) error {
		if IsActiveValidatorUsingTrie(val, epoch) {
			count++
		}
		return nil
	}); err != nil {
		return 0, err
	}

	if err := UpdateCommitteeCache(ctx, s, epoch); err != nil {
		return 0, errors.Wrap(err, "could not update committee cache")
	}

	return count, nil
}

// TotalEffectiveActivity returns sum of active validators' effective activities
func TotalEffectiveActivity(s state.ReadOnlyBeaconState, epoch primitives.Epoch) (uint64, error) {
	totalActivity := uint64(0)
	if err := s.ReadFromEveryValidator(func(idx int, val state.ReadOnlyValidator) error {
		if IsActiveValidatorUsingTrie(val, epoch) {
			totalActivity += val.EffectiveActivity()
		}
		return nil
	}); err != nil {
		return 0, err
	}

	return totalActivity, nil
}

// ActivationExitEpoch takes in epoch number and returns when
// the validator is eligible for activation and exit.
//
// Spec pseudocode definition:
//
//	def compute_activation_exit_epoch(epoch: Epoch) -> Epoch:
//	  """
//	  Return the epoch during which validator activations and exits initiated in ``epoch`` take effect.
//	  """
//	  return Epoch(epoch + 1 + MAX_SEED_LOOKAHEAD)
func ActivationExitEpoch(epoch primitives.Epoch) primitives.Epoch {
	return epoch + 1 + params.BeaconConfig().MaxSeedLookahead
}

// ValidatorChurnLimit returns the number of validators that are allowed to
// enter and exit validator pool for an epoch.
//
// Spec pseudocode definition:
//
//	def get_validator_churn_limit(state: BeaconState) -> uint64:
//	 """
//	 Return the validator churn limit for the current epoch.
//	 """
//	 active_validator_indices = get_active_validator_indices(state, get_current_epoch(state))
//	 return max(MIN_PER_EPOCH_CHURN_LIMIT, uint64(len(active_validator_indices)) // CHURN_LIMIT_QUOTIENT)
func ValidatorChurnLimit(activeValidatorCount uint64) (uint64, error) {
	churnLimit := activeValidatorCount / params.BeaconConfig().ChurnLimitQuotient
	if churnLimit < params.BeaconConfig().MinPerEpochChurnLimit {
		churnLimit = params.BeaconConfig().MinPerEpochChurnLimit
	}
	return churnLimit, nil
}

// BeaconProposerIndex returns proposer index of a current slot.
//
// Spec pseudocode definition:
//
//	def get_beacon_proposer_index(state: BeaconState) -> ValidatorIndex:
//	  """
//	  Return the beacon proposer index at the current slot.
//	  """
//	  epoch = get_current_epoch(state)
//	  seed = hash(get_seed(state, epoch, DOMAIN_BEACON_PROPOSER) + uint_to_bytes(state.slot))
//	  indices = get_active_validator_indices(state, epoch)
//	  return compute_proposer_index(state, indices, seed)
func BeaconProposerIndex(ctx context.Context, state state.ReadOnlyBeaconState) (primitives.ValidatorIndex, error) {
	e := time.CurrentEpoch(state)
	// The cache uses the state root of the previous epoch - minimum_seed_lookahead last slot as key. (e.g. Starting epoch 1, slot 32, the key would be block root at slot 31)
	// For simplicity, the node will skip caching of genesis epoch.
	if e > params.BeaconConfig().GenesisEpoch+params.BeaconConfig().MinSeedLookahead {
		wantedEpoch := time.PrevEpoch(state)
		s, err := slots.EpochEnd(wantedEpoch)
		if err != nil {
			return 0, err
		}
		r, err := StateRootAtSlot(state, s)
		if err != nil {
			return 0, err
		}
		if r != nil && !bytes.Equal(r, params.BeaconConfig().ZeroHash[:]) {
			proposerIndices, err := proposerIndicesCache.ProposerIndices(bytesutil.ToBytes32(r))
			if err != nil {
				return 0, errors.Wrap(err, "could not interface with committee cache")
			}
			if proposerIndices != nil {
				if len(proposerIndices) != int(params.BeaconConfig().SlotsPerEpoch) {
					return 0, errors.Errorf("length of proposer indices is not equal %d to slots per epoch", len(proposerIndices))
				}
				return proposerIndices[state.Slot()%params.BeaconConfig().SlotsPerEpoch], nil
			}
			if err := UpdateProposerIndicesInCache(ctx, state); err != nil {
				return 0, errors.Wrap(err, "could not update committee cache")
			}
		}
	}

	seed, err := Seed(state, e, params.BeaconConfig().DomainBeaconProposer)
	if err != nil {
		return 0, errors.Wrap(err, "could not generate seed")
	}

	seedWithSlot := append(seed[:], bytesutil.Bytes8(uint64(state.Slot()))...)
	seedWithSlotHash := hash.Hash(seedWithSlot)

	indices, err := ActiveValidatorIndices(ctx, state, e)
	if err != nil {
		return 0, errors.Wrap(err, "could not get active indices")
	}

	if state.Version() < version.FastexPhase1 {
		return ComputeProposerIndex(state, indices, seedWithSlotHash)
	}
	return ComputeProposerIndexFastexPhase1(state, indices, seedWithSlotHash)
}

// MaxPower find the validator with the highest power.
func MaxPower(
	bState state.ReadOnlyValidators,
	activeIndices []primitives.ValidatorIndex,
	totalBalance uint64,
	txGasPerPeriod uint64,
	nonStakersGasPerPeriod uint64,
) (*big.Int, error) {
	maxPower := big.NewInt(0)
	for i := 0; i < len(activeIndices); i++ {
		v, err := bState.ValidatorAtIndexReadOnly(activeIndices[i])
		if err != nil {
			return nil, err
		}

		totalBalanceBig := new(big.Int).SetUint64(totalBalance / params.BeaconConfig().EffectiveBalanceIncrement)
		effectiveBalanceBig := new(big.Int).SetUint64(v.EffectiveBalance() / params.BeaconConfig().EffectiveBalanceIncrement)
		effectiveActivityBig := new(big.Int).SetUint64(v.EffectiveActivity())
		txGasBig := new(big.Int).SetUint64(txGasPerPeriod)
		nonStakersGasBig := new(big.Int).SetUint64(nonStakersGasPerPeriod)

		var power *big.Int
		power = new(big.Int).Add(txGasBig, nonStakersGasBig)
		power = new(big.Int).Mul(power, effectiveBalanceBig)
		power = new(big.Int).Div(power, totalBalanceBig)
		power = new(big.Int).Add(power, effectiveActivityBig)

		if power.Cmp(maxPower) > 0 {
			maxPower = power
		}
	}
	if maxPower.Cmp(big.NewInt(0)) == 0 {
		maxPower = big.NewInt(1)
	}

	return maxPower, nil
}

// SumPowers returns sum of validators' powers.
func SumPowers(
	bState state.ReadOnlyValidators,
	activeIndices []primitives.ValidatorIndex,
	totalBalance uint64,
	txGasPerPeriod uint64,
) (*big.Float, error) {
	maxEffectiveBalanceBig := new(big.Float).SetUint64(params.BeaconConfig().MaxEffectiveBalance / params.BeaconConfig().EffectiveBalanceIncrement)
	totalBalanceBig := new(big.Float).SetUint64(totalBalance / params.BeaconConfig().EffectiveBalanceIncrement)
	txGasBig := new(big.Float).SetUint64(txGasPerPeriod)

	sumPowers := big.NewFloat(0)
	for i := 0; i < len(activeIndices); i++ {
		v, err := bState.ValidatorAtIndexReadOnly(activeIndices[i])
		if err != nil {
			return nil, err
		}

		effectiveBalanceBig := new(big.Float).SetUint64(v.EffectiveBalance() / params.BeaconConfig().EffectiveBalanceIncrement)
		effectiveActivityBig := new(big.Float).SetUint64(v.EffectiveActivity())

		// In post-FastexPhase1 fork power of i'th validator is
		// power_i = (tx_gas * effective_balance_i) / total_balance + effective_activity
		// effective_power_i = power_i * effective_balance_i / max_effective_balance
		power := new(big.Float).Set(txGasBig)
		power = power.Mul(power, effectiveBalanceBig)
		power = power.Quo(power, totalBalanceBig)
		power = power.Add(power, effectiveActivityBig)

		effectivePower := new(big.Float).Set(power)
		effectivePower = effectivePower.Mul(effectivePower, effectiveBalanceBig)
		effectivePower = effectivePower.Quo(effectivePower, maxEffectiveBalanceBig)

		sumPowers = sumPowers.Add(sumPowers, effectivePower)
	}

	return sumPowers, nil
}

// ComputeProposerIndex returns the index sampled by effective balance, which is used to calculate proposer.
func ComputeProposerIndex(bState state.ReadOnlyBeaconState, activeIndices []primitives.ValidatorIndex, seed [32]byte) (primitives.ValidatorIndex, error) {
	length := uint64(len(activeIndices))
	if length == 0 {
		return 0, errors.New("empty active indices list")
	}
	maxRandomByte := new(big.Float).SetUint64(1<<16 - 1)
	hashFunc := hash.CustomSHA256Hasher()

	txGasPerPeriod := bState.TransactionsGasPerPeriod()
	var nonStakersGasPerPeriod uint64
	// Ignore nonStakersGasPerPeriod in post-FastexPhase1 fork.
	if bState.Version() < version.FastexPhase1 {
		nonStakersGasPerPeriod = bState.NonStakersGasPerPeriod()
	}
	totalBalance := TotalBalance(bState, activeIndices)
	maxPower, err := MaxPower(bState, activeIndices, totalBalance, txGasPerPeriod, nonStakersGasPerPeriod)
	maxPowerFloat := new(big.Float).SetInt(maxPower)
	if err != nil {
		return 0, err
	}

	totalBalanceBig := new(big.Int).SetUint64(totalBalance / params.BeaconConfig().EffectiveBalanceIncrement)
	txGasBig := new(big.Int).SetUint64(txGasPerPeriod)
	nonStakersGasBig := new(big.Int).SetUint64(nonStakersGasPerPeriod)

	for i := uint64(0); ; i++ {
		candidateIndex, err := ComputeShuffledIndex(primitives.ValidatorIndex(i%length), length, seed, true /* shuffle */)
		if err != nil {
			return 0, err
		}
		candidateIndex = activeIndices[candidateIndex]
		if uint64(candidateIndex) >= uint64(bState.NumValidators()) {
			return 0, errors.New("active index out of range")
		}
		b := append(seed[:], bytesutil.Bytes8(i/16)...)
		hash := hashFunc(b)
		bytes2 := append([]byte{}, hash[i%16], hash[16+i%16])
		randomBytes := new(big.Float).SetUint64(uint64(bytesutil.FromBytes2(bytes2)))
		v, err := bState.ValidatorAtIndexReadOnly(candidateIndex)
		if err != nil {
			return 0, err
		}

		effectiveBalanceBig := new(big.Int).SetUint64(v.EffectiveBalance() / params.BeaconConfig().EffectiveBalanceIncrement)
		effectiveActivityBig := new(big.Int).SetUint64(v.EffectiveActivity())

		var power *big.Int
		power = new(big.Int).Add(txGasBig, nonStakersGasBig)
		power = new(big.Int).Mul(power, effectiveBalanceBig)
		power = new(big.Int).Div(power, totalBalanceBig)
		power = new(big.Int).Add(power, effectiveActivityBig)

		powerFloat := new(big.Float).SetInt(power)
		powerProportion := new(big.Float).Quo(powerFloat, maxPowerFloat)
		p, _ := powerProportion.Float64()
		if p == 0 {
			p = 1
		}

		powerSigmoid := (1 / (1 + math.Exp(params.BeaconConfig().SigmoidExpCoefficient*p)))

		var left *big.Float
		left = new(big.Float).Quo(randomBytes, maxRandomByte)
		left = new(big.Float).Mul(left, new(big.Float).SetFloat64(params.BeaconConfig().SigmoidLimit))

		right := new(big.Float).SetFloat64((2*powerSigmoid - 1) * float64(v.EffectiveBalance()) / float64(params.BeaconConfig().MaxEffectiveBalance))

		if right.Cmp(left) >= 0 {
			return candidateIndex, nil
		}
	}
}

// ComputeProposerIndexFastexPhase1 returns the index sampled by validators' power in post-FastexPhase1 forks.
func ComputeProposerIndexFastexPhase1(bState state.ReadOnlyBeaconState, activeIndices []primitives.ValidatorIndex, seed [32]byte) (primitives.ValidatorIndex, error) {
	length := uint64(len(activeIndices))
	if length == 0 {
		return 0, errors.New("empty active indices list")
	}
	maxRandomBytes := new(big.Float).SetUint64(1<<16 - 1)
	hashFunc := hash.CustomSHA256Hasher()

	txGasPerPeriod := bState.TransactionsGasPerPeriod()
	totalBalance := TotalBalance(bState, activeIndices)
	sumPower, err := SumPowers(bState, activeIndices, totalBalance, txGasPerPeriod)
	if err != nil {
		return 0, err
	}

	maxEffectiveBalanceBig := new(big.Float).SetUint64(params.BeaconConfig().MaxEffectiveBalance / params.BeaconConfig().EffectiveBalanceIncrement)
	totalBalanceBig := new(big.Float).SetUint64(totalBalance / params.BeaconConfig().EffectiveBalanceIncrement)
	txGasBig := new(big.Float).SetUint64(txGasPerPeriod)

	b := append(seed[:], bytesutil.Bytes8(0)...)
	hash := hashFunc(b)
	bytes2 := append([]byte{}, hash[0], hash[16])
	randomBytes := new(big.Float).SetUint64(uint64(bytesutil.FromBytes2(bytes2)))

	accumPower := big.NewFloat(0)
	for i := uint64(0); ; i++ {
		candidateIndex, err := ComputeShuffledIndex(primitives.ValidatorIndex(i%length), length, seed, true /* shuffle */)
		if err != nil {
			return 0, err
		}
		candidateIndex = activeIndices[candidateIndex]
		if uint64(candidateIndex) >= uint64(bState.NumValidators()) {
			return 0, errors.New("active index out of range")
		}
		v, err := bState.ValidatorAtIndexReadOnly(candidateIndex)
		if err != nil {
			return 0, err
		}

		effectiveBalanceBig := new(big.Float).SetUint64(v.EffectiveBalance() / params.BeaconConfig().EffectiveBalanceIncrement)
		effectiveActivityBig := new(big.Float).SetUint64(v.EffectiveActivity())

		// In post-FastexPhase1 fork power of i'th validator is
		// power_i = (tx_gas * effective_balance_i) / total_balance + effective_activity
		// effective_power_i = power_i * effective_balance_i / max_effective_balance
		power := new(big.Float).Set(txGasBig)
		power = power.Mul(power, effectiveBalanceBig)
		power = power.Quo(power, totalBalanceBig)
		power = power.Add(power, effectiveActivityBig)

		effectivePower := new(big.Float).Set(power)
		effectivePower = effectivePower.Mul(effectivePower, effectiveBalanceBig)
		effectivePower = effectivePower.Quo(effectivePower, maxEffectiveBalanceBig)

		accumPower = accumPower.Add(accumPower, effectivePower)

		left := new(big.Float).Mul(randomBytes, sumPower)
		right := new(big.Float).Mul(accumPower, maxRandomBytes)

		if right.Cmp(left) >= 0 {
			return candidateIndex, nil
		}
	}
}

// IsEligibleForActivationQueue checks if the validator is eligible to
// be placed into the activation queue.
//
// Spec pseudocode definition:
//
//	def is_eligible_for_activation_queue(validator: Validator) -> bool:
//	  """
//	  Check if ``validator`` is eligible to be placed into the activation queue.
//	  """
//	  return (
//	      validator.activation_eligibility_epoch == FAR_FUTURE_EPOCH
//	      and validator.effective_balance == MAX_EFFECTIVE_BALANCE
//	  )
func IsEligibleForActivationQueue(validator *ethpb.Validator) bool {
	return isEligibileForActivationQueue(validator.ActivationEligibilityEpoch, validator.EffectiveBalance)
}

// IsEligibleForActivationQueueUsingTrie checks if the read-only validator is eligible to
// be placed into the activation queue.
func IsEligibleForActivationQueueUsingTrie(validator state.ReadOnlyValidator) bool {
	return isEligibileForActivationQueue(validator.ActivationEligibilityEpoch(), validator.EffectiveBalance())
}

// isEligibleForActivationQueue carries out the logic for IsEligibleForActivationQueue*
func isEligibileForActivationQueue(activationEligibilityEpoch primitives.Epoch, effectiveBalance uint64) bool {
	return activationEligibilityEpoch == params.BeaconConfig().FarFutureEpoch &&
		effectiveBalance == params.BeaconConfig().MaxEffectiveBalance
}

// IsEligibleForActivation checks if the validator is eligible for activation.
//
// Spec pseudocode definition:
//
//	def is_eligible_for_activation(state: BeaconState, validator: Validator) -> bool:
//	  """
//	  Check if ``validator`` is eligible for activation.
//	  """
//	  return (
//	      # Placement in queue is finalized
//	      validator.activation_eligibility_epoch <= state.finalized_checkpoint.epoch
//	      # Has not yet been activated
//	      and validator.activation_epoch == FAR_FUTURE_EPOCH
//	  )
func IsEligibleForActivation(state state.ReadOnlyCheckpoint, validator *ethpb.Validator) bool {
	finalizedEpoch := state.FinalizedCheckpointEpoch()
	return isEligibleForActivation(validator.ActivationEligibilityEpoch, validator.ActivationEpoch, finalizedEpoch)
}

// IsEligibleForActivationUsingTrie checks if the validator is eligible for activation.
func IsEligibleForActivationUsingTrie(state state.ReadOnlyCheckpoint, validator state.ReadOnlyValidator) bool {
	cpt := state.FinalizedCheckpoint()
	if cpt == nil {
		return false
	}
	return isEligibleForActivation(validator.ActivationEligibilityEpoch(), validator.ActivationEpoch(), cpt.Epoch)
}

// isEligibleForActivation carries out the logic for IsEligibleForActivation*
func isEligibleForActivation(activationEligibilityEpoch, activationEpoch, finalizedEpoch primitives.Epoch) bool {
	return activationEligibilityEpoch <= finalizedEpoch &&
		activationEpoch == params.BeaconConfig().FarFutureEpoch
}
