package schema

type Type byte

const (
	TypeUndefined Type = 0

	TypeBool   Type = 1
	TypeI8     Type = 2
	TypeU8     Type = 3
	TypeI16    Type = 4
	TypeU16    Type = 5
	TypeI32    Type = 6
	TypeU32    Type = 7
	TypeI64    Type = 8
	TypeU64    Type = 9
	TypeF32    Type = 10
	TypeF64    Type = 11
	TypeString Type = 12

	TypeEnum      Type = 13
	TypeTimestamp Type = 14

	TypeArray  Type = 15
	TypeMap    Type = 16
	TypeObject Type = 17
)

func (t Type) IsPrimary() bool {
	return t >= TypeBool && t <= TypeString
}

var (
	PrimaryTypes = map[string]Type{
		"bool":      TypeBool,
		"i8":        TypeI8,
		"u8":        TypeU8,
		"i16":       TypeI16,
		"u16":       TypeU16,
		"i32":       TypeI32,
		"u32":       TypeU32,
		"i64":       TypeI64,
		"u64":       TypeU64,
		"f32":       TypeF32,
		"f64":       TypeF64,
		"string":    TypeString,
		"timestamp": TypeTimestamp,
	}
)

type TypeInfo struct {
	Type    Type
	SubType Type
	TypeRef string
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
	Imports  []string
	Options  map[string]string

	Enums    []*Enum
	Messages []*Message
	Services []*Service

	AbsImports      []string
	ResolvedImports []string
}
