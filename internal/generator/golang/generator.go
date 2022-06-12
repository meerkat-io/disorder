package golang

import (
	"bytes"
	"fmt"
	"go/format"
	"io/ioutil"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/meerkat-lib/disorder/internal/generator"
	"github.com/meerkat-lib/disorder/internal/schema"
	"github.com/meerkat-lib/disorder/internal/utils/folder"
	"github.com/meerkat-lib/disorder/internal/utils/strcase"
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

func (g *goGenerator) Generate(dir string, files map[string]*schema.File) error {
	for _, file := range files {
		resolvedImports := make(map[string]bool)
		for _, importPath := range file.AbsImports {
			resolvedImports[g.resolveImport(files[importPath])] = true
		}
		file.DefineImports = file.DefineImports[:0]
		file.RpcImports = file.RpcImports[:0]
		if len(file.Enums) > 0 {
			file.DefineImports = append(file.DefineImports, "fmt")
		}
		file.RpcImports = append(file.RpcImports, "fmt")
		for importPath := range resolvedImports {
			file.DefineImports = append(file.DefineImports, importPath)
			file.RpcImports = append(file.RpcImports, importPath)
		}
		file.RpcImports = append(file.RpcImports, "github.com/meerkat-lib/disorder")
		file.RpcImports = append(file.RpcImports, "github.com/meerkat-lib/disorder/rpc")
		file.RpcImports = append(file.RpcImports, "github.com/meerkat-lib/disorder/rpc/code")

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
		err = ioutil.WriteFile(filepath.Join(schemaDir, schemaFile), source, 0666)
		if err != nil {
			return err
		}

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
			err = ioutil.WriteFile(filepath.Join(schemaDir, schemaFile), source, 0666)
			if err != nil {
				return err
			}
		}
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
			switch typ.Type {
			case schema.TypeArray:
				return fmt.Sprintf("[]%s", goType(typ.SubType, typ.TypeRef))
			case schema.TypeMap:
				return fmt.Sprintf("map[string]%s", goType(typ.SubType, typ.TypeRef))
			default:
				return goType(typ.Type, typ.TypeRef)
			}
		},
		"IsPointer": func(typ *schema.TypeInfo) bool {
			switch typ.Type {
			case schema.TypeEnum, schema.TypeTimestamp, schema.TypeObject:
				return true
			default:
				return false
			}
		},
		"Tag": func(name string) string {
			return fmt.Sprintf("`disorder:\"%s\"`", name)
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
	schema.TypeI8:        "int8",
	schema.TypeU8:        "uint8",
	schema.TypeI16:       "int16",
	schema.TypeU16:       "uint16",
	schema.TypeI32:       "int32",
	schema.TypeU32:       "uint32",
	schema.TypeI64:       "int64",
	schema.TypeU64:       "uint64",
	schema.TypeF32:       "float32",
	schema.TypeF64:       "float64",
	schema.TypeString:    "string",
	schema.TypeTimestamp: "int64",
}

func goType(typ schema.Type, ref string) string {
	if typ.IsPrimary() {
		return goTypes[typ]
	}
	if strings.Contains(ref, ".") {
		names := strings.Split(ref, ".")
		pkg := strcase.SnakeCase(names[len(names)-2])
		obj := strcase.PascalCase(names[len(names)-1])
		return fmt.Sprintf("*%s.%s", pkg, obj)
	}
	return fmt.Sprintf("*%s", strcase.PascalCase(ref))
}
