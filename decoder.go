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
	reader   io.Reader
	warnings []error
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		reader: r,
	}
}

func (d *Decoder) Decode(value interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("recover panic when decoding disorder data: %s", r)
		}
	}()
	var t tag
	t, err = d.readTag()
	if err != nil {
		return err
	}
	return d.read(t, reflect.ValueOf(value))
}

func (d *Decoder) Warnings() []error {
	return d.warnings
}

func (d *Decoder) read(t tag, value reflect.Value) error {
	switch i := value.Interface().(type) {
	case *time.Time:
		if t == tagTimestamp {
			time, err := d.readTime()
			if err != nil {
				return err
			}
			value.Elem().Set(reflect.ValueOf(*time))
			return nil
		} else {
			return fmt.Errorf("type mismatch: assign time to %s", value.Type())
		}

	case Enum:
		if t == tagEnum {
			enum, err := d.readName()
			if err != nil {
				return err
			}
			return i.SetValue(enum)
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

	case tagInt:
		bytes = make([]byte, 4)
		_, err := d.reader.Read(bytes)
		if err != nil {
			return err
		}
		resolved = int32(binary.BigEndian.Uint32(bytes))
		if value.Kind() == reflect.Int32 {
			value.SetInt(int64(resolved.(int32)))
			return nil
		}

	case tagLong:
		bytes = make([]byte, 8)
		_, err := d.reader.Read(bytes)
		if err != nil {
			return err
		}
		resolved = int64(binary.BigEndian.Uint64(bytes))
		if value.Kind() == reflect.Int64 {
			value.SetInt(resolved.(int64))
			return nil
		}

	case tagFloat:
		bytes = make([]byte, 4)
		_, err := d.reader.Read(bytes)
		if err != nil {
			return err
		}
		resolved = math.Float32frombits(binary.BigEndian.Uint32(bytes))
		if value.Kind() == reflect.Float32 {
			value.SetFloat(float64(resolved.(float32)))
			return nil
		}

	case tagDouble:
		bytes = make([]byte, 8)
		_, err := d.reader.Read(bytes)
		if err != nil {
			return err
		}
		resolved = math.Float64frombits(binary.BigEndian.Uint64(bytes))
		if value.Kind() == reflect.Float64 {
			value.SetFloat(resolved.(float64))
			return nil
		}

	case tagBytes:
		bytes = make([]byte, 4)
		_, err := d.reader.Read(bytes)
		if err != nil {
			return err
		}
		count := binary.BigEndian.Uint32(bytes)
		resolved = []byte{}
		if count > 0 {
			bytes = make([]byte, count)
			_, err := d.reader.Read(bytes)
			if err != nil {
				return err
			}
			resolved = bytes
		}
		if value.Kind() == reflect.Slice && value.Type().Elem().Kind() == reflect.Uint8 {
			value.Set(reflect.ValueOf(resolved))
			return nil
		}

	case tagString:
		bytes = make([]byte, 4)
		_, err := d.reader.Read(bytes)
		if err != nil {
			return err
		}
		count := binary.BigEndian.Uint32(bytes)
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

	case tagArrayStart:
		return d.readArray(value)

	case tagObjectStart:
		return d.readObject(value)

	default:
		return fmt.Errorf("invalid tag: %d", t)
	}
	return fmt.Errorf("type mismatch: assign %s to %s", reflect.ValueOf(resolved).Type(), value.Type())
}

func (d *Decoder) readArray(value reflect.Value) error {
	if value.Kind() != reflect.Slice {
		return fmt.Errorf("type mismatch, assign slice to type %s ", value.Type())
	}
	elementType := value.Type().Elem()
	t, err := d.readTag()
	if err != nil {
		return err
	}
	values := []reflect.Value{}
	for t != tagArrayEnd {
		element := reflect.New(elementType).Elem()
		if element.Kind() == reflect.Ptr && element.IsNil() {
			elementValue := reflect.New(element.Type().Elem())
			element.Set(elementValue)
		}
		err = d.read(t, element)
		if err != nil {
			return err
		}
		t, err = d.readTag()
		if err != nil {
			return err
		}
		values = append(values, element)
	}
	count := len(values)
	value.Set(reflect.MakeSlice(value.Type(), count, count))
	for i, v := range values {
		value.Index(i).Set(v)
	}
	return nil
}

func (d *Decoder) readObject(value reflect.Value) error {
	switch value.Kind() {
	case reflect.Struct:
		info, err := getStructInfo(value.Type())
		if err != nil {
			return err
		}
		t, err := d.readTag()
		if err != nil {
			return err
		}
		for t != tagObjectEnd {
			name, err := d.readName()
			if err != nil {
				return err
			}
			if fieldInfo, exists := info.fieldsMap[name]; exists {
				field := value.Field(fieldInfo.index)
				if field.Kind() == reflect.Ptr && field.IsNil() {
					fieldValue := reflect.New(field.Type().Elem())
					field.Set(fieldValue)
				}
				err = d.read(t, field)
				if err != nil {
					return fmt.Errorf("assign \"%s\" to field \"%s\" in struct \"%s\" failed: %s", field.Type(), name, value.Type(), err.Error())
				}
			} else {
				d.warnings = append(d.warnings, fmt.Errorf("field %s not found in struct %s", name, value.Type()))
				err = d.skip(t)
				if err != nil {
					return err
				}
			}
			t, err = d.readTag()
			if err != nil {
				return err
			}
		}
		return nil

	case reflect.Map:
		if value.Type().Key().Kind() != reflect.String {
			return fmt.Errorf("key type of map must be string")
		}

	default:
		return fmt.Errorf("type mismatch, assign map to type %s ", value.Type())
	}

	valueType := value.Type()
	keyType := valueType.Key()
	elementType := valueType.Elem()
	if value.IsNil() {
		value.Set(reflect.MakeMap(valueType))
	}
	t, err := d.readTag()
	if err != nil {
		return err
	}
	for t != tagObjectEnd {
		name, err := d.readName()
		if err != nil {
			return err
		}
		key := reflect.New(keyType).Elem()
		key.SetString(name)
		element := reflect.New(elementType).Elem()
		if element.Kind() == reflect.Ptr && element.IsNil() {
			elementValue := reflect.New(element.Type().Elem())
			element.Set(elementValue)
		}
		err = d.read(t, element)
		if err != nil {
			return err
		}
		value.SetMapIndex(key, element)
		t, err = d.readTag()
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *Decoder) readTime() (*time.Time, error) {
	bytes := make([]byte, 8)
	_, err := d.reader.Read(bytes)
	if err != nil {
		return nil, err
	}
	timestamp := int64(binary.BigEndian.Uint64(bytes))
	t := time.UnixMilli(timestamp)
	return &t, nil
}

func (d *Decoder) readName() (string, error) {
	bytes := make([]byte, 1)
	_, err := d.reader.Read(bytes)
	if err != nil {
		return "", err
	}
	count := bytes[0]
	if count == 0 {
		return "", fmt.Errorf("empty name")
	}
	bytes = make([]byte, count)
	_, err = d.reader.Read(bytes)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (d *Decoder) readTag() (tag, error) {
	t := make([]byte, 1)
	_, err := d.reader.Read(t)
	return tag(t[0]), err
}

func (d *Decoder) skip(t tag) error {
	var bytes []byte
	switch t {
	case tagBool:
		return d.skipBytes(1)

	case tagInt, tagFloat:
		return d.skipBytes(4)

	case tagLong, tagDouble, tagTimestamp:
		return d.skipBytes(8)

	case tagString, tagBytes:
		bytes = make([]byte, 4)
		_, err := d.reader.Read(bytes)
		if err != nil {
			return err
		}
		count := binary.BigEndian.Uint32(bytes)
		return d.skipBytes(int(count))

	case tagEnum:
		return d.skipName()

	case tagArrayStart:
		return d.skipArray()

	case tagObjectStart:
		return d.skipObject()

	default:
		return fmt.Errorf("invalid tag: %d", t)
	}
}

func (d *Decoder) skipBytes(count int) error {
	bytes := make([]byte, count)
	_, err := d.reader.Read(bytes)
	return err
}

func (d *Decoder) skipName() error {
	bytes := make([]byte, 1)
	_, err := d.reader.Read(bytes)
	if err != nil {
		return err
	}
	count := bytes[0]
	if count == 0 {
		return nil
	}
	err = d.skipBytes(int(count))
	return err
}

func (d *Decoder) skipArray() error {
	t, err := d.readTag()
	if err != nil {
		return err
	}
	for t != tagArrayEnd {
		err = d.skip(t)
		if err != nil {
			return err
		}
		t, err = d.readTag()
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *Decoder) skipObject() error {
	t, err := d.readTag()
	if err != nil {
		return err
	}
	for t != tagObjectEnd {
		err = d.skipName()
		if err != nil {
			return err
		}
		err = d.skip(t)
		if err != nil {
			return err
		}
		t, err = d.readTag()
		if err != nil {
			return err
		}
	}
	return nil
}
