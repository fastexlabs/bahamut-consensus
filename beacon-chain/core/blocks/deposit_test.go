package blocks_test

import (
	"context"
	"testing"

	"github.com/prysmaticlabs/prysm/v4/beacon-chain/core/blocks"
	"github.com/prysmaticlabs/prysm/v4/beacon-chain/core/signing"
	state_native "github.com/prysmaticlabs/prysm/v4/beacon-chain/state/state-native"
	fieldparams "github.com/prysmaticlabs/prysm/v4/config/fieldparams"
	"github.com/prysmaticlabs/prysm/v4/config/params"
	"github.com/prysmaticlabs/prysm/v4/consensus-types/primitives"
	"github.com/prysmaticlabs/prysm/v4/container/trie"
	"github.com/prysmaticlabs/prysm/v4/crypto/bls"
	"github.com/prysmaticlabs/prysm/v4/encoding/bytesutil"
	ethpb "github.com/prysmaticlabs/prysm/v4/proto/prysm/v1alpha1"
	"github.com/prysmaticlabs/prysm/v4/testing/assert"
	"github.com/prysmaticlabs/prysm/v4/testing/require"
	"github.com/prysmaticlabs/prysm/v4/testing/util"
)

func TestProcessDeposits_SameValidatorMultipleDepositsSameBlock(t *testing.T) {
	// Same validator created 3 valid deposits within the same block

	dep, _, err := util.DeterministicDepositsAndKeysSameValidator(3)
	require.NoError(t, err)
	eth1Data, err := util.DeterministicEth1Data(len(dep))
	require.NoError(t, err)
	b := util.NewBeaconBlock()
	b.Block = &ethpb.BeaconBlock{
		Body: &ethpb.BeaconBlockBody{
			// 3 deposits from the same validator
			Deposits: []*ethpb.Deposit{dep[0], dep[1], dep[2]},
		},
	}
	registry := []*ethpb.Validator{
		{
			PublicKey:             []byte{1},
			WithdrawalCredentials: []byte{1, 2, 3},
		},
	}
	balances := []uint64{0}
	activities := []uint64{0}
	beaconState, err := state_native.InitializeFromProtoPhase0(&ethpb.BeaconState{
		Validators: registry,
		Balances:   balances,
		Activities: activities,
		Eth1Data:   eth1Data,
		Fork: &ethpb.Fork{
			PreviousVersion: params.BeaconConfig().GenesisForkVersion,
			CurrentVersion:  params.BeaconConfig().GenesisForkVersion,
		},
	})
	require.NoError(t, err)
	newState, err := blocks.ProcessDeposits(context.Background(), beaconState, b.Block.Body.Deposits)
	require.NoError(t, err, "Expected block deposits to process correctly")

	assert.Equal(t, 2, len(newState.Validators()), "Incorrect validator count")
	assert.Equal(t, len(activities)+1, len(newState.Activities()), "Incorrect activities count")
}

func TestProcessDeposits_MerkleBranchFailsVerification(t *testing.T) {
	deposit := &ethpb.Deposit{
		Data: &ethpb.Deposit_Data{
			PublicKey:             bytesutil.PadTo([]byte{1, 2, 3}, fieldparams.BLSPubkeyLength),
			Contract:              make([]byte, fieldparams.ContractAddressLength),
			WithdrawalCredentials: make([]byte, 32),
			Signature:             make([]byte, fieldparams.BLSSignatureLength),
		},
	}
	leaf, err := deposit.Data.HashTreeRoot()
	require.NoError(t, err)

	// We then create a merkle branch for the test.
	depositTrie, err := trie.GenerateTrieFromItems([][]byte{leaf[:]}, params.BeaconConfig().DepositContractTreeDepth)
	require.NoError(t, err, "Could not generate trie")
	proof, err := depositTrie.MerkleProof(0)
	require.NoError(t, err, "Could not generate proof")

	deposit.Proof = proof
	b := util.NewBeaconBlock()
	b.Block = &ethpb.BeaconBlock{
		Body: &ethpb.BeaconBlockBody{
			Deposits: []*ethpb.Deposit{deposit},
		},
	}
	beaconState, err := state_native.InitializeFromProtoPhase0(&ethpb.BeaconState{
		Eth1Data: &ethpb.Eth1Data{
			DepositRoot: []byte{0},
			BlockHash:   []byte{1},
		},
	})
	require.NoError(t, err)
	want := "deposit root did not verify"
	_, err = blocks.ProcessDeposits(context.Background(), beaconState, b.Block.Body.Deposits)
	assert.ErrorContains(t, want, err)
}

func TestProcessDeposits_AddsNewValidatorDeposit(t *testing.T) {
	dep, _, err := util.DeterministicDepositsAndKeys(1)
	require.NoError(t, err)
	eth1Data, err := util.DeterministicEth1Data(len(dep))
	require.NoError(t, err)

	b := util.NewBeaconBlock()
	b.Block = &ethpb.BeaconBlock{
		Body: &ethpb.BeaconBlockBody{
			Deposits: []*ethpb.Deposit{dep[0]},
		},
	}
	registry := []*ethpb.Validator{
		{
			PublicKey:             []byte{1},
			WithdrawalCredentials: []byte{1, 2, 3},
		},
	}
	balances := []uint64{0}
	activities := []uint64{0}
	beaconState, err := state_native.InitializeFromProtoPhase0(&ethpb.BeaconState{
		Validators: registry,
		Balances:   balances,
		Activities: activities,
		Eth1Data:   eth1Data,
		Fork: &ethpb.Fork{
			PreviousVersion: params.BeaconConfig().GenesisForkVersion,
			CurrentVersion:  params.BeaconConfig().GenesisForkVersion,
		},
	})
	require.NoError(t, err)
	newState, err := blocks.ProcessDeposits(context.Background(), beaconState, b.Block.Body.Deposits)
	require.NoError(t, err, "Expected block deposits to process correctly")
	if newState.Balances()[1] != dep[0].Data.Amount {
		t.Errorf(
			"Expected state validator balances index 0 to equal %d, received %d",
			dep[0].Data.Amount,
			newState.Balances()[1],
		)
	}
}

