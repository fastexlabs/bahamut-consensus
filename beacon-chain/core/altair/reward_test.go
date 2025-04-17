package altair_test

import (
	mathC "github.com/prysmaticlabs/prysm/v4/math"
	"math"
	"testing"

	"github.com/prysmaticlabs/prysm/v4/beacon-chain/core/altair"
	"github.com/prysmaticlabs/prysm/v4/beacon-chain/core/helpers"
	"github.com/prysmaticlabs/prysm/v4/beacon-chain/state"
	"github.com/prysmaticlabs/prysm/v4/config/params"
	"github.com/prysmaticlabs/prysm/v4/consensus-types/primitives"
	"github.com/prysmaticlabs/prysm/v4/testing/require"
	"github.com/prysmaticlabs/prysm/v4/testing/util"
)

func Test_BaseReward(t *testing.T) {
	helpers.ClearCache()
	genState := func(valCount uint64) state.ReadOnlyBeaconState {
		s, _ := util.DeterministicGenesisStateAltair(t, valCount)
		return s
	}
	tests := []struct {
		name      string
		valIdx    primitives.ValidatorIndex
		st        state.ReadOnlyBeaconState
		want      uint64
		errString string
	}{
		{
			name:      "unknown validator",
			valIdx:    2,
			st:        genState(1),
			want:      0,
			errString: "validator index 2 does not exist",
		},
		{
			name:      "active balance is 8192ftn",
			valIdx:    0,
			st:        genState(1),
			want:      8192 * (1e9 * 156 / mathC.CachedSquareRoot(8192*1e9)),
			errString: "",
		},
		{
			name:      "active balance is 8192ftn * target committee size",
			valIdx:    0,
			st:        genState(params.BeaconConfig().TargetCommitteeSize),
			want:      8192 * (1e9 * 156 / mathC.CachedSquareRoot(8192*1e9*params.BeaconConfig().TargetCommitteeSize)),
			errString: "",
		},
		{
			name:      "active balance is 8192ftn * max validator per  committee size",
			valIdx:    0,
			st:        genState(params.BeaconConfig().MaxValidatorsPerCommittee),
			want:      8192 * (1e9 * 156 / mathC.CachedSquareRoot(8192*1e9*params.BeaconConfig().MaxValidatorsPerCommittee)),
			errString: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := altair.BaseReward(tt.st, tt.valIdx)
			if (err != nil) && (tt.errString != "") {
				require.ErrorContains(t, tt.errString, err)
				return
			}
			require.Equal(t, tt.want, got)
		})
	}
}

func Test_BaseRewardWithTotalBalance(t *testing.T) {
	helpers.ClearCache()
	s, _ := util.DeterministicGenesisStateAltair(t, 1)
	tests := []struct {
		name          string
		valIdx        primitives.ValidatorIndex
		activeBalance uint64
		want          uint64
		errString     string
	}{
		{
			name:          "active balance is 0",
			valIdx:        0,
			activeBalance: 0,
			want:          0,
			errString:     "active balance can't be 0",
		},
		{
			name:          "unknown validator",
			valIdx:        2,
			activeBalance: 1,
			want:          0,
			errString:     "validator index 2 does not exist",
		},
		{
			name:          "active balance is 1",
			valIdx:        0,
			activeBalance: 1,
			want:          8192 * (1e9 * 156 / 1),
			errString:     "",
		},
		{
			name:          "active balance is 1ftn",
			valIdx:        0,
			activeBalance: params.BeaconConfig().EffectiveBalanceIncrement,
			want:          8192 * (1e9 * 156 / mathC.CachedSquareRoot(params.BeaconConfig().EffectiveBalanceIncrement)),
			errString:     "",
		},
		{
			name:          "active balance is 8192ftn",
			valIdx:        0,
			activeBalance: params.BeaconConfig().MaxEffectiveBalance,
			want:          8192 * (1e9 * 156 / mathC.CachedSquareRoot(params.BeaconConfig().MaxEffectiveBalance)),
			errString:     "",
		},
		{
			name:          "active balance is 8192ftn * 1m validators",
			valIdx:        0,
			activeBalance: params.BeaconConfig().MaxEffectiveBalance * 1e9,
			want:          8192 * (1e9 * 156 / mathC.CachedSquareRoot(params.BeaconConfig().MaxEffectiveBalance*1e9)),
			errString:     "",
		},
		{
			name:          "active balance is max uint64",
			valIdx:        0,
			activeBalance: math.MaxUint64,
			want:          8192 * (1e9 * 156 / mathC.CachedSquareRoot(math.MaxUint64)),
			errString:     "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := altair.BaseRewardWithTotalBalance(s, tt.valIdx, tt.activeBalance)
			if (err != nil) && (tt.errString != "") {
				require.ErrorContains(t, tt.errString, err)
				return
			}
			require.Equal(t, tt.want, got)
		})
	}
}

func Test_BaseRewardPerIncrement(t *testing.T) {
	helpers.ClearCache()
	tests := []struct {
		name          string
		activeBalance uint64
		want          uint64
		errString     string
	}{
		{
			name:          "active balance is 0",
			activeBalance: 0,
			want:          0,
			errString:     "active balance can't be 0",
		},
		{
			name:          "active balance is 1",
			activeBalance: 1,
			want:          156 * 1e9 / 1,
			errString:     "",
		},
		{
			name:          "active balance is 1ftn ",
			activeBalance: params.BeaconConfig().EffectiveBalanceIncrement,
			want:          156 * 1e9 / mathC.CachedSquareRoot(params.BeaconConfig().EffectiveBalanceIncrement),
			errString:     "",
		},
		{
			name:          "active balance is 8192ftn",
			activeBalance: params.BeaconConfig().MaxEffectiveBalance,
			want:          156 * 1e9 / mathC.CachedSquareRoot(params.BeaconConfig().MaxEffectiveBalance),
			errString:     "",
		},
		{
			name:          "active balance is 8192ftn * 1m validators",
			activeBalance: params.BeaconConfig().MaxEffectiveBalance * 1e9,
			want:          156 * 1e9 / mathC.CachedSquareRoot(params.BeaconConfig().MaxEffectiveBalance*1e9),
			errString:     "",
		},
		{
			name:          "active balance is max uint64",
			activeBalance: math.MaxUint64,
			want:          156 * 1e9 / mathC.CachedSquareRoot(math.MaxUint64),
			errString:     "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := altair.BaseRewardPerIncrement(tt.activeBalance)
			if (err != nil) && (tt.errString != "") {
				require.ErrorContains(t, tt.errString, err)
				return
			}
			require.Equal(t, tt.want, got)
		})
	}
}
