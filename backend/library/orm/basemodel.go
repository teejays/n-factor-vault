package orm

import (
	"time"

	"github.com/jinzhu/gorm"
	"github.com/teejays/clog"
	"github.com/teejays/n-factor-vault/backend/library/id"
)

// BaseModelORM is the parent model that any struct intending to use ORM should embed.
// This ensures that we have common meta fields across all entities and makes our code DRY.
//
// The reason we're not using the conventional gorm.Model is because that uses int as primary keys
// while we want to use uuids.
type BaseModelORM struct {
	ID        id.ID      `gorm:"PRIMARY_KEY;type:UUID;" json:"id"`
	CreatedAt time.Time  `gorm:"CREATED NOTNULL" json:"created_at"`
	UpdatedAt time.Time  `gorm:"UPDATED NOTNULL" json:"updated_at"`
	DeletedAt *time.Time `gorm:"DELETED NULL" json:"deleted_at"`
}

// BeforeCreate is run whenever a new instance of a model is created.
func (m *BaseModelORM) BeforeCreate(scope *gorm.Scope) error {
	clog.Warning("orm: BaseModelORM.BeforeCreate() running...")
	if m != nil && m.ID == "" {
		scope.SetColumn("ID", id.GetNewID())
	}
	return nil
}

// BaseModel is a wrapper around the BaseModelGORM, which implements the
type BaseModel struct {
	BaseModelORM `gorm:"embedded"`
	Status       EntityStatus `gorm:"NOTNULL" json:"status"`
}

func (m *BaseModel) GetBaseEntity() Entity {
	return m
}

// ValidationErrors function makes BaseModel implement the Entity interface
func (m *BaseModel) ValidationErrors() []error {
	var errs []error
	// if m.Status == "" {
	// 	errs = append(errs, fmt.Errorf("status cannot be empty"))
	// }
	return errs
}

// IsValid function makes BaseModel implement the Entity interface
func (m *BaseModel) IsValid() bool {
	return len(m.ValidationErrors()) == 0
}

// IsEmpty function makes BaseModel implement the Entity interface
func (m *BaseModel) IsEmpty() bool {
	return m.ID.IsEmpty()
}

// BeforeCreate function makes BaseModel implement the Entity interface
func (m *BaseModel) BeforeCreateX() error {
	m.Status = DefaultStatus
	if m.ID == "" {
		m.ID = id.GetNewID()
	}
	return nil
}

// AfterCreate function makes BaseModel implement the Entity interface
func (m *BaseModel) AfterCreateX() error {
	return nil
}

// BeforeSave function makes BaseModel implement the Entity interface
func (m *BaseModel) BeforeSave() error {
	return nil
}

// AfterSave function makes BaseModel implement the Entity interface
func (m *BaseModel) AfterSave() error {
	return nil
}

// BeforeDelete function makes BaseModel implement the Entity interface
func (m *BaseModel) BeforeDelete() error {
	return nil
}

// AfterDelete function makes BaseModel implement the Entity interface
func (m *BaseModel) AfterDelete() error {
	return nil
}

// EntityStatus represents mutually exclusive states of the entity
type EntityStatus string

const StatusCreated EntityStatus = "CREATED"

var DefaultStatus EntityStatus = StatusCreated

// Entity is an interface which should be implemented by all types that go into our database
type Entity interface {
	GetBaseEntity() Entity

	ValidationErrors() []error
	IsValid() bool
	IsEmpty() bool

	BeforeCreateX() error
	AfterCreateX() error

	BeforeSave() error
	AfterSave() error

	BeforeDelete() error
	AfterDelete() error
}
