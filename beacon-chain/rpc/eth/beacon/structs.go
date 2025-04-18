package beacon

import (
	"github.com/prysmaticlabs/prysm/v4/beacon-chain/rpc/eth/shared"
)

type BlockRootResponse struct {
	Data *struct {
		Root string `json:"root"`
	} `json:"data"`
	ExecutionOptimistic bool `json:"execution_optimistic"`
	Finalized           bool `json:"finalized"`
}

type GetCommitteesResponse struct {
	Data                []*shared.Committee `json:"data"`
	ExecutionOptimistic bool                `json:"execution_optimistic"`
	Finalized           bool                `json:"finalized"`
}

type DepositContractResponse struct {
	Data *struct {
		ChainId string `json:"chain_id"`
		Address string `json:"address"`
	} `json:"data"`
}

type ListAttestationsResponse struct {
	Data []*shared.Attestation `json:"data"`
}

type SubmitAttestationsRequest struct {
	Data []*shared.Attestation `json:"data"`
}

type ListVoluntaryExitsResponse struct {
	Data []*shared.SignedVoluntaryExit `json:"data"`
}

type SubmitSyncCommitteeSignaturesRequest struct {
	Data []*shared.SyncCommitteeMessage `json:"data"`
}

type GetStateForkResponse struct {
	Data                *shared.Fork `json:"data"`
	ExecutionOptimistic bool         `json:"execution_optimistic"`
	Finalized           bool         `json:"finalized"`
}

type GetFinalityCheckpointsResponse struct {
	ExecutionOptimistic bool                 `json:"execution_optimistic"`
	Finalized           bool                 `json:"finalized"`
	Data                *FinalityCheckpoints `json:"data"`
}

type FinalityCheckpoints struct {
	PreviousJustified *shared.Checkpoint `json:"previous_justified"`
	CurrentJustified  *shared.Checkpoint `json:"current_justified"`
	Finalized         *shared.Checkpoint `json:"finalized"`
}

type GetGenesisResponse struct {
	Data *Genesis `json:"data"`
}

type Genesis struct {
	GenesisTime           string `json:"genesis_time"`
	GenesisValidatorsRoot string `json:"genesis_validators_root"`
	GenesisForkVersion    string `json:"genesis_fork_version"`
}

type GetBlockHeadersResponse struct {
	Data                []*shared.SignedBeaconBlockHeaderContainer `json:"data"`
	ExecutionOptimistic bool                                       `json:"execution_optimistic"`
	Finalized           bool                                       `json:"finalized"`
}

type GetBlockHeaderResponse struct {
	ExecutionOptimistic bool                                     `json:"execution_optimistic"`
	Finalized           bool                                     `json:"finalized"`
	Data                *shared.SignedBeaconBlockHeaderContainer `json:"data"`
}

type GetValidatorsResponse struct {
	ExecutionOptimistic bool                  `json:"execution_optimistic"`
	Finalized           bool                  `json:"finalized"`
	Data                []*ValidatorContainer `json:"data"`
}

type GetValidatorResponse struct {
	ExecutionOptimistic bool                `json:"execution_optimistic"`
	Finalized           bool                `json:"finalized"`
	Data                *ValidatorContainer `json:"data"`
}

type GetValidatorBalancesResponse struct {
	ExecutionOptimistic bool                `json:"execution_optimistic"`
	Finalized           bool                `json:"finalized"`
	Data                []*ValidatorBalance `json:"data"`
}

// todo unit act
type GetValidatorActivitiesResponse struct {
	ExecutionOptimistic bool                 `json:"execution_optimistic"`
	Finalized           bool                 `json:"finalized"`
	Data                []*ValidatorActivity `json:"data"`
}

type GetValidatorPowersResponse struct {
	ExecutionOptimistic bool                      `json:"execution_optimistic"`
	Finalized           bool                      `json:"finalized"`
	Data                *ValidatorPowersContainer `json:"data"`
}

type GetSharedActivitiesResponse struct {
	ExecutionOptimistic bool              `json:"execution_optimistic"`
	Finalized           bool              `json:"finalized"`
	Data                *SharedActivities `json:"data"`
}

// todo unit act
type ValidatorContainer struct {
	Index     string     `json:"index"`
	Balance   string     `json:"balance"`
	Activity  string     `json:"activity"`
	Status    string     `json:"status"`
	Validator *Validator `json:"validator"`
}

// todo unit act
type Validator struct {
	Pubkey                     string `json:"pubkey"`
	WithdrawalCredentials      string `json:"withdrawal_credentials"`
	Contract                   string `json:"contract,omitempty" ssz-size:"20"`
	EffectiveBalance           string `json:"effective_balance"`
	EffectiveActivity          string `json:"effective_activity"`
	Slashed                    bool   `json:"slashed"`
	ActivationEligibilityEpoch string `json:"activation_eligibility_epoch"`
	ActivationEpoch            string `json:"activation_epoch"`
	ExitEpoch                  string `json:"exit_epoch"`
	WithdrawableEpoch          string `json:"withdrawable_epoch"`
}

type ValidatorBalance struct {
	Index   string `json:"index"`
	Balance string `json:"balance"`
}

// todo unit act
type ValidatorActivity struct {
	Index    string `json:"index"`
	Activity string `json:"activity"`
}

type ValidatorPowersContainer struct {
	Powers              []*ValidatorPower `json:"powers"`
	TotalEffectivePower string            `json:"total_effective_power"`
}

type ValidatorPower struct {
	Index          string `json:"index"`
	Power          string `json:"power"`
	EffectivePower string `json:"effective_power"`
}

type SharedActivities struct {
	TransactionsGas uint64 `json:"transactions_gas,omitempty"`
	BaseFee         uint64 `json:"base_fee,omitempty"`
}
