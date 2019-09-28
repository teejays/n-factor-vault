package json

import (
	"encoding/json"

	"github.com/Rican7/conjson"
	"github.com/Rican7/conjson/transform"
)

// Marshal encodes the struct into JSON
func Marshal(v interface{}) ([]byte, error) {

	marshaler := conjson.NewMarshaler(v, transform.ConventionalKeys())
	encoded, err := json.Marshal(marshaler)
	if err != nil {
		return nil, err
	}
	return encoded, nil

}

// Unmarshal deencodes JSON bytes into the provided struct
func Unmarshal(src []byte, v interface{}) error {

	unmarshaler := conjson.NewUnmarshaler(v, transform.ConventionalKeys())

	err := json.Unmarshal(src, unmarshaler)
	if err != nil {
		return err
	}
	return nil

}
