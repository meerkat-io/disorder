package disorder

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"reflect"
	"sort"
	"time"
)

// Encoder writes values in the disorder binary format.
type Encoder struct {
	writer io.Writer
}

// NewEncoder returns an Encoder that writes to w.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		writer: w,
	}
}

// Encode writes value to the underlying writer. Map keys are written in sorted
// order, so the output is deterministic.
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

	switch i := value.Interface().(type) {
	case *time.Time:
		return e.writeTimestamp(name, *i)
	case time.Time:
		return e.writeTimestamp(name, i)
	case Enum:
		enum, err := i.GetValue()
		if err != nil {
			return err
		}
		if len(enum) > 255 {
			return fmt.Errorf("enum length overflow. should less than 255")
		}
		return e.writeScalar(tagEnum, name, append([]byte{byte(len(enum))}, enum...))
	}

	switch value.Kind() {
	case reflect.Ptr, reflect.Interface:
		return e.write(name, value.Elem())

	case reflect.Bool:
		b := byte(0)
		if value.Bool() {
			b = 1
		}
		return e.writeScalar(tagBool, name, []byte{b})

	case reflect.Int32:
		return e.writeScalar(tagInt, name, binary.BigEndian.AppendUint32(nil, uint32(value.Int())))

	case reflect.Int64:
		return e.writeScalar(tagLong, name, binary.BigEndian.AppendUint64(nil, uint64(value.Int())))

	case reflect.Float32:
		return e.writeScalar(tagFloat, name, binary.BigEndian.AppendUint32(nil, math.Float32bits(float32(value.Float()))))

	case reflect.Float64:
		return e.writeScalar(tagDouble, name, binary.BigEndian.AppendUint64(nil, math.Float64bits(value.Float())))

	case reflect.String:
		str := value.String()
		if uint64(len(str)) > math.MaxUint32 {
			return fmt.Errorf("string length overflow")
		}
		return e.writeScalar(tagString, name, append(binary.BigEndian.AppendUint32(nil, uint32(len(str))), str...))

	case reflect.Slice:
		if value.Type().Elem().Kind() == reflect.Uint8 {
			data := value.Bytes()
			if uint64(len(data)) > math.MaxUint32 {
				return fmt.Errorf("bytes length overflow")
			}
			return e.writeScalar(tagBytes, name, append(binary.BigEndian.AppendUint32(nil, uint32(len(data))), data...))
		}
		return e.writeArray(name, value)

	case reflect.Array:
		return e.writeArray(name, value)

	case reflect.Map:
		return e.writeMap(name, value)

	case reflect.Struct:
		return e.writeObject(name, value)
	}
	return fmt.Errorf("unsupported type: %s", value.Type().String())
}

func (e *Encoder) writeTimestamp(name string, t time.Time) error {
	return e.writeScalar(tagTimestamp, name, binary.BigEndian.AppendUint64(nil, uint64(t.UnixMilli())))
}

// writeScalar writes tag + name + payload.
func (e *Encoder) writeScalar(t tag, name string, payload []byte) error {
	if err := e.writeHeader(t, name); err != nil {
		return err
	}
	_, err := e.writer.Write(payload)
	return err
}

func (e *Encoder) writeArray(name string, value reflect.Value) error {
	if err := e.writeHeader(tagArrayStart, name); err != nil {
		return err
	}
	count := value.Len()
	for i := 0; i < count; i++ {
		if err := e.write("", value.Index(i)); err != nil {
			return err
		}
	}
	return e.writeTag(tagArrayEnd)
}

func (e *Encoder) writeMap(name string, value reflect.Value) error {
	if value.Type().Key().Kind() != reflect.String {
		return fmt.Errorf("map key type must be string")
	}
	if err := e.writeHeader(tagObjectStart, name); err != nil {
		return err
	}

	keys := value.MapKeys()
	sort.Slice(keys, func(i, j int) bool { return keys[i].String() < keys[j].String() })
	for _, key := range keys {
		if key.String() == "" {
			return fmt.Errorf("map key cannot be empty")
		}
		if err := e.write(key.String(), value.MapIndex(key)); err != nil {
			return err
		}
	}
	return e.writeTag(tagObjectEnd)
}

func (e *Encoder) writeObject(name string, value reflect.Value) error {
	info, err := getStructInfo(value.Type())
	if err != nil {
		return err
	}
	if err := e.writeHeader(tagObjectStart, name); err != nil {
		return err
	}
	for _, field := range info.fieldsList {
		if err := e.write(field.key, value.Field(field.index)); err != nil {
			return err
		}
	}
	return e.writeTag(tagObjectEnd)
}

// writeHeader writes tag + optional name. The name is empty for the root value
// and array elements.
func (e *Encoder) writeHeader(t tag, name string) error {
	if err := e.writeTag(t); err != nil {
		return err
	}
	return e.writeName(name)
}

func (e *Encoder) writeName(value string) error {
	if len(value) == 0 {
		return nil
	}
	if len(value) > 255 {
		return fmt.Errorf("string length overflow. should less than 255")
	}
	_, err := e.writer.Write(append([]byte{byte(len(value))}, value...))
	return err
}

func (e *Encoder) writeTag(t tag) error {
	_, err := e.writer.Write([]byte{byte(t)})
	return err
}
