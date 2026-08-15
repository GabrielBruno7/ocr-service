package usecase

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"

	"github.com/ocr-service/ocr-service/internal/apperr"
	"github.com/ocr-service/ocr-service/internal/document/domain"
	"github.com/ocr-service/ocr-service/internal/platform/queue"
)

type ExtractDocumentUsecase struct {
	ctx                context.Context
	uploadDirectory    string
	queue              queue.Publisher
	documentRepository domain.DocumentRepositoryInterface
}

func NewExtractDocumentUsecase(
	ctx context.Context,
	queue queue.Publisher,
	uploadDirectory string,
	documentRepository domain.DocumentRepositoryInterface,
) *ExtractDocumentUsecase {
	return &ExtractDocumentUsecase{
		ctx:                ctx,
		uploadDirectory:    uploadDirectory,
		queue:              queue,
		documentRepository: documentRepository,
	}
}

func (uc *ExtractDocumentUsecase) Execute(input ExtractDocumentInput) (*string, *apperr.AppError) {
	id, err := uc.documentRepository.CreatePending(uc.ctx, input.FileName)
	if err != nil {
		return nil, apperr.New(apperr.CodeOCRFailed, err)
	}

	if err := uc.saveUploadedFile(input.File, id.String(), input.FileName); err != nil {
		return nil, apperr.New(apperr.CodeOCRFailed, err)
	}

	body, err := json.Marshal(ExtractDocumentJob{DocumentID: id})
	if err != nil {
		return nil, apperr.New(apperr.CodeOCRFailed, err)
	}

	if err := uc.queue.Publish(uc.ctx, body); err != nil {
		return nil, apperr.New(apperr.CodeOCRFailed, err)
	}

	idStr := id.String()

	return &idStr, nil
}

func (uc *ExtractDocumentUsecase) saveUploadedFile(file io.Reader, id, fileName string) error {
	destPath := filepath.Join(uc.uploadDirectory, id+filepath.Ext(fileName))

	dst, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return err
	}

	return nil
}
