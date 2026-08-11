package document

import (
	"time"

	"github.com/google/uuid"
)

// Status possíveis de um documento, do início ao fim do processamento.
const (
	StatusPending    = "pending"
	StatusProcessing = "processing"
	StatusDone       = "done"
	StatusFailed     = "failed"
)

// Document representa um documento enviado para processamento de OCR,
// com seu estado atual e resultado (quando concluído).
type Document struct {
	ID                  uuid.UUID
	Status              string
	Filename            string
	ExtractedText       *string
	ErrorMessage        *string
	ProcessingStartedAt *time.Time
}
