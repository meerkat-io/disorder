package schema

type Type byte

const (
	TypeUndefined Type = 0

	TypeBool   Type = 1
	TypeInt    Type = 2
	TypeLong   Type = 3
	TypeFloat  Type = 4
	TypeDouble Type = 5
	TypeBytes  Type = 6

	TypeString    Type = 11
	TypeTimestamp Type = 12
	TypeEnum      Type = 13

	TypeArray  Type = 21
	TypeMap    Type = 22
	TypeObject Type = 23
)

func (t Type) IsPrimary() bool {
	return t >= TypeBool && t <= TypeTimestamp
}

var (
	PrimaryTypes = map[string]Type{
		"bool":      TypeBool,
		"int":       TypeInt,
		"long":      TypeLong,
		"float":     TypeFloat,
		"double":    TypeDouble,
		"bytes":     TypeBytes,
		"string":    TypeString,
		"timestamp": TypeTimestamp,
	}
)

type TypeInfo struct {
	Type        Type
	TypeRef     string
	Qualified   string
	ElementType *TypeInfo
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

	AbsImports map[string]bool
}
