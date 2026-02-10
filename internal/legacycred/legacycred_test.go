package legacycred

import (
	"context"
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
)

func TestRotateAndRead(t *testing.T) {
	ctx := context.Background()
	storage := &logical.InmemStorage{}

	b := newBackend()
	conf := &logical.BackendConfig{StorageView: storage}
	if err := b.Setup(ctx, conf); err != nil {
		t.Fatalf("setup: %v", err)
	}

	// Configure to noop updater.
	_, err := b.HandleRequest(ctx, &logical.Request{
		Storage:   storage,
		Operation: logical.UpdateOperation,
		Path:      "config",
		Data: map[string]any{
			"updater_type":      "noop",
			"default_ttl":       60,
			"default_hash_type": "bcrypt",
		},
	})
	if err != nil {
		t.Fatalf("config write: %v", err)
	}

	// Rotate
	rotateResp, err := b.HandleRequest(ctx, &logical.Request{
		Storage:   storage,
		Operation: logical.UpdateOperation,
		Path:      "rotate/alice",
		Data:      map[string]any{},
	})
	if err != nil {
		t.Fatalf("rotate: %v", err)
	}
	if rotateResp == nil || rotateResp.Data["plaintext"] == "" {
		t.Fatalf("expected plaintext")
	}

	// Read creds
	credsResp, err := b.HandleRequest(ctx, &logical.Request{
		Storage:   storage,
		Operation: logical.ReadOperation,
		Path:      "creds/alice",
	})
	if err != nil {
		t.Fatalf("creds read: %v", err)
	}
	if credsResp == nil {
		t.Fatalf("expected creds response")
	}
	if credsResp.Data["plaintext"] != rotateResp.Data["plaintext"] {
		t.Fatalf("plaintext mismatch")
	}
}
