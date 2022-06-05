package loader

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/meerkat-lib/disorder/internal/schema"
)

type Loader interface {
	Load(file string) ([]*schema.File, error)
}

type unmarshaller interface {
	unmarshal([]byte, *proto) error
}

type rpc struct {
	Input  string `yaml:"input" json:"input"`
	Output string `yaml:"output" json:"output"`
}

type proto struct {
	FilePath string `yaml:"-" json:"-"`

	Package  string                       `yaml:"package" json:"package"`
	Imports  []string                     `yaml:"import" json:"import"`
	Enums    map[string][]string          `yaml:"enums" json:"enums"`
	Messages map[string]map[string]string `yaml:"messages" json:"messages"`
	Services map[string]map[string]*rpc   `yaml:"services" json:"services"`
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

func (l *loaderImpl) Load(file string) ([]*schema.File, error) {
	files := map[string]*schema.File{}
	err := l.load(file, files)
	if err != nil {
		return nil, err
	}
	schemas := []*schema.File{}
	for _, file := range files {
		schemas = append(schemas, file)
	}
	err = l.resolver.resolve(schemas)
	if err != nil {
		return nil, err
	}
	return schemas, nil
}

func (l *loaderImpl) load(file string, files map[string]*schema.File) error {
	file, err := filepath.Abs(file)
	if err != nil {
		return fmt.Errorf("yaml file not found: %s", err.Error())
	}
	if _, exists := files[file]; exists {
		return nil
	}
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		return fmt.Errorf("load yaml file [%s] failed: %s", file, err.Error())
	}
	p := &proto{
		FilePath: file,
	}
	err = l.unmarshaller.unmarshal(bytes, p)
	if err != nil {
		return fmt.Errorf("unmarshal yaml file [%s] failed: %s", file, err.Error())
	}
	schema, err := l.parser.parse(p)
	if err != nil {
		return fmt.Errorf("parse schema file [%s] failed: %s", file, err.Error())
	}
	files[file] = schema
	dir := filepath.Dir(file)
	for _, importPath := range p.Imports {
		path := filepath.Join(dir, importPath)
		err = l.load(path, files)
		if err != nil {
			return err
		}
	}
	return nil
}
