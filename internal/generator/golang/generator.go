package golang

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/meerkat-io/bloom/folder"
	"github.com/meerkat-io/bloom/format/strcase"
	"github.com/meerkat-io/disorder/internal/generator"
	"github.com/meerkat-io/disorder/internal/schema"
)

const (
	golang          = "golang"
	goPackagePrefix = "go_package_prefix"
)

func NewGoGenerator() generator.Generator {
	g := &goGenerator{}
	g.initTemplete()
	return g
}

type goGenerator struct {
	define *template.Template
	rpc    *template.Template
}

func (g *goGenerator) Generate(dir string, files map[string]*schema.File, qualifiedPath map[string]string) error {
	for _, file := range files {
		defineImports := make(map[string]bool)
		rpcImports := make(map[string]bool)
		for _, message := range file.Messages {
			for _, field := range message.Fields {
				if field.Type.Type == schema.TypeTimestamp {
					defineImports["time"] = true
				}
				//TO-DO add qualified
			}
		}
		for _, service := range file.Services {
			for _, rpc := range service.Rpc {
				if rpc.Input.Type == schema.TypeTimestamp || rpc.Output.Type == schema.TypeTimestamp {
					rpcImports["time"] = true
				}
				//TO-DO add qualified
			}
		}

		if len(file.Enums) > 0 {
			defineImports["fmt"] = true
		}
		rpcImports["fmt"] = true
		rpcImports["github.com/meerkat-io/disorder"] = true
		rpcImports["github.com/meerkat-io/disorder/rpc"] = true
		rpcImports["github.com/meerkat-io/disorder/rpc/code"] = true

		schemaDir, err := filepath.Abs(filepath.Join(dir, g.packageFolder(file.Package)))
		if err != nil {
			return err
		}
		err = folder.Create(schemaDir)
		if err != nil {
			return err
		}

		buf := &bytes.Buffer{}
		if err := g.define.Execute(buf, file); err != nil {
			return err
		}
		source := buf.Bytes()
		if source, err = format.Source(source); err != nil {
			return err
		}
		schemaFile := filepath.Base(file.FilePath)
		schemaFile = fmt.Sprintf("%s.go", strings.TrimSuffix(schemaFile, filepath.Ext(schemaFile)))
		err = os.WriteFile(filepath.Join(schemaDir, schemaFile), source, 0666)
		if err != nil {
			return err
		}
		/*
			if len(file.Services) > 0 {
				buf = &bytes.Buffer{}
				if err := g.rpc.Execute(buf, file); err != nil {
					return err
				}
				source = buf.Bytes()
				if source, err = format.Source(source); err != nil {
					return err
				}
				schemaFile = filepath.Base(file.FilePath)
				schemaFile = fmt.Sprintf("%s_rpc.go", strings.TrimSuffix(schemaFile, filepath.Ext(schemaFile)))
				err = os.WriteFile(filepath.Join(schemaDir, schemaFile), source, 0666)
				if err != nil {
					return err
				}
			}*/
	}
	return nil
}

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
			//TO-DO support "omitnil" in the future
			omitEmpty := ""
			switch typ.Type {
			case schema.TypeTimestamp, schema.TypeEnum, schema.TypeObject, schema.TypeArray, schema.TypeMap:
				omitEmpty = ",omitempty"
			}
			return fmt.Sprintf("`disorder:\"%s\" json:\"%s%s\"`", name, name, omitEmpty)
		},
	}
	g.define = template.New(fmt.Sprintf("%s_define", golang)).Funcs(funcMap)
	template.Must(g.define.Parse(defineTemplate))
	g.rpc = template.New(fmt.Sprintf("%s_rpc", golang)).Funcs(funcMap)
	template.Must(g.rpc.Parse(rpcTemplate))
}

func (g *goGenerator) packageFolder(pkg string) string {
	folders := strings.Split(pkg, ".")
	for i := range folders {
		folders[i] = strcase.SnakeCase(folders[i])
	}
	return strings.Join(folders, "/")
}

func (g *goGenerator) resolveImport(file *schema.File) string {
	prefix := ""
	if file.Options != nil {
		prefix = file.Options[goPackagePrefix]
	}
	return filepath.Join(prefix, g.packageFolder(file.Package))
}

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
