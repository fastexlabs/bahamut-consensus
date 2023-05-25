package validator

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	blockchainTest "github.com/prysmaticlabs/prysm/v4/beacon-chain/blockchain/testing"
	builderTest "github.com/prysmaticlabs/prysm/v4/beacon-chain/builder/testing"
	"github.com/prysmaticlabs/prysm/v4/beacon-chain/cache"
	dbTest "github.com/prysmaticlabs/prysm/v4/beacon-chain/db/testing"
	powtesting "github.com/prysmaticlabs/prysm/v4/beacon-chain/execution/testing"
	fieldparams "github.com/prysmaticlabs/prysm/v4/config/fieldparams"
	"github.com/prysmaticlabs/prysm/v4/config/params"
	"github.com/prysmaticlabs/prysm/v4/consensus-types/blocks"
	"github.com/prysmaticlabs/prysm/v4/consensus-types/primitives"
	"github.com/prysmaticlabs/prysm/v4/encoding/bytesutil"
	v1 "github.com/prysmaticlabs/prysm/v4/proto/engine/v1"
	ethpb "github.com/prysmaticlabs/prysm/v4/proto/prysm/v1alpha1"
	"github.com/prysmaticlabs/prysm/v4/testing/require"
	"github.com/prysmaticlabs/prysm/v4/testing/util"
)

func TestServer_setActivities(t *testing.T) {
	cfg := params.BeaconConfig().Copy()
	cfg.BellatrixForkEpoch = 0
	cfg.CapellaForkEpoch = 0
	params.OverrideBeaconConfig(cfg)
	params.SetupTestConfigCleanup(t)

	beaconDB := dbTest.SetupDB(t)
	capellaTransitionState, _ := util.DeterministicGenesisStateCapella(t, 1)
	wrappedHeaderCapella, err := blocks.WrappedExecutionPayloadHeaderCapella(&v1.ExecutionPayloadHeaderCapella{BlockNumber: 1}, big.NewInt(0))
	require.NoError(t, err)
	require.NoError(t, capellaTransitionState.SetLatestExecutionPayloadHeader(wrappedHeaderCapella))
	b2pbCapella := util.NewBeaconBlockCapella()
	b2rCapella, err := b2pbCapella.Block.HashTreeRoot()
	require.NoError(t, err)
	util.SaveBlock(t, context.Background(), beaconDB, b2pbCapella)
	require.NoError(t, capellaTransitionState.SetFinalizedCheckpoint(&ethpb.Checkpoint{
		Root: b2rCapella[:],
	}))
	require.NoError(t, beaconDB.SaveFeeRecipientsByValidatorIDs(context.Background(), []primitives.ValidatorIndex{0}, []common.Address{{}}))
	withdrawals := []*v1.Withdrawal{{
		Index:          1,
		ValidatorIndex: 2,
		Address:        make([]byte, fieldparams.FeeRecipientLength),
		Amount:         3,
	}}
	id := &v1.PayloadIDBytes{0x1}
	vs := &Server{
		ExecutionEngineCaller: &powtesting.EngineClient{PayloadIDBytes: id, ExecutionPayloadCapella: &v1.ExecutionPayloadCapella{BlockNumber: 1, Withdrawals: withdrawals}, BlockValue: big.NewInt(0)}, HeadFetcher: &blockchainTest.ChainService{State: capellaTransitionState},
		FinalizationFetcher:    &blockchainTest.ChainService{},
		BeaconDB:               beaconDB,
		ProposerSlotIndexCache: cache.NewProposerPayloadIDsCache(),
		BlockBuilder:           &builderTest.MockBuilderService{HasConfigured: true},
	}
	activityChangesFromEL := []*ethpb.ActivityChange{
		{
			ContractAddress: bytesutil.PadTo([]byte("contract-1"), 20),
			DeltaActivity:   123,
		},
		{
			ContractAddress: bytesutil.PadTo([]byte("contract-2"), 20),
			DeltaActivity:   123,
		},
	}

	t.Run("No builder configured. Use local block", func(t *testing.T) {
		blk, err := blocks.NewSignedBeaconBlock(util.NewBeaconBlockCapella())
		require.NoError(t, err)
		require.NoError(t, vs.setActivities(context.Background(), blk, capellaTransitionState))
		baseFee := blk.Block().Body().BaseFee()
		transactionCount := blk.Block().Body().TransactionsCount()
		activityChanges := blk.Block().Body().ActivityChanges()
		executionHeight := blk.Block().Body().ExecutionHeight()
		require.Equal(t, uint64(123), baseFee)
		require.Equal(t, uint64(123), transactionCount)
		require.Equal(t, uint64(1), executionHeight)
		require.DeepSSZEqual(t, activityChangesFromEL, activityChanges)
	})
}
