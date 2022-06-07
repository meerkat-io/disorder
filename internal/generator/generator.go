package generator

import (
	"bytes"
	"fmt"
	"path/filepath"
	"text/template"

	"github.com/meerkat-lib/disorder/internal/schema"
)

type Generator interface {
	Generate(folder string, files map[string]*schema.File) error
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
	template() *template.Template
	folder(pkg string) string
}

func (g *generatorImpl) Generate(folder string, files map[string]*schema.File) error {
	for _, file := range files {
		path, err := filepath.Abs(filepath.Join(folder, g.language.folder(file.Package)))
		if err != nil {
			return err
		}
		fmt.Println(path)
		template := g.language.template()
		buf := &bytes.Buffer{}
		if err := template.Execute(buf, file); err != nil {
			return err
		}
		fmt.Println(buf.String())
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
