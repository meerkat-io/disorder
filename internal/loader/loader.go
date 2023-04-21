package loader

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/meerkat-io/disorder/internal/schema"
	"gopkg.in/yaml.v3"
)

type Loader interface {
	Load(file string) (map[string]*schema.File, map[string]string, error)
}

func NewLoader() Loader {
	return &loader{
		parser:   newParser(),
		resolver: newResolver(),
	}
}

type rpc struct {
	Input  string `yaml:"input" json:"input" toml:"input"`
	Output string `yaml:"output" json:"output" toml:"output"`
}

type proto struct {
	FilePath string            `yaml:"-" json:"-" toml:"-"`
	Schema   string            `yaml:"schema" json:"schema" toml:"schema"`
	Version  string            `yaml:"version" json:"version" toml:"version"`
	Package  string            `yaml:"package" json:"package" toml:"package"`
	Imports  []string          `yaml:"import" json:"import" toml:"import"`
	Options  map[string]string `yaml:"option" json:"option" toml:"option"`

	Enums    map[string][]string        `yaml:"enums" json:"enums" toml:"enums"`
	Messages mapSlice                   `yaml:"messages" json:"messages" toml:"messages"`
	Services map[string]map[string]*rpc `yaml:"services" json:"services" toml:"services"`
}

type loader struct {
	parser   *parser
	resolver *resolver
}

func (l *loader) Load(file string) (map[string]*schema.File, map[string]string, error) {
	files := map[string]*schema.File{}
	err := l.load(file, files)
	if err != nil {
		return nil, nil, err
	}
	err = l.resolver.resolve(files)
	if err != nil {
		return nil, nil, err
	}
	return files, l.resolver.qualified, nil
}

func (l *loader) load(file string, files map[string]*schema.File) error {
	file, err := filepath.Abs(file)
	if err != nil {
		return fmt.Errorf("schema file not found: %s", err.Error())
	}
	if _, exists := files[file]; exists {
		return nil
	}
	bytes, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("load schema file [%s] failed: %s", file, err.Error())
	}

	p := &proto{
		FilePath: file,
	}
	err = yaml.Unmarshal(bytes, p)
	if err != nil {
		return fmt.Errorf("unmarshal schema file [%s] failed: %s", file, err.Error())
	}
	if p.Schema != "disorder" {
		return fmt.Errorf("invalid disorder schema file")
	}
	//TO-DO support version in the future

	schemaFile, err := l.parser.parse(p)
	if err != nil {
		return fmt.Errorf("parse schema file [%s] failed: %s", file, err.Error())
	}
	files[file] = schemaFile

	dir := filepath.Dir(file)
	schemaFile.AbsImports = map[string]bool{}
	for _, importPath := range p.Imports {
		path := filepath.Join(dir, importPath)
		schemaFile.AbsImports[path] = true
		err = l.load(path, files)
		if err != nil {
			return err
		}
	}
	return nil
}
