package usecase

import "io"

type ExtractDocumentInput struct {
	File     io.Reader
	FileName string
	FileSize int64
}
