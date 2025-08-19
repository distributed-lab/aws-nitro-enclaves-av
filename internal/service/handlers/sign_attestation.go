package handlers

import (
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/distributed-lab/aws-nitro-enclaves-av/internal/pkg/icrypto"
	"github.com/distributed-lab/aws-nitro-enclaves-av/internal/pkg/utils"
	"github.com/distributed-lab/aws-nitro-enclaves-av/internal/service/requests"
	"github.com/distributed-lab/aws-nitro-enclaves-av/resources"
	"github.com/distributed-lab/enclave-extras/attestation"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
)

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
		return
	}

	fields := make(map[string]struct{}, len(req.Data.Attributes.FieldsToSign))
	for _, field := range req.Data.Attributes.FieldsToSign {
		fields[field] = struct{}{}
	}

	typedDataMessage, err := utils.BuildTypedDataAttestationMessage(attestationDocument, *primaryType, fields)
	if err != nil {
		ape.RenderErr(w, problems.BadRequest(validation.Errors{
			"data/attributes": err,
		})...)
		return
	}

	domain := icrypto.GetDomain(req.Data.Attributes.Domain)
	sig, _, err := domain.SignTypedDataWithSigner(typedDataMessage, Signer(r))
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
