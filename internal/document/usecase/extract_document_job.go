package usecase

import "github.com/google/uuid"

type ExtractDocumentJob struct {
	DocumentID uuid.UUID `json:"document_id"`
}
