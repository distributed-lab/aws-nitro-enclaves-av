package icrypto

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

var AddressRegexp = regexp.MustCompile("^0x[0-9a-fA-F]{40}$")
var SignatureRegexp = regexp.MustCompile("^0x[0-9a-fA-F]{130}$")
var SaltRegexp = regexp.MustCompile("^0x[0-9a-fA-F]{64}$")
var HexRegexp = regexp.MustCompile("^0x[0-9a-fA-F]+$")

var (
	ErrDecodeHex      = errors.New("failed to decode hex string")
	ErrBadLength      = errors.New("bad signature length")
	ErrBadRecoverByte = errors.New("bad recovery byte")
	ErrMissMatched    = errors.New("recovered address didn't match any of the given ones")
	ErrNoAddress      = errors.New("no addresses provided for signature verification")
)

func VerifySignature(hash, signature []byte, signer common.Address) error {
	if len(signature) != 65 {
		return ErrBadLength
	}

	sig := make([]byte, len(signature))
	copy(sig, signature)

	if sig[64] != 0 && sig[64] != 1 {
		sig[64] = sig[64] - 27
	}

	recoveredPubkey, err := crypto.SigToPub(hash, sig)
	if err != nil {
		return fmt.Errorf("failed to recover pubkey from signed message: %w", err)
	}

	recoveredAddress := crypto.PubkeyToAddress(*recoveredPubkey)
	if !bytes.Equal(signer[:], recoveredAddress[:]) {
		return ErrMissMatched
	}

	return nil
}
