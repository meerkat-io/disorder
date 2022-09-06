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
	return e.writeValue("", v)
}

func (e *Encoder) writeValue(name string, value reflect.Value) error {
	if isNull(value) {
		return nil
	}

	t := tagUndefined
	var bytes []byte
	var err error

	switch i := value.Interface().(type) {
	case time.Time:
		t = tagTimestamp
		bytes, err = e.timestampBits(&i)

	case *time.Time:
		t = tagTimestamp
		bytes, err = e.timestampBits(i)

	case Enum:
		t = tagEnum
		bytes, err = i.Encode()

	default:
		switch value.Kind() {
		case reflect.Ptr, reflect.Interface:
			return e.writeValue(name, value.Elem())

		case reflect.Bool:
			t = tagBool
			bytes = make([]byte, 1)
			if value.Bool() {
				bytes[0] = 1
			} else {
				bytes[0] = 0
			}

		case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint:
			t = tagInt
			bytes = make([]byte, 4)
			binary.BigEndian.PutUint32(bytes, uint32(value.Uint()))

		case reflect.Uint64:
			t = tagLong
			bytes = make([]byte, 8)
			binary.BigEndian.PutUint64(bytes, value.Uint())

		case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int:
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

		case reflect.Slice, reflect.Array:
			//TO-DO check bytes
			return e.writeArrayValue(name, value)

		case reflect.Map:
			return e.writeMapValue(name, value)

		case reflect.Struct:
			return e.writeObjectValue(name, value)
		}
	}

	if t == tagUndefined {
		return fmt.Errorf("unsupported type: %s", value.Type().String())
	}
	_, err = e.writer.Write([]byte{byte(t)})
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

func (e *Encoder) writeArrayValue(name string, value reflect.Value) error {
	_, err := e.writer.Write([]byte{byte(tagArrayStart)})
	if err != nil {
		return err
	}
	err = e.writeName(name)
	if err != nil {
		return err
	}

	count := value.Len()
	for i := 0; i < count; i++ {
		err = e.writeValue("", value.Index(i))
		if err != nil {
			return err
		}
	}

	_, err = e.writer.Write([]byte{byte(tagArrayEnd)})
	return err
}

func (e *Encoder) writeMapValue(name string, value reflect.Value) error {
	_, err := e.writer.Write([]byte{byte(tagObjectStart)})
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
		err = e.writeValue(key.String(), value.MapIndex(key))
		if err != nil {
			return err
		}
	}

	_, err = e.writer.Write([]byte{byte(tagObjectEnd)})
	return err
}

func (e *Encoder) writeObjectValue(name string, value reflect.Value) error {
	_, err := e.writer.Write([]byte{byte(tagObjectStart)})
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
		err = e.writeValue(field.key, fieldValue)
		if err != nil {
			return err
		}
	}

	_, err = e.writer.Write([]byte{byte(tagObjectEnd)})
	return err
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

func (e *Encoder) timestampBits(t *time.Time) ([]byte, error) {
	bytes := make([]byte, 8)
	binary.BigEndian.PutUint64(bytes, uint64(t.Unix()))
	_, err := e.writer.Write(bytes)
	return bytes, err
}
