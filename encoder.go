package disorder

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"reflect"
	"time"
)

type Encoder struct {
	writer io.Writer
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		writer: w,
	}
}

func (e *Encoder) Encode(value interface{}) error {
	v := reflect.ValueOf(value)
	if isNull(v) {
		return fmt.Errorf("null value cannot be marshal")
	}
	return e.write("", v)
}

func (e *Encoder) write(name string, value reflect.Value) error {
	if isNull(value) {
		return nil
	}

	t := tagUndefined
	var bytes []byte
	var err error

	switch i := value.Interface().(type) {
	case *time.Time:
		t = tagTimestamp
		bytes = make([]byte, 8)
		binary.BigEndian.PutUint64(bytes, uint64(i.Unix()))

	case Enum:
		t = tagEnum
		var enum string
		enum, err = i.GetValue()
		if err != nil {
			return err
		}
		bytes = []byte{byte(len(enum))}
		bytes = append(bytes, []byte(enum)...)

	default:
		switch value.Kind() {
		case reflect.Ptr, reflect.Interface:
			return e.write(name, value.Elem())

		case reflect.Bool:
			t = tagBool
			bytes = make([]byte, 1)
			if value.Bool() {
				bytes[0] = 1
			} else {
				bytes[0] = 0
			}

		case reflect.Int32:
			t = tagInt
			bytes = make([]byte, 4)
			binary.BigEndian.PutUint32(bytes, uint32(value.Int()))

		case reflect.Int64:
			t = tagLong
			bytes = make([]byte, 8)
			binary.BigEndian.PutUint64(bytes, uint64(value.Int()))

		case reflect.Float32:
			t = tagFloat
			bytes = make([]byte, 4)
			binary.BigEndian.PutUint32(bytes, math.Float32bits(float32(value.Float())))

		case reflect.Float64:
			t = tagDouble
			bytes = make([]byte, 8)
			binary.BigEndian.PutUint64(bytes, math.Float64bits(value.Float()))

		case reflect.String:
			t = tagString
			bytes = make([]byte, 4)
			str := value.String()
			binary.BigEndian.PutUint32(bytes, uint32(len(str)))
			bytes = append(bytes, []byte(str)...)

		case reflect.Slice:
			if value.Type().Elem().Kind() == reflect.Uint8 {
				t = tagBytes
				bytes = make([]byte, 4)
				binary.BigEndian.PutUint32(bytes, uint32(value.Len()))
				array := value.Interface().([]byte)
				bytes = append(bytes, array...)
			} else {
				return e.writeArray(name, value)
			}

		case reflect.Map:
			return e.writeMap(name, value)

		case reflect.Struct:
			return e.writeObject(name, value)
		}
	}

	if err != nil {
		return err
	}
	if t == tagUndefined {
		return fmt.Errorf("unsupported type: %s", value.Type().String())
	}
	err = e.writeTag(t)
	if err != nil {
		return err
	}
	err = e.writeName(name)
	if err != nil {
		return err
	}
	_, err = e.writer.Write(bytes)
	return err
}

func (e *Encoder) writeArray(name string, value reflect.Value) error {
	err := e.writeTag(tagArrayStart)
	if err != nil {
		return err
	}
	err = e.writeName(name)
	if err != nil {
		return err
	}

	count := value.Len()
	for i := 0; i < count; i++ {
		err = e.write("", value.Index(i))
		if err != nil {
			return err
		}
	}

	err = e.writeTag(tagArrayEnd)
	return err
}

func (e *Encoder) writeMap(name string, value reflect.Value) error {
	err := e.writeTag(tagObjectStart)
	if err != nil {
		return err
	}
	err = e.writeName(name)
	if err != nil {
		return err
	}

	keys := value.MapKeys()
	for i, key := range keys {
		if i == 0 && key.Kind() != reflect.String {
			return fmt.Errorf("map key type must be string")
		}
		if isNull(value.MapIndex(key)) {
			continue
		}
		err = e.write(key.String(), value.MapIndex(key))
		if err != nil {
			return err
		}
	}

	return e.writeTag(tagObjectEnd)
}

func (e *Encoder) writeObject(name string, value reflect.Value) error {
	err := e.writeTag(tagObjectStart)
	if err != nil {
		return err
	}
	err = e.writeName(name)
	if err != nil {
		return err
	}

	info, err := getStructInfo(value.Type())
	if err != nil {
		return err
	}
	for _, field := range info.fieldsList {
		fieldValue := value.Field(field.index)
		if isNull(fieldValue) {
			continue
		}
		err = e.write(field.key, fieldValue)
		if err != nil {
			return err
		}
	}

	return e.writeTag(tagObjectEnd)
}

func (e *Encoder) writeName(value string) error {
	if len(value) == 0 {
		return nil
	}
	if len(value) > 255 {
		return fmt.Errorf("string length overflow. should less than 255")
	}
	_, err := e.writer.Write([]byte{byte(len(value))})
	if err != nil {
		return err
	}
	_, err = e.writer.Write([]byte(value))
	return err
}

func (e *Encoder) writeTag(t tag) error {
	_, err := e.writer.Write([]byte{byte(t)})
	return err
}
