package beacon

import (
	"fmt"
	corehelpers "github.com/prysmaticlabs/prysm/v4/beacon-chain/core/helpers"
	"github.com/prysmaticlabs/prysm/v4/config/params"
	"net/http"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gorilla/mux"
	"github.com/prysmaticlabs/prysm/v4/beacon-chain/rpc/eth/helpers"
	"github.com/prysmaticlabs/prysm/v4/beacon-chain/rpc/eth/shared"
	"github.com/prysmaticlabs/prysm/v4/beacon-chain/state"
	statenative "github.com/prysmaticlabs/prysm/v4/beacon-chain/state/state-native"
	fieldparams "github.com/prysmaticlabs/prysm/v4/config/fieldparams"
	"github.com/prysmaticlabs/prysm/v4/consensus-types/primitives"
	"github.com/prysmaticlabs/prysm/v4/consensus-types/validator"
	"github.com/prysmaticlabs/prysm/v4/encoding/bytesutil"
	http2 "github.com/prysmaticlabs/prysm/v4/network/http"
	"github.com/prysmaticlabs/prysm/v4/time/slots"
	"go.opencensus.io/trace"
)

// todo unit act
// GetValidators returns filterable list of validators with their balance, status and index.
func (s *Server) GetValidators(w http.ResponseWriter, r *http.Request) {
	ctx, span := trace.StartSpan(r.Context(), "beacon.GetValidators")
	defer span.End()

	stateId := mux.Vars(r)["state_id"]
	if stateId == "" {
		http2.HandleError(w, "state_id is required in URL params", http.StatusBadRequest)
		return
	}
	st, err := s.Stater.State(ctx, []byte(stateId))
	if err != nil {
		shared.WriteStateFetchError(w, err)
		return
	}

	isOptimistic, err := helpers.IsOptimistic(ctx, []byte(stateId), s.OptimisticModeFetcher, s.Stater, s.ChainInfoFetcher, s.BeaconDB)
	if err != nil {
		http2.HandleError(w, "Could not check optimistic status: "+err.Error(), http.StatusInternalServerError)
		return
	}
	blockRoot, err := st.LatestBlockHeader().HashTreeRoot()
	if err != nil {
		http2.HandleError(w, "Could not calculate root of latest block header: "+err.Error(), http.StatusInternalServerError)
		return
	}
	isFinalized := s.FinalizationFetcher.IsFinalized(ctx, blockRoot)

	rawIds := r.URL.Query()["id"]
	ids, ok := decodeIds(w, st, rawIds, true /* ignore unknown */)
	if !ok {
		return
	}
	// return no data if all IDs are ignored
	if len(rawIds) > 0 && len(ids) == 0 {
		resp := &GetValidatorsResponse{
			Data:                []*ValidatorContainer{},
			ExecutionOptimistic: isOptimistic,
			Finalized:           isFinalized,
		}
		http2.WriteJson(w, resp)
		return
	}

	readOnlyVals, ok := valsFromIds(w, st, ids)
	if !ok {
		return
	}
	epoch := slots.ToEpoch(st.Slot())
	allBalances := st.Balances()
	allActivities := st.Activities()

	statuses := r.URL.Query()["status"]
	for i, ss := range statuses {
		statuses[i] = strings.ToLower(ss)
	}

	// Exit early if no matching validators were found or we don't want to further filter validators by status.
	if len(readOnlyVals) == 0 || len(statuses) == 0 {
		containers := make([]*ValidatorContainer, len(readOnlyVals))
		for i, val := range readOnlyVals {
			valStatus, err := helpers.ValidatorSubStatus(val, epoch)
			if err != nil {
				http2.HandleError(w, "Could not get validator status: "+err.Error(), http.StatusInternalServerError)
				return
			}
			if len(ids) == 0 {
				containers[i] = valContainerFromReadOnlyVal(val, primitives.ValidatorIndex(i), allBalances[i], allActivities[i], valStatus)
			} else {
				containers[i] = valContainerFromReadOnlyVal(val, ids[i], allBalances[ids[i]], allActivities[ids[i]], valStatus)
			}
		}
		resp := &GetValidatorsResponse{
			Data:                containers,
			ExecutionOptimistic: isOptimistic,
			Finalized:           isFinalized,
		}
		http2.WriteJson(w, resp)
		return
	}

	filteredStatuses := make(map[validator.ValidatorStatus]bool, len(statuses))
	for _, ss := range statuses {
		ok, vs := validator.ValidatorStatusFromString(ss)
		if !ok {
			http2.HandleError(w, "Invalid status "+ss, http.StatusBadRequest)
			return
		}
		filteredStatuses[vs] = true
	}
	valContainers := make([]*ValidatorContainer, 0, len(readOnlyVals))
	for i, val := range readOnlyVals {
		valStatus, err := helpers.ValidatorStatus(val, epoch)
		if err != nil {
			http2.HandleError(w, "Could not get validator status: "+err.Error(), http.StatusInternalServerError)
			return
		}
		valSubStatus, err := helpers.ValidatorSubStatus(val, epoch)
		if err != nil {
			http2.HandleError(w, "Could not get validator status: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if filteredStatuses[valStatus] || filteredStatuses[valSubStatus] {
			var container *ValidatorContainer
			if len(ids) == 0 {
				container = valContainerFromReadOnlyVal(val, primitives.ValidatorIndex(i), allBalances[i], allActivities[i], valSubStatus)
			} else {
				container = valContainerFromReadOnlyVal(val, ids[i], allBalances[ids[i]], allActivities[ids[i]], valSubStatus)
			}
			valContainers = append(valContainers, container)
		}
	}

	resp := &GetValidatorsResponse{
		Data:                valContainers,
		ExecutionOptimistic: isOptimistic,
		Finalized:           isFinalized,
	}
	http2.WriteJson(w, resp)
}

// todo unit act
// GetValidator returns a validator specified by state and id or public key along with status and balance.
func (s *Server) GetValidator(w http.ResponseWriter, r *http.Request) {
	ctx, span := trace.StartSpan(r.Context(), "beacon.GetValidator")
	defer span.End()

	stateId := mux.Vars(r)["state_id"]
	if stateId == "" {
		http2.HandleError(w, "state_id is required in URL params", http.StatusBadRequest)
		return
	}
	valId := mux.Vars(r)["validator_id"]
	if valId == "" {
		http2.HandleError(w, "validator_id is required in URL params", http.StatusBadRequest)
		return
	}

	st, err := s.Stater.State(ctx, []byte(stateId))
	if err != nil {
		shared.WriteStateFetchError(w, err)
		return
	}
	ids, ok := decodeIds(w, st, []string{valId}, false /* ignore unknown */)
	if !ok {
		return
	}
	readOnlyVals, ok := valsFromIds(w, st, ids)
	if !ok {
		return
	}
	if len(ids) == 0 || len(readOnlyVals) == 0 {
		http2.HandleError(w, "No validator returned for the given ID", http.StatusInternalServerError)
		return
	}
	valSubStatus, err := helpers.ValidatorSubStatus(readOnlyVals[0], slots.ToEpoch(st.Slot()))
	if err != nil {
		http2.HandleError(w, "Could not get validator status: "+err.Error(), http.StatusInternalServerError)
		return
	}
	bal, err := st.BalanceAtIndex(ids[0])
	if err != nil {
		http2.HandleError(w, "Could not get validator balance: "+err.Error(), http.StatusInternalServerError)
		return
	}
	act, err := st.ActivityAtIndex(ids[0])
	if err != nil {
		http2.HandleError(w, "Could not get validator activity: "+err.Error(), http.StatusInternalServerError)
		return
	}
	container := valContainerFromReadOnlyVal(readOnlyVals[0], ids[0], bal, act, valSubStatus)

	isOptimistic, err := helpers.IsOptimistic(ctx, []byte(stateId), s.OptimisticModeFetcher, s.Stater, s.ChainInfoFetcher, s.BeaconDB)
	if err != nil {
		http2.HandleError(w, "Could not check optimistic status: "+err.Error(), http.StatusInternalServerError)
		return
	}
	blockRoot, err := st.LatestBlockHeader().HashTreeRoot()
	if err != nil {
		http2.HandleError(w, "Could not calculate root of latest block header: "+err.Error(), http.StatusInternalServerError)
		return
	}
	isFinalized := s.FinalizationFetcher.IsFinalized(ctx, blockRoot)

	resp := &GetValidatorResponse{
		Data:                container,
		ExecutionOptimistic: isOptimistic,
		Finalized:           isFinalized,
	}
	http2.WriteJson(w, resp)
}

// GetValidatorBalances returns a filterable list of validator balances.
func (bs *Server) GetValidatorBalances(w http.ResponseWriter, r *http.Request) {
	ctx, span := trace.StartSpan(r.Context(), "beacon.GetValidatorBalances")
	defer span.End()

	stateId := mux.Vars(r)["state_id"]
	if stateId == "" {
		http2.HandleError(w, "state_id is required in URL params", http.StatusBadRequest)
		return
	}
	st, err := bs.Stater.State(ctx, []byte(stateId))
	if err != nil {
		shared.WriteStateFetchError(w, err)
		return
	}

	isOptimistic, err := helpers.IsOptimistic(ctx, []byte(stateId), bs.OptimisticModeFetcher, bs.Stater, bs.ChainInfoFetcher, bs.BeaconDB)
	if err != nil {
		http2.HandleError(w, "Could not check optimistic status: "+err.Error(), http.StatusInternalServerError)
		return
	}
	blockRoot, err := st.LatestBlockHeader().HashTreeRoot()
	if err != nil {
		http2.HandleError(w, "Could not calculate root of latest block header: "+err.Error(), http.StatusInternalServerError)
		return
	}
	isFinalized := bs.FinalizationFetcher.IsFinalized(ctx, blockRoot)

	rawIds := r.URL.Query()["id"]
	ids, ok := decodeIds(w, st, rawIds, true /* ignore unknown */)
	if !ok {
		return
	}
	// return no data if all IDs are ignored
	if len(rawIds) > 0 && len(ids) == 0 {
		resp := &GetValidatorBalancesResponse{
			Data:                []*ValidatorBalance{},
			ExecutionOptimistic: isOptimistic,
			Finalized:           isFinalized,
		}
		http2.WriteJson(w, resp)
		return
	}

	bals := st.Balances()
	var valBalances []*ValidatorBalance
	if len(ids) == 0 {
		valBalances = make([]*ValidatorBalance, len(bals))
		for i, b := range bals {
			valBalances[i] = &ValidatorBalance{
				Index:   strconv.FormatUint(uint64(i), 10),
				Balance: strconv.FormatUint(b, 10),
			}
		}
	} else {
		valBalances = make([]*ValidatorBalance, len(ids))
		for i, id := range ids {
			valBalances[i] = &ValidatorBalance{
				Index:   strconv.FormatUint(uint64(id), 10),
				Balance: strconv.FormatUint(bals[id], 10),
			}
		}
	}

	resp := &GetValidatorBalancesResponse{
		Data:                valBalances,
		ExecutionOptimistic: isOptimistic,
		Finalized:           isFinalized,
	}
	http2.WriteJson(w, resp)
}

// todo unit act
// GetValidatorActivities returns a filterable list of validator activities.
func (bs *Server) GetValidatorActivities(w http.ResponseWriter, r *http.Request) {
	ctx, span := trace.StartSpan(r.Context(), "beacon.GetValidatorActivities")
	defer span.End()

	stateId := mux.Vars(r)["state_id"]
	if stateId == "" {
		http2.HandleError(w, "state_id is required in URL params", http.StatusBadRequest)
		return
	}
	st, err := bs.Stater.State(ctx, []byte(stateId))
	if err != nil {
		shared.WriteStateFetchError(w, err)
		return
	}

	isOptimistic, err := helpers.IsOptimistic(ctx, []byte(stateId), bs.OptimisticModeFetcher, bs.Stater, bs.ChainInfoFetcher, bs.BeaconDB)
	if err != nil {
		http2.HandleError(w, "Could not check optimistic status: "+err.Error(), http.StatusInternalServerError)
		return
	}
	blockRoot, err := st.LatestBlockHeader().HashTreeRoot()
	if err != nil {
		http2.HandleError(w, "Could not calculate root of latest block header: "+err.Error(), http.StatusInternalServerError)
		return
	}
	isFinalized := bs.FinalizationFetcher.IsFinalized(ctx, blockRoot)

	rawIds := r.URL.Query()["id"]
	ids, ok := decodeIds(w, st, rawIds, true /* ignore unknown */)
	if !ok {
		return
	}

	// return no data if all IDs are ignored
	if len(rawIds) > 0 && len(ids) == 0 {
		resp := &GetValidatorActivitiesResponse{
			Data:                []*ValidatorActivity{},
			ExecutionOptimistic: isOptimistic,
			Finalized:           isFinalized,
		}
		http2.WriteJson(w, resp)
		return
	}

	vals := st.Validators()
	var valActivities []*ValidatorActivity
	if len(ids) == 0 {
		valActivities = make([]*ValidatorActivity, len(vals))
		for i, val := range vals {
			valActivities[i] = &ValidatorActivity{
				Index:    strconv.FormatUint(uint64(i), 10),
				Activity: strconv.FormatUint(val.EffectiveActivity, 10),
			}
		}
	} else {
		valActivities = make([]*ValidatorActivity, len(ids))
		for i, id := range ids {
			valActivities[i] = &ValidatorActivity{
				Index:    strconv.FormatUint(uint64(id), 10),
				Activity: strconv.FormatUint(vals[id].EffectiveActivity, 10),
			}
		}
	}

	resp := &GetValidatorActivitiesResponse{
		Data:                valActivities,
		ExecutionOptimistic: isOptimistic,
		Finalized:           isFinalized,
	}
	http2.WriteJson(w, resp)
}

// GetValidatorPowers returns a filterable list of validator powers.
func (bs *Server) GetValidatorPowers(w http.ResponseWriter, r *http.Request) {
	ctx, span := trace.StartSpan(r.Context(), "beacon.GetValidatorActivities")
	defer span.End()

	stateId := mux.Vars(r)["state_id"]
	if stateId == "" {
		http2.HandleError(w, "state_id is required in URL params", http.StatusBadRequest)
		return
	}
	st, err := bs.Stater.State(ctx, []byte(stateId))
	if err != nil {
		shared.WriteStateFetchError(w, err)
		return
	}

	isOptimistic, err := helpers.IsOptimistic(ctx, []byte(stateId), bs.OptimisticModeFetcher, bs.Stater, bs.ChainInfoFetcher, bs.BeaconDB)
	if err != nil {
		http2.HandleError(w, "Could not check optimistic status: "+err.Error(), http.StatusInternalServerError)
		return
	}
	blockRoot, err := st.LatestBlockHeader().HashTreeRoot()
	if err != nil {
		http2.HandleError(w, "Could not calculate root of latest block header: "+err.Error(), http.StatusInternalServerError)
		return
	}
	isFinalized := bs.FinalizationFetcher.IsFinalized(ctx, blockRoot)

	rawIds := r.URL.Query()["id"]
	ids, ok := decodeIds(w, st, rawIds, true /* ignore unknown */)
	if !ok {
		return
	}

	// return no data if all IDs are ignored
	if len(rawIds) > 0 && len(ids) == 0 {
		resp := &GetValidatorPowersResponse{
			Data:                &ValidatorPowersContainer{},
			ExecutionOptimistic: isOptimistic,
			Finalized:           isFinalized,
		}
		http2.WriteJson(w, resp)
		return
	}

	maxEffectiveBalance := params.BeaconConfig().MaxEffectiveBalance / params.BeaconConfig().EffectiveBalanceIncrement
	var totalEffectivePower uint64
	activeIndices, err := corehelpers.ActiveValidatorIndices(ctx, st, slots.ToEpoch(st.Slot()))
	if err != nil {
		http2.HandleError(w, "Could not get active validators indices: "+err.Error(), http.StatusInternalServerError)
		return
	}

	length := uint64(len(activeIndices))
	if length == 0 {
		http2.HandleError(w, "Empty active indices list", http.StatusInternalServerError)
		return
	}

	sharedActivity := st.SharedActivity()
	if sharedActivity == nil {
		http2.HandleError(w, "Nil shared activity in state", http.StatusInternalServerError)
		return
	}
	transactionsGas := sharedActivity.TransactionsGasPerPeriod / length

	valPowers := make([]*ValidatorPower, len(activeIndices))

	for i, idx := range activeIndices {
		v, err := st.ValidatorAtIndexReadOnly(idx)
		if err != nil {
			http2.HandleError(w, "Could not get validator at index: "+err.Error(), http.StatusInternalServerError)
			return
		}
		effectiveBalance := v.EffectiveBalance() / params.BeaconConfig().EffectiveBalanceIncrement
		effectiveActivity := v.EffectiveActivity()
		power := effectiveActivity + transactionsGas
		effectivePower := power * effectiveBalance / maxEffectiveBalance
		valPowers[i] = &ValidatorPower{
			Index:          strconv.FormatUint(uint64(idx), 10),
			Power:          strconv.FormatUint(power, 10),
			EffectivePower: strconv.FormatUint(effectivePower, 10),
		}
		totalEffectivePower += effectivePower
	}

	data := &ValidatorPowersContainer{
		Powers:              valPowers,
		TotalEffectivePower: strconv.FormatUint(totalEffectivePower, 10),
	}

	resp := &GetValidatorPowersResponse{
		Data:                data,
		ExecutionOptimistic: isOptimistic,
		Finalized:           isFinalized,
	}
	http2.WriteJson(w, resp)
}

// decodeIds takes in a list of validator ID strings (as either a pubkey or a validator index)
// and returns the corresponding validator indices. It can be configured to ignore well-formed but unknown indices.
func decodeIds(w http.ResponseWriter, st state.BeaconState, rawIds []string, ignoreUnknown bool) ([]primitives.ValidatorIndex, bool) {
	ids := make([]primitives.ValidatorIndex, 0, len(rawIds))
	numVals := uint64(st.NumValidators())
	for _, rawId := range rawIds {
		pubkey, err := hexutil.Decode(rawId)
		if err == nil {
			if len(pubkey) != fieldparams.BLSPubkeyLength {
				http2.HandleError(w, fmt.Sprintf("Pubkey length is %d instead of %d", len(pubkey), fieldparams.BLSPubkeyLength), http.StatusBadRequest)
				return nil, false
			}
			valIndex, ok := st.ValidatorIndexByPubkey(bytesutil.ToBytes48(pubkey))
			if !ok {
				if ignoreUnknown {
					continue
				}
				http2.HandleError(w, fmt.Sprintf("Unknown pubkey %s", pubkey), http.StatusBadRequest)
				return nil, false
			}
			ids = append(ids, valIndex)
			continue
		}

		index, err := strconv.ParseUint(rawId, 10, 64)
		if err != nil {
			http2.HandleError(w, fmt.Sprintf("Invalid validator index %s", rawId), http.StatusBadRequest)
			return nil, false
		}
		if index >= numVals {
			if ignoreUnknown {
				continue
			}
			http2.HandleError(w, fmt.Sprintf("Invalid validator index %d", index), http.StatusBadRequest)
			return nil, false
		}
		ids = append(ids, primitives.ValidatorIndex(index))
	}
	return ids, true
}

// valsFromIds returns read-only validators based on the supplied validator indices.
func valsFromIds(w http.ResponseWriter, st state.BeaconState, ids []primitives.ValidatorIndex) ([]state.ReadOnlyValidator, bool) {
	var vals []state.ReadOnlyValidator
	if len(ids) == 0 {
		allVals := st.Validators()
		vals = make([]state.ReadOnlyValidator, len(allVals))
		for i, val := range allVals {
			readOnlyVal, err := statenative.NewValidator(val)
			if err != nil {
				http2.HandleError(w, "Could not convert validator: "+err.Error(), http.StatusInternalServerError)
				return nil, false
			}
			vals[i] = readOnlyVal
		}
	} else {
		vals = make([]state.ReadOnlyValidator, 0, len(ids))
		for _, id := range ids {
			val, err := st.ValidatorAtIndex(id)
			if err != nil {
				http2.HandleError(w, fmt.Sprintf("Could not get validator at index %d: %s", id, err.Error()), http.StatusInternalServerError)
				return nil, false
			}

			readOnlyVal, err := statenative.NewValidator(val)
			if err != nil {
				http2.HandleError(w, "Could not convert validator: "+err.Error(), http.StatusInternalServerError)
				return nil, false
			}
			vals = append(vals, readOnlyVal)
		}
	}

	return vals, true
}

// todo unit act
func valContainerFromReadOnlyVal(
	val state.ReadOnlyValidator,
	id primitives.ValidatorIndex,
	bal uint64,
	act uint64,
	valStatus validator.ValidatorStatus,
) *ValidatorContainer {
	pubkey := val.PublicKey()
	contract := val.Contract()
	return &ValidatorContainer{
		Index:    strconv.FormatUint(uint64(id), 10),
		Balance:  strconv.FormatUint(bal, 10),
		Activity: strconv.FormatUint(act, 10),
		Status:   valStatus.String(),
		Validator: &Validator{
			Pubkey:                     hexutil.Encode(pubkey[:]),
			WithdrawalCredentials:      hexutil.Encode(val.WithdrawalCredentials()),
			Contract:                   hexutil.Encode(contract[:]),
			EffectiveBalance:           strconv.FormatUint(val.EffectiveBalance(), 10),
			EffectiveActivity:          strconv.FormatUint(val.EffectiveActivity(), 10),
			Slashed:                    val.Slashed(),
			ActivationEligibilityEpoch: strconv.FormatUint(uint64(val.ActivationEligibilityEpoch()), 10),
			ActivationEpoch:            strconv.FormatUint(uint64(val.ActivationEpoch()), 10),
			ExitEpoch:                  strconv.FormatUint(uint64(val.ExitEpoch()), 10),
			WithdrawableEpoch:          strconv.FormatUint(uint64(val.WithdrawableEpoch()), 10),
		},
	}
}