func TestProcessDeposits_RepeatedDeposit_IncreasesValidatorBalance(t *testing.T) {
	sk, err := bls.RandKey()
	require.NoError(t, err)
	deposit := &ethpb.Deposit{
		Data: &ethpb.Deposit_Data{
			PublicKey:             sk.PublicKey().Marshal(),
			Contract:              make([]byte, 20),
			Amount:                1000,
			WithdrawalCredentials: make([]byte, 32),
			Signature:             make([]byte, fieldparams.BLSSignatureLength),
		},
	}
	sr, err := signing.ComputeSigningRoot(deposit.Data, bytesutil.ToBytes(3, 32))
	require.NoError(t, err)
	sig := sk.Sign(sr[:])
	deposit.Data.Signature = sig.Marshal()
	leaf, err := deposit.Data.HashTreeRoot()
	require.NoError(t, err)

	// We then create a merkle branch for the test.
	depositTrie, err := trie.GenerateTrieFromItems([][]byte{leaf[:]}, params.BeaconConfig().DepositContractTreeDepth)
	require.NoError(t, err, "Could not generate trie")
	proof, err := depositTrie.MerkleProof(0)
	require.NoError(t, err, "Could not generate proof")

	deposit.Proof = proof
	b := util.NewBeaconBlock()
	b.Block = &ethpb.BeaconBlock{
		Body: &ethpb.BeaconBlockBody{
			Deposits: []*ethpb.Deposit{deposit},
		},
	}
	registry := []*ethpb.Validator{
		{
			PublicKey: []byte{1, 2, 3},
		},
		{
			PublicKey:             sk.PublicKey().Marshal(),
			WithdrawalCredentials: []byte{1},
		},
	}
	balances := []uint64{0, 50}
	root, err := depositTrie.HashTreeRoot()
	require.NoError(t, err)
	beaconState, err := state_native.InitializeFromProtoPhase0(&ethpb.BeaconState{
		Validators: registry,
		Balances:   balances,
		Eth1Data: &ethpb.Eth1Data{
			DepositRoot: root[:],
			BlockHash:   root[:],
		},
	})
	require.NoError(t, err)
	newState, err := blocks.ProcessDeposits(context.Background(), beaconState, b.Block.Body.Deposits)
	require.NoError(t, err, "Process deposit failed")
	assert.Equal(t, uint64(1000+50), newState.Balances()[1], "Expected balance at index 1 to be 1050")
}

func TestProcessDeposit_TwoValidatorsWithSameContract(t *testing.T) {
	contract := []byte{0x42, 0x42, 0x42}
	dep, _, err := util.DeterministicDepositsAndKeysWithContract(2, [][]byte{contract, contract})
	require.NoError(t, err)
	eth1Data, err := util.DeterministicEth1Data(len(dep))
	require.NoError(t, err)

	beaconState, err := state_native.InitializeFromProtoPhase0(&ethpb.BeaconState{
		Eth1Data: eth1Data,
		Fork: &ethpb.Fork{
			PreviousVersion: params.BeaconConfig().GenesisForkVersion,
			CurrentVersion:  params.BeaconConfig().GenesisForkVersion,
		},
	})
	require.NoError(t, err)
	newState, isNewValidator, err := blocks.ProcessDeposit(beaconState, dep[0], true)
	require.NoError(t, err, "Process deposit failed")
	assert.Equal(t, true, isNewValidator, "Expected isNewValidator to be true")
	assert.Equal(t, 1, len(newState.Validators()), "Expected validator list to have length 1")
	assert.Equal(t, 1, len(newState.Balances()), "Expected validator balances list to have length 1")
	assert.Equal(t, 1, len(newState.Activities()), "Expected validator activities list to have length 1")
	if newState.Balances()[0] != dep[0].Data.Amount {
		t.Errorf(
			"Expected state validator balances index 0 to equal %d, received %d",
			dep[0].Data.Amount,
			newState.Balances()[0],
		)
	}
	owner, exist := newState.ValidatorIndexByContract(bytesutil.ToBytes20(contract))
	assert.Equal(t, true, exist, "Expected exist to be true")
	assert.Equal(t, primitives.ValidatorIndex(0), owner, "Expected owner of the contract is 0")

	newState, isNewValidator, err = blocks.ProcessDeposit(beaconState, dep[1], true)
	require.NoError(t, err, "Process deposit failed")
	assert.Equal(t, true, isNewValidator, "Expected isNewValidator to be true")
	assert.Equal(t, 2, len(newState.Validators()), "Expected validator list to have length 2")
	assert.Equal(t, 2, len(newState.Balances()), "Expected validator balances list to have length 2")
	assert.Equal(t, 2, len(newState.Activities()), "Expected validator activities list to have length 2")
	if newState.Balances()[1] != dep[1].Data.Amount {
		t.Errorf(
			"Expected state validator balances index 1 to equal %d, received %d",
			dep[1].Data.Amount,
			newState.Balances()[1],
		)
	}

	owner, exist = newState.ValidatorIndexByContract(bytesutil.ToBytes20(contract))
	assert.Equal(t, true, exist, "Expected exist to be true")
	assert.Equal(t, primitives.ValidatorIndex(0), owner, "Expected owner of the contract is 0")

	contractAtIndex, ok := newState.ContractAtIndex(primitives.ValidatorIndex(0))
	assert.Equal(t, true, ok, "Expected ok to be true")
	assert.DeepEqual(t, bytesutil.ToBytes20(contract), contractAtIndex, "Expected validator 0 to save the old contract")

	contractAtIndex, ok = newState.ContractAtIndex(primitives.ValidatorIndex(1))
	assert.Equal(t, true, ok, "Expected ok to be true")
	assert.DeepEqual(t, params.BeaconConfig().ZeroContract, contractAtIndex, "Expected validator 1 to have zero-contract")
}

