package integration_test

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/zjoart/docai/internal/config"
	"github.com/zjoart/docai/internal/database"
	"github.com/zjoart/docai/internal/documents"
	"github.com/zjoart/docai/internal/documents/analyzer"
	"github.com/zjoart/docai/internal/storage"
)

func TestDOCXUploadFlow(t *testing.T) {

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	db, err := database.Connect(cfg.DBURL)
	if err != nil {
		t.Fatalf("DB connect failed: %v", err)
	}

	// clean db
	db.Exec("DELETE FROM documents")

	minioClient, err := storage.NewMinioClient(cfg.MinioEndpoint, cfg.MinioAccessKey, cfg.MinioSecretKey, cfg.MinioBucket)
	if err != nil {
		t.Fatalf("Minio init failed: %v", err)
	}

	err = minioClient.EnsureBucket(context.Background())
	if err != nil {
		t.Fatalf("Failed to ensure bucket: %v", err)
	}

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

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "test.docx")
	part.Write(docxBytes)
	writer.Close()

	req := httptest.NewRequest("POST", "/documents/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	repo := documents.NewRepository(db)
	ai := analyzer.NewAnalyzer(cfg.OpenRouterAPIKey)
	svc := documents.NewService(repo, minioClient, ai)
	h := documents.NewHandler(svc)

	r := mux.NewRouter()
	documents.RegisterRoutes(r, h)

	r.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		t.Fatalf("Upload failed: status %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var doc documents.Document
	json.NewDecoder(resp.Body).Decode(&doc)

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
