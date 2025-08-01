package config

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/distributed-lab/aws-nitro-enclaves-av/internal/pkg/nitro"
	"github.com/ethereum/go-ethereum/crypto"
	"gitlab.com/distributed_lab/figure/v3"
	"gitlab.com/distributed_lab/kit/kv"
)

type Signer struct {
	pk *ecdsa.PrivateKey
}

func (s *Signer) Sign(data []byte) ([]byte, error) {
	return crypto.Sign(data, s.pk)
}

func (c *config) GetSigner() *Signer {
	return c.signerConfigurator.Do(func() interface{} {
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

		rootArn, principalArn, err := nitro.GetArns(awsConfig)
		if err != nil {
			panic(fmt.Errorf("failed to get arns: %w", err))
		}

		_, pcr0Value, err := nitro.DescribePCR(0)
		if err != nil {
			panic(fmt.Errorf("failed to get PCR0 value: %w", err))
		}

		kmsKeyPolicy := nitro.DefaultPolicies(rootArn, principalArn, map[int][]byte{0: pcr0Value})

		kmsEnclaveClient, err := nitro.GetKMSEnclaveClient(awsConfig)
		if err != nil {
			panic(fmt.Errorf("failed to get kms enclave client: %w", err))
		}

		createKeyOutput, err := kmsEnclaveClient.CreateKey(context.Background(), &kms.CreateKeyInput{
			// DANGER: The key may become unmanageable
			BypassPolicyLockoutSafetyCheck: true,
			Description:                    aws.String("Nitro Enclave Key"),
			Policy:                         aws.String(kmsKeyPolicy),
		})

		return &cfg
	}).(*Signer)
}
