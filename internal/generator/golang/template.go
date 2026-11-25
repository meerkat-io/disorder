package golang

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/meerkat-io/bloom/strcase"
	goTemplate "github.com/meerkat-io/disorder/internal/generator/golang/template"
	"github.com/meerkat-io/disorder/internal/schema"
)

func (g *goGenerator) initTemplete() {
	funcMap := template.FuncMap{
		"PascalCase": func(name string) string {
			return strcase.PascalCase(name)
		},
		"SnakeCase": func(name string) string {
			return strcase.SnakeCase(name)
		},
		"CamelCase": func(name string) string {
			return strcase.CamelCase(name)
		},
		"PackageName": func(pkg string) string {
			names := strings.Split(pkg, ".")
			return strcase.SnakeCase(names[len(names)-1])
		},
		"Type": func(typ *schema.TypeInfo) string {
			return goType(typ)
		},
		"Tag": func(typ *schema.TypeInfo, name string) string {
			omitEmpty := ""
			switch typ.Type {
			case schema.TypeTimestamp, schema.TypeEnum, schema.TypeObject, schema.TypeArray, schema.TypeMap:
				omitEmpty = ",omitempty"
			}
			return fmt.Sprintf("`disorder:\"%s\" json:\"%s%s\"`", name, name, omitEmpty)
		},
	}
	g.define = template.New(fmt.Sprintf("%s_define", golang)).Funcs(funcMap)
	template.Must(g.define.Parse(goTemplate.DefineTemplate))
}
