package disorder

import (
	"bytes"
	"fmt"
)

type Enum interface {
	Enum()
	Decode(enum string) error
	Encode() (string, error)
}

type EnumBase string

func (*EnumBase) Enum() {}

func (enum *EnumBase) Decode(value string) error {
	if value == "" {
		return fmt.Errorf("invalid enum value")
	}
	*enum = EnumBase(value)
	return nil
}

func (enum *EnumBase) Encode() (string, error) {
	if string(*enum) == "" {
		return "", fmt.Errorf("invalid enum value")
	}
	return string(*enum), nil
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
	tagUndefined tag = 0

	tagBool      tag = 1
	tagInt       tag = 2
	tagLong      tag = 3
	tagFloat     tag = 4
	tagDouble    tag = 5
	tagString    tag = 6
	tagBytes     tag = 7
	tagTimestamp tag = 8

	tagEnum   tag = 10
	tagArray  tag = 11
	tagMap    tag = 12
	tagObject tag = 13
)
