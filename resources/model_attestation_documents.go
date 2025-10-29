/*
 * GENERATED. Do not modify. Your changes might be overwritten!
 */

package resources

import "encoding/json"

type AttestationDocuments struct {
	Key
	Attributes AttestationDocumentsAttributes `json:"attributes"`
}
type AttestationDocumentsResponse struct {
	Data     AttestationDocuments `json:"data"`
	Included Included             `json:"included"`
}

type AttestationDocumentsListResponse struct {
	Data     []AttestationDocuments `json:"data"`
	Included Included               `json:"included"`
	Links    *Links                 `json:"links"`
	Meta     json.RawMessage        `json:"meta,omitempty"`
}

func (r *AttestationDocumentsListResponse) PutMeta(v interface{}) (err error) {
	r.Meta, err = json.Marshal(v)
	return err
}

func (r *AttestationDocumentsListResponse) GetMeta(out interface{}) error {
	return json.Unmarshal(r.Meta, out)
}

// MustAttestationDocuments - returns AttestationDocuments from include collection.
// if entry with specified key does not exist - returns nil
// if entry with specified key exists but type or ID mismatches - panics
func (c *Included) MustAttestationDocuments(key Key) *AttestationDocuments {
	var attestationDocuments AttestationDocuments
	if c.tryFindEntry(key, &attestationDocuments) {
		return &attestationDocuments
	}
	return nil
}
