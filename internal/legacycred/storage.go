package legacycred

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/vault/sdk/logical"

	"github.com/kimthiphuongthao/vaulthashicorp/internal/legacycred/config"
)

func (b *backend) loadConfig(ctx context.Context, s logical.Storage) (config.Config, error) {
	raw, err := s.Get(ctx, storageKeyConfigPrefix)
	if err != nil {
		return config.Config{}, err
	}
	if raw == nil {
		return config.DefaultConfig(), nil
	}
	var cfg config.Config
	if err := json.Unmarshal(raw.Value, &cfg); err != nil {
		return config.Config{}, err
	}
	if cfg.DefaultTTL == 0 {
		cfg.DefaultTTL = config.DefaultConfig().DefaultTTL
	}
	if cfg.PasswordLength == 0 {
		cfg.PasswordLength = config.DefaultConfig().PasswordLength
	}
	if cfg.PasswordCharset == "" {
		cfg.PasswordCharset = config.DefaultConfig().PasswordCharset
	}
	if cfg.DefaultHashType == "" {
		cfg.DefaultHashType = config.DefaultConfig().DefaultHashType
	}
	if cfg.BcryptCost == 0 {
		cfg.BcryptCost = config.DefaultConfig().BcryptCost
	}
	if cfg.ShaSaltPosition == "" {
		cfg.ShaSaltPosition = config.DefaultConfig().ShaSaltPosition
	}
	if cfg.ShaOutputEncoding == "" {
		cfg.ShaOutputEncoding = config.DefaultConfig().ShaOutputEncoding
	}
	if cfg.PBKDF2Iterations == 0 {
		cfg.PBKDF2Iterations = config.DefaultConfig().PBKDF2Iterations
	}
	if cfg.PBKDF2KeyLength == 0 {
		cfg.PBKDF2KeyLength = config.DefaultConfig().PBKDF2KeyLength
	}
	if cfg.PBKDF2SaltLength == 0 {
		cfg.PBKDF2SaltLength = config.DefaultConfig().PBKDF2SaltLength
	}
	if cfg.PBKDF2Hash == "" {
		cfg.PBKDF2Hash = config.DefaultConfig().PBKDF2Hash
	}
	if cfg.UpdaterType == "" {
		cfg.UpdaterType = config.DefaultConfig().UpdaterType
	}
	if cfg.WebhookMethod == "" {
		cfg.WebhookMethod = config.DefaultConfig().WebhookMethod
	}
	if cfg.SQLPlaceholderStyle == "" {
		cfg.SQLPlaceholderStyle = config.DefaultConfig().SQLPlaceholderStyle
	}
	return cfg, nil
}

func (b *backend) saveConfig(ctx context.Context, s logical.Storage, cfg config.Config) error {
	buf, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	return s.Put(ctx, &logical.StorageEntry{Key: storageKeyConfigPrefix, Value: buf})
}

func subjectKey(subject string) string {
	return storageKeySubjectsPrefx + subject
}

func (b *backend) loadSubject(ctx context.Context, s logical.Storage, subject string) (*SubjectRecord, error) {
	raw, err := s.Get(ctx, subjectKey(subject))
	if err != nil {
		return nil, err
	}
	if raw == nil {
		return nil, nil
	}
	var rec SubjectRecord
	if err := json.Unmarshal(raw.Value, &rec); err != nil {
		return nil, fmt.Errorf("invalid subject record: %w", err)
	}
	return &rec, nil
}

func (b *backend) saveSubject(ctx context.Context, s logical.Storage, rec SubjectRecord) error {
	buf, err := json.Marshal(rec)
	if err != nil {
		return err
	}
	return s.Put(ctx, &logical.StorageEntry{Key: subjectKey(rec.Subject), Value: buf})
}

func (b *backend) deleteSubject(ctx context.Context, s logical.Storage, subject string) error {
	return s.Delete(ctx, subjectKey(subject))
}
