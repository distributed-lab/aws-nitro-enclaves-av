package nitro

import (
	"errors"
	"fmt"

	"github.com/hf/nsm"
	"github.com/hf/nsm/request"
)

func GetAttestationDoc(nonce, userData, publicKey []byte) ([]byte, error) {
	sess, err := nsm.OpenDefaultSession()
	if err != nil {
		return nil, fmt.Errorf("failed to open nsm session: %w", err)
	}
	defer sess.Close()

	res, err := sess.Send(&request.Attestation{
		Nonce:     nonce,
		UserData:  userData,
		PublicKey: publicKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to send attestation request: %w", err)
	}

	if res.Error != "" {
		return nil, errors.New(string(res.Error))
	}

	if res.Attestation == nil || res.Attestation.Document == nil {
		return nil, errors.New("NSM device did not return an attestation")
	}

	return res.Attestation.Document, nil
}

func DescribePCR(pcrIndex int) (bool, []byte, error) {
	sess, err := nsm.OpenDefaultSession()
	if err != nil {
		return false, nil, fmt.Errorf("failed to open nsm session: %w", err)
	}
	defer sess.Close()

	res, err := sess.Send(&request.DescribePCR{
		Index: uint16(pcrIndex),
	})
	if err != nil {
		return false, nil, fmt.Errorf("failed to send describe pcr request: %w", err)
	}

	if res.Error != "" {
		return false, nil, errors.New(string(res.Error))
	}

	if res.DescribePCR.Data == nil {
		return false, nil, errors.New("NSM device did not return an pcr")
	}

	return res.DescribePCR.Lock, res.DescribePCR.Data, nil
}
