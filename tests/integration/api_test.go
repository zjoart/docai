package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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

func TestDocumentFlow(t *testing.T) {

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

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "test.txt")
	part.Write([]byte("This is a test invoice. Date: 2023-10-27. Total: $500."))
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

	t.Logf("Uploaded Doc ID: %s", doc.ID)

	analyzeReq := httptest.NewRequest("POST", fmt.Sprintf("/documents/%s/analyze", doc.ID), nil)
	analyzeW := httptest.NewRecorder()
	r.ServeHTTP(analyzeW, analyzeReq)

	analyzeResp := analyzeW.Result()
	if analyzeResp.StatusCode != http.StatusOK {
		t.Fatalf("Analyze step failed: Status %d. Body: %s", analyzeResp.StatusCode, analyzeW.Body.String())
	}

	var analyzeDoc documents.Document
	if err := json.NewDecoder(analyzeResp.Body).Decode(&analyzeDoc); err != nil {
		t.Fatalf("Failed to decode analyze response: %v", err)
	}
	t.Logf("Analysis Summary: %s", analyzeDoc.Summary)

	getReq := httptest.NewRequest("GET", fmt.Sprintf("/documents/%s", doc.ID), nil)
	getW := httptest.NewRecorder()
	r.ServeHTTP(getW, getReq)

	getResp := getW.Result()
	if getResp.StatusCode != http.StatusOK {
		t.Fatalf("Get failed: status %d", getResp.StatusCode)
	}
}
