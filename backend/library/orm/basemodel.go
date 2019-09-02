package orm

import (
	"time"

	"github.com/jinzhu/gorm"
	"github.com/teejays/n-factor-vault/backend/library/id"
)

// BaseModel is the parent model that any struct intending to use ORM should embed.
// This ensures that we have common meta fields across all entities and makes our code DRY.
//
// The reason we're not using the conventional gorm.Model is because that uses int as primary keys
// while we want to use uuids.
type BaseModel struct {
	ID         id.ID      `gorm:"type:UUID;" json:"id"` // gorm by defaults treats field with name ID as primary key (unless specified)
	CreatedAt  time.Time  `gorm:"CREATED NOTNULL" json:"created_at"`
	UpdatedAt  time.Time  `gorm:"UPDATED NOTNULL" json:"updated_at"`
	DeletedAt  *time.Time `gorm:"DELETED NULL" json:"deleted_at"`
	RowVersion int        `gorm:"AUTO_INCREMENT"`
}

// BeforeCreate is run whenever a new instance of a model is created.
func (m *BaseModel) BeforeCreate(scope *gorm.Scope) error {
	if m != nil && m.ID == "" {
		scope.SetColumn("ID", id.GetNewID())
	}
	return nil
}
