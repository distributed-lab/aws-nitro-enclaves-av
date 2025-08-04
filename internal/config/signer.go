package config

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"os"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/distributed-lab/aws-nitro-enclaves-av/internal/pkg/nitro"
	"github.com/ethereum/go-ethereum/crypto"
	figure "gitlab.com/distributed_lab/figure/v3"
	"gitlab.com/distributed_lab/kit/kv"
)

type Signer struct {
	pk *ecdsa.PrivateKey
}

func (s *Signer) Sign(data []byte) ([]byte, error) {
	return crypto.Sign(data, s.pk)
}

func (c *config) GetSigner() *Signer {
	return c.signerConfigurator.Do(func() any {
		var cfg struct {
			AttestationsDirectory string `fig:"attestations_directory,required"`
		}

		err := figure.
			Out(&cfg).
			From(kv.MustGetStringMap(c.getter, "signer")).
			Please()
		if err != nil {
			panic(fmt.Errorf("failed to figure out signer config: %w", err))
		}

		if err = os.MkdirAll(cfg.AttestationsDirectory, os.ModePerm); err != nil {
			panic(fmt.Errorf("failed to create attestation target directory %s with error: %w", cfg.AttestationsDirectory, err))
		}

		awsConfig, err := awsconfig.LoadDefaultConfig(context.Background())
		if err != nil {
			panic(fmt.Errorf("failed to load AWS config: %w", err))
		}

		kmsKeyID, err := nitro.GetAttestedKMSKeyID(awsConfig, cfg.AttestationsDirectory)
		if err != nil {
			panic(fmt.Errorf("failed to get attested KMS Key ID: %w", err))
		}

		privateKey, err := nitro.GetAttestedPrivateKey(awsConfig, kmsKeyID, cfg.AttestationsDirectory)
		if err != nil {
			panic(fmt.Errorf("failed to get attested private key: %w", err))
		}

		publicKey, err := nitro.GetAttestedPublicKey(privateKey, cfg.AttestationsDirectory)
		if err != nil {
			panic(fmt.Errorf("failed to get attested public key: %w", err))
		}

		if _, err = nitro.GetAttestedAddress(publicKey, cfg.AttestationsDirectory); err != nil {
			panic(fmt.Errorf("failed to get attested address: %w", err))
		}

		return &Signer{
			pk: privateKey,
		}
	}).(*Signer)
}
