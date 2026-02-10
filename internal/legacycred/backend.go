package legacycred

import (
	"context"
	"time"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

const (
	storageKeyConfigPrefix  = "config"
	storageKeySubjectsPrefx = "subjects/"
)

type backend struct {
	*framework.Backend
	realClock func() time.Time
}

func Factory(ctx context.Context, conf *logical.BackendConfig) (logical.Backend, error) {
	b := newBackend()
	if err := b.Setup(ctx, conf); err != nil {
		return nil, err
	}
	return b, nil
}

func newBackend() *backend {
	b := &backend{realClock: time.Now}

	b.Backend = &framework.Backend{
		Help: "Legacy credential rotator: generates plaintext, hashes it, updates legacy credential store, and (optionally) serves plaintext to OpenIG.",
		Paths: []*framework.Path{
			b.pathConfig(),
			b.pathRotate(),
			b.pathCreds(),
		},
		BackendType: logical.TypeLogical,
	}

	return b
}
