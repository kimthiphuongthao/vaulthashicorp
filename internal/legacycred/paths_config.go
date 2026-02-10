package legacycred

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"

	"github.com/kimthiphuongthao/vaulthashicorp/internal/legacycred/config"
)

func (b *backend) pathConfig() *framework.Path {
	return &framework.Path{
		Pattern: "config",
		HelpSynopsis: "Configure the legacy-cred secrets engine.",
		Fields: map[string]*framework.FieldSchema{
			"default_ttl": {Type: framework.TypeDurationSecond, Description: "Default TTL for stored plaintext (e.g. 3h)."},
			"password_length": {Type: framework.TypeInt, Description: "Generated password length."},
			"password_charset": {Type: framework.TypeString, Description: "Password charset: alnum|ascii|custom."},
			"one_time_read": {Type: framework.TypeBool, Description: "If true, delete plaintext after a successful creds read."},

			"default_hash_type": {Type: framework.TypeString, Description: "Default hash type: bcrypt|sha256|sha512|pbkdf2."},
			"bcrypt_cost": {Type: framework.TypeInt, Description: "Bcrypt cost."},
			"sha_salt": {Type: framework.TypeString, Description: "Optional salt used for sha256/sha512."},
			"sha_salt_position": {Type: framework.TypeString, Description: "Salt position: prefix|suffix."},
			"sha_output_encoding": {Type: framework.TypeString, Description: "SHA output encoding: hex|base64."},

			"pbkdf2_iterations": {Type: framework.TypeInt, Description: "PBKDF2 iterations."},
			"pbkdf2_key_length": {Type: framework.TypeInt, Description: "PBKDF2 derived key length (bytes)."},
			"pbkdf2_salt_length": {Type: framework.TypeInt, Description: "PBKDF2 salt length (bytes)."},
			"pbkdf2_hash": {Type: framework.TypeString, Description: "PBKDF2 hash: sha256|sha512."},

			"updater_type": {Type: framework.TypeString, Description: "Updater type: noop|webhook|sql."},
			"webhook_url": {Type: framework.TypeString, Description: "Webhook URL to update legacy credential store."},
			"webhook_method": {Type: framework.TypeString, Description: "Webhook method (default POST)."},
			"webhook_headers": {Type: framework.TypeString, Description: "Webhook headers JSON object string."},

			"sql_driver": {Type: framework.TypeString, Description: "SQL driver name (e.g. postgres, mysql, sqlserver)."},
			"sql_dsn": {Type: framework.TypeString, Description: "SQL DSN/connection string."},
			"sql_update_query_named": {Type: framework.TypeString, Description: "SQL update query using :hash and :subject placeholders."},
			"sql_placeholder_style": {Type: framework.TypeString, Description: "SQL placeholder style: question|dollar."},
		},
		Operations: map[logical.Operation]framework.OperationHandler{
			logical.ReadOperation: &framework.PathOperation{Callback: b.handleConfigRead},
			logical.UpdateOperation: &framework.PathOperation{Callback: b.handleConfigWrite},
		},
	}
}

func (b *backend) handleConfigRead(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	cfg, err := b.loadConfig(ctx, req.Storage)
	if err != nil {
		return nil, err
	}
	return &logical.Response{Data: map[string]any{
		"default_ttl":             int64(cfg.DefaultTTL.Seconds()),
		"password_length":         cfg.PasswordLength,
		"password_charset":        cfg.PasswordCharset,
		"one_time_read":           cfg.OneTimeRead,
		"default_hash_type":       cfg.DefaultHashType,
		"bcrypt_cost":             cfg.BcryptCost,
		"sha_salt":                cfg.ShaSalt,
		"sha_salt_position":       cfg.ShaSaltPosition,
		"sha_output_encoding":     cfg.ShaOutputEncoding,
		"pbkdf2_iterations":       cfg.PBKDF2Iterations,
		"pbkdf2_key_length":       cfg.PBKDF2KeyLength,
		"pbkdf2_salt_length":      cfg.PBKDF2SaltLength,
		"pbkdf2_hash":             cfg.PBKDF2Hash,
		"updater_type":            cfg.UpdaterType,
		"webhook_url":             cfg.WebhookURL,
		"webhook_method":          cfg.WebhookMethod,
		"webhook_headers":         cfg.WebhookHeaders,
		"sql_driver":              cfg.SQLDriver,
		"sql_dsn":                 maskDSN(cfg.SQLDSN),
		"sql_update_query_named":  cfg.SQLUpdateQueryNamed,
		"sql_placeholder_style":   cfg.SQLPlaceholderStyle,
	}}, nil
}

