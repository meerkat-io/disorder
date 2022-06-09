package disorder

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"reflect"
	"time"

	"github.com/meerkat-lib/disorder/internal/schema"
)

type Encoder struct {
	writer io.Writer
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		writer: w,
	}
}

func (e *Encoder) Encode(v interface{}) error {
	return e.write(reflect.ValueOf(v))
}

func (e *Encoder) write(value reflect.Value) error {
	if !value.IsValid() || value.Kind() == reflect.Ptr && value.IsNil() {
		return fmt.Errorf("invalid value or type")
	}

	switch i := value.Interface().(type) {
	case time.Time:
		return e.writeTime(&i)

	case *time.Time:
		return e.writeTime(i)

	case Enum:
		_, err := e.writer.Write([]byte{byte(schema.TypeEnum)})
		if err != nil {
			return err
		}
		name, err := i.ToString()
		if err != nil {
			return err
		}
		return e.writeName(name)

	case Marshaler:
		return i.MarshalDO(e.writer)
	}

	var bytes []byte
	switch value.Kind() {
	case reflect.Ptr:
		return e.write(value.Elem())

	case reflect.Bool:
		bytes = make([]byte, 2)
		bytes[0] = byte(schema.TypeBool)
		if value.Bool() {
			bytes[1] = 1
		} else {
			bytes[1] = 0
		}

	case reflect.Uint8:
		bytes = make([]byte, 2)
		bytes[0] = byte(schema.TypeU8)
		bytes[1] = byte(uint8(value.Int()))

	case reflect.Uint16:
		bytes = make([]byte, 3)
		bytes[0] = byte(schema.TypeU16)
		binary.LittleEndian.PutUint16(bytes[1:], uint16(value.Int()))

	case reflect.Uint32:
		bytes = make([]byte, 5)
		bytes[0] = byte(schema.TypeU32)
		binary.LittleEndian.PutUint32(bytes[1:], uint32(value.Int()))

	case reflect.Uint64:
		bytes = make([]byte, 9)
		bytes[0] = byte(schema.TypeU64)
		binary.LittleEndian.PutUint64(bytes[1:], uint64(value.Int()))

	case reflect.Int8:
		bytes = make([]byte, 2)
		bytes[0] = byte(schema.TypeI8)
		bytes[1] = byte(int8(value.Int()))

	case reflect.Int16:
		bytes = make([]byte, 3)
		bytes[0] = byte(schema.TypeI16)
		binary.LittleEndian.PutUint16(bytes[1:], uint16(value.Int()))

	case reflect.Int32:
		bytes = make([]byte, 5)
		bytes[0] = byte(schema.TypeI32)
		binary.LittleEndian.PutUint32(bytes[1:], uint32(value.Int()))

	case reflect.Int64:
		bytes = make([]byte, 9)
		bytes[0] = byte(schema.TypeI64)
		binary.LittleEndian.PutUint64(bytes[1:], uint64(value.Int()))

	case reflect.Float32:
		bytes = make([]byte, 5)
		bytes[0] = byte(schema.TypeF32)
		binary.LittleEndian.PutUint32(bytes[1:], math.Float32bits(float32(value.Float())))

	case reflect.Float64:
		bytes = make([]byte, 9)
		bytes[0] = byte(schema.TypeF64)
		binary.LittleEndian.PutUint64(bytes[1:], math.Float64bits(value.Float()))

	case reflect.String:
		bytes = make([]byte, 5)
		bytes[0] = byte(schema.TypeString)
		str := value.String()
		binary.LittleEndian.PutUint32(bytes[1:], uint32(len(str)))
		bytes = append(bytes, []byte(str)...)

	case reflect.Slice, reflect.Array:
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
	return nil
}

func (e *Encoder) writeMap(value reflect.Value) error {
	return nil
}

func (e *Encoder) writeObject(value reflect.Value) error {
	return nil
}

func (e *Encoder) writeTime(t *time.Time) error {
	bytes := make([]byte, 9)
	bytes[0] = byte(schema.TypeTimestamp)
	binary.LittleEndian.PutUint64(bytes[1:], uint64(t.Unix()))
	_, err := e.writer.Write(bytes)
	return err
}

func (e *Encoder) writeName(value string) error {
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
