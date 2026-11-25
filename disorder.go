package disorder

import (
	"bytes"
)

// Marshal returns the binary encoding of value.
func Marshal(value interface{}) ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := NewEncoder(buffer)
	err := encoder.Encode(value)
	if err == nil {
		return buffer.Bytes(), nil
	}
	return nil, err
}

// Unmarshal decodes binary data into value, which must be a non-nil pointer.
func Unmarshal(data []byte, value interface{}) error {
	buffer := bytes.NewBuffer(data)
	decoder := NewDecoder(buffer)
	return decoder.Decode(value)
}
