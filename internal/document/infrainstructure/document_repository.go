package infrainstructure

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ocr-service/ocr-service/internal/apperr"
	"github.com/ocr-service/ocr-service/internal/document/domain"
)

type documentRepository struct {
	db *pgxpool.Pool
}

func NewDocumentRepository(db *pgxpool.Pool) domain.DocumentRepositoryInterface {
	return &documentRepository{db: db}
}

func (r *documentRepository) CreatePending(ctx context.Context, filename string) (uuid.UUID, error) {
	var id uuid.UUID

	query := `
		INSERT INTO documents (status, filename)
		VALUES ($1, $2)
		RETURNING id
	`

	err := r.db.QueryRow(ctx, query, domain.StatusPending, filename).Scan(&id)
	if err != nil {
		return uuid.Nil, apperr.New(apperr.CodeInternal,
			fmt.Errorf("erro ao criar documento pendente: %w", err))
	}

	return id, nil
}

func (r *documentRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Document, error) {
	var doc domain.Document
	var fieldsJSON []byte

	query := `
		SELECT id, status, filename, document_type, extracted_text, extracted_fields, error_message, processing_started_at
		FROM documents
		WHERE id = $1
	`

	err := r.db.QueryRow(ctx, query, id).Scan(
		&doc.ID, &doc.Status, &doc.Filename, &doc.DocumentType, &doc.ExtractedText, &fieldsJSON, &doc.ErrorMessage, &doc.ProcessingStartedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.New(apperr.CodeNotFound,
				fmt.Errorf("documento %s não encontrado: %w", id, err))
		}
		return nil, apperr.New(apperr.CodeInternal,
			fmt.Errorf("erro ao buscar documento %s: %w", id, err))
	}

	if fieldsJSON != nil {
		var fields domain.ExtractedFields
		if err := json.Unmarshal(fieldsJSON, &fields); err == nil {
			doc.ExtractedFields = &fields
		}
	}

	return &doc, nil
}

func (r *documentRepository) MarkAsProcessing(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE documents SET status = $1, processing_started_at = now(), updated_at = now() WHERE id = $2`
	_, err := r.db.Exec(ctx, query, domain.StatusProcessing, id)
	if err != nil {
		return apperr.New(apperr.CodeInternal,
			fmt.Errorf("erro ao marcar documento como em processamento: %w", err))
	}
	return nil
}

func (r *documentRepository) MarkAsDone(
	ctx context.Context,
	id uuid.UUID,
	extractedText string,
	documentType domain.DocumentType,
	fields domain.ExtractedFields,
) error {
	fieldsJSON, err := json.Marshal(fields)
	if err != nil {
		return apperr.New(apperr.CodeInternal,
			fmt.Errorf("erro ao serializar campos extraídos: %w", err))
	}

	query := `
		UPDATE documents
		SET status = $1, extracted_text = $2, document_type = $3, extracted_fields = $4, updated_at = now()
		WHERE id = $5
	`

	_, err = r.db.Exec(ctx, query, domain.StatusDone, extractedText, string(documentType), fieldsJSON, id)
	if err != nil {
		return apperr.New(apperr.CodeInternal,
			fmt.Errorf("erro ao marcar documento como concluído: %w", err))
	}
	return nil
}

func (r *documentRepository) MarkAsFailed(ctx context.Context, id uuid.UUID, errMsg string) error {
	query := `UPDATE documents SET status = $1, error_message = $2, updated_at = now() WHERE id = $3`
	_, err := r.db.Exec(ctx, query, domain.StatusFailed, errMsg, id)
	if err != nil {
		return apperr.New(apperr.CodeInternal,
			fmt.Errorf("erro ao marcar documento como falho: %w", err))
	}
	return nil
}
