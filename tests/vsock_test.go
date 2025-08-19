package tests

import (
	"encoding/base64"
	"testing"

	"github.com/distributed-lab/aws-nitro-enclaves-av/sdk"
	"github.com/stretchr/testify/assert"
)

func TestVsockHttpAttestations(t *testing.T) {
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := sdk.NewVsockClient(16, 8000, testDomain, test.primaryType)

			attestationDocument, err := base64.StdEncoding.DecodeString(test.attestationDocument)
			assert.NoError(t, err, "failed to decode base64 attestation document")

			sig, err := client.SignAttestationDocument(attestationDocument, test.fields)
			assert.Equal(t, test.wantErr, err != nil, "unexpected result")

			sigB64 := base64.StdEncoding.EncodeToString(sig)
			if test.want != sigB64 {
				t.Errorf("signature mismatch, want %s, get %s", test.want, sigB64)
			}
		})
	}
}
