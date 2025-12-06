package documents

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/zjoart/docai/pkg/id"
	"github.com/zjoart/docai/pkg/logger"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) UploadDocument(w http.ResponseWriter, r *http.Request) {

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "File is required", http.StatusBadRequest)
		return
	}

	defer file.Close()

	if header.Size > 5*1024*1024 {
		http.Error(w, "File too large (max 5MB)", http.StatusBadRequest)
		return
	}

	doc, err := h.service.UploadDocument(r.Context(), header.Filename, file, header.Size, header.Header.Get("Content-Type"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(doc)
}

func (h *Handler) AnalyzeDocument(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	id, err := id.IsValidUUID(vars["id"])
	if err != nil {
		logger.Error("Invalid file ID format", logger.Fields{"id": id})
		http.Error(w, "Invalid file ID format", http.StatusBadRequest)
		return
	}

	doc, err := h.service.AnalyzeDocument(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(doc)
}

func (h *Handler) GetDocument(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	id, err := id.IsValidUUID(vars["id"])
	if err != nil {
		logger.Error("Invalid file ID format", logger.Fields{"id": id})
		http.Error(w, "Invalid file ID format", http.StatusBadRequest)
		return
	}

	doc, err := h.service.GetDocument(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(doc)
}
