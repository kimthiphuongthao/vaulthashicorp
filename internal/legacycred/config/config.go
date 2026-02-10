package config

import "time"

type Config struct {
	DefaultTTL        time.Duration `json:"default_ttl"`
	PasswordLength    int           `json:"password_length"`
	PasswordCharset   string        `json:"password_charset"`
	OneTimeRead       bool          `json:"one_time_read"`
	DefaultHashType   string        `json:"default_hash_type"`
	BcryptCost        int           `json:"bcrypt_cost"`
	ShaSalt           string        `json:"sha_salt"`
	ShaSaltPosition   string        `json:"sha_salt_position"`   // prefix|suffix
	ShaOutputEncoding string        `json:"sha_output_encoding"` // hex|base64

	PBKDF2Iterations int    `json:"pbkdf2_iterations"`
	PBKDF2KeyLength  int    `json:"pbkdf2_key_length"`
	PBKDF2SaltLength int    `json:"pbkdf2_salt_length"`
	PBKDF2Hash       string `json:"pbkdf2_hash"` // sha256|sha512

	UpdaterType string `json:"updater_type"` // noop|webhook|sql
	// Webhook updater
	WebhookURL     string `json:"webhook_url"`
	WebhookMethod  string `json:"webhook_method"`
	WebhookHeaders string `json:"webhook_headers"` // JSON string

	// SQL updater
	SQLDriver           string `json:"sql_driver"`
	SQLDSN              string `json:"sql_dsn"`
	SQLUpdateQueryNamed string `json:"sql_update_query_named"` // supports :hash and :subject
	SQLPlaceholderStyle string `json:"sql_placeholder_style"`    // question|dollar
}

func DefaultConfig() Config {
	return Config{
		DefaultTTL:          3 * time.Hour,
		PasswordLength:      32,
		PasswordCharset:     "alnum",
		OneTimeRead:         false,
		DefaultHashType:     "bcrypt",
		BcryptCost:          12,
		ShaSaltPosition:     "prefix",
		ShaOutputEncoding:   "hex",
		PBKDF2Iterations:    600000,
		PBKDF2KeyLength:     32,
		PBKDF2SaltLength:    16,
		PBKDF2Hash:          "sha256",
		UpdaterType:         "noop",
		WebhookMethod:       "POST",
		SQLPlaceholderStyle: "question",
	}
}
