package document

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gabrielbruno7/ocr-service/internal/platform/queue"
)

type OCRJob struct {
	DocumentID string `json:"document_id"`
}

func PublishOCRJob(ctx context.Context, q *queue.Queue, documentID string) error {
	job := OCRJob{DocumentID: documentID}

	body, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("erro ao serializar job: %w", err)
	}

	return q.Publish(ctx, body)
}
