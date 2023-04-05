package validator_client_factory

import (
	"github.com/prysmaticlabs/prysm/v3/config/features"
	beaconApi "github.com/prysmaticlabs/prysm/v3/validator/client/beacon-api"
	grpcApi "github.com/prysmaticlabs/prysm/v3/validator/client/grpc-api"
	"github.com/prysmaticlabs/prysm/v3/validator/client/iface"
	validatorHelpers "github.com/prysmaticlabs/prysm/v3/validator/helpers"
)

func NewValidatorClient(validatorConn validatorHelpers.NodeConnection) iface.ValidatorClient {
	featureFlags := features.Get()

	if featureFlags.EnableBeaconRESTApi {
		return beaconApi.NewBeaconApiValidatorClient(validatorConn.GetBeaconApiUrl(), validatorConn.GetBeaconApiTimeout())
	} else {
		return grpcApi.NewGrpcValidatorClient(validatorConn.GetGrpcClientConn())
	}
}
