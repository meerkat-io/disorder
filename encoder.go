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
		return fmt.Errorf("invalid value or type")
	}
	return e.write(reflect.ValueOf(value))
}

func (e *Encoder) write(value reflect.Value) error {
	switch i := value.Interface().(type) {
	case time.Time:
		return e.writeTime(&i)

	case *time.Time:
		return e.writeTime(i)

	case Enum:
		_, err := e.writer.Write([]byte{byte(tagEnum)})
		if err != nil {
			return err
		}
		name, err := i.Encode()
		if err != nil {
			return err
		}
		return e.writeName(name)

	case Marshaler:
		return i.MarshalDO(e.writer)
	}

	var bytes []byte
	switch value.Kind() {
	case reflect.Ptr, reflect.Interface:
		return e.write(value.Elem())

	case reflect.Bool:
		bytes = make([]byte, 2)
		bytes[0] = byte(tagBool)
		if value.Bool() {
			bytes[1] = 1
		} else {
			bytes[1] = 0
		}

	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint:
		bytes = make([]byte, 5)
		bytes[0] = byte(tagInt)
		binary.BigEndian.PutUint32(bytes[1:], uint32(value.Uint()))

	case reflect.Uint64:
		bytes = make([]byte, 9)
		bytes[0] = byte(tagLong)
		binary.BigEndian.PutUint64(bytes[1:], value.Uint())

	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int:
		bytes = make([]byte, 5)
		bytes[0] = byte(tagInt)
		binary.BigEndian.PutUint32(bytes[1:], uint32(value.Int()))

	case reflect.Int64:
		bytes = make([]byte, 9)
		bytes[0] = byte(tagLong)
		binary.BigEndian.PutUint64(bytes[1:], uint64(value.Int()))

	case reflect.Float32:
		bytes = make([]byte, 5)
		bytes[0] = byte(tagFloat)
		binary.BigEndian.PutUint32(bytes[1:], math.Float32bits(float32(value.Float())))

	case reflect.Float64:
		bytes = make([]byte, 9)
		bytes[0] = byte(tagDouble)
		binary.BigEndian.PutUint64(bytes[1:], math.Float64bits(value.Float()))

	case reflect.String:
		bytes = make([]byte, 5)
		bytes[0] = byte(tagString)
		str := value.String()
		binary.BigEndian.PutUint32(bytes[1:], uint32(len(str)))
		bytes = append(bytes, []byte(str)...)

	case reflect.Slice, reflect.Array:
		//TO-DO check bytes
		return e.writeArray(value)

	case reflect.Map:
		return e.writeMap(value)

	case reflect.Struct:
		return e.writeObject(value)

	default:
		return fmt.Errorf("invalid type: %s", value.Type().String())
	}

	_, err := e.writer.Write(bytes)
	return err
}

func (e *Encoder) writeArray(value reflect.Value) error {
	count := value.Len()
	_, err := e.writer.Write([]byte{byte(tagStartArray)})
	if err != nil {
		return err
	}

	for i := 0; i < count; i++ {
		err = e.write(value.Index(i))
		if err != nil {
			return err
		}
	}

	_, err = e.writer.Write([]byte{byte(tagEndArray)})
	return err
}

func (e *Encoder) writeMap(value reflect.Value) error {
	_, err := e.writer.Write([]byte{byte(tagStartObject)})
	if err != nil {
		return err
	}

	keys := value.MapKeys()
	for _, key := range keys {
		if key.Kind() != reflect.String {
			return fmt.Errorf("map key type must be string")
		}
		err := e.writeName(key.String())
		if err != nil {
			return err
		}
		err = e.write(value.MapIndex(key))
		if err != nil {
			return err
		}
	}

	_, err = e.writer.Write([]byte{byte(tagEndObject)})
	return err
}

func (e *Encoder) writeObject(value reflect.Value) error {
	info, err := getStructInfo(value.Type())
	if err != nil {
		return err
	}

	_, err = e.writer.Write([]byte{byte(tagStartObject)})
	if err != nil {
		return err
	}

	for _, field := range info.fieldsList {
		fieldValue := value.Field(field.index)
		if !isNull(fieldValue) {
			err = e.writeName(field.key)
			if err != nil {
				return err
			}
			err = e.write(fieldValue)
			if err != nil {
				return err
			}
		}
	}

	_, err = e.writer.Write([]byte{byte(tagEndObject)})
	return err
}

func (e *Encoder) writeTime(t *time.Time) error {
	bytes := make([]byte, 9)
	bytes[0] = byte(tagTimestamp)
	binary.BigEndian.PutUint64(bytes[1:], uint64(t.Unix()))
	_, err := e.writer.Write(bytes)
	return err
}

func (e *Encoder) writeName(value string) error {
	if len(value) == 0 {
		return fmt.Errorf("empty name")
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
