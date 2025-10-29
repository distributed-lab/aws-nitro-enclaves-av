package handlers

import (
	"bytes"
	"encoding/base64"
	"net/http"
	"os"
	"path"

	"github.com/distributed-lab/aws-nitro-enclaves-av/resources"
	"github.com/distributed-lab/enclave-extras/attestation"
	"github.com/distributed-lab/enclave-extras/nitro"
	"github.com/distributed-lab/enclave-extras/nsm"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
)

func GetPublicKeyAttestationDoc(w http.ResponseWriter, r *http.Request) {
	attestationDirectory := Signer(r).AttestationsDirectory

	publicKeyAttestationDocPath := path.Join(attestationDirectory, nitro.PublicKeyFile)

	publicKeyAttestationDocRaw, err := os.ReadFile(publicKeyAttestationDocPath)
	if err != nil {
		Log(r).WithError(err).Errorf("Failed to read: %s", publicKeyAttestationDocPath)
		ape.RenderErr(w, problems.InternalError())
		return
	}

	publicKeyAttestationDoc, err := attestation.ParseNSMAttestationDoc(publicKeyAttestationDocRaw)
	if err != nil {
		Log(r).WithError(err).Errorf("Failed to parse: %s", publicKeyAttestationDocPath)
		ape.RenderErr(w, problems.InternalError())
		return
	}

	if err = publicKeyAttestationDoc.Verify(); err != nil {
		Log(r).WithError(err).Errorf("Failed to verify signature: %s", publicKeyAttestationDocPath)
		ape.RenderErr(w, problems.InternalError())
		return
	}

	_, pcr0Actual, err := nsm.DescribePCR(0)
	if err != nil {
		Log(r).WithError(err).Error("Failed to get PCR0")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	if pcr0Stored, ok := publicKeyAttestationDoc.PCRs[0]; !ok || !bytes.Equal(pcr0Stored, pcr0Actual) {
		Log(r).WithError(err).Errorf("PCR0 from %s mismatch with actual PCR0 value", publicKeyAttestationDocPath)
		ape.RenderErr(w, problems.InternalError())
		return
	}

	publicKey := hexutil.Encode(publicKeyAttestationDoc.UserData)

	ape.Render(w, resources.AttestationDocumentsResponse{
		Data: resources.AttestationDocuments{
			Key: resources.Key{
				Type: resources.PUBLIC_KEY_ATTESTATION_DOCUMENTS,
			},
			Attributes: resources.AttestationDocumentsAttributes{
				PublicKey: &publicKey,
				Document:  base64.StdEncoding.EncodeToString(publicKeyAttestationDocRaw),
			},
		},
	})
}
