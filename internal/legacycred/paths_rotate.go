package legacycred

import (
	"context"
	"time"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"

	"github.com/kimthiphuongthao/vaulthashicorp/internal/legacycred/hash"
	"github.com/kimthiphuongthao/vaulthashicorp/internal/legacycred/updater"
)

func (b *backend) pathRotate() *framework.Path {
	return &framework.Path{
		Pattern: "rotate/" + framework.GenericNameRegex("subject"),
		HelpSynopsis: "Rotate plaintext + hash for a subject, update legacy credential store, then store plaintext for OpenIG to use.",
		Fields: map[string]*framework.FieldSchema{
			"subject": {Type: framework.TypeString, Description: "Subject identifier (username/email).", Required: true},
			"hash_type": {Type: framework.TypeString, Description: "Override hash type for this rotation: bcrypt|sha256|sha512|pbkdf2."},
			"ttl": {Type: framework.TypeDurationSecond, Description: "Override TTL for this subject plaintext."},
		},
		Operations: map[logical.Operation]framework.OperationHandler{
			logical.UpdateOperation: &framework.PathOperation{Callback: b.handleRotate},
		},
	}
}

func (b *backend) handleRotate(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	subject := d.Get("subject").(string)
	cfg, err := b.loadConfig(ctx, req.Storage)
	if err != nil {
		return nil, err
	}

	hashType := cfg.DefaultHashType
	if v, ok := d.GetOk("hash_type"); ok {
		hashType = v.(string)
	}

	ttl := cfg.DefaultTTL
	if v, ok := d.GetOk("ttl"); ok {
		ttl = time.Duration(v.(int)) * time.Second
	}

	plaintext, err := generatePassword(cfg.PasswordLength, cfg.PasswordCharset)
	if err != nil {
		return nil, err
	}

	h, err := hash.Hash(cfg, hashType, plaintext)
	if err != nil {
		return logical.ErrorResponse(err.Error()), nil
	}

	up, err := updater.FromConfig(cfg)
	if err != nil {
		return logical.ErrorResponse(err.Error()), nil
	}

	prev, err := b.loadSubject(ctx, req.Storage, subject)
	if err != nil {
		return nil, err
	}
	var nextVersion int64 = 1
	if prev != nil {
		nextVersion = prev.Version + 1
	}

	now := b.realClock().UTC()
	rec := SubjectRecord{
		Subject:   subject,
		Plaintext: plaintext,
		Hash:      h,
		HashType:  hashType,
		Version:   nextVersion,
		UpdatedAt: now,
		ExpiresAt: now.Add(ttl),
	}

	if err := up.Update(ctx, updater.UpdateRequest{
		Subject:   subject,
		Hash:      h,
		HashType:  hashType,
		Version:   nextVersion,
		UpdatedAt: now,
	}); err != nil {
		return logical.ErrorResponse("legacy update failed: %v", err), nil
	}

	if err := b.saveSubject(ctx, req.Storage, rec); err != nil {
		return nil, err
	}

	return &logical.Response{Data: map[string]any{
		"subject":    subject,
		"plaintext":  plaintext,
		"hash":       h,
		"hash_type":  hashType,
		"version":    nextVersion,
		"updated_at": rec.UpdatedAt.Format(time.RFC3339),
		"expires_at": rec.ExpiresAt.Format(time.RFC3339),
	}}, nil
}
