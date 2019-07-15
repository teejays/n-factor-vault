package orm

import (
	"time"
)

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
	// If we are saving an entity, and it doesn't have an ID,
	// then get a new ID and add it
	if m.ID.IsEmpty() {
		m.ID = GetNewID()
	}
}
