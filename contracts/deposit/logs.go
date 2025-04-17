package deposit

import (
	"bytes"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/pkg/errors"
)

// todo unit act
// UnpackDepositLogData unpacks the data from a deposit log using the ABI decoder.
func UnpackDepositLogData(data []byte) (pubkey, withdrawalCredentials, contractAddress, amount, signature, index []byte, err error) {
	reader := bytes.NewReader([]byte(DepositContractABI))
	contractAbi, err := abi.JSON(reader)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, errors.Wrap(err, "unable to generate contract abi")
	}

	unpackedLogs, err := contractAbi.Unpack("DepositEvent", data)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, errors.Wrap(err, "unable to unpack logs")
	}

	return unpackedLogs[0].([]byte), unpackedLogs[1].([]byte), unpackedLogs[2].([]byte), unpackedLogs[3].([]byte), unpackedLogs[4].([]byte), unpackedLogs[5].([]byte), nil
}
