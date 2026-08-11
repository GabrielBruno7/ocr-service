package document

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	CreatePending(ctx context.Context, filename string) (uuid.UUID, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Document, error)
	MarkAsProcessing(ctx context.Context, id uuid.UUID) error
	MarkAsDone(ctx context.Context, id uuid.UUID, extractedText string) error
	MarkAsFailed(ctx context.Context, id uuid.UUID, errMsg string) error
}
