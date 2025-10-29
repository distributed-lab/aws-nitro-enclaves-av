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

func GetAddressAttestationDoc(w http.ResponseWriter, r *http.Request) {
	attestationDirectory := Signer(r).AttestationsDirectory

	addressAttestationDocPath := path.Join(attestationDirectory, nitro.AddressFile)

	addressAttestationDocRaw, err := os.ReadFile(addressAttestationDocPath)
	if err != nil {
		Log(r).WithError(err).Errorf("Failed to read: %s", addressAttestationDocPath)
		ape.RenderErr(w, problems.InternalError())
		return
	}

	addressAttestationDoc, err := attestation.ParseNSMAttestationDoc(addressAttestationDocRaw)
	if err != nil {
		Log(r).WithError(err).Errorf("Failed to parse: %s", addressAttestationDocPath)
		ape.RenderErr(w, problems.InternalError())
		return
	}

	if err = addressAttestationDoc.Verify(); err != nil {
		Log(r).WithError(err).Errorf("Failed to verify signature: %s", addressAttestationDocPath)
		ape.RenderErr(w, problems.InternalError())
		return
	}

	_, pcr0Actual, err := nsm.DescribePCR(0)
	if err != nil {
		Log(r).WithError(err).Error("Failed to get PCR0")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	if pcr0Stored, ok := addressAttestationDoc.PCRs[0]; !ok || !bytes.Equal(pcr0Stored, pcr0Actual) {
		Log(r).WithError(err).Errorf("PCR0 from %s mismatch with actual PCR0 value", addressAttestationDocPath)
		ape.RenderErr(w, problems.InternalError())
		return
	}

	address := hexutil.Encode(addressAttestationDoc.UserData)
	ape.Render(w, resources.AttestationDocumentsResponse{
		Data: resources.AttestationDocuments{
			Key: resources.Key{
				Type: resources.ADDRESS_ATTESTATION_DOCUMENTS,
			},
			Attributes: resources.AttestationDocumentsAttributes{
				UserData: &address,
				Document: base64.StdEncoding.EncodeToString(addressAttestationDocRaw),
			},
		},
	})
}
