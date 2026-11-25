package disorder

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"reflect"
	"time"
)

var (
	timeType = reflect.TypeOf(time.Time{})
	enumType = reflect.TypeOf((*Enum)(nil)).Elem()
)

// smallRead is the largest length allocated up front. Larger lengths are read
// incrementally, so a corrupt length prefix cannot trigger a huge allocation.
const smallRead = 64 * 1024

// Decoder reads values in the disorder binary format.
type Decoder struct {
	reader   io.Reader
	warnings []error
}

// NewDecoder returns a Decoder that reads from r.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		reader: r,
	}
}

// Decode reads one value into value, which must be a non-nil pointer.
// A target of type interface{} receives bool, int32, int64, float32, float64,
// string, []byte, time.Time, []interface{} and map[string]interface{}; enums
// are decoded as their string value.
func (d *Decoder) Decode(value interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("recover panic when decoding disorder data: %v", r)
		}
	}()
	v := reflect.ValueOf(value)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return fmt.Errorf("decode target must be a non-nil pointer")
	}
	// io.EOF here means a clean end of stream; EOF inside a value is
	// io.ErrUnexpectedEOF.
	first := make([]byte, 1)
	if _, err = io.ReadFull(d.reader, first); err != nil {
		return err
	}
	return d.read(tag(first[0]), v)
}

// Warnings returns non-fatal problems found while decoding, e.g. fields that
// do not exist in the target struct.
func (d *Decoder) Warnings() []error {
	return d.warnings
}

func (d *Decoder) read(t tag, value reflect.Value) error {
	if value.Kind() == reflect.Ptr {
		if value.IsNil() {
			value.Set(reflect.New(value.Type().Elem()))
		}
		if value.Type().Implements(enumType) {
			return d.readEnum(t, value.Interface().(Enum))
		}
		return d.read(t, value.Elem())
	}
	if value.CanAddr() && value.Addr().Type().Implements(enumType) {
		return d.readEnum(t, value.Addr().Interface().(Enum))
	}

	if value.Kind() == reflect.Interface {
		if value.NumMethod() != 0 {
			return fmt.Errorf("type mismatch: cannot assign to non-empty interface %s", value.Type())
		}
		resolved, err := d.readAny(t)
		if err != nil {
			return err
		}
		value.Set(reflect.ValueOf(resolved))
		return nil
	}

	switch t {
	case tagArrayStart:
		return d.readArray(value)
	case tagObjectStart:
		return d.readObject(value)
	}

	if t == tagEnum && value.Kind() == reflect.String {
		name, err := d.readName()
		if err != nil {
			return err
		}
		value.SetString(name)
		return nil
	}

	resolved, err := d.readScalar(t)
	if err != nil {
		return err
	}
	rv := reflect.ValueOf(resolved)
	if rv.Type() == timeType && value.Type() == timeType {
		value.Set(rv)
		return nil
	}
	if rv.Kind() == value.Kind() && rv.Kind() != reflect.Struct &&
		(rv.Kind() != reflect.Slice || value.Type().Elem().Kind() == reflect.Uint8) {
		value.Set(rv.Convert(value.Type()))
		return nil
	}
	return fmt.Errorf("type mismatch: assign %s to %s", rv.Type(), value.Type())
}

// readEnum accepts both enum and string tags, so generic data (where enums are
// plain strings) can be decoded into typed enums. SetValue validates the value.
func (d *Decoder) readEnum(t tag, enum Enum) error {
	var name string
	switch t {
	case tagEnum:
		var err error
		if name, err = d.readName(); err != nil {
			return err
		}
	case tagString:
		b, err := d.readLengthPrefixed()
		if err != nil {
			return err
		}
		name = string(b)
	default:
		return fmt.Errorf("type mismatch: assign tag %d to enum %T", t, enum)
	}
	return enum.SetValue(name)
}

