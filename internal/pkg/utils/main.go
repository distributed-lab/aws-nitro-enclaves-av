package utils

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/distributed-lab/aws-nitro-enclaves-av/internal/pkg/icrypto"
	"github.com/distributed-lab/enclave-extras/attestation"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

const DefaultPrimaryType string = "Register"

var DefaultFieldsToSign = []string{"pcr0", "public_key"}

var fieldToType = map[string]apitypes.Type{
	"public_key": {Name: "public_key", Type: "bytes"},
	"user_data":  {Name: "user_data", Type: "bytes"},
	"nonce":      {Name: "nonce", Type: "bytes"},
	"timestamp":  {Name: "timestamp", Type: "uint64"},
	"digest":     {Name: "digest", Type: "string"},
	"module_id":  {Name: "module_id", Type: "string"},
}

var (
	ErrAbsentField  = errors.New("field not present in attestation document")
	ErrInvalidField = errors.New("invalid attestation document field")
)

// fields must not have duplicate items
func BuildTypedDataAttestationMessage(attestationDocument *attestation.NSMAttestationDoc, primaryType string, fields []string) (*icrypto.Message, error) {
	if attestationDocument == nil {
		return nil, fmt.Errorf("attestation document shouldn't be nil")
	}

	dataTypes := make([]apitypes.Type, 0, len(fields))
	dataValues := make(apitypes.TypedDataMessage, len(fields))
	for _, field := range fields {
		if dataType, ok := fieldToType[field]; ok {
			dataTypes = append(dataTypes, dataType)
		}

		switch {
		case strings.HasPrefix(field, "pcr"):
			pcr, err := strconv.ParseUint(field[3:], 10, 5)
			if err != nil {
				return nil, fmt.Errorf("invalid attestation document pcr: %s", field)
			}

			pcrValue, ok := attestationDocument.PCRs[int(pcr)]
			if !ok {
				return nil, fmt.Errorf("%w: %s", ErrAbsentField, field)
			}

			dataTypes = append(dataTypes, apitypes.Type{Name: field, Type: "bytes"})
			dataValues[field] = pcrValue
		case field == "public_key":
			if attestationDocument.PublicKey == nil {
				return nil, fmt.Errorf("%w: %s", ErrAbsentField, field)
			}
			dataValues[field] = attestationDocument.PublicKey
		case field == "user_data":
			if attestationDocument.UserData == nil {
				return nil, fmt.Errorf("%w: %s", ErrAbsentField, field)
			}
			dataValues[field] = attestationDocument.UserData
		case field == "nonce":
			if attestationDocument.Nonce == nil {
				return nil, fmt.Errorf("%w: %s", ErrAbsentField, field)
			}
			dataValues[field] = attestationDocument.Nonce
		case field == "timestamp":
			dataValues[field] = uint64(attestationDocument.Timestamp.Unix())
		case field == "module_id":
			dataValues[field] = attestationDocument.ModuleID
		case field == "digest":
			dataValues[field] = attestationDocument.Digest
		default:
			return nil, fmt.Errorf("%w: %s", ErrInvalidField, field)
		}
	}

	return &icrypto.Message{
		TypedDataMessage: dataValues,
		DataTypes:        dataTypes,
		PrimaryType:      primaryType,
	}, nil
}

func AsPointer[T any](v T) *T {
	return &v
}
