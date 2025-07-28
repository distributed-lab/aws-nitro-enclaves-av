/*
 * GENERATED. Do not modify. Your changes might be overwritten!
 */

package resources

import "encoding/json"

type SignAttestations struct {
	Key
	Attributes SignAttestationsAttributes `json:"attributes"`
}
type SignAttestationsRequest struct {
	Data     SignAttestations `json:"data"`
	Included Included         `json:"included"`
}

type SignAttestationsListRequest struct {
	Data     []SignAttestations `json:"data"`
	Included Included           `json:"included"`
	Links    *Links             `json:"links"`
	Meta     json.RawMessage    `json:"meta,omitempty"`
}

func (r *SignAttestationsListRequest) PutMeta(v interface{}) (err error) {
	r.Meta, err = json.Marshal(v)
	return err
}

func (r *SignAttestationsListRequest) GetMeta(out interface{}) error {
	return json.Unmarshal(r.Meta, out)
}

// MustSignAttestations - returns SignAttestations from include collection.
// if entry with specified key does not exist - returns nil
// if entry with specified key exists but type or ID mismatches - panics
func (c *Included) MustSignAttestations(key Key) *SignAttestations {
	var signAttestations SignAttestations
	if c.tryFindEntry(key, &signAttestations) {
		return &signAttestations
	}
	return nil
}
