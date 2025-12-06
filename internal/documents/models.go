package documents

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Document struct {
	ID            uuid.UUID       `gorm:"type:uuid;primary_key;" json:"id"`
	Filename      string          `json:"filename"`
	ContentType   string          `json:"content_type"`
	StoragePath   string          `json:"-"`
	ExtractedText string          `json:"extracted_text"`
	Summary       string          `json:"summary"`
	DocType       string          `json:"doc_type"`
	Metadata      json.RawMessage `gorm:"type:jsonb" json:"metadata"`
	Status        string          `json:"status"` // uploaded, processed, analyzed
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

func (d *Document) BeforeCreate(tx *gorm.DB) (err error) {
	d.ID = uuid.New()
	return
}
