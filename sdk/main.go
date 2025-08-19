package sdk

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/distributed-lab/aws-nitro-enclaves-av/internal/pkg/utils"
	"github.com/distributed-lab/aws-nitro-enclaves-av/resources"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/mdlayher/vsock"
)

var vsockTarget, _ = url.Parse("http://vsock")

type Client struct {
	base        *url.URL
	domain      apitypes.TypedDataDomain
	primaryType string

	c *http.Client
}

func NewInetClient(target string, domain apitypes.TypedDataDomain, primaryType *string) (*Client, error) {
	base, err := url.Parse(target)
	if err != nil {
		return nil, fmt.Errorf("failed to parse target url: %w", err)
	}

	if primaryType == nil || len(*primaryType) == 0 {
		primaryType = utils.AsPointer(utils.DefaultPrimaryType)
	}

	return &Client{
		base:        base,
		domain:      domain,
		primaryType: *primaryType,
		c:           http.DefaultClient,
	}, nil
}

func NewVsockClient(contextID uint32, port uint32, domain apitypes.TypedDataDomain, primaryType *string) *Client {
	if primaryType == nil || len(*primaryType) == 0 {
		primaryType = utils.AsPointer(utils.DefaultPrimaryType)
	}

	return &Client{
		base:        vsockTarget,
		domain:      domain,
		primaryType: *primaryType,
		c: &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
					return vsock.Dial(contextID, port, nil)
				},
				ForceAttemptHTTP2:     true,
				MaxIdleConns:          100,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
		},
	}
}

func (c *Client) SignAttestationDocument(attestationDocument []byte, fields []string) (sig []byte, err error) {
	reqResource := newSignAttestationRequest(base64.StdEncoding.EncodeToString(attestationDocument), fields, &c.primaryType, c.domain)
	reqBody, err := json.Marshal(reqResource)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, c.base.JoinPath("v1/attestations").String(), bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create POST request: %w", err)
	}

	res, err := c.c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to Do request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var resResource resources.SignedAttestationsResponse
	if err := json.Unmarshal(resBody, &resResource); err != nil {
		return nil, fmt.Errorf("failed to unmarshal signed attestation response: %w", err)
	}

	if sig, err = base64.StdEncoding.DecodeString(resResource.Data.Attributes.Signature); err != nil {
		return nil, fmt.Errorf("invalid base64 signature: %w", err)
	}

	return sig, nil
}

func newSignAttestationRequest(attestationB64 string, fieldsToSign []string, primaryType *string, domain apitypes.TypedDataDomain) resources.SignAttestationsRequest {
	return resources.SignAttestationsRequest{
		Data: resources.SignAttestations{
			Key: resources.Key{
				Type: resources.ATTESTATIONS,
			},
			Attributes: resources.SignAttestationsAttributes{
				Attestation:  attestationB64,
				Domain:       domain,
				PrimaryType:  primaryType,
				FieldsToSign: fieldsToSign,
			},
		},
	}
}
