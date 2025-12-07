package extractor

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
)

func ExtractTextFromDOCX(reader io.ReaderAt, size int64) (string, error) {
	r, err := zip.NewReader(reader, size)
	if err != nil {
		return "", fmt.Errorf("failed to open docx zip: %w", err)
	}

	var docXML *zip.File
	for _, f := range r.File {
		if f.Name == "word/document.xml" {
			docXML = f
			break
		}
	}

	if docXML == nil {
		return "", fmt.Errorf("word/document.xml not found in docx")
	}

	rc, err := docXML.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open document.xml: %w", err)
	}
	defer rc.Close()

	decoder := xml.NewDecoder(rc)
	var extractedText bytes.Buffer

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("error parsing xml: %w", err)
		}

		switch t := token.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "t":

				var text string
				if err := decoder.DecodeElement(&text, &t); err != nil {
					return "", fmt.Errorf("failed to decode text element: %w", err)
				}
				extractedText.WriteString(text)
			case "p":

				extractedText.WriteString("\n")
			}
		}
	}

	return extractedText.String(), nil
}
