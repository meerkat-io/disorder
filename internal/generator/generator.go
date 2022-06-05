package generator

import (
	"fmt"
	"path/filepath"
	"text/template"

	"github.com/meerkat-lib/disorder/internal/schema"
	"github.com/meerkat-lib/disorder/internal/utils/strcase"
)

type Generator interface {
	Generate(folder string, files []*schema.File) error
}

type generatorImpl struct {
	language language
}

func newGeneratorImpl(language language) Generator {
	return &generatorImpl{
		language: language,
	}
}

type language interface {
	getTemplate() *template.Template
	packageFolder(packageName string) string
}

func caseConversion() template.FuncMap {
	return template.FuncMap{
		"PascalCase": func(name string) string {
			return strcase.PascalCase(name)
		},
		"CamelCase": func(name string) string {
			return strcase.CamelCase(name)
		},
		"SnakeCase": func(name string) string {
			return strcase.SnakeCase(name)
		},
	}
}

func (g *generatorImpl) Generate(folder string, files []*schema.File) error {
	for _, file := range files {
		fmt.Println(filepath.Join(folder, g.language.packageFolder(file.Package)))
	}
	return nil
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
