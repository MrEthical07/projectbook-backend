package docsync

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/MrEthical07/superapi/internal/core/storage"
)

const queryClaimOutboxItems = `
WITH candidates AS (
	SELECT id
	FROM document_sync_outbox
	WHERE status IN ('pending', 'failed')
		AND next_attempt_at <= NOW()
		AND attempt_count < $1
	ORDER BY id ASC
	LIMIT $2
	FOR UPDATE SKIP LOCKED
)
UPDATE document_sync_outbox o
SET
	status = 'processing',
	attempt_count = o.attempt_count + 1,
	updated_at = NOW()
FROM candidates c
WHERE o.id = c.id
RETURNING
	o.id,
	o.project_id::text,
	o.artifact_type::text,
	o.artifact_id::text,
	o.operation,
	COALESCE(o.document_id, ''),
	o.document_revision,
	o.payload,
	o.attempt_count
`

const queryMarkOutboxCompleted = `
UPDATE document_sync_outbox
SET
	status = 'completed',
	last_error = NULL,
	updated_at = NOW()
WHERE id = $1
`

const queryMarkOutboxFailed = `
UPDATE document_sync_outbox
SET
	status = 'failed',
	last_error = $2,
	next_attempt_at = $3,
	updated_at = NOW()
WHERE id = $1
`

// Config controls processing cadence and retry behavior for outbox records.
type Config struct {
	PollInterval time.Duration
	BatchSize    int
	MaxAttempts  int
	RetryDelay   time.Duration
}

// DefaultConfig returns conservative defaults for background outbox processing.
func DefaultConfig() Config {
	return Config{
		PollInterval: 500 * time.Millisecond,
		BatchSize:    25,
		MaxAttempts:  8,
		RetryDelay:   3 * time.Second,
	}
}

// Processor consumes SQL outbox rows and applies document operations to Mongo.
type Processor struct {
	relationalStore storage.RelationalStore
	documentStore   storage.DocumentStore
	cfg             Config
}

// NewProcessor validates dependencies and creates one processor instance.
func NewProcessor(relationalStore storage.RelationalStore, documentStore storage.DocumentStore, cfg Config) (*Processor, error) {
	if relationalStore == nil {
		return nil, errors.New("outbox processor requires relational store")
	}
	if documentStore == nil {
		return nil, errors.New("outbox processor requires document store")
	}
	if cfg.PollInterval <= 0 {
		cfg.PollInterval = 500 * time.Millisecond
	}
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 25
	}
	if cfg.MaxAttempts <= 0 {
		cfg.MaxAttempts = 8
	}
	if cfg.RetryDelay <= 0 {
		cfg.RetryDelay = 3 * time.Second
	}

	return &Processor{relationalStore: relationalStore, documentStore: documentStore, cfg: cfg}, nil
}

// Run starts a processing loop and stops when the context is canceled.
func (p *Processor) Run(ctx context.Context, onError func(error)) {
	if p == nil {
		return
	}

	ticker := time.NewTicker(p.cfg.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := p.ProcessOnce(ctx); err != nil && onError != nil {
				onError(err)
			}
		}
	}
}

// ProcessOnce performs one claim-and-process pass.
func (p *Processor) ProcessOnce(ctx context.Context) error {
	if p == nil {
		return nil
	}

	items := make([]outboxItem, 0, p.cfg.BatchSize)
	err := p.relationalStore.Execute(ctx, storage.RelationalQueryMany(
		queryClaimOutboxItems,
		func(row storage.RowScanner) error {
			var item outboxItem
			if err := row.Scan(
				&item.ID,
				&item.ProjectID,
				&item.ArtifactType,
				&item.ArtifactID,
				&item.Operation,
				&item.DocumentID,
				&item.DocumentRevision,
				&item.Payload,
				&item.AttemptCount,
			); err != nil {
				return err
			}
			items = append(items, item)
			return nil
		},
		p.cfg.MaxAttempts,
		p.cfg.BatchSize,
	))
	if err != nil {
		return fmt.Errorf("claim outbox rows: %w", err)
	}
	if len(items) == 0 {
		return nil
	}

	for _, item := range items {
		if err := p.processItem(ctx, item); err != nil {
			msg := trimError(err)
			nextAttemptAt := time.Now().UTC().Add(p.cfg.RetryDelay)
			_ = p.relationalStore.Execute(ctx, storage.RelationalExec(queryMarkOutboxFailed, item.ID, msg, nextAttemptAt))
		}
	}

	return nil
}

