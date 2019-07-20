package orm

import (
	"time"

	"github.com/jinzhu/gorm"
	"github.com/teejays/n-factor-vault/backend/library/id"
)

// BaseModel is the parent model that any struct intending to use ORM should embed.
// This ensures that we have common meta fields across all entities and makes our code DRY.
type BaseModel struct {
	ID        id.ID      `gorm:"primary_key;type:uuid;"`
	CreatedAt time.Time  `gorm:"created notnull" json:"created_at"`
	UpdatedAt time.Time  `gorm:"updated notnull" json:"updated_at"`
	DeletedAt *time.Time `gorm:"deleted null" json:"deleted_at"`
}

func (m *BaseModel) BeforeCreate(scope *gorm.Scope) error {
	if m.ID == "" {
		scope.SetColumn("ID", id.GetNewID())
	}
	return nil
}
