package disorder

import (
	"bytes"
	"fmt"
)

//TO-DO import not used error (sub yaml) // check define and rpc separately

type Enum interface {
	Enum()
	Decode(enum string) error
	Encode() ([]byte, error)
}

type EnumBase string

func (*EnumBase) Enum() {}

func (enum *EnumBase) Decode(value string) error {
	if value == "" {
		return fmt.Errorf("empty enum value")
	}
	*enum = EnumBase(value)
	return nil
}

func (enum *EnumBase) Encode() ([]byte, error) {
	name := string(*enum)
	if len(name) == 0 {
		return nil, fmt.Errorf("empty enum value")
	}
	if len(name) > 255 {
		return nil, fmt.Errorf("string length overflow. should less than 255")
	}
	data := []byte{byte(len(name))}
	data = append(data, []byte(name)...)
	return data, nil
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

	tagBool   tag = 1
	tagInt    tag = 2
	tagLong   tag = 3
	tagFloat  tag = 4
	tagDouble tag = 5
	tagString tag = 6

	tagBytes     tag = 7
	tagTimestamp tag = 8
	tagEnum      tag = 9

	tagArrayStart  tag = 10
	tagArrayEnd    tag = 11
	tagObjectStart tag = 12
	tagObjectEnd   tag = 13
)
