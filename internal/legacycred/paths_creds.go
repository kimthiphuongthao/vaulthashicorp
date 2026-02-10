package legacycred

import (
	"context"
	"time"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

func (b *backend) pathCreds() *framework.Path {
	return &framework.Path{
		Pattern: "creds/" + framework.GenericNameRegex("subject"),
		HelpSynopsis: "Read plaintext credentials for a subject (for OpenIG).",
		Fields: map[string]*framework.FieldSchema{
			"subject": {Type: framework.TypeString, Description: "Subject identifier (username/email).", Required: true},
		},
		Operations: map[logical.Operation]framework.OperationHandler{
			logical.ReadOperation: &framework.PathOperation{Callback: b.handleCredsRead},
		},
	}
}

func (b *backend) handleCredsRead(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	subject := d.Get("subject").(string)
	cfg, err := b.loadConfig(ctx, req.Storage)
	if err != nil {
		return nil, err
	}

	rec, err := b.loadSubject(ctx, req.Storage, subject)
	if err != nil {
		return nil, err
	}
	if rec == nil {
		return nil, nil
	}

	now := b.realClock().UTC()
	if !rec.ExpiresAt.IsZero() && now.After(rec.ExpiresAt) {
		_ = b.deleteSubject(ctx, req.Storage, subject)
		return nil, nil
	}

	if cfg.OneTimeRead {
		_ = b.deleteSubject(ctx, req.Storage, subject)
	}

	return &logical.Response{Data: map[string]any{
		"subject":    rec.Subject,
		"plaintext":  rec.Plaintext,
		"hash_type":  rec.HashType,
		"version":    rec.Version,
		"updated_at": rec.UpdatedAt.Format(time.RFC3339),
		"expires_at": rec.ExpiresAt.Format(time.RFC3339),
	}}, nil
}
