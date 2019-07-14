package orm

import "github.com/google/uuid"

// ID is the type we should use for primary keys for entities. We make an alias for it here
// so outside code doesn't have to worry about it, and we can eventually choose to change it
// if needed
type ID string

// IsEmpty return true if the id has an empty or default value i.e. id is an empty string
func (id ID) IsEmpty() bool {
	return id == ""
}

// GetNewID creates a new unique ID
func GetNewID() ID {
	return ID(uuid.New().String())
}

// StrToID takes a string and converts it to type ID.
// It returns an error is the string cannot be converted to type ID.
func StrToID(str string) (ID, error) {
	return ID(str), nil
}
