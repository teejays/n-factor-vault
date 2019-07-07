package orm

import (
	"time"

	"github.com/google/uuid"
)

// ID is the type we should use for primary keys for entities. We make an alias for it here
// so outside code doesn't have to worry about it, and we can eventually choose to change it
// if needed
type ID string

func (id ID) IsEmpty() bool {
	return id == ""
}

// BaseModel is the parent model that any struct intending to use ORM should embed.
// This ensures that we have common meta fields across all entities and makes our code DRY.
type BaseModel struct {
	ID         ID         `xorm:"pk UUID notnull" json:"id"`
	CreatedAt  time.Time  `xorm:"created notnull" json:"created_at"`
	UpdatedAt  time.Time  `xorm:"updated notnull" json:"updated_at"`
	RowVersion int        `xorm:"version notnull" json:"row_version"`
	DeletedAt  *time.Time `xorm:"deleted null" json:"deleted_at"`
}

// BeforeInsert is called by xorm before each insert. This function can be used to edit
// or overwrite fields.
func (m *BaseModel) BeforeInsert() {
	if m.ID == "" {
		m.ID = ID(uuid.New().String())
	}
}
