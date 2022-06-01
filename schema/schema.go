package schema

//"github.com/meerkat-lib/cellar/util/logger"
//"github.com/meerkat-lib/cellar/util/strcase"

type Type int

const (
	TypeNull Type = 0

	TypeBool   Type = 1
	TypeSbyte  Type = 2
	TypeByte   Type = 3
	TypeShort  Type = 4
	TypeUshort Type = 5
	TypeInt    Type = 6
	TypeUint   Type = 7
	TypeLong   Type = 8
	TypeUlong  Type = 9
	TypeFloat  Type = 10
	TypeDouble Type = 11
	TypeString Type = 12
	TypeEnum   Type = 13
	TypeObject Type = 14

	TypeBoolArray   Type = 17
	TypeSbyteArray  Type = 18
	TypeByteArray   Type = 19
	TypeShortArray  Type = 20
	TypeUshortArray Type = 21
	TypeIntArray    Type = 22
	TypeUintArray   Type = 23
	TypeLongArray   Type = 24
	TypeUlongArray  Type = 25
	TypeFloatArray  Type = 26
	TypeDoubleArray Type = 27
	TypeStringArray Type = 28
	TypeEnumArray   Type = 29
	TypeObjectArray Type = 30

	TypeBoolMap   Type = 33
	TypeSbyteMap  Type = 34
	TypeByteMap   Type = 35
	TypeShortMap  Type = 36
	TypeUshortMap Type = 37
	TypeIntMap    Type = 38
	TypeUintMap   Type = 39
	TypeLongMap   Type = 40
	TypeUlongMap  Type = 41
	TypeFloatMap  Type = 42
	TypeDoubleMap Type = 43
	TypeStringMap Type = 44
	TypeEnumMap   Type = 45
	TypeObjectMap Type = 46
)

const (
	PackageGlobal = "global"
)

type Field struct {
	Name    string
	Type    Type
	SubType string
}

type Message struct {
	Name    string
	Package string
	Fields  []*Field
}

type Enum struct {
	Name    string
	Package string
	Values  []string
}

type Param struct {
	Type    Type
	SubType string
}

type Rpc struct {
	Name   string
	Input  *Param
	Output *Param
}

type SchemaFile struct {
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