// readScalar reads the payload of a non-container tag.
func (d *Decoder) readScalar(t tag) (interface{}, error) {
	switch t {
	case tagBool:
		b, err := d.readFull(1)
		if err != nil {
			return nil, err
		}
		return b[0] == 1, nil

	case tagInt:
		b, err := d.readFull(4)
		if err != nil {
			return nil, err
		}
		return int32(binary.BigEndian.Uint32(b)), nil

	case tagLong:
		b, err := d.readFull(8)
		if err != nil {
			return nil, err
		}
		return int64(binary.BigEndian.Uint64(b)), nil

	case tagFloat:
		b, err := d.readFull(4)
		if err != nil {
			return nil, err
		}
		return math.Float32frombits(binary.BigEndian.Uint32(b)), nil

	case tagDouble:
		b, err := d.readFull(8)
		if err != nil {
			return nil, err
		}
		return math.Float64frombits(binary.BigEndian.Uint64(b)), nil

	case tagBytes:
		b, err := d.readLengthPrefixed()
		if err != nil {
			return nil, err
		}
		return b, nil

	case tagString:
		b, err := d.readLengthPrefixed()
		if err != nil {
			return nil, err
		}
		return string(b), nil

	case tagTimestamp:
		b, err := d.readFull(8)
		if err != nil {
			return nil, err
		}
		return time.UnixMilli(int64(binary.BigEndian.Uint64(b))), nil
	}
	return nil, fmt.Errorf("invalid tag: %d", t)
}

// readAny reads a value of any type into its generic Go representation.
func (d *Decoder) readAny(t tag) (interface{}, error) {
	switch t {
	case tagEnum:
		return d.readName()

	case tagArrayStart:
		array := []interface{}{}
		for {
			t, err := d.readTag()
			if err != nil {
				return nil, err
			}
			if t == tagArrayEnd {
				return array, nil
			}
			element, err := d.readAny(t)
			if err != nil {
				return nil, err
			}
			array = append(array, element)
		}

	case tagObjectStart:
		object := map[string]interface{}{}
		for {
			t, err := d.readTag()
			if err != nil {
				return nil, err
			}
			if t == tagObjectEnd {
				return object, nil
			}
			name, err := d.readName()
			if err != nil {
				return nil, err
			}
			element, err := d.readAny(t)
			if err != nil {
				return nil, err
			}
			object[name] = element
		}
	}
	return d.readScalar(t)
}

func (d *Decoder) readArray(value reflect.Value) error {
	if value.Kind() != reflect.Slice {
		return fmt.Errorf("type mismatch, assign slice to type %s ", value.Type())
	}
	elements := reflect.MakeSlice(value.Type(), 0, 0)
	for {
		t, err := d.readTag()
		if err != nil {
			return err
		}
		if t == tagArrayEnd {
			break
		}
		element := reflect.New(value.Type().Elem()).Elem()
		if err = d.read(t, element); err != nil {
			return err
		}
		elements = reflect.Append(elements, element)
	}
	value.Set(elements)
	return nil
}

func (d *Decoder) readObject(value reflect.Value) error {
	switch value.Kind() {
	case reflect.Struct:
		return d.readStruct(value)
	case reflect.Map:
		if value.Type().Key().Kind() != reflect.String {
			return fmt.Errorf("key type of map must be string")
		}
	default:
		return fmt.Errorf("type mismatch, assign map to type %s ", value.Type())
	}

	valueType := value.Type()
	if value.IsNil() {
		value.Set(reflect.MakeMap(valueType))
	}
	for {
		t, err := d.readTag()
		if err != nil {
			return err
		}
		if t == tagObjectEnd {
			return nil
		}
		name, err := d.readName()
		if err != nil {
			return err
		}
		key := reflect.New(valueType.Key()).Elem()
		key.SetString(name)
		element := reflect.New(valueType.Elem()).Elem()
		if err = d.read(t, element); err != nil {
			return fmt.Errorf("assign \"%s\" to map element \"%s\" failed: %s", element.Type(), name, err.Error())
		}
		value.SetMapIndex(key, element)
	}
}

