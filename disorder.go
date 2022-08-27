package disorder

import (
	"bytes"
	"io"
)

type Enum interface {
	Enum()
	Decode(enum string) error
	Encode() (string, error)
}

type EnumBase string

func (*EnumBase) Enum() {}

func (enum *EnumBase) Decode(value string) error {
	*enum = EnumBase(value)
	return nil
}

func (enum *EnumBase) Encode() (string, error) {
	return string(*enum), nil
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
	tagBool      tag = 1
	tagInt       tag = 2
	tagLong      tag = 3
	tagFloat     tag = 4
	tagDouble    tag = 5
	tagString    tag = 6
	tagBytes     tag = 7
	tagTimestamp tag = 8

	tagEnum        tag = 10
	tagStartArray  tag = 11
	tagEndArray    tag = 12
	tagStartObject tag = 13
	tagEndObject   tag = 0
)
