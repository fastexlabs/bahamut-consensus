package payloadattribute

import (
	"github.com/pkg/errors"
	"github.com/prysmaticlabs/prysm/v3/consensus-types/blocks"
	enginev1 "github.com/prysmaticlabs/prysm/v3/proto/engine/v1"
)

var (
	_ = Attributer(&data{})
)

type data struct {
	version               int
	timeStamp             uint64
	prevRandao            []byte
	suggestedFeeRecipient []byte
	withdrawals           []*enginev1.Withdrawal
}

var (
	errNilPayloadAttribute         = errors.New("received nil payload attribute")
	errUnsupportedPayloadAttribute = errors.New("unsupported payload attribute")
)

// New returns a new payload attribute with the given input object.
func New(i interface{}, v int) (Attributer, error) {
	switch a := i.(type) {
	case nil:
		return nil, blocks.ErrNilObject
	case *enginev1.PayloadAttributes:
		return initPayloadAttributeFromV1(a, v)
	case *enginev1.PayloadAttributesV2:
		return initPayloadAttributeFromV2(a, v)
	default:
		return nil, errors.Wrapf(errUnsupportedPayloadAttribute, "unable to create payload attribute from type %T", i)
	}
}

// EmptyWithVersion returns an empty payload attribute with the given version.
func EmptyWithVersion(version int) Attributer {
	return &data{
		version: version,
	}
}

func initPayloadAttributeFromV1(a *enginev1.PayloadAttributes, v int) (Attributer, error) {
	if a == nil {
		return nil, errNilPayloadAttribute
	}

	return &data{
		version:               v,
		prevRandao:            a.PrevRandao,
		timeStamp:             a.Timestamp,
		suggestedFeeRecipient: a.SuggestedFeeRecipient,
	}, nil
}

func initPayloadAttributeFromV2(a *enginev1.PayloadAttributesV2, v int) (Attributer, error) {
	if a == nil {
		return nil, errNilPayloadAttribute
	}

	return &data{
		version:               v,
		prevRandao:            a.PrevRandao,
		timeStamp:             a.Timestamp,
		suggestedFeeRecipient: a.SuggestedFeeRecipient,
		withdrawals:           a.Withdrawals,
	}, nil
}
