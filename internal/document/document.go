package document

import (
	"time"

	"github.com/google/uuid"
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
	DocumentType        *string
	ExtractedText       *string
	ErrorMessage        *string
	ProcessingStartedAt *time.Time
}
