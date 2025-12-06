package documents

import (
	"github.com/gorilla/mux"
)

func RegisterRoutes(r *mux.Router, h *Handler) {
	r.HandleFunc("/documents/upload", h.UploadDocument).Methods("POST")
	r.HandleFunc("/documents/{id}/analyze", h.AnalyzeDocument).Methods("POST")
	r.HandleFunc("/documents/{id}", h.GetDocument).Methods("GET")
}