func TestProcessDeposit_AddsNewValidatorDeposit(t *testing.T) {
	// Similar to TestProcessDeposits_AddsNewValidatorDeposit except that this test directly calls ProcessDeposit
	dep, _, err := util.DeterministicDepositsAndKeys(1)
	require.NoError(t, err)
	eth1Data, err := util.DeterministicEth1Data(len(dep))
	require.NoError(t, err)

	registry := []*ethpb.Validator{
		{
			PublicKey:             []byte{1},
			WithdrawalCredentials: []byte{1, 2, 3},
		},
	}
	balances := []uint64{0}
	beaconState, err := state_native.InitializeFromProtoPhase0(&ethpb.BeaconState{
		Validators: registry,
		Balances:   balances,
		Eth1Data:   eth1Data,
		Fork: &ethpb.Fork{
			PreviousVersion: params.BeaconConfig().GenesisForkVersion,
			CurrentVersion:  params.BeaconConfig().GenesisForkVersion,
		},
	})
	require.NoError(t, err)
	newState, isNewValidator, err := blocks.ProcessDeposit(beaconState, dep[0], true)
	require.NoError(t, err, "Process deposit failed")
	assert.Equal(t, true, isNewValidator, "Expected isNewValidator to be true")
	assert.Equal(t, 2, len(newState.Validators()), "Expected validator list to have length 2")
	assert.Equal(t, 2, len(newState.Balances()), "Expected validator balances list to have length 2")
	if newState.Balances()[1] != dep[0].Data.Amount {
		t.Errorf(
			"Expected state validator balances index 1 to equal %d, received %d",
			dep[0].Data.Amount,
			newState.Balances()[1],
		)
	}
}

func TestProcessDeposit_SkipsInvalidDeposit(t *testing.T) {
	// Same test settings as in TestProcessDeposit_AddsNewValidatorDeposit, except that we use an invalid signature
	contract := []byte{0x42, 0x42, 0x42}
	dep, _, err := util.DeterministicDepositsAndKeysWithContract(1, [][]byte{contract})
	require.NoError(t, err)
	dep[0].Data.Signature = make([]byte, 96)
	dt, _, err := util.DepositTrieFromDeposits(dep)
	require.NoError(t, err)
	root, err := dt.HashTreeRoot()
	require.NoError(t, err)
	eth1Data := &ethpb.Eth1Data{
		DepositRoot:  root[:],
		DepositCount: 1,
	}
	registry := []*ethpb.Validator{
		{
			PublicKey:             []byte{1},
			WithdrawalCredentials: []byte{1, 2, 3},
		},
	}
	balances := []uint64{0}
	beaconState, err := state_native.InitializeFromProtoPhase0(&ethpb.BeaconState{
		Validators: registry,
		Balances:   balances,
		Eth1Data:   eth1Data,
		Fork: &ethpb.Fork{
			PreviousVersion: params.BeaconConfig().GenesisForkVersion,
			CurrentVersion:  params.BeaconConfig().GenesisForkVersion,
		},
	})
	require.NoError(t, err)
	newState, isNewValidator, err := blocks.ProcessDeposit(beaconState, dep[0], true)
	require.NoError(t, err, "Expected invalid block deposit to be ignored without error")
	assert.Equal(t, false, isNewValidator, "Expected isNewValidator to be false")

	if newState.Eth1DepositIndex() != 1 {
		t.Errorf(
			"Expected Eth1DepositIndex to be increased by 1 after processing an invalid deposit, received change: %v",
			newState.Eth1DepositIndex(),
		)
	}
	if len(newState.Validators()) != 1 {
		t.Errorf("Expected validator list to have length 1, received: %v", len(newState.Validators()))
	}
	if len(newState.Balances()) != 1 {
		t.Errorf("Expected validator balances list to have length 1, received: %v", len(newState.Balances()))
	}
	if newState.Balances()[0] != 0 {
		t.Errorf("Expected validator balance at index 0 to stay 0, received: %v", newState.Balances()[0])
	}
	if _, ok := newState.ValidatorIndexByContract(bytesutil.ToBytes20(contract)); ok {
		t.Error("Expected contractIndexMap don't have invalid deposit contract")
	}
}

func TestPreGenesisDeposits_SkipInvalidDeposit(t *testing.T) {
	contracts := [][]byte{{0x1}, {0x2}, {0x3}}
	dep, _, err := util.DeterministicDepositsAndKeysWithContract(100, contracts)
	require.NoError(t, err)
	dep[0].Data.Signature = make([]byte, 96)
	dt, _, err := util.DepositTrieFromDeposits(dep)
	require.NoError(t, err)

	for i := range dep {
		proof, err := dt.MerkleProof(i)
		require.NoError(t, err)
		dep[i].Proof = proof
	}
	root, err := dt.HashTreeRoot()
	require.NoError(t, err)

	eth1Data := &ethpb.Eth1Data{
		DepositRoot:  root[:],
		DepositCount: 1,
	}
	registry := []*ethpb.Validator{
		{
			PublicKey:             []byte{1},
			WithdrawalCredentials: []byte{1, 2, 3},
		},
	}
	balances := []uint64{0}
	beaconState, err := state_native.InitializeFromProtoPhase0(&ethpb.BeaconState{
		Validators: registry,
		Balances:   balances,
		Eth1Data:   eth1Data,
		Fork: &ethpb.Fork{
			PreviousVersion: params.BeaconConfig().GenesisForkVersion,
			CurrentVersion:  params.BeaconConfig().GenesisForkVersion,
		},
	})
	require.NoError(t, err)
	newState, err := blocks.ProcessPreGenesisDeposits(context.Background(), beaconState, dep)
	require.NoError(t, err, "Expected invalid block deposit to be ignored without error")

	_, ok := newState.ValidatorIndexByPubkey(bytesutil.ToBytes48(dep[0].Data.PublicKey))
	require.Equal(t, false, ok, "bad pubkey should not exist in state")

	for i := 1; i < newState.NumValidators(); i++ {
		val, err := newState.ValidatorAtIndex(primitives.ValidatorIndex(i))
		require.NoError(t, err)
		require.Equal(t, params.BeaconConfig().MaxEffectiveBalance, val.EffectiveBalance, "unequal effective balance")
		require.Equal(t, primitives.Epoch(0), val.ActivationEpoch)
		require.Equal(t, primitives.Epoch(0), val.ActivationEligibilityEpoch)
	}
	if newState.Eth1DepositIndex() != 100 {
		t.Errorf(
			"Expected Eth1DepositIndex to be increased by 99 after processing an invalid deposit, received change: %v",
			newState.Eth1DepositIndex(),
		)
	}
	if len(newState.Validators()) != 100 {
		t.Errorf("Expected validator list to have length 100, received: %v", len(newState.Validators()))
	}
	if len(newState.Balances()) != 100 {
		t.Errorf("Expected validator balances list to have length 100, received: %v", len(newState.Balances()))
	}
	if newState.Balances()[0] != 0 {
		t.Errorf("Expected validator balance at index 0 to stay 0, received: %v", newState.Balances()[0])
	}

	_, ok = newState.ValidatorIndexByContract(bytesutil.ToBytes20(contracts[0]))
	if ok {
		t.Error("Expected contractIndexMap don't have invalid contract")
	}
	_, ok = newState.ValidatorIndexByContract(bytesutil.ToBytes20(contracts[1]))
	if !ok {
		t.Errorf("Expected contractIndexMap to have contract: 0x%x", bytesutil.ToBytes20(contracts[1]))
	}
	_, ok = newState.ValidatorIndexByContract(bytesutil.ToBytes20(contracts[2]))
	if !ok {
		t.Errorf("Expected contractIndexMap to have contract: 0x%x", bytesutil.ToBytes20(contracts[2]))
	}
}

