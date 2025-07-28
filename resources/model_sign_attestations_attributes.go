/*
 * GENERATED. Do not modify. Your changes might be overwritten!
 */

package resources

import "github.com/ethereum/go-ethereum/signer/core/apitypes"

type SignAttestationsAttributes struct {
	// Standard base64-encoded EIP712 AWS Nitro Enclave attestation document
	Attestation  string                   `json:"attestation"`
	Domain       apitypes.TypedDataDomain `json:"domain"`
	PrimaryType  *string                  `json:"primary_type"`
	FieldsToSign []string                 `json:"fields_to_sign"`
}
