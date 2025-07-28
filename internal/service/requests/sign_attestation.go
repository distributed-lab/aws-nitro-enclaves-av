package requests

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/distributed-lab/aws-nitro-enclaves-av/resources"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

func NewSignAttestation(r *http.Request) (req resources.SignAttestationsRequest, err error) {
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		err = newDecodeError("body", err)
		return req, err
	}

	attr := req.Data.Attributes
	errs := validation.Errors{
		"data/type":                   validation.Validate(req.Data.Type, validation.Required, validation.In(resources.ATTESTATIONS)),
		"data/attributes/attestation": validation.Validate(attr.Attestation, validation.Required, is.Base64),
	}

	return req, errs.Filter()
}

func newDecodeError(what string, err error) error {
	return validation.Errors{
		what: fmt.Errorf("decode request %s: %w", what, err),
	}
}