type outboxItem struct {
	ID               int64
	ProjectID        string
	ArtifactType     string
	ArtifactID       string
	Operation        string
	DocumentID       string
	DocumentRevision int
	Payload          []byte
	AttemptCount     int
}

func (p *Processor) processItem(ctx context.Context, item outboxItem) error {
	collection, err := collectionForArtifactType(item.ArtifactType)
	if err != nil {
		return err
	}

	op := strings.ToLower(strings.TrimSpace(item.Operation))
	switch op {
	case "upsert":
		if err := p.upsertDocument(ctx, collection, item); err != nil {
			return err
		}
	case "delete":
		if err := p.deleteDocument(ctx, collection, item); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported outbox operation %q", item.Operation)
	}

	if err := p.relationalStore.Execute(ctx, storage.RelationalExec(queryMarkOutboxCompleted, item.ID)); err != nil {
		return fmt.Errorf("mark outbox completed: %w", err)
	}
	return nil
}

func (p *Processor) upsertDocument(ctx context.Context, collection string, item outboxItem) error {
	payload := make(map[string]any)
	if len(item.Payload) > 0 {
		if err := json.Unmarshal(item.Payload, &payload); err != nil {
			return fmt.Errorf("decode outbox payload: %w", err)
		}
	}

	now := time.Now().UTC()
	doc := map[string]any{
		"artifact_id": item.ArtifactID,
		"project_id":  item.ProjectID,
		"revision":    item.DocumentRevision,
		"updated_at":  now,
	}
	if item.DocumentID != "" {
		doc["document_id"] = item.DocumentID
	}

	for k, v := range payload {
		doc[k] = v
	}
	if _, ok := doc["schema_version"]; !ok {
		doc["schema_version"] = 1
	}
	if _, ok := doc["content"]; !ok {
		doc["content"] = map[string]any{}
	}

	err := p.documentStore.Execute(ctx, storage.DocumentRun(
		collection+":update_one",
		map[string]any{
			"filter":  bson.M{"artifact_id": item.ArtifactID},
			"update":  bson.M{"$set": doc},
			"options": options.Update().SetUpsert(true),
		},
		nil,
	))
	if err != nil {
		return fmt.Errorf("upsert mongo document: %w", err)
	}
	return nil
}

func (p *Processor) deleteDocument(ctx context.Context, collection string, item outboxItem) error {
	err := p.documentStore.Execute(ctx, storage.DocumentRun(
		collection+":delete_one",
		map[string]any{
			"filter": bson.M{"artifact_id": item.ArtifactID},
		},
		nil,
	))
	if err != nil {
		return fmt.Errorf("delete mongo document: %w", err)
	}
	return nil
}

func collectionForArtifactType(artifactType string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(artifactType)) {
	case "story":
		return "story_documents", nil
	case "journey":
		return "journey_documents", nil
	case "problem":
		return "problem_documents", nil
	case "idea":
		return "idea_documents", nil
	case "task":
		return "task_documents", nil
	case "feedback":
		return "feedback_documents", nil
	case "page":
		return "page_documents", nil
	case "resource":
		return "resource_documents", nil
	default:
		return "", fmt.Errorf("unsupported artifact type %q", artifactType)
	}
}

func trimError(err error) string {
	if err == nil {
		return ""
	}
	msg := strings.TrimSpace(err.Error())
	if len(msg) <= 4000 {
		return msg
	}
	return msg[:4000]
}
