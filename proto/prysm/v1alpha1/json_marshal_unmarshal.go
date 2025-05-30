// todo unit act
package eth

import (
	"encoding/json"
	"errors"

	"github.com/ethereum/go-ethereum/common"
)

// activityChangeJSON is the custom json representation of ActivityChange structure.
type activityChangeJSON struct {
	ContractAddress *common.Address `json:"contract_address"`
	DeltaActivtity  *uint64         `json:"delta_activity"`
}

// MarshalJSON --
func (c *ActivityChange) MarshalJSON() ([]byte, error) {
	contractAddress := common.BytesToAddress(c.ContractAddress)
	deltaActivity := c.DeltaActivity
	return json.Marshal(&activityChangeJSON{
		ContractAddress: &contractAddress,
		DeltaActivtity:  &deltaActivity,
	})
}

// UnmarshalJSON --
func (c *ActivityChange) UnmarshalJSON(enc []byte) error {
	dec := activityChangeJSON{}
	if err := json.Unmarshal(enc, &dec); err != nil {
		return err
	}

	if dec.ContractAddress == nil {
		return errors.New("missing required field 'contracts_address' for ActivityChange")
	}
	if dec.DeltaActivtity == nil {
		return errors.New("missing required field 'delta_activity' for ActivityChange")
	}

	*c = ActivityChange{}
	c.ContractAddress = dec.ContractAddress.Bytes()
	c.DeltaActivity = *(dec.DeltaActivtity)
	return nil
}

// BlockActivities is the response kind received by the eth_getBlockActivities
// endpoint via JSON-RPC.
type BlockActivities struct {
	BaseFee    uint64            `json:"baseFee"`
	TxCount    uint64            `json:"txCount"`
	Activities []*ActivityChange `json:"activities"`
}
