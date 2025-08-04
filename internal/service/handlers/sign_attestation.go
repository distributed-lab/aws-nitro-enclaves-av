package handlers

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/distributed-lab/aws-nitro-enclaves-av/internal/pkg/icrypto"
	"github.com/distributed-lab/aws-nitro-enclaves-av/internal/service/requests"
	"github.com/distributed-lab/aws-nitro-enclaves-av/resources"
	"github.com/distributed-lab/enclave-extras/attestation"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
)

var fieldToType = map[string]apitypes.Type{
	"public_key": {Name: "public_key", Type: "bytes"},
	"user_data":  {Name: "user_data", Type: "bytes"},
	"nonce":      {Name: "nonce", Type: "bytes"},
	"timestamp":  {Name: "timestamp", Type: "uint64"},
	"digest":     {Name: "digest", Type: "string"},
	"module_id":  {Name: "module_id", Type: "string"},
}

func VerifyAttestation(w http.ResponseWriter, r *http.Request) {
	req, err := requests.NewSignAttestation(r)
	if err != nil {
		ape.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	var (
		// Should never panic because of request validation
		attestationDocumentBytes, _ = base64.StdEncoding.DecodeString(req.Data.Attributes.Attestation)
		primaryType                 = req.Data.Attributes.PrimaryType
	)

	attestationDocument, err := attestation.ParseNSMAttestationDoc(attestationDocumentBytes)
	if err != nil {
		ape.RenderErr(w, problems.BadRequest(validation.Errors{
			"data/attributes/attestation": fmt.Errorf("failed to parse attestation document: %w", err),
		})...)
		return
	}
	if err = attestationDocument.Verify(); err != nil {
		ape.RenderErr(w, problems.BadRequest(validation.Errors{
			"data/attributes/attestation": fmt.Errorf("invalid signature: %w", err),
		})...)
	}

	usedFields := make(map[string]struct{}, len(req.Data.Attributes.FieldsToSign))
	dataTypes := make([]apitypes.Type, 0, len(req.Data.Attributes.FieldsToSign))
	dataValues := make(apitypes.TypedDataMessage, len(req.Data.Attributes.FieldsToSign))
	for _, field := range req.Data.Attributes.FieldsToSign {
		if _, ok := usedFields[field]; ok {
			continue
		}

		usedFields[field] = struct{}{}
		dataType, ok := fieldToType[field]
		if ok {
			dataTypes = append(dataTypes, dataType)
		} else {
			// pcrX types
			dataTypes = append(dataTypes, apitypes.Type{Name: field, Type: "bytes"})
		}

		switch {
		case strings.HasPrefix(field, "pcr"):
			pcrNum := field[3:]
			// Should never panic because of request validation
			pcr, _ := strconv.ParseUint(pcrNum, 10, 5)

			pcrValue, ok := attestationDocument.PCRs[int(pcr)]
			if !ok {
				err = fmt.Errorf("%s is not present in attestation document", field)
				break
			}

			dataValues[field] = pcrValue
		case field == "public_key":
			if attestationDocument.PublicKey != nil {
				dataValues[field] = attestationDocument.PublicKey
			}
			err = fmt.Errorf("%s is not present in attestation document", field)
		case field == "user_data":
			if attestationDocument.UserData != nil {
				dataValues[field] = attestationDocument.UserData
			}
			err = fmt.Errorf("%s is not present in attestation document", field)
		case field == "nonce":
			if attestationDocument.Nonce != nil {
				dataValues[field] = attestationDocument.Nonce
			}
			err = fmt.Errorf("%s is not present in attestation document", field)
		case field == "timestamp":
			dataValues[field] = uint64(attestationDocument.Timestamp.Unix())
		case field == "module_id":
			dataValues[field] = attestationDocument.ModuleID
		case field == "digest":
			dataValues[field] = attestationDocument.Digest
		}
		if err != nil {
			ape.RenderErr(w, problems.BadRequest(validation.Errors{
				"data/attributes/fields_to_sign": err,
			})...)
			return
		}
	}

	typedDataMessage := icrypto.Message{
		TypedDataMessage: dataValues,
		DataTypes:        dataTypes,
		// Should never panic because of request validation
		PrimaryType: *primaryType,
	}

	domain := icrypto.GetDomain(req.Data.Attributes.Domain)
	sig, _, err := domain.SignTypedDataWithSigner(&typedDataMessage, Signer(r))
	if err != nil {
		Log(r).WithError(err).Errorf("Failed to sign attestation typed data")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	ape.Render(w, resources.SignedAttestationsResponse{
		Data: resources.SignedAttestations{
			Key: resources.Key{
				Type: resources.ATTESTATIONS,
			},
			Attributes: resources.SignedAttestationsAttributes{
				Signature: base64.StdEncoding.EncodeToString(sig),
			},
		},
	})
}
