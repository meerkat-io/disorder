package disorder

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"reflect"
	"time"
)

type Decoder struct {
	reader io.Reader
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		reader: r,
	}
}

func (d *Decoder) Decode(value interface{}) error {
	//TO-DO recover panic
	t := make([]byte, 1)
	_, err := d.reader.Read(t)
	if err != nil {
		return err
	}
	return d.read(tag(t[0]), reflect.ValueOf(value))
}

func (d *Decoder) read(t tag, value reflect.Value) error {
	if !value.IsValid() || value.Kind() == reflect.Ptr && value.IsNil() {
		return fmt.Errorf("invalid value or type")
	}

	switch i := value.Interface().(type) {
	case Unmarshaler:
		return i.UnmarshalDO(d.reader)

	case Enum:
		if t == tagEnum {
			enum, err := d.readName()
			if err != nil {
				return err
			}
			return i.FromString(enum)
		} else {
			return fmt.Errorf("type mismatch: assign enum to %s", value.Type())
		}
	}

	if value.Kind() == reflect.Ptr {
		return d.read(t, value.Elem())
	}

	var bytes []byte
	var resolved interface{}
	switch t {
	case tagBool:
		bytes = make([]byte, 1)
		_, err := d.reader.Read(bytes)
		if err != nil {
			return err
		}
		resolved = bytes[0] == 1
		if value.Kind() == reflect.Bool {
			value.SetBool(resolved.(bool))
			return nil
		}

	case tagU8:
		bytes = make([]byte, 1)
		_, err := d.reader.Read(bytes)
		if err != nil {
			return err
		}
		resolved = uint8(bytes[0])
		if value.Kind() == reflect.Uint8 {
			value.SetUint(uint64(resolved.(uint8)))
			return nil
		}

	case tagU16:
		bytes = make([]byte, 2)
		_, err := d.reader.Read(bytes)
		if err != nil {
			return err
		}
		resolved = binary.LittleEndian.Uint16(bytes)
		if value.Kind() == reflect.Uint16 {
			value.SetUint(uint64(resolved.(uint16)))
			return nil
		}

	case tagU32:
		bytes = make([]byte, 4)
		_, err := d.reader.Read(bytes)
		if err != nil {
			return err
		}
		resolved = binary.LittleEndian.Uint32(bytes)
		if value.Kind() == reflect.Uint32 {
			value.SetUint(uint64(resolved.(uint32)))
			return nil
		}

	case tagU64:
		bytes = make([]byte, 8)
		_, err := d.reader.Read(bytes)
		if err != nil {
			return err
		}
		resolved = binary.LittleEndian.Uint64(bytes)
		if value.Kind() == reflect.Uint64 {
			value.SetUint(resolved.(uint64))
			return nil
		}

	case tagI8:
		bytes = make([]byte, 1)
		_, err := d.reader.Read(bytes)
		if err != nil {
			return err
		}
		resolved = int8(bytes[0])
		if value.Kind() == reflect.Int8 {
			value.SetInt(int64(resolved.(uint8)))
			return nil
		}

	case tagI16:
		bytes = make([]byte, 2)
		_, err := d.reader.Read(bytes)
		if err != nil {
			return err
		}
		resolved = int16(binary.LittleEndian.Uint16(bytes))
		if value.Kind() == reflect.Int16 {
			value.SetInt(int64(resolved.(int16)))
			return nil
		}

	case tagI32:
		bytes = make([]byte, 4)
		_, err := d.reader.Read(bytes)
		if err != nil {
			return err
		}
		resolved = int32(binary.LittleEndian.Uint32(bytes))
		if value.Kind() == reflect.Int32 {
			value.SetInt(int64(resolved.(int32)))
			return nil
		}

	case tagI64:
		bytes = make([]byte, 8)
		_, err := d.reader.Read(bytes)
		if err != nil {
			return err
		}
		resolved = int64(binary.LittleEndian.Uint64(bytes))
		if value.Kind() == reflect.Int64 {
			value.SetInt(resolved.(int64))
			return nil
		}

	case tagF32:
		bytes = make([]byte, 4)
		_, err := d.reader.Read(bytes)
		if err != nil {
			return err
		}
		resolved = math.Float32frombits(binary.LittleEndian.Uint32(bytes))
		if value.Kind() == reflect.Float32 {
			value.SetFloat(float64(resolved.(float32)))
			return nil
		}

	case tagF64:
		bytes = make([]byte, 8)
		_, err := d.reader.Read(bytes)
		if err != nil {
			return err
		}
		resolved = math.Float64frombits(binary.LittleEndian.Uint64(bytes))
		if value.Kind() == reflect.Float64 {
			value.SetFloat(resolved.(float64))
			return nil
		}

	case tagString:
		bytes = make([]byte, 4)
		_, err := d.reader.Read(bytes)
		if err != nil {
			return err
		}
		count := binary.LittleEndian.Uint32(bytes)
		resolved = ""
		if count > 0 {
			bytes = make([]byte, count)
			_, err := d.reader.Read(bytes)
			if err != nil {
				return err
			}
			resolved = string(bytes)
		}
		if value.Kind() == reflect.String {
			value.SetString(resolved.(string))
			return nil
		}

		/*
			case time.Time:
				return e.writeTime(&i)

			case *time.Time:
				return e.writeTime(i)
		*/
	case tagEnum:
		var err error
		resolved, err = d.readName()
		if err != nil {
			return err
		}

		/*
			case reflect.Slice, reflect.Array:
				return e.writeArray(value)

			case reflect.Map:
				return e.writeMap(value)

			case reflect.Struct:
				return e.writeObject(value)
		*/
	default:
		return fmt.Errorf("invalid tag: %d", t)
	}

	if value.Kind() == reflect.Interface {
		value.Set(reflect.ValueOf(resolved))
		return nil
	}

	return fmt.Errorf("type mismatch: assign %s to %s", reflect.ValueOf(resolved).Type(), value.Type())
}

func (d *Decoder) readArray(value reflect.Value) error {
	/*
		count := value.Len()
		if count == 0 {
			return nil
		}

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
		return err*/
	return nil
}

/*
func (d *Decoder) writeMap(value reflect.Value) error {
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
}*/

func (d *Decoder) readObject(value reflect.Value) error {
	/*
		info, err := getStructInfo(value.Type())
		if err != nil {
			return err
		}

		_, err = d.reader.Read([]byte{byte(tagStartObject)})
		if err != nil {
			return err
		}

		for _, field := range info.fieldsList {
			fieldValue := value.Field(field.index)
			if field.omitEmpty && isZero(fieldValue) {
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

		_, err = d.reader.Read([]byte{byte(tagEndObject)})
		return err*/

	return nil
}

func (d *Decoder) readTime() (*time.Time, error) {
	/*
		bytes := make([]byte, 9)
		bytes[0] = byte(tagTimestamp)
		binary.LittleEndian.PutUint64(bytes[1:], uint64(t.Unix()))
		_, err := e.writer.Write(bytes)
		return err*/
	return nil, nil
}

func (d *Decoder) readName() (string, error) {
	bytes := make([]byte, 1)
	_, err := d.reader.Read(bytes)
	if err != nil {
		return "", err
	}
	count := uint8(bytes[0])
	if count > 0 {
		bytes = make([]byte, count)
		_, err := d.reader.Read(bytes)
		if err != nil {
			return "", err
		}
		return string(bytes), nil
	}
	return "", nil
}
