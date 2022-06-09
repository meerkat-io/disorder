package disorder

import (
	"bytes"
	"io"
)

type Enum interface {
	Enum()
	FromString(enum string) error
	ToString() (string, error)
}

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
	buffer := bytes.NewBuffer(data)
	decoder := NewDecoder(buffer)
	return decoder.Decode(value)
}

type tag byte

const (
	tagBool   tag = 1
	tagI8     tag = 2
	tagU8     tag = 3
	tagI16    tag = 4
	tagU16    tag = 5
	tagI32    tag = 6
	tagU32    tag = 7
	tagI64    tag = 8
	tagU64    tag = 9
	tagF32    tag = 10
	tagF64    tag = 11
	tagString tag = 12

	tagEnum      tag = 13
	tagTimestamp tag = 14

	tagStartArray  tag = 13
	tagEndArray    tag = 14
	tagStartObject tag = 15
	tagEndObject   tag = 16
)
