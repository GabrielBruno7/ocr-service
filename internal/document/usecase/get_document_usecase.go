package usecase

import (
	"context"
	"errors"

	"github.com/gabrielbruno7/ocr-service/internal/apperr"
	"github.com/gabrielbruno7/ocr-service/internal/document/domain"
	"github.com/google/uuid"
)

type GetDocumentUsecase struct {
	ctx                context.Context
	documentRepository domain.DocumentRepositoryInterface
}

func NewGetDocumentUsecase(
	ctx context.Context,
	documentRepository domain.DocumentRepositoryInterface,
) *GetDocumentUsecase {
	return &GetDocumentUsecase{
		ctx:                ctx,
		documentRepository: documentRepository,
	}
}

func (uc *GetDocumentUsecase) Execute(id uuid.UUID) (*domain.Document, *apperr.AppError) {
	doc, err := uc.documentRepository.GetByID(uc.ctx, id)
	if err != nil {
		var appErr *apperr.AppError

		if errors.As(err, &appErr) {
			return nil, appErr
		}

		return nil, apperr.New(apperr.CodeInternal, err)
	}

	return doc, nil
}