func TestProcessDeposit_RepeatedDeposit_IncreasesValidatorBalance(t *testing.T) {
	sk, err := bls.RandKey()
	require.NoError(t, err)
	deposit := &ethpb.Deposit{
		Data: &ethpb.Deposit_Data{
			PublicKey:             sk.PublicKey().Marshal(),
			Contract:              make([]byte, 20),
			Amount:                1000,
			WithdrawalCredentials: make([]byte, 32),
			Signature:             make([]byte, 96),
		},
	}
	sr, err := signing.ComputeSigningRoot(deposit.Data, bytesutil.ToBytes(3, 32))
	require.NoError(t, err)
	sig := sk.Sign(sr[:])
	deposit.Data.Signature = sig.Marshal()
	leaf, err := deposit.Data.HashTreeRoot()
	require.NoError(t, err)

	// We then create a merkle branch for the test.
	depositTrie, err := trie.GenerateTrieFromItems([][]byte{leaf[:]}, params.BeaconConfig().DepositContractTreeDepth)
	require.NoError(t, err, "Could not generate trie")
	proof, err := depositTrie.MerkleProof(0)
	require.NoError(t, err, "Could not generate proof")

	deposit.Proof = proof
	registry := []*ethpb.Validator{
		{
			PublicKey: []byte{1, 2, 3},
		},
		{
			PublicKey:             sk.PublicKey().Marshal(),
			WithdrawalCredentials: []byte{1},
		},
	}
	balances := []uint64{0, 50}
	root, err := depositTrie.HashTreeRoot()
	require.NoError(t, err)

	beaconState, err := state_native.InitializeFromProtoPhase0(&ethpb.BeaconState{
		Validators: registry,
		Balances:   balances,
		Eth1Data: &ethpb.Eth1Data{
			DepositRoot: root[:],
			BlockHash:   root[:],
		},
	})
	require.NoError(t, err)
	newState, isNewValidator, err := blocks.ProcessDeposit(beaconState, deposit, true /*verifySignature*/)
	require.NoError(t, err, "Process deposit failed")
	assert.Equal(t, false, isNewValidator, "Expected isNewValidator to be false")
	assert.Equal(t, uint64(1000+50), newState.Balances()[1], "Expected balance at index 1 to be 1050")
}

func TestProcessDeposits_RepeatedDeposit_UpdateContract_NoContract_ZeroContract(t *testing.T) {
	sk, err := bls.RandKey()
	contract := params.BeaconConfig().ZeroContract
	require.NoError(t, err)
	deposit := &ethpb.Deposit{
		Data: &ethpb.Deposit_Data{
			PublicKey:             sk.PublicKey().Marshal(),
			Contract:              contract[:],
			Amount:                1000,
			WithdrawalCredentials: make([]byte, 32),
			Signature:             make([]byte, fieldparams.BLSSignatureLength),
		},
	}
	sr, err := signing.ComputeSigningRoot(deposit.Data, bytesutil.ToBytes(3, 32))
	require.NoError(t, err)
	sig := sk.Sign(sr[:])
	deposit.Data.Signature = sig.Marshal()
	leaf, err := deposit.Data.HashTreeRoot()
	require.NoError(t, err)

	// We then create a merkle branch for the test.
	depositTrie, err := trie.GenerateTrieFromItems([][]byte{leaf[:]}, params.BeaconConfig().DepositContractTreeDepth)
	require.NoError(t, err, "Could not generate trie")
	proof, err := depositTrie.MerkleProof(0)
	require.NoError(t, err, "Could not generate proof")

	deposit.Proof = proof
	b := util.NewBeaconBlock()
	b.Block = &ethpb.BeaconBlock{
		Body: &ethpb.BeaconBlockBody{
			Deposits: []*ethpb.Deposit{deposit},
		},
	}
	registry := []*ethpb.Validator{
		{
			PublicKey: []byte{1, 2, 3},
		},
		{
			PublicKey:             sk.PublicKey().Marshal(),
			Contract:              params.BeaconConfig().ZeroContract[:],
			WithdrawalCredentials: []byte{1},
		},
	}
	balances := []uint64{0, 50}
	root, err := depositTrie.HashTreeRoot()
	require.NoError(t, err)
	beaconState, err := state_native.InitializeFromProtoPhase0(&ethpb.BeaconState{
		Validators: registry,
		Balances:   balances,
		Eth1Data: &ethpb.Eth1Data{
			DepositRoot: root[:],
			BlockHash:   root[:],
		},
	})
	require.NoError(t, err)
	newState, err := blocks.ProcessDeposits(context.Background(), beaconState, b.Block.Body.Deposits)
	require.NoError(t, err, "Process deposit failed")
	assert.Equal(t, uint64(1000+50), newState.Balances()[1], "Expected balance at index 1 to be 1050")
	assert.DeepEqual(t, contract[:], newState.Validators()[1].Contract, "Expected contracts to be equal")
	index, exist := newState.ValidatorIndexByContract(contract)
	assert.Equal(t, false, exist)
	assert.Equal(t, primitives.ValidatorIndex(0), index)
}

