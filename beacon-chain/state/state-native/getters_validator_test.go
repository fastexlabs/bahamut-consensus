package state_native_test

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/prysmaticlabs/prysm/v4/beacon-chain/state"
	statenative "github.com/prysmaticlabs/prysm/v4/beacon-chain/state/state-native"
	testtmpl "github.com/prysmaticlabs/prysm/v4/beacon-chain/state/testing"
	fieldparams "github.com/prysmaticlabs/prysm/v4/config/fieldparams"
	"github.com/prysmaticlabs/prysm/v4/consensus-types/primitives"
	"github.com/prysmaticlabs/prysm/v4/encoding/bytesutil"
	ethpb "github.com/prysmaticlabs/prysm/v4/proto/prysm/v1alpha1"
	"github.com/prysmaticlabs/prysm/v4/testing/require"
	"github.com/prysmaticlabs/prysm/v4/testing/util"
)

func TestBeaconState_ValidatorAtIndexReadOnly_HandlesNilSlice_Phase0(t *testing.T) {
	testtmpl.VerifyBeaconStateValidatorAtIndexReadOnlyHandlesNilSlice(t, func() (state.BeaconState, error) {
		return statenative.InitializeFromProtoUnsafePhase0(&ethpb.BeaconState{
			Validators: nil,
		})
	})
}

func TestBeaconState_ValidatorAtIndexReadOnly_HandlesNilSlice_Altair(t *testing.T) {
	testtmpl.VerifyBeaconStateValidatorAtIndexReadOnlyHandlesNilSlice(t, func() (state.BeaconState, error) {
		return statenative.InitializeFromProtoUnsafeAltair(&ethpb.BeaconStateAltair{
			Validators: nil,
		})
	})
}

func TestBeaconState_ValidatorAtIndexReadOnly_HandlesNilSlice_Bellatrix(t *testing.T) {
	testtmpl.VerifyBeaconStateValidatorAtIndexReadOnlyHandlesNilSlice(t, func() (state.BeaconState, error) {
		return statenative.InitializeFromProtoUnsafeBellatrix(&ethpb.BeaconStateBellatrix{
			Validators: nil,
		})
	})
}

func TestBeaconState_ValidatorAtIndexReadOnly_HandlesNilSlice_Capella(t *testing.T) {
	testtmpl.VerifyBeaconStateValidatorAtIndexReadOnlyHandlesNilSlice(t, func() (state.BeaconState, error) {
		return statenative.InitializeFromProtoUnsafeCapella(&ethpb.BeaconStateCapella{
			Validators: nil,
		})
	})
}

func TestBeaconState_ValidatorAtIndexReadOnly_HandlesNilSlice_Deneb(t *testing.T) {
	testtmpl.VerifyBeaconStateValidatorAtIndexReadOnlyHandlesNilSlice(t, func() (state.BeaconState, error) {
		return statenative.InitializeFromProtoUnsafeDeneb(&ethpb.BeaconStateDeneb{
			Validators: nil,
		})
	})
}

func TestValidatorIndexes(t *testing.T) {
	dState, _ := util.DeterministicGenesisState(t, 10)
	byteValue := dState.PubkeyAtIndex(1)
	t.Run("ValidatorIndexByPubkey", func(t *testing.T) {
		require.Equal(t, hexutil.Encode(byteValue[:]), "0xb89bebc699769726a318c8e9971bd3171297c61aea4a6578a7a4f94b547dcba5bac16a89108b6b6a1fe3695d1a874a0b")
	})
	t.Run("ValidatorAtIndexReadOnly", func(t *testing.T) {
		readOnlyState, err := dState.ValidatorAtIndexReadOnly(1)
		require.NoError(t, err)
		readOnlyBytes := readOnlyState.PublicKey()
		require.NotEmpty(t, readOnlyBytes)
		require.Equal(t, hexutil.Encode(readOnlyBytes[:]), hexutil.Encode(byteValue[:]))
	})
}

func TestContractIndexes(t *testing.T) {
	expecetedContracts := [][]byte{
		bytesutil.PadTo([]byte("contract_1"), 20),
		bytesutil.PadTo([]byte("contract_2"), 20),
		bytesutil.PadTo([]byte("contract_3"), 20),
	}
	dState, _ := util.DeterministicGenesisStateWithContracts(t, 3, expecetedContracts)
	t.Run("ValidatorIndexByContract", func(t *testing.T) {
		contract, ok := dState.ContractAtIndex(primitives.ValidatorIndex(1))
		require.Equal(t, true, ok)
		require.Equal(t, common.Bytes2Hex(expecetedContracts[1]), common.Bytes2Hex(contract[:]))

		// Not found case.
		_, ok = dState.ContractAtIndex(primitives.ValidatorIndex(4))
		require.Equal(t, false, ok)
	})

	t.Run("ValidatorIndexByContract", func(t *testing.T) {
		validatorIndex, ok := dState.ValidatorIndexByContract([fieldparams.ContractAddressLength]byte(expecetedContracts[2]))
		require.Equal(t, true, ok)
		require.Equal(t, primitives.ValidatorIndex(2), validatorIndex)

		// Not found case.
		_, ok = dState.ValidatorIndexByContract([fieldparams.ContractAddressLength]byte{'A'})
		require.Equal(t, false, ok)
	})
}
