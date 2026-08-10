package document

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Document struct {
	ID            uuid.UUID
	Status        string
	Filename      string
	ExtractedText *string
	ErrorMessage  *string
}

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, filename, extractedText string) (uuid.UUID, error) {
	var id uuid.UUID

	query := `
		INSERT INTO documents (status, filename, extracted_text)
		VALUES ('done', $1, $2)
		RETURNING id
	`

	err := r.db.QueryRow(ctx, query, filename, extractedText).Scan(&id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("erro ao inserir documento: %w", err)
	}

	return id, nil
}
