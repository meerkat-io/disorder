package golang

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/meerkat-io/bloom/folder"
	"github.com/meerkat-io/bloom/strcase"
	"github.com/meerkat-io/disorder/internal/generator"
	"github.com/meerkat-io/disorder/internal/schema"
)

const goPackagePrefix = "go_package_prefix"

// NewGoGenerator returns a generator that writes Go types, one file per schema.
func NewGoGenerator() generator.Generator {
	return &goGenerator{define: newTemplate()}
}

type goGenerator struct {
	define *template.Template
}

type goSchema struct {
	Schema  *schema.File
	Imports []string
}

func (g *goGenerator) Generate(dir string, files map[string]*schema.File, qualifiedPath map[string]string) error {
	for _, file := range files {
		if err := g.generateFile(dir, file, files, qualifiedPath); err != nil {
			return fmt.Errorf("generate %s failed: %w", file.FilePath, err)
		}
	}
	return nil
}

func (g *goGenerator) generateFile(dir string, file *schema.File, files map[string]*schema.File, qualifiedPath map[string]string) error {
	data := &goSchema{
		Schema:  file,
		Imports: collectImports(file, files, qualifiedPath),
	}
	buf := &bytes.Buffer{}
	if err := g.define.Execute(buf, data); err != nil {
		return err
	}
	source, err := format.Source(buf.Bytes())
	if err != nil {
		return err
	}

	schemaDir, err := filepath.Abs(filepath.Join(dir, packageFolder(file.Package)))
	if err != nil {
		return err
	}
	if err := folder.Create(schemaDir); err != nil {
		return err
	}
	name := strings.TrimSuffix(filepath.Base(file.FilePath), filepath.Ext(file.FilePath)) + ".go"
	return os.WriteFile(filepath.Join(schemaDir, name), source, 0666)
}

// collectImports returns the sorted import paths needed by the generated file.
func collectImports(file *schema.File, files map[string]*schema.File, qualifiedPath map[string]string) []string {
	imports := map[string]bool{}
	for _, message := range file.Messages {
		for _, field := range message.Fields {
			addImports(field.Type, imports, file, files, qualifiedPath)
		}
	}
	if len(file.Enums) > 0 {
		imports["fmt"] = true
		imports["encoding/json"] = true
	}
	paths := make([]string, 0, len(imports))
	for path := range imports {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return paths
}

func addImports(typeInfo *schema.TypeInfo, imports map[string]bool, current *schema.File, files map[string]*schema.File, qualifiedPath map[string]string) {
	if typeInfo.Type == schema.TypeTimestamp {
		imports["time"] = true
	}
	if typeInfo.ElementType != nil {
		addImports(typeInfo.ElementType, imports, current, files, qualifiedPath)
	}
	if typeInfo.Qualified != "" && current.FilePath != qualifiedPath[typeInfo.Qualified] {
		target := files[qualifiedPath[typeInfo.Qualified]]
		imports[filepath.Join(target.Options[goPackagePrefix], packageFolder(target.Package))] = true
	}
}

func packageFolder(pkg string) string {
	folders := strings.Split(pkg, ".")
	for i := range folders {
		folders[i] = strcase.SnakeCase(folders[i])
	}
	return strings.Join(folders, "/")
}
