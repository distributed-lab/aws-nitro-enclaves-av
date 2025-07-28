package icrypto

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

const DomainType = "EIP712Domain"

type Domain struct {
	apitypes.TypedDataDomain
	DomainTypes []apitypes.Type
}

type Message struct {
	apitypes.TypedDataMessage
	DataTypes   []apitypes.Type
	PrimaryType string
}

type DomainProvider interface {
	Eip712Domain(opts *bind.CallOpts) (struct {
		Fields            [1]byte
		Name              string
		Version           string
		ChainId           *big.Int //nolint
		VerifyingContract common.Address
		Salt              [32]byte
		Extensions        []*big.Int
	}, error)
}

func GetDomain(provider DomainProvider) (*Domain, error) {
	domainData, err := provider.Eip712Domain(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get eip712 domain: %w", err)
	}

	domain := &Domain{}
	domainTypes := make([]apitypes.Type, 0, 5)
	if domainData.Fields[0]&0b00001 != 0 {
		domain.Name = domainData.Name
		domainTypes = append(domainTypes, apitypes.Type{Name: "name", Type: "string"})
	}
	if domainData.Fields[0]&0b00010 != 0 {
		domain.Version = domainData.Version
		domainTypes = append(domainTypes, apitypes.Type{Name: "version", Type: "string"})
	}
	if domainData.Fields[0]&0b00100 != 0 {
		domain.ChainId = (*math.HexOrDecimal256)(domainData.ChainId)
		domainTypes = append(domainTypes, apitypes.Type{Name: "chainId", Type: "uint256"})
	}
	if domainData.Fields[0]&0b01000 != 0 {
		domain.VerifyingContract = domainData.VerifyingContract.Hex()
		domainTypes = append(domainTypes, apitypes.Type{Name: "verifyingContract", Type: "address"})
	}
	if domainData.Fields[0]&0b10000 != 0 {
		domain.Salt = hexutil.Encode(domainData.Salt[:])
		domainTypes = append(domainTypes, apitypes.Type{Name: "salt", Type: "bytes32"})
	}
	domain.DomainTypes = domainTypes

	return domain, nil
}

func (d *Domain) TypedDataAndHash(message *Message) (hash []byte, rawData string, err error) {
	primaryType := message.PrimaryType
	types := apitypes.Types{
		primaryType: message.DataTypes,
		DomainType:  d.DomainTypes,
	}

	data := apitypes.TypedData{
		Types:       types,
		PrimaryType: primaryType,
		Domain:      d.TypedDataDomain,
		Message:     message.TypedDataMessage,
	}

	return apitypes.TypedDataAndHash(data)
}

func (d *Domain) SignTypedData(message *Message, pk *ecdsa.PrivateKey) ([]byte, []byte, error) {
	hash, _, err := d.TypedDataAndHash(message)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get typed data hash: %w", err)
	}
	sig, err := crypto.Sign(hash, pk)
	if err != nil {
		return nil, nil, fmt.Errorf("faield to sign: %w", err)
	}

	// set recovery byte
	sig[64] += 0x1b
	return sig, hash, nil
}

func (d *Domain) SignTypedDataWithSigner(message *Message, signer interface{ Sign([]byte) ([]byte, error) }) ([]byte, []byte, error) {
	hash, _, err := d.TypedDataAndHash(message)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get typed data hash: %w", err)
	}
	sig, err := signer.Sign(hash)
	if err != nil {
		return nil, nil, fmt.Errorf("faield to sign: %w", err)
	}

	// set recovery byte
	sig[64] += 0x1b
	return sig, hash, nil
}

func (d *Domain) VerifyTypedData(message *Message, signature []byte, signer common.Address) error {
	hash, _, err := d.TypedDataAndHash(message)
	if err != nil {
		return fmt.Errorf("failed to get typed data and hash: %w", err)
	}

	if err = VerifySignature(hash, signature, signer); err != nil {
		return fmt.Errorf("invalid signature: %w", err)
	}

	return nil
}
