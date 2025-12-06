package documents

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/zjoart/docai/internal/documents/analyzer"
	"github.com/zjoart/docai/internal/documents/extractor"
	"github.com/zjoart/docai/internal/storage"
	"github.com/zjoart/docai/pkg/logger"
)

type Service struct {
	repo     Repository
	storage  *storage.Client
	analyzer *analyzer.Analyzer
}

func NewService(repo Repository, storage *storage.Client, analyzer *analyzer.Analyzer) *Service {
	return &Service{
		repo:     repo,
		storage:  storage,
		analyzer: analyzer,
	}
}

func (s *Service) UploadDocument(ctx context.Context, filename string, reader io.Reader, size int64, contentType string) (*Document, error) {

	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, reader); err != nil {
		logger.Error("Failed to read upload content", logger.WithError(err))
		return nil, err
	}

	fileBytes := buf.Bytes()

	ext := filepath.Ext(filename)
	objectName := fmt.Sprintf("%d_%s", time.Now().Unix(), filename)

	if err := s.storage.UploadFile(ctx, objectName, bytes.NewReader(fileBytes), int64(len(fileBytes)), contentType); err != nil {
		return nil, fmt.Errorf("failed to upload: %w", err)
	}

	var extractedText string
	var err error
	switch ext {
	case ".pdf":
		extractedText, err = extractor.ExtractTextFromPDF(bytes.NewReader(fileBytes), int64(len(fileBytes)))
		if err != nil {
			logger.Warn("Failed to extract text from PDF", logger.Merge(logger.Fields{"filename": filename}, logger.WithError(err)))
		}
	case ".txt":
		extractedText = string(fileBytes)
	}

	doc := &Document{
		Filename:      filename,
		ContentType:   contentType,
		StoragePath:   objectName,
		ExtractedText: extractedText,
		Status:        "uploaded",
	}

	if err := s.repo.Create(doc); err != nil {
		logger.Error("Failed to create document record", logger.WithError(err))
		return nil, err
	}

	logger.Info("Document uploaded successfully", logger.Fields{"id": doc.ID, "filename": filename})

	return doc, nil
}

func (s *Service) AnalyzeDocument(ctx context.Context, id uuid.UUID) (*Document, error) {
	doc, err := s.repo.FindByID(id)
	if err != nil {
		logger.Error("Document not found for analysis", logger.Merge(logger.Fields{"id": id}, logger.WithError(err)))
		return nil, err
	}

	if doc.ExtractedText == "" {
		logger.Warn("No text extracted for document", logger.Fields{"id": id})
		return nil, fmt.Errorf("no extracted text available for analysis")
	}

	result, err := s.analyzer.AnalyzeText(ctx, doc.ExtractedText)
	if err != nil {
		logger.Error("LLM analysis failed", logger.Merge(logger.Fields{"id": id}, logger.WithError(err)))
		return nil, err
	}

	metaBytes, _ := json.Marshal(result.Metadata)

	doc.Summary = result.Summary
	doc.DocType = result.Type
	doc.Metadata = metaBytes
	doc.Status = "analyzed"

	if err := s.repo.Update(doc); err != nil {
		return nil, err
	}

	return doc, nil
}

func (s *Service) GetDocument(ctx context.Context, id uuid.UUID) (*Document, error) {

	return s.repo.FindByID(id)
}