func TestProcessDeposits_RepeatedDeposit_UpdateContract_HasContract_ZeroContract(t *testing.T) {
	sk, err := bls.RandKey()
	contract := bytesutil.ToBytes20([]byte("some_contract"))
	require.NoError(t, err)
	deposit := &ethpb.Deposit{
		Data: &ethpb.Deposit_Data{
			PublicKey:             sk.PublicKey().Marshal(),
			Contract:              params.BeaconConfig().ZeroContract[:],
			Amount:                1000,
			WithdrawalCredentials: make([]byte, 32),
			Signature:             make([]byte, fieldparams.BLSSignatureLength),
		},
	}
	sr, err := signing.ComputeSigningRoot(deposit.Data, bytesutil.ToBytes(3, 32))
	require.NoError(t, err)
	sig := sk.Sign(sr[:])
	deposit.Data.Signature = sig.Marshal()
	leaf, err := deposit.Data.HashTreeRoot()
	require.NoError(t, err)

	// We then create a merkle branch for the test.
	depositTrie, err := trie.GenerateTrieFromItems([][]byte{leaf[:]}, params.BeaconConfig().DepositContractTreeDepth)
	require.NoError(t, err, "Could not generate trie")
	proof, err := depositTrie.MerkleProof(0)
	require.NoError(t, err, "Could not generate proof")

	deposit.Proof = proof
	b := util.NewBeaconBlock()
	b.Block = &ethpb.BeaconBlock{
		Body: &ethpb.BeaconBlockBody{
			Deposits: []*ethpb.Deposit{deposit},
		},
	}
	registry := []*ethpb.Validator{
		{
			PublicKey: []byte{1, 2, 3},
		},
		{
			PublicKey:             sk.PublicKey().Marshal(),
			Contract:              contract[:],
			WithdrawalCredentials: []byte{1},
		},
	}
	balances := []uint64{0, 50}
	root, err := depositTrie.HashTreeRoot()
	require.NoError(t, err)
	beaconState, err := state_native.InitializeFromProtoPhase0(&ethpb.BeaconState{
		Validators: registry,
		Balances:   balances,
		Eth1Data: &ethpb.Eth1Data{
			DepositRoot: root[:],
			BlockHash:   root[:],
		},
	})
	require.NoError(t, err)
	newState, err := blocks.ProcessDeposits(context.Background(), beaconState, b.Block.Body.Deposits)
	require.NoError(t, err, "Process deposit failed")
	assert.Equal(t, uint64(1000+50), newState.Balances()[1], "Expected balance at index 1 to be 1050")
	assert.DeepEqual(t, contract[:], newState.Validators()[1].Contract, "Expected contracts to be equal")
	index, exist := newState.ValidatorIndexByContract(contract)
	assert.Equal(t, true, exist)
	assert.Equal(t, primitives.ValidatorIndex(1), index)
}

func TestProcessDeposits_RepeatedDeposit_UpdateContract_NoContract_NewContract(t *testing.T) {
	sk, err := bls.RandKey()
	contract := bytesutil.ToBytes20([]byte("some_contract"))
	require.NoError(t, err)
	deposit := &ethpb.Deposit{
		Data: &ethpb.Deposit_Data{
			PublicKey:             sk.PublicKey().Marshal(),
			Contract:              contract[:],
			Amount:                1000,
			WithdrawalCredentials: make([]byte, 32),
			Signature:             make([]byte, fieldparams.BLSSignatureLength),
		},
	}
	sr, err := signing.ComputeSigningRoot(deposit.Data, bytesutil.ToBytes(3, 32))
	require.NoError(t, err)
	sig := sk.Sign(sr[:])
	deposit.Data.Signature = sig.Marshal()
	leaf, err := deposit.Data.HashTreeRoot()
	require.NoError(t, err)

	// We then create a merkle branch for the test.
	depositTrie, err := trie.GenerateTrieFromItems([][]byte{leaf[:]}, params.BeaconConfig().DepositContractTreeDepth)
	require.NoError(t, err, "Could not generate trie")
	proof, err := depositTrie.MerkleProof(0)
	require.NoError(t, err, "Could not generate proof")

	deposit.Proof = proof
	b := util.NewBeaconBlock()
	b.Block = &ethpb.BeaconBlock{
		Body: &ethpb.BeaconBlockBody{
			Deposits: []*ethpb.Deposit{deposit},
		},
	}
	registry := []*ethpb.Validator{
		{
			PublicKey: []byte{1, 2, 3},
		},
		{
			PublicKey:             sk.PublicKey().Marshal(),
			Contract:              params.BeaconConfig().ZeroContract[:],
			WithdrawalCredentials: []byte{1},
		},
	}
	balances := []uint64{0, 50}
	root, err := depositTrie.HashTreeRoot()
	require.NoError(t, err)
	beaconState, err := state_native.InitializeFromProtoPhase0(&ethpb.BeaconState{
		Validators: registry,
		Balances:   balances,
		Eth1Data: &ethpb.Eth1Data{
			DepositRoot: root[:],
			BlockHash:   root[:],
		},
	})
	require.NoError(t, err)
	newState, err := blocks.ProcessDeposits(context.Background(), beaconState, b.Block.Body.Deposits)
	require.NoError(t, err, "Process deposit failed")
	assert.Equal(t, uint64(1000+50), newState.Balances()[1], "Expected balance at index 1 to be 1050")
	assert.DeepEqual(t, contract[:], newState.Validators()[1].Contract, "Expected contracts to be equal")
	index, exist := newState.ValidatorIndexByContract(contract)
	assert.Equal(t, true, exist)
	assert.Equal(t, primitives.ValidatorIndex(1), index)
}

