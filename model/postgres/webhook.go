package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
	"github.com/pkg/errors"
)

// ErrWebhookExists webhook already exists
var ErrWebhookExists = errors.New("postgres: webhook exists")

// ErrWebhookNotFound webhook not found error
var ErrWebhookNotFound = errors.New("postgres: webhook not found")

// WebhookRow holds a single row in the webhook table.
type WebhookRow struct {
	id       int
	UUID     string
	URL      string
	Events   []string
	Enabled  bool
	Created  time.Time
	Modified time.Time
}

// CreateWebhook adds a new row to the webhook table.
func (m *PgModel) CreateWebhook(ctx context.Context, url string, events []string) (*WebhookRow, error) {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "postgres: db.BeginTx failed")
	}

	// 1. Check if the webhook already exists
	q1 := `SELECT EXISTS(SELECT 1 FROM webhook WHERE url = $1) AS exists`
	var exists bool
	err = tx.QueryRowContext(ctx, q1, url).Scan(&exists)
	if err != nil {
		return nil, errors.Wrapf(err, "postgres: tx.QueryRowContext(ctx, q1=%q, url=%q)", q1, url)
	}
	if exists {
		return nil, ErrWebhookExists
	}

	// 2. Insert the new webhook
	q2 := `
		INSERT INTO webhook (url, events, enabled)
		VALUES ($1, $2, true)
		RETURNING id, uuid, url, events, enabled, created, modified
	`
	var w WebhookRow
	row := tx.QueryRowContext(ctx, q2, url, pq.Array(events))
	if err := row.Scan(&w.id, &w.UUID, &w.URL, pq.Array(&w.Events), &w.Enabled, &w.Created, &w.Modified); err != nil {
		return nil, errors.Wrapf(err, "postgres: query row context q2=%q", q2)
	}

	if err = tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "postgres: tx.Commit failed")
	}
	return &w, nil
}

// GetWebhook retrieves a single webhook by uuid.
func (m *PgModel) GetWebhook(ctx context.Context, webhookUUID string) (*WebhookRow, error) {
	q1 := `
		SELECT id, uuid, url, events, enabled, created, modified
		FROM webhook
		WHERE uuid = $1
	`
	var w WebhookRow
	err := m.db.QueryRowContext(ctx, q1, webhookUUID).Scan(&w.id, &w.UUID, &w.URL, pq.Array(&w.Events), &w.Enabled, &w.Created, &w.Modified)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrWebhookNotFound
		}
		return nil, errors.Wrapf(err, "postgres: query row context failed for q1=%q", q1)
	}
	return &w, nil
}

// GetWebhooks retrieves all webhooks.
func (m *PgModel) GetWebhooks(ctx context.Context) ([]*WebhookRow, error) {
	q1 := "SELECT id, uuid, url, events, enabled, created, modified FROM webhook"
	rows, err := m.db.QueryContext(ctx, q1)
	if err != nil {
		return nil, errors.Wrapf(err, "postgres: m.db.QueryContext(ctx) failed")
	}
	defer rows.Close()

	webhooks := make([]*WebhookRow, 0)
	for rows.Next() {
		var w WebhookRow
		if err = rows.Scan(&w.id, &w.UUID, &w.URL, pq.Array(&w.Events), &w.Enabled, &w.Created, &w.Modified); err != nil {
			return nil, errors.Wrap(err, "postgres: scan failed")
		}
		webhooks = append(webhooks, &w)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "postgres: rows.Err()")
	}
	return webhooks, nil
}

// UpdateWebhook does a partial update to a row in the webhook table.
func (m *PgModel) UpdateWebhook(ctx context.Context, webhookUUID string, url *string, events []string, enabled *bool) (*WebhookRow, error) {
	// 1. Check the webhook exists
	q1 := "SELECT id FROM webhook WHERE uuid = $1"
	var webhookID int
	err := m.db.QueryRowContext(ctx, q1, webhookUUID).Scan(&webhookID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrWebhookNotFound
		}
		return nil, errors.Wrapf(err, "postgres: query row context failed for q1=%q", q1)
	}

	// 2. Update the webhook
	argCounter := 1

	var set []string
	var queryArgs []interface{}

	if url != nil {
		set = append(set, fmt.Sprintf("url = $%d", argCounter))
		queryArgs = append(queryArgs, *url)
		argCounter++
	}

	if events != nil {
		set = append(set, fmt.Sprintf("events = $%d", argCounter))
		queryArgs = append(queryArgs, pq.Array(events))
		argCounter++
	}

	if enabled != nil {
		set = append(set, fmt.Sprintf("enabled = $%d", argCounter))
		queryArgs = append(queryArgs, *enabled)
		argCounter++
	}
	set = append(set, "modified = NOW()")
	queryArgs = append(queryArgs, webhookID)

	q2 := fmt.Sprintf(`
		UPDATE webhook
		SET %s
		WHERE id = $%d
		RETURNING id, uuid, url, events, enabled, created, modified
	`, strings.Join(set, ", "), argCounter)
	w := WebhookRow{}
	row := m.db.QueryRowContext(ctx, q2, queryArgs...)
	if err := row.Scan(&w.id, &w.UUID, &w.URL, pq.Array(&w.Events), &w.Enabled, &w.Created, &w.Modified); err != nil {
		return nil, errors.Wrapf(err, "postgres: query row context q2=%q failed", q2)
	}
	return &w, nil
}

// DeleteWebhook deletes a webhook by uuid.
func (m *PgModel) DeleteWebhook(ctx context.Context, webhookUUID string) error {
	// 1. Check if the webhook exists
	q1 := "SELECT id FROM webhook WHERE uuid = $1"
	var webhookID int
	err := m.db.QueryRowContext(ctx, q1, webhookUUID).Scan(&webhookID)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrWebhookNotFound
		}
		return errors.Wrapf(err, "postgres: m.db.QueryRowContext(ctx, q1=%q, webhookUUID=%q)", q1, webhookUUID)
	}

	// 2. Delete the webhook
	q2 := "DELETE FROM webhook WHERE id = $1"
	_, err = m.db.ExecContext(ctx, q2, webhookID)
	if err != nil {
		return errors.Wrapf(err, "postgres: m.db.ExecContext(ctx, q2=%q, webhookID=%d) failed", q2, webhookID)
	}
	return nil
}
