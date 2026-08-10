package document

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	StatusPending    = "pending"
	StatusProcessing = "processing"
	StatusDone       = "done"
	StatusFailed     = "failed"
)

type Document struct {
	ID                  uuid.UUID
	Status              string
	Filename            string
	ExtractedText       *string
	ErrorMessage        *string
	ProcessingStartedAt *time.Time
}

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreatePending(ctx context.Context, filename string) (uuid.UUID, error) {
	var id uuid.UUID

	query := `
		INSERT INTO documents (status, filename)
		VALUES ($1, $2)
		RETURNING id
	`

	err := r.db.QueryRow(ctx, query, StatusPending, filename).Scan(&id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("erro ao criar documento pendente: %w", err)
	}

	return id, nil
}

func (r *Repository) MarkAsDone(ctx context.Context, id uuid.UUID, extractedText string) error {
	query := `
		UPDATE documents
		SET status = $1, extracted_text = $2, updated_at = now()
		WHERE id = $3
	`

	_, err := r.db.Exec(ctx, query, StatusDone, extractedText, id)
	if err != nil {
		return fmt.Errorf("erro ao marcar documento como concluído: %w", err)
	}

	return nil
}

func (r *Repository) MarkAsFailed(ctx context.Context, id uuid.UUID, errMsg string) error {
	query := `
		UPDATE documents
		SET status = $1, error_message = $2, updated_at = now()
		WHERE id = $3
	`

	_, err := r.db.Exec(ctx, query, StatusFailed, errMsg, id)
	if err != nil {
		return fmt.Errorf("erro ao marcar documento como falho: %w", err)
	}

	return nil
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*Document, error) {
	var doc Document

	query := `
		SELECT id, status, filename, extracted_text, error_message, processing_started_at
		FROM documents
		WHERE id = $1
	`

	err := r.db.QueryRow(ctx, query, id).Scan(
		&doc.ID, &doc.Status, &doc.Filename, &doc.ExtractedText, &doc.ErrorMessage, &doc.ProcessingStartedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar documento: %w", err)
	}

	return &doc, nil
}

func (r *Repository) MarkAsProcessing(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE documents SET status = $1, processing_started_at = now(), updated_at = now() WHERE id = $2`
	_, err := r.db.Exec(ctx, query, StatusProcessing, id)
	if err != nil {
		return fmt.Errorf("erro ao marcar documento como em processamento: %w", err)
	}
	return nil
}
