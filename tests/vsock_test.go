package tests

import (
	"encoding/base64"
	"os"
	"testing"

	"github.com/distributed-lab/aws-nitro-enclaves-av/internal/pkg/utils"
	"github.com/distributed-lab/aws-nitro-enclaves-av/sdk"
	"github.com/distributed-lab/enclave-extras/attestation"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

func TestVsockHttpAttestations(t *testing.T) {
	attestationDocRaw, err := os.ReadFile(addressAttDocPath)
	assert.NoError(t, err, "failed to read attestation document with address")

	attestationDoc, err := attestation.ParseNSMAttestationDoc(attestationDocRaw)
	assert.NoError(t, err, "failed to parse attestation document with address")

	address := common.Address(attestationDoc.UserData)

	for _, test := range tests {
		primaryType := utils.DefaultPrimaryType
		if test.primaryType != nil {
			primaryType = *test.primaryType
		}

		fields := append([]string{}, utils.DefaultFieldsToSign...)
		if len(test.fields) != 0 {
			fields = append([]string{}, test.fields...)
		}

		t.Run(test.name, func(t *testing.T) {
			client := sdk.NewVsockClient(16, 8000, domain.TypedDataDomain, test.primaryType)

			attestationDocumentRaw, err := base64.StdEncoding.DecodeString(test.attestationDocument)
			assert.NoError(t, err, "failed to decode base64 attestation document")

			sig, err := client.SignAttestationDocument(attestationDocumentRaw, test.fields)
			assert.Equal(t, test.wantErr, err != nil, "unexpected result")

			attestationDocument, err := attestation.ParseNSMAttestationDoc(attestationDocumentRaw)
			assert.NoError(t, err, "failed to parse attestation document")

			msg, err := utils.BuildTypedDataAttestationMessage(attestationDocument, primaryType, fields)
			assert.NoError(t, err, "failed to build typed data message")

			err = domain.VerifyTypedData(msg, sig, address)
			assert.NoError(t, err, "invalid signature")
		})
	}
}
