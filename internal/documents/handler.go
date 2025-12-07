package documents

import (
	"context"
	"encoding/json"
	"net/http"
	"path/filepath"

	"github.com/gorilla/mux"
	"github.com/zjoart/docai/pkg/id"
	"github.com/zjoart/docai/pkg/logger"
)

type Handler struct {
	service *Service
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeErrorJSON(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"message": message})
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) UploadDocument(w http.ResponseWriter, r *http.Request) {

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "Failed to parse form")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "File is required")
		return
	}

	defer file.Close()

	ext := filepath.Ext(header.Filename)
	if ext != ".pdf" && ext != ".txt" && ext != ".docx" {
		writeErrorJSON(w, http.StatusBadRequest, "File type not supported")
		return
	}

	if header.Size > 5*1024*1024 {
		writeErrorJSON(w, http.StatusBadRequest, "File too large (max 5MB)")
		return
	}

	doc, err := h.service.UploadDocument(r.Context(), header.Filename, file, header.Size, header.Header.Get("Content-Type"))
	if err != nil {
		writeErrorJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	processImmediately := r.FormValue("processImmediately") == "true"
	message := "Document uploaded successfully"

	if processImmediately {

		if err := h.service.UpdateStatus(r.Context(), doc.ID, "processing"); err != nil {
			logger.Error("Failed to update status to processing", logger.WithError(err))

		} else {
			doc.Status = "processing"
			message = "Document uploaded and analysis started"

			go func() {

				if _, err := h.service.AnalyzeDocument(context.Background(), doc.ID); err != nil {
					logger.Error("Background analysis failed", logger.WithError(err))
				}
			}()
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message":  message,
		"document": doc,
	})
}

func (h *Handler) AnalyzeDocument(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	id, err := id.IsValidUUID(vars["id"])
	if err != nil {
		logger.Error("Invalid file ID format", logger.Fields{"id": id})
		writeErrorJSON(w, http.StatusBadRequest, "Invalid file ID format")
		return
	}

	currentDoc, err := h.service.GetDocument(r.Context(), id)
	if err != nil {
		writeErrorJSON(w, http.StatusNotFound, "Document not found")
		return
	}

	if currentDoc.Status == "processing" {
		writeErrorJSON(w, http.StatusConflict, "Document is already being processed")
		return
	}

	doc, err := h.service.AnalyzeDocument(r.Context(), id)
	if err != nil {
		writeErrorJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, doc)
}

func (h *Handler) GetDocument(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	id, err := id.IsValidUUID(vars["id"])
	if err != nil {
		logger.Error("Invalid file ID format", logger.Fields{"id": id})
		writeErrorJSON(w, http.StatusBadRequest, "Invalid file ID format")
		return
	}

	doc, err := h.service.GetDocument(r.Context(), id)
	if err != nil {
		writeErrorJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, doc)
}
