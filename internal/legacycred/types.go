package legacycred

import "time"

type SubjectRecord struct {
	Subject   string    `json:"subject"`
	Plaintext string    `json:"plaintext"`
	Hash      string    `json:"hash"`
	HashType  string    `json:"hash_type"`
	Version   int64     `json:"version"`
	UpdatedAt time.Time `json:"updated_at"`
	ExpiresAt time.Time `json:"expires_at"`
}