func (d *Decoder) readStruct(value reflect.Value) error {
	info, err := getStructInfo(value.Type())
	if err != nil {
		return err
	}
	for {
		t, err := d.readTag()
		if err != nil {
			return err
		}
		if t == tagObjectEnd {
			return nil
		}
		name, err := d.readName()
		if err != nil {
			return err
		}
		fieldInfo, exists := info.fieldsMap[name]
		if !exists {
			d.warnings = append(d.warnings, fmt.Errorf("field %s not found in struct %s", name, value.Type()))
			if err = d.skip(t); err != nil {
				return err
			}
			continue
		}
		field := value.Field(fieldInfo.index)
		if err = d.read(t, field); err != nil {
			return fmt.Errorf("assign \"%s\" to field \"%s\" in struct \"%s\" failed: %s", field.Type(), name, value.Type(), err.Error())
		}
	}
}

func (d *Decoder) readTag() (tag, error) {
	b, err := d.readFull(1)
	if err != nil {
		return tagUndefined, err
	}
	return tag(b[0]), nil
}

func (d *Decoder) readName() (string, error) {
	b, err := d.readFull(1)
	if err != nil {
		return "", err
	}
	if b[0] == 0 {
		return "", fmt.Errorf("empty name")
	}
	b, err = d.readFull(int(b[0]))
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// readLengthPrefixed reads a 4 bytes length followed by that many bytes. The
// result is never nil.
func (d *Decoder) readLengthPrefixed() ([]byte, error) {
	b, err := d.readFull(4)
	if err != nil {
		return nil, err
	}
	count := binary.BigEndian.Uint32(b)
	if count <= smallRead {
		return d.readFull(int(count))
	}
	data, err := io.ReadAll(io.LimitReader(d.reader, int64(count)))
	if err != nil {
		return nil, err
	}
	if len(data) != int(count) {
		return nil, io.ErrUnexpectedEOF
	}
	return data, nil
}

// readFull reads exactly count bytes. EOF in the middle of a value is reported
// as io.ErrUnexpectedEOF.
func (d *Decoder) readFull(count int) ([]byte, error) {
	b := make([]byte, count)
	if _, err := io.ReadFull(d.reader, b); err != nil {
		if err == io.EOF && count > 0 {
			return nil, io.ErrUnexpectedEOF
		}
		return nil, err
	}
	return b, nil
}

func (d *Decoder) skip(t tag) error {
	switch t {
	case tagBool:
		return d.skipBytes(1)

	case tagInt, tagFloat:
		return d.skipBytes(4)

	case tagLong, tagDouble, tagTimestamp:
		return d.skipBytes(8)

	case tagString, tagBytes:
		b, err := d.readFull(4)
		if err != nil {
			return err
		}
		return d.skipBytes(int64(binary.BigEndian.Uint32(b)))

	case tagEnum:
		b, err := d.readFull(1)
		if err != nil {
			return err
		}
		return d.skipBytes(int64(b[0]))

	case tagArrayStart:
		for {
			t, err := d.readTag()
			if err != nil {
				return err
			}
			if t == tagArrayEnd {
				return nil
			}
			if err = d.skip(t); err != nil {
				return err
			}
		}

	case tagObjectStart:
		for {
			t, err := d.readTag()
			if err != nil {
				return err
			}
			if t == tagObjectEnd {
				return nil
			}
			if _, err = d.readName(); err != nil {
				return err
			}
			if err = d.skip(t); err != nil {
				return err
			}
		}
	}
	return fmt.Errorf("invalid tag: %d", t)
}

func (d *Decoder) skipBytes(count int64) error {
	n, err := io.CopyN(io.Discard, d.reader, count)
	if err == io.EOF && n < count {
		return io.ErrUnexpectedEOF
	}
	return err
}
