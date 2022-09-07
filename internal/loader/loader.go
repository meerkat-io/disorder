package loader

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/meerkat-io/disorder/internal/schema"
)

type Loader interface {
	Load(file string) (map[string]*schema.File, error)
}

type unmarshaller interface {
	unmarshal([]byte, *proto) error
}

type rpc struct {
	Input  string `yaml:"input" json:"input" toml:"input"`
	Output string `yaml:"output" json:"output" toml:"output"`
}

type proto struct {
	FilePath string            `yaml:"-" json:"-" toml:"-"`
	Package  string            `yaml:"package" json:"package" toml:"package"`
	Imports  []string          `yaml:"import" json:"import" toml:"import"`
	Options  map[string]string `yaml:"option" json:"option" toml:"option"`

	Enums    map[string][]string          `yaml:"enums" json:"enums" toml:"enums"`
	Messages map[string]map[string]string `yaml:"messages" json:"messages" toml:"messages"`
	Services map[string]map[string]*rpc   `yaml:"services" json:"services" toml:"services"`
}

type loaderImpl struct {
	parser       *parser
	unmarshaller unmarshaller
	resolver     *resolver
}

func newLoaderImpl(unmarshaller unmarshaller) Loader {
	return &loaderImpl{
		parser:       newParser(),
		unmarshaller: unmarshaller,
		resolver:     newResolver(),
	}
}

func (l *loaderImpl) Load(file string) (map[string]*schema.File, error) {
	files := map[string]*schema.File{}
	err := l.load(file, files)
	if err != nil {
		return nil, err
	}
	err = l.resolver.resolve(files)
	if err != nil {
		return nil, err
	}
	return files, nil
}

func (l *loaderImpl) load(file string, files map[string]*schema.File) error {
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
	err = l.unmarshaller.unmarshal(bytes, p)
	if err != nil {
		return fmt.Errorf("unmarshal schema file [%s] failed: %s", file, err.Error())
	}

	schema, err := l.parser.parse(p)
	if err != nil {
		return fmt.Errorf("parse schema file [%s] failed: %s", file, err.Error())
	}
	files[file] = schema

	dir := filepath.Dir(file)
	for _, importPath := range p.Imports {
		path := filepath.Join(dir, importPath)
		schema.AbsImports = append(schema.AbsImports, path)
		err = l.load(path, files)
		if err != nil {
			return err
		}
	}
	return nil
}
