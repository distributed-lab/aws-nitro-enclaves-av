package nitro

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	kmstypes "github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/distributed-lab/enclave-extras/attestation"
	"github.com/distributed-lab/enclave-extras/attestedkms"
	"github.com/distributed-lab/enclave-extras/nsm"
	"github.com/ethereum/go-ethereum/common"
)

const (
	AwsIamServiceID = "iam"
	AwsStsServiceID = "sts"
)

const (
	// Attestation document with the KMS KeyID
	// in UserData attestation doc field.
	kmsKeyIDFile = "kms_key_id.coses1"
	// Attestation document with the validator's encrypted
	// private key in UserData attestation doc field.
	privateKeyFile = "private_key.coses1"
	// Attestation document with the validator's public key
	// in UserData and PublicKey attestation doc fields.
	publickKeyFile = "kms_key_id.coses1"
	// Attestation document with the validator's
	// address in UserData attestation doc field.
	addressFile = "address.coses1"
)

func EnsureArnIsIam(v string) (string, error) {
	resourceArn, err := arn.Parse(v)
	if err != nil {
		return "", fmt.Errorf("failed to parse resource ARN: %w", err)
	}

	// If ARN service already IAM just return it
	if resourceArn.Service == AwsIamServiceID {
		return v, nil
	}

	if resourceArn.Service != AwsStsServiceID || !strings.HasPrefix(resourceArn.Resource, "assumed-role/") {
		return "", fmt.Errorf("unsuported conversion, can convert only STS assumed-role in IAM role")
	}

	resourceArn.Service = AwsIamServiceID
	// Should never be out of range, because of AWS guarantee that role can't be empty string
	resourceArn.Resource = "role/" + strings.Split(resourceArn.Resource, "/")[1]

	return resourceArn.String(), nil
}

func ToRootArn(v string) (string, error) {
	resourceArn, err := arn.Parse(v)
	if err != nil {
		return "", fmt.Errorf("failed to parse resource ARN: %w", err)
	}

	resourceArn.Service = AwsIamServiceID
	resourceArn.Resource = "root"

	return resourceArn.String(), nil
}

func GetArns(cfg aws.Config) (rootArn string, principalArn string, err error) {
	stsClient := sts.NewFromConfig(cfg)
	callerIdentityResponse, err := stsClient.GetCallerIdentity(context.Background(), &sts.GetCallerIdentityInput{})
	if err != nil {
		return "", "", fmt.Errorf("failed to get caller identity: %w", err)
	}

	principalArn, err = EnsureArnIsIam(Deref(callerIdentityResponse.Arn))
	if err != nil {
		return "", "", fmt.Errorf("failed to cast arn: %w", err)
	}

	rootArn, err = ToRootArn(principalArn)
	if err != nil {
		return "", "", fmt.Errorf("failed to make root arn: %w", err)
	}

	return rootArn, principalArn, nil
}

func GetKMSEnclaveClient(cfg aws.Config) (*attestedkms.KMSEnclaveClient, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA private key: %w", err)
	}

	derEncodedPublicKey, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal public key PKIX: %w", err)
	}

	attestationDoc, err := GetAttestationDoc(nil, nil, derEncodedPublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get attestation document: %w", err)
	}

	return attestedkms.NewFromConfig(cfg, attestationDoc, privateKey), nil
}

// Safely pointer dereference
func Deref[T any](p *T) T {
	if p != nil {
		return *p
	}
	// Declares a variable of type T, initialized to its zero value
	var zero T
	return zero
}

func GetAttestedKMSKeyID(cfg aws.Config, attestationsPath string) (string, error) {
	kmsKeyIDPath := path.Join(attestationsPath, kmsKeyIDFile)

	_, pcr0Actual, err := DescribePCR(0)
	if err != nil {
		return "", fmt.Errorf("failed to get PCR0: %w", err)
	}

	kmsKeyIDAttestationDocRaw, err := os.ReadFile(kmsKeyIDPath)
	// if attestation document exist just read KMS Key ID
	if err == nil {
		kmsKeyIDAttestationDoc, err := attestation.ParseNSMAttestationDoc(kmsKeyIDAttestationDocRaw)
		if err != nil {
			return "", fmt.Errorf("failed to parse %s: %w", kmsKeyIDPath, err)
		}
		if err = kmsKeyIDAttestationDoc.Verify(); err != nil {
			return "", fmt.Errorf("%s have invalid signature: %w", kmsKeyIDPath, err)
		}

		if pcr0Stored, ok := kmsKeyIDAttestationDoc.PCRs[0]; !ok || !bytes.Equal(pcr0Stored, pcr0Actual) {
			return "", fmt.Errorf("PCR0 from %s mismatch with actual PCR0 value", kmsKeyIDPath)
		}
		return string(kmsKeyIDAttestationDoc.UserData), nil
	}

	// if attestation document exists, but we can't open file
	if !os.IsNotExist(err) {
		return "", fmt.Errorf("failed to read %s, check file permissions. err: %w", kmsKeyIDPath, err)
	}

	rootArn, principalArn, err := GetArns(cfg)
	if err != nil {
		return "", fmt.Errorf("failed to get root and principal arns for kms key policy: %w", err)
	}

	kmsEnclaveClient, err := GetKMSEnclaveClient(cfg)
	if err != nil {
		return "", fmt.Errorf("failed to get kms enclave client: %w", err)
	}

	kmsKeyPolicy := DefaultPolicies(rootArn, principalArn, map[int][]byte{0: pcr0Actual})

	createKeyOutput, err := kmsEnclaveClient.CreateKey(context.Background(), &kms.CreateKeyInput{
		// DANGER: The key may become unmanageable
		BypassPolicyLockoutSafetyCheck: true,
		Description:                    aws.String("Nitro Enclave Key"),
		Policy:                         aws.String(kmsKeyPolicy),
	})
	if err != nil {
		return "", fmt.Errorf("failed to create KMS key: %w", err)
	}

	kmsKeyID := Deref(createKeyOutput.KeyMetadata.KeyId)

	// Save KMS Key
	kmsKeyIDAttestationDocRaw, err = GetAttestationDoc([]byte(kmsKeyID), nil, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get attestation document for %s: %w", kmsKeyIDPath, err)
	}
	if err = os.WriteFile(kmsKeyIDPath, kmsKeyIDAttestationDocRaw, 0644); err != nil {
		return "", fmt.Errorf("failed to write %s: %w", kmsKeyIDPath, err)
	}

	return kmsKeyID, nil
}

