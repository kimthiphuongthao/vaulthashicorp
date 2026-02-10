package updater

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/microsoft/go-mssqldb"

	"github.com/kimthiphuongthao/vaulthashicorp/internal/legacycred/config"
)

type UpdateRequest struct {
	Subject   string    `json:"subject"`
	Hash      string    `json:"hash"`
	HashType  string    `json:"hash_type"`
	Version   int64     `json:"version"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Updater interface {
	Update(ctx context.Context, req UpdateRequest) error
}

type NoopUpdater struct{}

func (n NoopUpdater) Update(ctx context.Context, req UpdateRequest) error { return nil }

type WebhookUpdater struct {
	URL     string
	Method  string
	Headers map[string]string
	Client  *http.Client
}

func (w WebhookUpdater) Update(ctx context.Context, req UpdateRequest) error {
	payload, err := json.Marshal(req)
	if err != nil {
		return err
	}
	method := w.Method
	if method == "" {
		method = "POST"
	}
	client := w.Client
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}

	httpReq, err := http.NewRequestWithContext(ctx, method, w.URL, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	for k, v := range w.Headers {
		httpReq.Header.Set(k, v)
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook status %d", resp.StatusCode)
	}
	return nil
}

type SQLUpdater struct {
	Driver           string
	DSN              string
	UpdateQueryNamed string
	PlaceholderStyle string // question|dollar
}

func (s SQLUpdater) Update(ctx context.Context, req UpdateRequest) error {
	query, args, err := compileNamedQuery(s.UpdateQueryNamed, s.PlaceholderStyle, req)
	if err != nil {
		return err
	}

	db, err := sql.Open(s.Driver, s.DSN)
	if err != nil {
		return err
	}
	defer db.Close()

	if _, err := db.ExecContext(ctx, query, args...); err != nil {
		return err
	}
	return nil
}

func FromConfig(cfg config.Config) (Updater, error) {
	switch cfg.UpdaterType {
	case "", "noop":
		return NoopUpdater{}, nil
	case "webhook":
		headers := map[string]string{}
		if strings.TrimSpace(cfg.WebhookHeaders) != "" {
			if err := json.Unmarshal([]byte(cfg.WebhookHeaders), &headers); err != nil {
				return nil, fmt.Errorf("invalid webhook_headers JSON: %w", err)
			}
		}
		return WebhookUpdater{URL: cfg.WebhookURL, Method: cfg.WebhookMethod, Headers: headers}, nil
	case "sql":
		return SQLUpdater{
			Driver:           cfg.SQLDriver,
			DSN:              cfg.SQLDSN,
			UpdateQueryNamed: cfg.SQLUpdateQueryNamed,
			PlaceholderStyle: cfg.SQLPlaceholderStyle,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported updater_type")
	}
}

func compileNamedQuery(named, style string, req UpdateRequest) (string, []any, error) {
	if named == "" {
		return "", nil, fmt.Errorf("sql_update_query_named is empty")
	}

	repl := func(n int) string {
		switch style {
		case "dollar":
			return fmt.Sprintf("$%d", n)
		case "", "question":
			return "?"
		default:
			return "?"
		}
	}

	args := make([]any, 0, 2)
	query := named

	// Fixed ordering: hash first, subject second.
	if strings.Contains(query, ":hash") {
		query = strings.ReplaceAll(query, ":hash", repl(1))
		args = append(args, req.Hash)
	}
	if strings.Contains(query, ":subject") {
		idx := len(args) + 1
		query = strings.ReplaceAll(query, ":subject", repl(idx))
		args = append(args, req.Subject)
	}

	if len(args) == 0 {
		return "", nil, fmt.Errorf("sql_update_query_named must contain :hash and/or :subject")
	}
	return query, args, nil
}
