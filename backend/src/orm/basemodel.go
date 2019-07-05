package orm

import (
	"time"

	"github.com/google/uuid"
)

type BaseModel struct {
	ID         string    `xorm:"'id' pk UUID notnull" json:"id"`
	CreatedAt  time.Time `xorm:"created notnull" json:"created_at"`
	UpdatedAt  time.Time `xorm:"updated notnull" json:"updated_at"`
	IsDeleted  bool      `xorm:"default false notnull" json:"is_deleted"`
	RowVersion int       `xorm:"version notnull" json:"row_version"`
}

func (m *BaseModel) BeforeInsert() {
	m.ID = uuid.New().String()
}