func GetAttestedPrivateKey(cfg aws.Config, attestationsPath string) (*ecdsa.PrivateKey, error) {
	kmsKeyID, err := GetAttestedKMSKeyID(cfg, attestationsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get KMS key ID: %w", err)
	}
	kmsEnclaveClient, err := GetKMSEnclaveClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to get kms enclave client: %w", err)
	}

	privateKeyPath := path.Join(attestationsPath, privateKeyFile)
	privateKeyAttestationDocRaw, err := os.ReadFile(privateKeyPath)
	// if attestation document exist just read KMS Key ID
	if err == nil {
		privateKeyAttestationDoc, err := attestation.ParseNSMAttestationDoc(privateKeyAttestationDocRaw)
		if err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w", privateKeyPath, err)
		}
		if err = privateKeyAttestationDoc.Verify(); err != nil {
			return nil, fmt.Errorf("%s have invalid signature: %w", privateKeyPath, err)
		}
		_, pcr0Actual, err := DescribePCR(0)
		if err != nil {
			return nil, fmt.Errorf("failed to get PCR0: %w", err)
		}
		if pcr0Stored, ok := privateKeyAttestationDoc.PCRs[0]; !ok || !bytes.Equal(pcr0Stored, pcr0Actual) {
			return nil, fmt.Errorf("PCR0 from %s mismatch with actual PCR0 value", privateKeyPath)
		}

		decryptResp, err := kmsEnclaveClient.Decrypt(context.Background(), &kms.DecryptInput{
			KeyId:          aws.String(kmsKeyID),
			CiphertextBlob: privateKeyAttestationDoc.UserData,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt private key: %w", err)
		}

		privateKey, err := parsePKCS8ECPrivateKey(decryptResp.Plaintext)
		if err != nil {
			return nil, fmt.Errorf("failed to parse secp256k1: %w", err)
		}

		return privateKey, nil
	}

	// if attestation document exists, but we can't open file
	if !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to read %s, check file permissions. err: %w", privateKeyPath, err)
	}

	// Create private key
	generateDataKeyPairResp, err := kmsEnclaveClient.GenerateDataKeyPair(context.Background(), &kms.GenerateDataKeyPairInput{
		KeyId:       aws.String(kmsKeyID),
		KeyPairSpec: kmstypes.DataKeyPairSpecEccSecgP256k1,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate secp256k1 in KMS: %w", err)
	}
	privateKey, err := parsePKCS8ECPrivateKey(generateDataKeyPairResp.PrivateKeyPlaintext)
	if err != nil {
		return nil, fmt.Errorf("failed to parse secp256k1: %w", err)
	}

	// Save private key
	privateKeyAttestationDocRaw, err = nsm.GetAttestationDoc(generateDataKeyPairResp.PrivateKeyCiphertextBlob, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get attestation doc for %s: %w", privateKeyPath, err)
	}
	if err = os.WriteFile(privateKeyPath, privateKeyAttestationDocRaw, 0600); err != nil {
		return nil, fmt.Errorf("failed to write %s: %w", privateKeyPath, err)
	}

	return privateKey, nil
}

func GetAttestedPublicKey(kmsEnclaveClient *attestedkms.KMSEnclaveClient, attestationsPath string) (ecdsa.PublicKey, error)
func getAttestedAddress(kmsEnclaveClient *attestedkms.KMSEnclaveClient, attestationsPath string) (common.Address, error)

func parsePKCS8ECPrivateKey(pcks8PrivateKey []byte) (*ecdsa.PrivateKey, error) {
	privateKeyAny, err := attestedkms.ParsePKCS8PrivateKey(pcks8PrivateKey)
	if err != nil {
		return nil, err
	}

	privateKey, ok := privateKeyAny.(*ecdsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("invalid EC private key")
	}

	return privateKey, nil
}
