package golang

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/meerkat-io/bloom/format/strcase"
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
		"IsPointer": func(typ *schema.TypeInfo) bool {
			switch typ.Type {
			case schema.TypeEnum, schema.TypeTimestamp, schema.TypeObject:
				return true
			default:
				return false
			}
		},
		"InitType": func(typ *schema.TypeInfo) string {
			switch typ.Type {
			case schema.TypeEnum:
				return fmt.Sprintf(" = new(%s)", goType(typ)[1:])
			case schema.TypeTimestamp:
				return " = &time.Time{}"
			case schema.TypeObject:
				return fmt.Sprintf(" = &%s{}", goType(typ)[1:])
			case schema.TypeMap:
				return fmt.Sprintf(" = make(map[string]%s)", goType(typ.ElementType))
			default:
				return ""
			}
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
	g.rpc = template.New(fmt.Sprintf("%s_rpc", golang)).Funcs(funcMap)
	template.Must(g.rpc.Parse(goTemplate.RpcTemplate))
}
