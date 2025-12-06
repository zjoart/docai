package extractor

import (
	"bytes"
	"io"

	"github.com/ledongthuc/pdf"
	"github.com/zjoart/docai/pkg/logger"
)

func ExtractTextFromPDF(reader io.ReaderAt, size int64) (string, error) {
	r, err := pdf.NewReader(reader, size)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer

	totalPages := r.NumPage()

	for pageIndex := 1; pageIndex <= totalPages; pageIndex++ {
		p := r.Page(pageIndex)
		if p.V.IsNull() {
			continue
		}

		text, err := p.GetPlainText(nil)
		if err != nil {
			logger.Error("Error extracting text from PDF", logger.WithError(err))
			continue
		}
		buf.WriteString(text)
		buf.WriteString("\n")
	}

	return buf.String(), nil
}
