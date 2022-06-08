package disorder

import (
	"encoding/binary"
	"io"
	"math"

	"github.com/meerkat-lib/disorder/internal/schema"
)

type writer struct {
	writer io.Writer
}

func newWriter(w io.Writer) *writer {
	return &writer{
		writer: w,
	}
}

func (w *writer) write(data interface{}) error {
	var bytes []byte
	switch value := data.(type) {
	case bool:
		bytes = make([]byte, 2)
		bytes[0] = byte(schema.TypeBool)
		if value {
			bytes[1] = 1
		} else {
			bytes[1] = 0
		}

	case uint8:
		bytes = make([]byte, 2)
		bytes[0] = byte(schema.TypeU8)
		bytes[1] = byte(value)

	case uint16:
		bytes = make([]byte, 3)
		bytes[0] = byte(schema.TypeU16)
		binary.LittleEndian.PutUint16(bytes[1:], value)

	case uint32:
		bytes = make([]byte, 5)
		bytes[0] = byte(schema.TypeU32)
		binary.LittleEndian.PutUint32(bytes[1:], value)

	case uint64:
		bytes = make([]byte, 9)
		bytes[0] = byte(schema.TypeU64)
		binary.LittleEndian.PutUint64(bytes[1:], value)

	case int8:
		bytes = make([]byte, 2)
		bytes[0] = byte(schema.TypeI8)
		bytes[1] = byte(value)

	case int16:
		bytes = make([]byte, 3)
		bytes[0] = byte(schema.TypeI16)
		binary.LittleEndian.PutUint16(bytes[1:], uint16(value))

	case int32:
		bytes = make([]byte, 5)
		bytes[0] = byte(schema.TypeI32)
		binary.LittleEndian.PutUint32(bytes[1:], uint32(value))

	case int64:
		bytes = make([]byte, 9)
		bytes[0] = byte(schema.TypeI64)
		binary.LittleEndian.PutUint64(bytes[1:], uint64(value))

	case float32:
		bytes = make([]byte, 5)
		bytes[0] = byte(schema.TypeF32)
		binary.LittleEndian.PutUint32(bytes[1:], math.Float32bits(value))

	case float64:
		bytes = make([]byte, 9)
		bytes[0] = byte(schema.TypeF64)
		binary.LittleEndian.PutUint64(bytes[1:], math.Float64bits(value))
	}
	_, err := w.writer.Write(bytes)
	return err
}
