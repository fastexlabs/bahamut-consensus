package fastexphase1

import (
	"context"
	"math/big"

	"github.com/pkg/errors"
	"github.com/prysmaticlabs/prysm/v3/beacon-chain/core/altair"
	"github.com/prysmaticlabs/prysm/v3/beacon-chain/core/helpers"
	"github.com/prysmaticlabs/prysm/v3/beacon-chain/core/time"
	"github.com/prysmaticlabs/prysm/v3/beacon-chain/state"
	"github.com/prysmaticlabs/prysm/v3/config/params"
	"github.com/prysmaticlabs/prysm/v3/consensus-types/interfaces"
	"github.com/prysmaticlabs/prysm/v3/crypto/bls"
	"github.com/prysmaticlabs/prysm/v3/encoding/bytesutil"
	ethpb "github.com/prysmaticlabs/prysm/v3/proto/prysm/v1alpha1"
	"go.opencensus.io/trace"
)

// ProcessBaseFeePerEpoch calculates new baseFeePerEpoch value
func ProcessBaseFeePerEpoch(
	ctx context.Context,
	s state.BeaconState,
	blk interfaces.ReadOnlySignedBeaconBlock,
) (state.BeaconState, error) {
	ctx, span := trace.StartSpan(ctx, "core.ProcessBaseFeePerEpoch")
	defer span.End()

	val, err := blk.Block().Body().BaseFee()
	if err != nil {
		return nil, err
	}

	bfe, err := s.BaseFeePerEpoch()
	if err != nil {
		return nil, err
	}

	if err := s.SetBaseFeePerEpoch(bfe + val); err != nil {
		return nil, err
	}

	return s, nil
}

// ProcessSyncAggregate verifies sync committee aggregate signature signing over the previous slot block root.
func ProcessSyncAggregate(ctx context.Context, s state.BeaconState, sync *ethpb.SyncAggregate) (state.BeaconState, uint64, error) {
	s, votedKeys, reward, err := processSyncAggregate(ctx, s, sync)
	if err != nil {
		return nil, 0, errors.Wrap(err, "could not filter sync committee votes")
	}

	if err := altair.VerifySyncCommitteeSig(s, votedKeys, sync.SyncCommitteeSignature); err != nil {
		return nil, 0, errors.Wrap(err, "could not verify sync committee signature")
	}
	return s, reward, nil
}

// processSyncAggregate applies all the logic in the spec function `process_sync_aggregate` except
// verifying the BLS signatures. It returns the modified beacons state, the list of validators'
// public keys that voted (for future signature verification) and the proposer reward for including
// sync aggregate messages.
func processSyncAggregate(
	ctx context.Context,
	s state.BeaconState,
	sync *ethpb.SyncAggregate,
) (state.BeaconState, []bls.PublicKey, uint64, error) {
	currentSyncCommittee, err := s.CurrentSyncCommittee()
	if err != nil {
		return nil, nil, 0, err
	}
	if currentSyncCommittee == nil {
		return nil, nil, 0, errors.New("nil current sync committee in state")
	}
	committeeKeys := currentSyncCommittee.Pubkeys
	if sync.SyncCommitteeBits.Len() > uint64(len(committeeKeys)) {
		return nil, nil, 0, errors.New("bits length exceeds committee length")
	}
	votedKeys := make([]bls.PublicKey, 0, len(committeeKeys))

	activeBalance, err := helpers.TotalActiveBalance(s)
	if err != nil {
		return nil, nil, 0, err
	}
	epoch := time.CurrentEpoch(s)
	validatorsCount, err := helpers.ActiveValidatorCount(ctx, s, epoch)
	if err != nil {
		return nil, nil, 0, err
	}
	baseProposerReward, err := BaseProposerReward(s, validatorsCount)
	if err != nil {
		return nil, nil, 0, err
	}
	proposerReward, participantReward, err := SyncRewards(activeBalance, baseProposerReward)
	if err != nil {
		return nil, nil, 0, err
	}
	proposerIndex, err := helpers.BeaconProposerIndex(ctx, s)
	if err != nil {
		return nil, nil, 0, err
	}

	earnedProposerReward := uint64(0)
	for i := uint64(0); i < sync.SyncCommitteeBits.Len(); i++ {
		vIdx, exists := s.ValidatorIndexByPubkey(bytesutil.ToBytes48(committeeKeys[i]))
		// Impossible scenario.
		if !exists {
			return nil, nil, 0, errors.New("validator public key does not exist in state")
		}

		if sync.SyncCommitteeBits.BitAt(i) {
			pubKey, err := bls.PublicKeyFromBytes(committeeKeys[i])
			if err != nil {
				return nil, nil, 0, err
			}
			votedKeys = append(votedKeys, pubKey)
			if err := helpers.IncreaseBalance(s, vIdx, participantReward); err != nil {
				return nil, nil, 0, err
			}
			earnedProposerReward += proposerReward
		} else {
			if err := helpers.DecreaseBalance(s, vIdx, participantReward); err != nil {
				return nil, nil, 0, err
			}
		}
	}
	if err := helpers.IncreaseBalance(s, proposerIndex, earnedProposerReward); err != nil {
		return nil, nil, 0, err
	}
	return s, votedKeys, earnedProposerReward, err
}

// SyncRewards returns the proposer reward and the sync participant reward given the total active balance in state.
func SyncRewards(activeBalance, baseProposerReward uint64) (proposerReward, participantReward uint64, err error) {
	cfg := params.BeaconConfig()
	totalActiveIncrements := activeBalance / cfg.EffectiveBalanceIncrement
	baseRewardPerInc, err := BaseRewardPerIncrement(activeBalance)
	if err != nil {
		return 0, 0, err
	}
	totalBaseRewards := baseRewardPerInc * totalActiveIncrements
	maxParticipantRewards := totalBaseRewards * cfg.SyncRewardWeight / cfg.WeightDenominator / uint64(cfg.SlotsPerEpoch)
	participantReward = maxParticipantRewards / cfg.SyncCommitteeSize

	var bigProposerReward *big.Int
	bigBaseProposerReward := new(big.Int).SetUint64(baseProposerReward)
	bigWeightNumerator := new(big.Int).SetUint64(cfg.SyncRewardWeight)
	bigWeightDenominatorFTN := new(big.Int).SetUint64(cfg.WeightDenominatorFTN)
	// bigSlotsPerEpoch := new(big.Int).SetUint64(uint64(cfg.SlotsPerEpoch))
	bigSyncCommitteeSize := new(big.Int).SetUint64(cfg.SyncCommitteeSize)

	bigProposerReward = new(big.Int).Mul(bigBaseProposerReward, bigWeightNumerator)
	bigProposerReward = bigProposerReward.Div(bigProposerReward, bigWeightDenominatorFTN)
	// bigProposerReward = bigProposerReward.Div(bigProposerReward, bigSlotsPerEpoch)
	bigProposerReward = bigProposerReward.Div(bigProposerReward, bigSyncCommitteeSize)
	proposerReward = bigProposerReward.Uint64()
	return
}