func TestProcessDeposits_RepeatedDeposit_UpdateContract_NoContract_ContractExist(t *testing.T) {
	sk, err := bls.RandKey()
	contract := bytesutil.ToBytes20([]byte("some_contract"))
	require.NoError(t, err)
	deposit := &ethpb.Deposit{
		Data: &ethpb.Deposit_Data{
			PublicKey:             sk.PublicKey().Marshal(),
			Contract:              contract[:],
			Amount:                1000,
			WithdrawalCredentials: make([]byte, 32),
			Signature:             make([]byte, fieldparams.BLSSignatureLength),
		},
	}
	sr, err := signing.ComputeSigningRoot(deposit.Data, bytesutil.ToBytes(3, 32))
	require.NoError(t, err)
	sig := sk.Sign(sr[:])
	deposit.Data.Signature = sig.Marshal()
	leaf, err := deposit.Data.HashTreeRoot()
	require.NoError(t, err)

	// We then create a merkle branch for the test.
	depositTrie, err := trie.GenerateTrieFromItems([][]byte{leaf[:]}, params.BeaconConfig().DepositContractTreeDepth)
	require.NoError(t, err, "Could not generate trie")
	proof, err := depositTrie.MerkleProof(0)
	require.NoError(t, err, "Could not generate proof")

	deposit.Proof = proof
	b := util.NewBeaconBlock()
	b.Block = &ethpb.BeaconBlock{
		Body: &ethpb.BeaconBlockBody{
			Deposits: []*ethpb.Deposit{deposit},
		},
	}
	registry := []*ethpb.Validator{
		{
			Contract:  contract[:],
			PublicKey: []byte{1, 2, 3},
		},
		{
			PublicKey:             sk.PublicKey().Marshal(),
			Contract:              params.BeaconConfig().ZeroContract[:],
			WithdrawalCredentials: []byte{1},
		},
	}
	balances := []uint64{0, 50}
	root, err := depositTrie.HashTreeRoot()
	require.NoError(t, err)
	beaconState, err := state_native.InitializeFromProtoPhase0(&ethpb.BeaconState{
		Validators: registry,
		Balances:   balances,
		Eth1Data: &ethpb.Eth1Data{
			DepositRoot: root[:],
			BlockHash:   root[:],
		},
	})
	require.NoError(t, err)
	newState, err := blocks.ProcessDeposits(context.Background(), beaconState, b.Block.Body.Deposits)
	require.NoError(t, err, "Process deposit failed")
	assert.Equal(t, uint64(1000+50), newState.Balances()[1], "Expected balance at index 1 to be 1050")
	assert.DeepEqual(t, params.BeaconConfig().ZeroContract[:], newState.Validators()[1].Contract, "Expected contract to be equal ZeroContract")
	index, exist := newState.ValidatorIndexByContract(contract)
	assert.Equal(t, true, exist)
	assert.Equal(t, primitives.ValidatorIndex(0), index)
}

func TestProcessDeposits_RepeatedDeposit_UpdateContract_HasContract_ContractExist(t *testing.T) {
	sk, err := bls.RandKey()
	existed := bytesutil.ToBytes20([]byte("existed_contract"))
	contract := bytesutil.ToBytes20([]byte("some_contract"))
	require.NoError(t, err)
	deposit := &ethpb.Deposit{
		Data: &ethpb.Deposit_Data{
			PublicKey:             sk.PublicKey().Marshal(),
			Contract:              existed[:],
			Amount:                1000,
			WithdrawalCredentials: make([]byte, 32),
			Signature:             make([]byte, fieldparams.BLSSignatureLength),
		},
	}
	sr, err := signing.ComputeSigningRoot(deposit.Data, bytesutil.ToBytes(3, 32))
	require.NoError(t, err)
	sig := sk.Sign(sr[:])
	deposit.Data.Signature = sig.Marshal()
	leaf, err := deposit.Data.HashTreeRoot()
	require.NoError(t, err)

	// We then create a merkle branch for the test.
	depositTrie, err := trie.GenerateTrieFromItems([][]byte{leaf[:]}, params.BeaconConfig().DepositContractTreeDepth)
	require.NoError(t, err, "Could not generate trie")
	proof, err := depositTrie.MerkleProof(0)
	require.NoError(t, err, "Could not generate proof")

	deposit.Proof = proof
	b := util.NewBeaconBlock()
	b.Block = &ethpb.BeaconBlock{
		Body: &ethpb.BeaconBlockBody{
			Deposits: []*ethpb.Deposit{deposit},
		},
	}
	registry := []*ethpb.Validator{
		{
			Contract:  existed[:],
			PublicKey: []byte{1, 2, 3},
		},
		{
			PublicKey:             sk.PublicKey().Marshal(),
			Contract:              contract[:],
			WithdrawalCredentials: []byte{1},
		},
	}
	balances := []uint64{0, 50}
	root, err := depositTrie.HashTreeRoot()
	require.NoError(t, err)
	beaconState, err := state_native.InitializeFromProtoPhase0(&ethpb.BeaconState{
		Validators: registry,
		Balances:   balances,
		Eth1Data: &ethpb.Eth1Data{
			DepositRoot: root[:],
			BlockHash:   root[:],
		},
	})
	require.NoError(t, err)
	newState, err := blocks.ProcessDeposits(context.Background(), beaconState, b.Block.Body.Deposits)
	require.NoError(t, err, "Process deposit failed")
	assert.Equal(t, uint64(1000+50), newState.Balances()[1], "Expected balance at index 1 to be 1050")
	assert.DeepEqual(t, contract[:], newState.Validators()[1].Contract, "Expected contract to be equal ZeroContract")
	index, exist := newState.ValidatorIndexByContract(contract)
	assert.Equal(t, true, exist)
	assert.Equal(t, primitives.ValidatorIndex(1), index)
}

