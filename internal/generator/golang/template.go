package golang

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/meerkat-io/bloom/strcase"
	goTemplate "github.com/meerkat-io/disorder/internal/generator/golang/template"
	"github.com/meerkat-io/disorder/internal/schema"
)

func newTemplate() *template.Template {
	funcMap := template.FuncMap{
		"PascalCase": strcase.PascalCase,
		"CamelCase":  strcase.CamelCase,
		"PackageName": func(pkg string) string {
			names := strings.Split(pkg, ".")
			return strcase.SnakeCase(names[len(names)-1])
		},
		"Type": goType,
		"Tag":  fieldTag,
	}
	return template.Must(template.New("define").Funcs(funcMap).Parse(goTemplate.DefineTemplate))
}

// fieldTag returns the struct tag of a field. Types that have a natural empty
// value in Go are omitted from JSON when empty.
func fieldTag(typ *schema.TypeInfo, name string) string {
	omitEmpty := ""
	switch typ.Type {
	case schema.TypeTimestamp, schema.TypeEnum, schema.TypeObject, schema.TypeArray, schema.TypeMap:
		omitEmpty = ",omitempty"
	}
	return fmt.Sprintf("`disorder:\"%s\" json:\"%s%s\"`", name, name, omitEmpty)
}
