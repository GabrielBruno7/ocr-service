package document

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/gabrielbruno7/ocr-service/internal/apperr"
)

type postgresRepository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) CreatePending(ctx context.Context, filename string) (uuid.UUID, error) {
	var id uuid.UUID

	query := `
		INSERT INTO documents (status, filename)
		VALUES ($1, $2)
		RETURNING id
	`

	err := r.db.QueryRow(ctx, query, StatusPending, filename).Scan(&id)
	if err != nil {
		return uuid.Nil, apperr.New(apperr.CodeInternal,
			fmt.Errorf("erro ao criar documento pendente: %w", err))
	}

	return id, nil
}

func (r *postgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*Document, error) {
	var doc Document

	query := `
		SELECT id, status, filename, document_type, extracted_text, error_message, processing_started_at
		FROM documents
		WHERE id = $1
	`

	err := r.db.QueryRow(ctx, query, id).Scan(
		&doc.ID, &doc.Status, &doc.Filename, &doc.DocumentType, &doc.ExtractedText, &doc.ErrorMessage, &doc.ProcessingStartedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.New(apperr.CodeNotFound,
				fmt.Errorf("documento %s não encontrado: %w", id, err))
		}
		return nil, apperr.New(apperr.CodeInternal,
			fmt.Errorf("erro ao buscar documento %s: %w", id, err))
	}

	return &doc, nil
}

func (r *postgresRepository) MarkAsProcessing(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE documents SET status = $1, processing_started_at = now(), updated_at = now() WHERE id = $2`
	_, err := r.db.Exec(ctx, query, StatusProcessing, id)
	if err != nil {
		return apperr.New(apperr.CodeInternal,
			fmt.Errorf("erro ao marcar documento como em processamento: %w", err))
	}
	return nil
}

func (r *postgresRepository) MarkAsDone(ctx context.Context, id uuid.UUID, extractedText string, documentType DocumentType) error {
	query := `UPDATE documents SET status = $1, extracted_text = $2, document_type = $3, updated_at = now() WHERE id = $4`
	_, err := r.db.Exec(ctx, query, StatusDone, extractedText, string(documentType), id)
	if err != nil {
		return apperr.New(apperr.CodeInternal,
			fmt.Errorf("erro ao marcar documento como concluído: %w", err))
	}
	return nil
}

func (r *postgresRepository) MarkAsFailed(ctx context.Context, id uuid.UUID, errMsg string) error {
	query := `UPDATE documents SET status = $1, error_message = $2, updated_at = now() WHERE id = $3`
	_, err := r.db.Exec(ctx, query, StatusFailed, errMsg, id)
	if err != nil {
		return apperr.New(apperr.CodeInternal,
			fmt.Errorf("erro ao marcar documento como falho: %w", err))
	}
	return nil
}
