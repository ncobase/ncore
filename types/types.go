package types

import "encoding/json"

// JSON represents a generic JSON object, useful for handling dynamic key-value pairs.
type JSON = map[string]any

// JSONArray represents an array of JSON objects.
type JSONArray = []JSON

// StringArray represents a slice of strings, typically used for lists of string data.
type StringArray = []string

// AnyArray represents a slice of any type, useful for generic lists.
type AnyArray = []any

// Struct represents an empty struct, often used for signals or stateless markers.
type Struct struct{}

// JSONData provides utility methods for working with JSON objects.
type JSONData struct {
	Data JSON `json:"data,omitempty"`
}

// ToBytes serializes JSONData to a byte slice.
func (j *JSONData) ToBytes() ([]byte, error) {
	return json.Marshal(j.Data)
}

// FromBytes deserializes a byte slice into JSONData.
func (j *JSONData) FromBytes(data []byte) error {
	return json.Unmarshal(data, &j.Data)
}
