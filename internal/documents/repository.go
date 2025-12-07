package documents

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	Create(doc *Document) error
	FindByID(id uuid.UUID) (*Document, error)
	FindByFilename(filename string) (*Document, error)
	IsNotFoundError(err error) bool
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

func (r *repository) FindByFilename(filename string) (*Document, error) {
	var doc Document
	err := r.db.First(&doc, "filename = ?", filename).Error
	return &doc, err
}

func (r *repository) Update(doc *Document) error {
	return r.db.Save(doc).Error
}

func (r *repository) IsNotFoundError(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
