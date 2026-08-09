package ocr

type Processor interface {
	Process(imagePath string) (string, error)
}