func (b *backend) handleConfigWrite(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	cfg, err := b.loadConfig(ctx, req.Storage)
	if err != nil {
		return nil, err
	}

	if v, ok := d.GetOk("default_ttl"); ok {
		cfg.DefaultTTL = time.Duration(v.(int)) * time.Second
	}
	if v, ok := d.GetOk("password_length"); ok {
		cfg.PasswordLength = v.(int)
	}
	if v, ok := d.GetOk("password_charset"); ok {
		cfg.PasswordCharset = v.(string)
	}
	if v, ok := d.GetOk("one_time_read"); ok {
		cfg.OneTimeRead = v.(bool)
	}
	if v, ok := d.GetOk("default_hash_type"); ok {
		cfg.DefaultHashType = v.(string)
	}
	if v, ok := d.GetOk("bcrypt_cost"); ok {
		cfg.BcryptCost = v.(int)
	}
	if v, ok := d.GetOk("sha_salt"); ok {
		cfg.ShaSalt = v.(string)
	}
	if v, ok := d.GetOk("sha_salt_position"); ok {
		cfg.ShaSaltPosition = v.(string)
	}
	if v, ok := d.GetOk("sha_output_encoding"); ok {
		cfg.ShaOutputEncoding = v.(string)
	}
	if v, ok := d.GetOk("pbkdf2_iterations"); ok {
		cfg.PBKDF2Iterations = v.(int)
	}
	if v, ok := d.GetOk("pbkdf2_key_length"); ok {
		cfg.PBKDF2KeyLength = v.(int)
	}
	if v, ok := d.GetOk("pbkdf2_salt_length"); ok {
		cfg.PBKDF2SaltLength = v.(int)
	}
	if v, ok := d.GetOk("pbkdf2_hash"); ok {
		cfg.PBKDF2Hash = v.(string)
	}
	if v, ok := d.GetOk("updater_type"); ok {
		cfg.UpdaterType = v.(string)
	}
	if v, ok := d.GetOk("webhook_url"); ok {
		cfg.WebhookURL = v.(string)
	}
	if v, ok := d.GetOk("webhook_method"); ok {
		cfg.WebhookMethod = v.(string)
	}
	if v, ok := d.GetOk("webhook_headers"); ok {
		cfg.WebhookHeaders = v.(string)
	}
	if v, ok := d.GetOk("sql_driver"); ok {
		cfg.SQLDriver = v.(string)
	}
	if v, ok := d.GetOk("sql_dsn"); ok {
		cfg.SQLDSN = v.(string)
	}
	if v, ok := d.GetOk("sql_update_query_named"); ok {
		cfg.SQLUpdateQueryNamed = v.(string)
	}
	if v, ok := d.GetOk("sql_placeholder_style"); ok {
		cfg.SQLPlaceholderStyle = v.(string)
	}

	if err := validateConfig(cfg); err != nil {
		return logical.ErrorResponse(err.Error()), nil
	}
	if err := b.saveConfig(ctx, req.Storage, cfg); err != nil {
		return nil, err
	}
	return &logical.Response{Data: map[string]any{"ok": true}}, nil
}

func validateConfig(cfg config.Config) error {
	if cfg.PasswordLength < 8 {
		return fmt.Errorf("password_length must be >= 8")
	}
	if cfg.DefaultTTL < 0 {
		return fmt.Errorf("default_ttl must be >= 0")
	}
	switch cfg.DefaultHashType {
	case "bcrypt", "sha256", "sha512", "pbkdf2":
	default:
		return fmt.Errorf("unsupported default_hash_type")
	}
	switch cfg.ShaSaltPosition {
	case "prefix", "suffix":
	default:
		return fmt.Errorf("sha_salt_position must be prefix|suffix")
	}
	switch cfg.ShaOutputEncoding {
	case "hex", "base64":
	default:
		return fmt.Errorf("sha_output_encoding must be hex|base64")
	}
	switch cfg.UpdaterType {
	case "noop", "webhook", "sql":
	default:
		return fmt.Errorf("updater_type must be noop|webhook|sql")
	}
	if cfg.UpdaterType == "webhook" && cfg.WebhookURL == "" {
		return fmt.Errorf("webhook_url is required when updater_type=webhook")
	}
	if cfg.UpdaterType == "sql" {
		if cfg.SQLDriver == "" || cfg.SQLDSN == "" || cfg.SQLUpdateQueryNamed == "" {
			return fmt.Errorf("sql_driver, sql_dsn, sql_update_query_named are required when updater_type=sql")
		}
		switch cfg.SQLPlaceholderStyle {
		case "question", "dollar":
		default:
			return fmt.Errorf("sql_placeholder_style must be question|dollar")
		}
	}
	return nil
}

func maskDSN(dsn string) string {
	if dsn == "" {
		return ""
	}
	return "***"
}
