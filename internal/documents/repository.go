package documents

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	Create(doc *Document) error
	FindByID(id uuid.UUID) (*Document, error)
	Update(doc *Document) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(doc *Document) error {
	return r.db.Create(doc).Error
}

func (r *repository) FindByID(id uuid.UUID) (*Document, error) {
	var doc Document
	err := r.db.First(&doc, "id = ?", id).Error
	return &doc, err
}

func (r *repository) Update(doc *Document) error {
	return r.db.Save(doc).Error
}
