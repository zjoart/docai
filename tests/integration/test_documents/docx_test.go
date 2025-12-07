package test_documents

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/zjoart/docai/internal/documents"
)

func TestDOCXUploadFlow(t *testing.T) {

	env := SetupTestEnv(t)
	r := env.Router

	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	f, err := zipWriter.Create("word/document.xml")
	if err != nil {
		t.Fatalf("Failed to create zip entry: %v", err)
	}

	xmlContent := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
	<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
		<w:body>
			<w:p>
				<w:r>
					<w:t>Hello</w:t>
				</w:r>
			</w:p>
			<w:p>
				<w:r>
					<w:t>World</w:t>
				</w:r>
			</w:p>
		</w:body>
	</w:document>`

	if _, err := f.Write([]byte(xmlContent)); err != nil {
		t.Fatalf("Failed to write xml content: %v", err)
	}

	zipWriter.Close()
	docxBytes := buf.Bytes()
	// Prepare multipart request
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	filename := fmt.Sprintf("test_%s.docx", uuid.New().String())
	part, _ := writer.CreateFormFile("file", filename)
	part.Write(docxBytes)
	writer.Close()

	req := httptest.NewRequest("POST", "/documents/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		t.Fatalf("Upload failed: status %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var respData struct {
		Message  string             `json:"message"`
		Document documents.Document `json:"document"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	doc := respData.Document

	if doc.ID == uuid.Nil {
		t.Fatal("Expected valid ID")
	}

	if doc.Status != "uploaded" {
		t.Errorf("Expected status uploaded, got %s", doc.Status)
	}

	t.Logf("Uploaded Doc ID: %s", doc.ID)

	if doc.ExtractedText == "" {

		getReq := httptest.NewRequest("GET", "/documents/"+doc.ID.String(), nil)
		getW := httptest.NewRecorder()
		r.ServeHTTP(getW, getReq)
		json.NewDecoder(getW.Body).Decode(&doc)
	}

	expectedText := "\nHello\nWorld"
	if doc.ExtractedText != expectedText {
		t.Errorf("Expected extracted text %q, got %q", expectedText, doc.ExtractedText)
	}
}
