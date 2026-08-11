package document

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/gabrielbruno7/ocr-service/internal/apperr"
)

type memoryRepository struct {
	mu   sync.Mutex
	docs map[uuid.UUID]*Document
}

func newMemoryRepository() *memoryRepository {
	return &memoryRepository{docs: make(map[uuid.UUID]*Document)}
}

func (r *memoryRepository) CreatePending(ctx context.Context, filename string) (uuid.UUID, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	id := uuid.New()
	r.docs[id] = &Document{ID: id, Status: StatusPending, Filename: filename}
	return id, nil
}

func (r *memoryRepository) GetByID(ctx context.Context, id uuid.UUID) (*Document, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	doc, ok := r.docs[id]
	if !ok {
		return nil, apperr.New(apperr.CodeNotFound, fmt.Errorf("documento %s não encontrado", id))
	}
	return doc, nil
}

func (r *memoryRepository) MarkAsProcessing(ctx context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	doc, ok := r.docs[id]
	if !ok {
		return apperr.New(apperr.CodeNotFound, fmt.Errorf("documento %s não encontrado", id))
	}
	now := time.Now()
	doc.Status = StatusProcessing
	doc.ProcessingStartedAt = &now
	return nil
}

func (r *memoryRepository) MarkAsDone(ctx context.Context, id uuid.UUID, extractedText string, documentType DocumentType) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	doc, ok := r.docs[id]
	if !ok {
		return apperr.New(apperr.CodeNotFound, fmt.Errorf("documento %s não encontrado", id))
	}
	doc.Status = StatusDone
	doc.ExtractedText = &extractedText
	docType := string(documentType)
	doc.DocumentType = &docType
	return nil
}

func (r *memoryRepository) MarkAsFailed(ctx context.Context, id uuid.UUID, errMsg string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	doc, ok := r.docs[id]
	if !ok {
		return apperr.New(apperr.CodeNotFound, fmt.Errorf("documento %s não encontrado", id))
	}
	doc.Status = StatusFailed
	doc.ErrorMessage = &errMsg
	return nil
}

func (r *memoryRepository) seed(doc *Document) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.docs[doc.ID] = doc
}
