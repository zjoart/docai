package test_documents

import (
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

func TestDocumentFlow(t *testing.T) {

	env := SetupTestEnv(t)
	r := env.Router

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	filename := fmt.Sprintf("test_%s.txt", uuid.New().String())
	part, _ := writer.CreateFormFile("file", filename)
	part.Write([]byte("This is a test invoice. Date: 2023-10-27. Total: $500."))
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

	var doc documents.Document
	json.NewDecoder(resp.Body).Decode(&doc)

	if doc.ID == uuid.Nil {
		t.Fatal("Expected valid ID")
	}

	t.Logf("Uploaded Doc ID: %s", doc.ID)

	if doc.FileUrl == "" {
		t.Error("Expected FileUrl to be set")
	}

	{
		bodyDup := new(bytes.Buffer)
		writerDup := multipart.NewWriter(bodyDup)
		partDup, _ := writerDup.CreateFormFile("file", filename) // Use same filename
		partDup.Write([]byte("Duplicate content"))
		writerDup.Close()

		reqDup := httptest.NewRequest("POST", "/documents/upload", bodyDup)
		reqDup.Header.Set("Content-Type", writerDup.FormDataContentType())
		wDup := httptest.NewRecorder()

		r.ServeHTTP(wDup, reqDup)

		if wDup.Code != http.StatusOK {
			t.Fatalf("Expected duplicate upload to return 200 OK, got %d. Body: %s", wDup.Code, wDup.Body.String())
		}

		var docDup documents.Document
		if err := json.NewDecoder(wDup.Body).Decode(&docDup); err != nil {
			t.Fatalf("Failed to decode duplicate response: %v", err)
		}

		if docDup.ID != doc.ID {
			t.Errorf("Expected duplicate doc ID to match original %s, got %s", doc.ID, docDup.ID)
		}
	}

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
