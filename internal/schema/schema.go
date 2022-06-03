package schema

//"github.com/meerkat-lib/cellar/util/logger"
//"github.com/meerkat-lib/cellar/util/strcase"

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

	TypeEnum   Type = 14
	TypeObject Type = 15
	TypeArray  Type = 16
	TypeMap    Type = 17
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
	}
)

const (
	PackageGlobal = "global"
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

type File struct {
	FilePath string

	Enums    []*Enum
	Messages []*Message
	Services []*Rpc
}

/*
var (
	types = map[string]int{
		"bool":   1,
		"sbyte":  1,
		"byte":   1,
		"short":  2,
		"ushort": 2,
		"int":    4,
		"uint":   4,
		"long":   8,
		"ulong":  8,
		"float":  4,
		"double": 8,
		"string": 0,
	}
)
*/
