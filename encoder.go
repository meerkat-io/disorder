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
	return e.write(reflect.ValueOf(value))
}

func (e *Encoder) writeTag(value reflect.Value) error {
	t := tagUndefined

	switch value.Interface().(type) {
	case time.Time:
		t = tagTimestamp

	case *time.Time:
		t = tagTimestamp

	case Enum:
		t = tagEnum
	}

	switch value.Kind() {
	case reflect.Ptr, reflect.Interface:
		return e.writeTag(value.Elem())

	case reflect.Bool:
		t = tagBool

	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint,
		reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int:
		t = tagInt

	case reflect.Uint64, reflect.Int64:
		t = tagLong

	case reflect.Float32:
		t = tagFloat

	case reflect.Float64:
		t = tagDouble

	case reflect.String:
		t = tagString

	case reflect.Slice, reflect.Array:
		//TO-DO check bytes
		t = tagArray

	case reflect.Map:
		t = tagMap

	case reflect.Struct:
		t = tagObject
	}

	if t == tagUndefined {
		return fmt.Errorf("unsupported type: %s", value.Type().String())
	}

	_, err := e.writer.Write([]byte{byte(t)})
	return err
}

func (e *Encoder) writeValue(value reflect.Value) error {
	switch i := value.Interface().(type) {
	case time.Time:
		return e.writeTimeValue(&i)

	case *time.Time:
		return e.writeTimeValue(i)

	case Enum:
		name, err := i.Encode()
		if err != nil {
			return err
		}
		return e.writeName(name)
	}

	var bytes []byte
	switch value.Kind() {
	case reflect.Ptr, reflect.Interface:
		return e.writeValue(value.Elem())

	case reflect.Bool:
		bytes = make([]byte, 1)
		if value.Bool() {
			bytes[0] = 1
		} else {
			bytes[0] = 0
		}

	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint:
		bytes = make([]byte, 4)
		binary.BigEndian.PutUint32(bytes, uint32(value.Uint()))

	case reflect.Uint64:
		bytes = make([]byte, 8)
		bytes[0] = byte(tagLong)
		binary.BigEndian.PutUint64(bytes, value.Uint())

	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int:
		bytes = make([]byte, 4)
		binary.BigEndian.PutUint32(bytes, uint32(value.Int()))

	case reflect.Int64:
		bytes = make([]byte, 8)
		binary.BigEndian.PutUint64(bytes, uint64(value.Int()))

	case reflect.Float32:
		bytes = make([]byte, 4)
		binary.BigEndian.PutUint32(bytes, math.Float32bits(float32(value.Float())))

	case reflect.Float64:
		bytes = make([]byte, 8)
		binary.BigEndian.PutUint64(bytes, math.Float64bits(value.Float()))

	case reflect.String:
		bytes = make([]byte, 4)
		str := value.String()
		binary.BigEndian.PutUint32(bytes, uint32(len(str)))
		bytes = append(bytes, []byte(str)...)

	case reflect.Slice, reflect.Array:
		//TO-DO check bytes
		return e.writeArrayValue(value)

	case reflect.Map:
		return e.writeMapValue(value)

	case reflect.Struct:
		return e.writeObjectValue(value)

	default:
		return fmt.Errorf("unsupported type: %s", value.Type().String())
	}

	_, err := e.writer.Write(bytes)
	return err
}

func (e *Encoder) writeArrayValue(value reflect.Value) error {
	count := value.Len()
	if count == 0 {
		return nil
	}

	err := e.writeTag(value.Index(0))
	if err != nil {
		return err
	}
	bytes := make([]byte, 4)
	binary.BigEndian.PutUint32(bytes, uint32(count))
	_, err = e.writer.Write(bytes)
	if err != nil {
		return err
	}

	for i := 0; i < count; i++ {
		err = e.writeValue(value.Index(i))
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *Encoder) writeMapValue(value reflect.Value) error {
	count := value.Len()
	if count == 0 {
		return nil
	}
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

func (e *Encoder) writeObjectValue(value reflect.Value) error {
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
		if isNull(fieldValue) {
			continue
		}
		err = e.writeName(field.key)
		if err != nil {
			return err
		}
		err = e.write(fieldValue)
		if err != nil {
			return err
		}
	}

	_, err = e.writer.Write([]byte{byte(tagEndObject)})
	return err
}

func (e *Encoder) writeTimeValue(t *time.Time) error {
	bytes := make([]byte, 8)
	binary.BigEndian.PutUint64(bytes, uint64(t.Unix()))
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