func TestProcessDeposits_RepeatedDeposit_UpdateContract_HasContract_SameContract(t *testing.T) {
	sk, err := bls.RandKey()
	oldContract := bytesutil.ToBytes20([]byte("old_contract"))
	require.NoError(t, err)
	deposit := &ethpb.Deposit{
		Data: &ethpb.Deposit_Data{
			PublicKey:             sk.PublicKey().Marshal(),
			Contract:              oldContract[:],
			Amount:                1000,
			WithdrawalCredentials: make([]byte, 32),
			Signature:             make([]byte, fieldparams.BLSSignatureLength),
		},
	}
	sr, err := signing.ComputeSigningRoot(deposit.Data, bytesutil.ToBytes(3, 32))
	require.NoError(t, err)
	sig := sk.Sign(sr[:])
	deposit.Data.Signature = sig.Marshal()
	leaf, err := deposit.Data.HashTreeRoot()
	require.NoError(t, err)

	// We then create a merkle branch for the test.
	depositTrie, err := trie.GenerateTrieFromItems([][]byte{leaf[:]}, params.BeaconConfig().DepositContractTreeDepth)
	require.NoError(t, err, "Could not generate trie")
	proof, err := depositTrie.MerkleProof(0)
	require.NoError(t, err, "Could not generate proof")

	deposit.Proof = proof
	b := util.NewBeaconBlock()
	b.Block = &ethpb.BeaconBlock{
		Body: &ethpb.BeaconBlockBody{
			Deposits: []*ethpb.Deposit{deposit},
		},
	}
	registry := []*ethpb.Validator{
		{
			PublicKey: []byte{1, 2, 3},
		},
		{
			PublicKey:             sk.PublicKey().Marshal(),
			Contract:              oldContract[:],
			WithdrawalCredentials: []byte{1},
		},
	}
	balances := []uint64{0, 50}
	root, err := depositTrie.HashTreeRoot()
	require.NoError(t, err)
	beaconState, err := state_native.InitializeFromProtoPhase0(&ethpb.BeaconState{
		Validators: registry,
		Balances:   balances,
		Eth1Data: &ethpb.Eth1Data{
			DepositRoot: root[:],
			BlockHash:   root[:],
		},
	})
	require.NoError(t, err)
	newState, err := blocks.ProcessDeposits(context.Background(), beaconState, b.Block.Body.Deposits)
	require.NoError(t, err, "Process deposit failed")
	assert.Equal(t, uint64(1000+50), newState.Balances()[1], "Expected balance at index 1 to be 1050")
	assert.DeepEqual(t, oldContract[:], newState.Validators()[1].Contract, "Expected contract to be equal ZeroContract")
	index, exist := newState.ValidatorIndexByContract(oldContract)
	assert.Equal(t, true, exist)
	assert.Equal(t, primitives.ValidatorIndex(1), index)
}

func TestProcessDeposits_RepeatedDeposit_UpdateContract_NoContract_PrevOwnerExited(t *testing.T) {
	sk, err := bls.RandKey()
	contract := bytesutil.ToBytes20([]byte("some_contract"))
	require.NoError(t, err)
	deposit := &ethpb.Deposit{
		Data: &ethpb.Deposit_Data{
			PublicKey:             sk.PublicKey().Marshal(),
			Contract:              contract[:],
			Amount:                1000,
			WithdrawalCredentials: make([]byte, 32),
			Signature:             make([]byte, fieldparams.BLSSignatureLength),
		},
	}
	sr, err := signing.ComputeSigningRoot(deposit.Data, bytesutil.ToBytes(3, 32))
	require.NoError(t, err)
	sig := sk.Sign(sr[:])
	deposit.Data.Signature = sig.Marshal()
	leaf, err := deposit.Data.HashTreeRoot()
	require.NoError(t, err)

	// We then create a merkle branch for the test.
	depositTrie, err := trie.GenerateTrieFromItems([][]byte{leaf[:]}, params.BeaconConfig().DepositContractTreeDepth)
	require.NoError(t, err, "Could not generate trie")
	proof, err := depositTrie.MerkleProof(0)
	require.NoError(t, err, "Could not generate proof")

	deposit.Proof = proof
	b := util.NewBeaconBlock()
	b.Block = &ethpb.BeaconBlock{
		Body: &ethpb.BeaconBlockBody{
			Deposits: []*ethpb.Deposit{deposit},
		},
	}
	registry := []*ethpb.Validator{
		{
			Contract:  contract[:],
			PublicKey: []byte{1, 2, 3},
			ExitEpoch: 1,
		},
		{
			PublicKey:             sk.PublicKey().Marshal(),
			Contract:              params.BeaconConfig().ZeroContract[:],
			WithdrawalCredentials: []byte{1},
		},
	}
	balances := []uint64{0, 50}
	root, err := depositTrie.HashTreeRoot()
	require.NoError(t, err)
	beaconState, err := state_native.InitializeFromProtoPhase0(&ethpb.BeaconState{
		Validators: registry,
		Balances:   balances,
		Eth1Data: &ethpb.Eth1Data{
			DepositRoot: root[:],
			BlockHash:   root[:],
		},
	})
	require.NoError(t, err)
	require.NoError(t, beaconState.SetSlot(params.BeaconConfig().SlotsPerEpoch.Mul(3)))
	newState, err := blocks.ProcessDeposits(context.Background(), beaconState, b.Block.Body.Deposits)
	require.NoError(t, err, "Process deposit failed")
	assert.Equal(t, uint64(1000+50), newState.Balances()[1], "Expected balance at index 1 to be 1050")
	assert.DeepEqual(t, contract[:], newState.Validators()[1].Contract, "Expected contract to be equal ZeroContract")
	index, exist := newState.ValidatorIndexByContract(contract)
	assert.Equal(t, true, exist)
	assert.Equal(t, primitives.ValidatorIndex(1), index)
}

