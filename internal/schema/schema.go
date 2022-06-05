package schema

type Type int

const (
	TypeUndefined Type = 0

	TypeBool      Type = 1
	TypeSbyte     Type = 2
	TypeByte      Type = 3
	TypeShort     Type = 4
	TypeUshort    Type = 5
	TypeInt       Type = 6
	TypeUint      Type = 7
	TypeLong      Type = 8
	TypeUlong     Type = 9
	TypeFloat     Type = 10
	TypeDouble    Type = 11
	TypeString    Type = 12
	TypeTimestamp Type = 13
	TypeBytes     Type = 14

	TypeEnum   Type = 15
	TypeObject Type = 16
	TypeArray  Type = 17
	TypeMap    Type = 18
)

var (
	PrimaryTypes = map[string]Type{
		"bool":      TypeBool,
		"sbyte":     TypeSbyte,
		"byte":      TypeByte,
		"short":     TypeShort,
		"ushort":    TypeUshort,
		"int":       TypeInt,
		"uint":      TypeUint,
		"long":      TypeLong,
		"ulong":     TypeUlong,
		"float":     TypeFloat,
		"double":    TypeDouble,
		"string":    TypeString,
		"timestamp": TypeTimestamp,
		"bytes":     TypeBytes,
	}
)

type TypeInfo struct {
	Type       Type
	TypeRef    string
	SubType    Type
	SubTypeRef string
}

type Field struct {
	Name string
	Type *TypeInfo
}

type Message struct {
	Name   string
	Fields []*Field
}

type Enum struct {
	Name   string
	Values []string
}

type Rpc struct {
	Name   string
	Input  *TypeInfo
	Output *TypeInfo
}

type Service struct {
	Name string
	Rpc  []*Rpc
}

type File struct {
	FilePath string
	Package  string

	Enums    []*Enum
	Messages []*Message
	Services []*Service
}
