/*
 * GENERATED. Do not modify. Your changes might be overwritten!
 */

package resources

import "encoding/json"

type SignedAttestations struct {
	Key
	Attributes SignedAttestationsAttributes `json:"attributes"`
}
type SignedAttestationsResponse struct {
	Data     SignedAttestations `json:"data"`
	Included Included           `json:"included"`
}

type SignedAttestationsListResponse struct {
	Data     []SignedAttestations `json:"data"`
	Included Included             `json:"included"`
	Links    *Links               `json:"links"`
	Meta     json.RawMessage      `json:"meta,omitempty"`
}

func (r *SignedAttestationsListResponse) PutMeta(v interface{}) (err error) {
	r.Meta, err = json.Marshal(v)
	return err
}

func (r *SignedAttestationsListResponse) GetMeta(out interface{}) error {
	return json.Unmarshal(r.Meta, out)
}

// MustSignedAttestations - returns SignedAttestations from include collection.
// if entry with specified key does not exist - returns nil
// if entry with specified key exists but type or ID mismatches - panics
func (c *Included) MustSignedAttestations(key Key) *SignedAttestations {
	var signedAttestations SignedAttestations
	if c.tryFindEntry(key, &signedAttestations) {
		return &signedAttestations
	}
	return nil
}