func TestProcessDeposits_RepeatedDeposit_UpdateContract_HasContract_PrevOwnerExited(t *testing.T) {
	sk, err := bls.RandKey()
	oldContract := bytesutil.ToBytes20([]byte("old_contract"))
	contract := bytesutil.ToBytes20([]byte("some_contract"))
	require.NoError(t, err)
	deposit := &ethpb.Deposit{
		Data: &ethpb.Deposit_Data{
			PublicKey:             sk.PublicKey().Marshal(),
			Contract:              contract[:],
			Amount:                1000,
			WithdrawalCredentials: make([]byte, 32),
			Signature:             make([]byte, fieldparams.BLSSignatureLength),
		},
	}
	sr, err := signing.ComputeSigningRoot(deposit.Data, bytesutil.ToBytes(3, 32))
	require.NoError(t, err)
	sig := sk.Sign(sr[:])
	deposit.Data.Signature = sig.Marshal()
	leaf, err := deposit.Data.HashTreeRoot()
	require.NoError(t, err)

	// We then create a merkle branch for the test.
	depositTrie, err := trie.GenerateTrieFromItems([][]byte{leaf[:]}, params.BeaconConfig().DepositContractTreeDepth)
	require.NoError(t, err, "Could not generate trie")
	proof, err := depositTrie.MerkleProof(0)
	require.NoError(t, err, "Could not generate proof")

	deposit.Proof = proof
	b := util.NewBeaconBlock()
	b.Block = &ethpb.BeaconBlock{
		Body: &ethpb.BeaconBlockBody{
			Deposits: []*ethpb.Deposit{deposit},
		},
	}
	registry := []*ethpb.Validator{
		{
			Contract:  contract[:],
			PublicKey: []byte{1, 2, 3},
			ExitEpoch: 1,
		},
		{
			PublicKey:             sk.PublicKey().Marshal(),
			Contract:              oldContract[:],
			WithdrawalCredentials: []byte{1},
		},
	}
	balances := []uint64{0, 50}
	root, err := depositTrie.HashTreeRoot()
	require.NoError(t, err)
	beaconState, err := state_native.InitializeFromProtoPhase0(&ethpb.BeaconState{
		Validators: registry,
		Balances:   balances,
		Eth1Data: &ethpb.Eth1Data{
			DepositRoot: root[:],
			BlockHash:   root[:],
		},
	})
	require.NoError(t, err)
	require.NoError(t, beaconState.SetSlot(params.BeaconConfig().SlotsPerEpoch.Mul(3)))
	newState, err := blocks.ProcessDeposits(context.Background(), beaconState, b.Block.Body.Deposits)
	require.NoError(t, err, "Process deposit failed")
	assert.Equal(t, uint64(1000+50), newState.Balances()[1], "Expected balance at index 1 to be 1050")
	assert.DeepEqual(t, oldContract[:], newState.Validators()[1].Contract, "Expected contract to be equal ZeroContract")
	index, exist := newState.ValidatorIndexByContract(oldContract)
	assert.Equal(t, true, exist)
	assert.Equal(t, primitives.ValidatorIndex(1), index)
	index, exist = newState.ValidatorIndexByContract(contract)
	assert.Equal(t, true, exist)
	assert.Equal(t, primitives.ValidatorIndex(0), index)
}

func TestProcessDeposits_RepeatedDeposit_UpdateContract_HasContract_NewContract(t *testing.T) {
	sk, err := bls.RandKey()
	oldContract := bytesutil.ToBytes20([]byte("old_contract"))
	newContract := bytesutil.ToBytes20([]byte("new_contract"))
	require.NoError(t, err)
	deposit := &ethpb.Deposit{
		Data: &ethpb.Deposit_Data{
			PublicKey:             sk.PublicKey().Marshal(),
			Contract:              newContract[:],
			Amount:                1000,
			WithdrawalCredentials: make([]byte, 32),
			Signature:             make([]byte, fieldparams.BLSSignatureLength),
		},
	}
	sr, err := signing.ComputeSigningRoot(deposit.Data, bytesutil.ToBytes(3, 32))
	require.NoError(t, err)
	sig := sk.Sign(sr[:])
	deposit.Data.Signature = sig.Marshal()
	leaf, err := deposit.Data.HashTreeRoot()
	require.NoError(t, err)

	// We then create a merkle branch for the test.
	depositTrie, err := trie.GenerateTrieFromItems([][]byte{leaf[:]}, params.BeaconConfig().DepositContractTreeDepth)
	require.NoError(t, err, "Could not generate trie")
	proof, err := depositTrie.MerkleProof(0)
	require.NoError(t, err, "Could not generate proof")

	deposit.Proof = proof
	b := util.NewBeaconBlock()
	b.Block = &ethpb.BeaconBlock{
		Body: &ethpb.BeaconBlockBody{
			Deposits: []*ethpb.Deposit{deposit},
		},
	}
	registry := []*ethpb.Validator{
		{
			PublicKey: []byte{1, 2, 3},
		},
		{
			PublicKey:             sk.PublicKey().Marshal(),
			Contract:              oldContract[:],
			WithdrawalCredentials: []byte{1},
		},
	}
	balances := []uint64{0, 50}
	root, err := depositTrie.HashTreeRoot()
	require.NoError(t, err)
	beaconState, err := state_native.InitializeFromProtoPhase0(&ethpb.BeaconState{
		Validators: registry,
		Balances:   balances,
		Eth1Data: &ethpb.Eth1Data{
			DepositRoot: root[:],
			BlockHash:   root[:],
		},
	})
	require.NoError(t, err)
	newState, err := blocks.ProcessDeposits(context.Background(), beaconState, b.Block.Body.Deposits)
	require.NoError(t, err, "Process deposit failed")
	assert.Equal(t, uint64(1000+50), newState.Balances()[1], "Expected balance at index 1 to be 1050")
	assert.DeepEqual(t, oldContract[:], newState.Validators()[1].Contract, "Expected contract to be equal ZeroContract")
	index, exist := newState.ValidatorIndexByContract(oldContract)
	assert.Equal(t, true, exist)
	assert.Equal(t, primitives.ValidatorIndex(1), index)
}
