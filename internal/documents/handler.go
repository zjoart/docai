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

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) UploadDocument(w http.ResponseWriter, r *http.Request) {

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		writeJSON(w, http.StatusBadRequest, "Failed to parse form")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, "File is required")
		return
	}

	defer file.Close()

	if header.Size > 5*1024*1024 {
		writeJSON(w, http.StatusBadRequest, "File too large (max 5MB)")
		return
	}

	doc, err := h.service.UploadDocument(r.Context(), header.Filename, file, header.Size, header.Header.Get("Content-Type"))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, doc)
}

func (h *Handler) AnalyzeDocument(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	id, err := id.IsValidUUID(vars["id"])
	if err != nil {
		logger.Error("Invalid file ID format", logger.Fields{"id": id})
		writeJSON(w, http.StatusBadRequest, "Invalid file ID format")
		return
	}

	doc, err := h.service.AnalyzeDocument(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, doc)
}

func (h *Handler) GetDocument(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	id, err := id.IsValidUUID(vars["id"])
	if err != nil {
		logger.Error("Invalid file ID format", logger.Fields{"id": id})
		writeJSON(w, http.StatusBadRequest, "Invalid file ID format")
		return
	}

	doc, err := h.service.GetDocument(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, doc)
}
