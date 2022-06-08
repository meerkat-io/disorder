package disorder

import (
	"bytes"
	"io"
)

type Marshaler interface {
	MarshalDO(w io.Writer) error
}

type Unmarshaler interface {
	UnmarshalDO(r io.Reader) error
}

func Marshal(value interface{}) ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := NewEncoder(buffer)
	err := encoder.Encode(value)
	if err == nil {
		return buffer.Bytes(), nil
	}
	return nil, err
}

func Unmarshal(data []byte, value interface{}) error {
	return nil
}
