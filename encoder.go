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
	var bytes []byte
	switch i := value.Interface().(type) {
	case time.Time:
		return e.writeTime(value)

	case *time.Time:
		return e.writeTime(value.Elem())

	case Marshaler:
		return i.MarshalDO(e.writer)

	default:
		fmt.Printf("%v === \n", i)
	}
	/*
		switch value.Kind() {
		case reflect.Map:
			e.mapv(tag, in)
		case reflect.Ptr:
			e.marshal(tag, in.Elem())
		case reflect.Struct:
			e.structv(tag, in)
		case reflect.Slice, reflect.Array:
			e.slicev(tag, in)
		default:
			panic("cannot marshal type: " + in.Type().String())
		}*/
	switch value.Kind() {
	case reflect.Interface:
		return e.write(value.Elem())

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
		/*
			case []byte:
				bytes = make([]byte, 5)
				bytes[0] = byte(schema.TypeBytes)
				binary.LittleEndian.PutUint32(bytes[1:], uint32(len(value)))
				bytes = append(bytes, value...)*/
		//value.MethodByName()

	default:
		return fmt.Errorf("invalid type: %s", value.Type().String())
	}
	_, err := e.writer.Write(bytes)
	return err
}

func (e *Encoder) writeTime(value reflect.Value) error {
	fmt.Println("------------")
	t := value.Interface().(time.Time)
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
