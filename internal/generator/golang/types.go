package golang

import (
	"fmt"
	"strings"

	"github.com/meerkat-io/bloom/format/strcase"
	"github.com/meerkat-io/disorder/internal/schema"
)

var goTypes = map[schema.Type]string{
	schema.TypeBool:      "bool",
	schema.TypeInt:       "int32",
	schema.TypeLong:      "int64",
	schema.TypeFloat:     "float32",
	schema.TypeDouble:    "float64",
	schema.TypeString:    "string",
	schema.TypeBytes:     "[]byte",
	schema.TypeTimestamp: "*time.Time",
}

func goType(typ *schema.TypeInfo) string {
	switch typ.Type {
	case schema.TypeArray:
		return fmt.Sprintf("[]%s", goType(typ.ElementType))
	case schema.TypeMap:
		return fmt.Sprintf("map[string]%s", goType(typ.ElementType))
	default:
		if typ.Type.IsPrimary() {
			return goTypes[typ.Type]
		}
	}
	if strings.Contains(typ.TypeRef, ".") {
		names := strings.Split(typ.TypeRef, ".")
		pkg := strcase.SnakeCase(names[len(names)-2])
		obj := strcase.PascalCase(names[len(names)-1])
		return fmt.Sprintf("*%s.%s", pkg, obj)
	}
	return fmt.Sprintf("*%s", strcase.PascalCase(typ.TypeRef))
}
