package id

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
// It returns an error if the string cannot be converted to type ID.
func StrToID(str string) (ID, error) {
	return ID(str), nil
}

// StrToIDMust takes a string and converts it to type ID.
// It panics if the string cannot be converted to type ID.
func StrToIDMust(str string) ID {
	return ID(str)
}

// IDToStr takes an ID and converts it to type string.
// It returns an error if the ID cannot be converted to type string.
func IDToStr(id ID) (string, error) {
	return string(id), nil
}

// IDToStrMust takes an ID and converts it to type string.
// It panics if the ID cannot be converted to type string.
func IDToStrMust(id ID) string {
	return string(id)
}
