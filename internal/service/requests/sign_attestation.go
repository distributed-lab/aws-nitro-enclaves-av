package requests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/distributed-lab/aws-nitro-enclaves-av/internal/pkg/utils"
	"github.com/distributed-lab/aws-nitro-enclaves-av/resources"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

func NewSignAttestation(r *http.Request) (req resources.SignAttestationsRequest, err error) {
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		err = newDecodeError("body", err)
		return req, err
	}

	attr := &req.Data.Attributes
	errs := validation.Errors{
		"data/type":                   validation.Validate(req.Data.Type, validation.Required, validation.In(resources.ATTESTATIONS)),
		"data/attributes/attestation": validation.Validate(attr.Attestation, validation.Required, is.Base64),
	}

	if len(attr.FieldsToSign) == 0 {
		attr.FieldsToSign = append([]string{}, utils.DefaultFieldsToSign...)
	}

	if attr.PrimaryType == nil || len(*attr.PrimaryType) == 0 {
		attr.PrimaryType = utils.AsPointer(utils.DefaultPrimaryType)
	}

	errs["data/attributes/fields_to_sign"] = validateAttestationFields(attr.FieldsToSign)

	return req, errs.Filter()
}

func newDecodeError(what string, err error) error {
	return validation.Errors{
		what: fmt.Errorf("decode request %s: %w", what, err),
	}
}

func validateAttestationFields(fields []string) error {
	if len(fields) == 0 {
		return fmt.Errorf("fields to sign cannot be empty")
	}

	for _, field := range fields {
		if field == "public_key" || field == "user_data" ||
			field == "nonce" || field == "module_id" ||
			field == "digest" || field == "timestamp" {
			continue
		}
		if !strings.HasPrefix(field, "pcr") {
			return fmt.Errorf("invalid field to sign: %s, must be one of [pcr0, pcr1, ..., pcr31, public_key, user_data, nonce, module_id, digest, timestamp]", field)
		}

		pcrNum := field[3:]

		// 5 bit because currently maximum count of pcr in nsm module is 32
		if _, err := strconv.ParseUint(pcrNum, 10, 5); err != nil {
			return fmt.Errorf("invalid field to sign: %s, must be one of [pcr0, pcr1, ..., pcr31, public_key, user_data, nonce, module_id, digest, timestamp]", field)
		}
	}

	return nil
}
