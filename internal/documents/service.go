package documents

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"
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

	existingDoc, err := s.repo.FindByFilename(filename)
	if err == nil {
		logger.Info("Document already exists, returning existing record", logger.Fields{"filename": filename, "id": existingDoc.ID})
		return existingDoc, nil
	}

	if !s.repo.IsNotFoundError(err) {
		return nil, fmt.Errorf("failed to check for duplicates: %w", err)
	}

	fileBytes := buf.Bytes()

	ext := filepath.Ext(filename)
	objectName := fmt.Sprintf("%d_%s", time.Now().Unix(), filename)

	var extractedText string
	switch ext {
	case ".pdf":
		extractedText, err = extractor.ExtractTextFromPDF(bytes.NewReader(fileBytes), int64(len(fileBytes)))
		if err != nil {
			logger.Warn("Failed to extract text from PDF", logger.Merge(logger.Fields{"filename": filename}, logger.WithError(err)))
			return nil, fmt.Errorf("failed to extract text from PDF/Image")
		}

	case ".docx":
		extractedText, err = extractor.ExtractTextFromDOCX(bytes.NewReader(fileBytes), int64(len(fileBytes)))
		if err != nil {
			logger.Warn("Failed to extract text from DOCX", logger.Merge(logger.Fields{"filename": filename}, logger.WithError(err)))
			return nil, fmt.Errorf("failed to extract text from DOCX: %w", err)
		}

	case ".txt":
		extractedText = string(fileBytes)
	}

	if strings.TrimSpace(extractedText) == "" {
		return nil, fmt.Errorf("upload rejected: no text could be extracted from document")
	}

	fileUrl, err := s.storage.UploadFile(ctx, objectName, bytes.NewReader(fileBytes), int64(len(fileBytes)), contentType)
	if err != nil {
		return nil, fmt.Errorf("failed to upload: %w", err)
	}

	doc := &Document{
		Filename:      filename,
		ContentType:   contentType,
		StoragePath:   objectName,
		FileUrl:       fileUrl,
		ExtractedText: extractedText,
		Status:        "uploaded",
	}

	if err := s.repo.Create(doc); err != nil {
		logger.Error("Failed to create document record", logger.WithError(err))

		//  delete file from storage
		if delErr := s.storage.DeleteFile(ctx, objectName); delErr != nil {
			logger.Error("Failed to delete orphaned file", logger.Merge(logger.Fields{"object": objectName}, logger.WithError(delErr)))
		}
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

	if strings.TrimSpace(doc.ExtractedText) == "" {
		logger.Warn("Skipping analysis: No text extracted", logger.Fields{"id": id})

		return doc, fmt.Errorf("analysis skipped: no text extracted from document (likely scanned PDF or image)")
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
