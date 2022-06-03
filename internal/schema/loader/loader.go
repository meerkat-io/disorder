package loader

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/meerkat-lib/disorder/internal/schema"
	"github.com/meerkat-lib/disorder/internal/utils"
)

type Loader interface {
	Load(file string) ([]*schema.File, error)
}

type unmarshaller interface {
	Unmarshal([]byte, *schemaFile) error
}

type rpc struct {
	Input  string `yaml:"input" json:"input"`
	Output string `yaml:"output" json:"output"`
}

type schemaFile struct {
	FilePath string `yaml:"-" json:"-"`

	Package  string                       `yaml:"package" json:"package"`
	Imports  []string                     `yaml:"import" json:"import"`
	Enums    map[string][]string          `yaml:"enums" json:"enums"`
	Messages map[string]map[string]string `yaml:"messages" json:"messages"`
	Services map[string]*rpc              `yaml:"services" json:"services"`
}

type loaderImpl struct {
	parser       *parser
	unmarshaller unmarshaller
}

func newLoaderImpl(unmarshaller unmarshaller) Loader {
	return &loaderImpl{
		parser:       newParser(),
		unmarshaller: unmarshaller,
	}
}

func (l *loaderImpl) Load(filePath string) ([]*schema.File, error) {
	files := map[string]*schema.File{}
	err := l.load(filePath, files)
	if err != nil {
		return nil, err
	}

	result := []*schema.File{}
	for _, file := range files {
		result = append(result, file)
	}

	r := newResolver()
	err = r.resolve(result)
	if err != nil {
		return nil, err
	}

	for _, f := range result {
		utils.PrettyPrint(f)
	}

	return result, nil
}

func (l *loaderImpl) load(filePath string, files map[string]*schema.File) error {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("yaml file not found: %s", err.Error())
	}

	if _, exists := files[absPath]; exists {
		return nil
	}

	bytes, err := ioutil.ReadFile(absPath)
	if err != nil {
		return fmt.Errorf("load yaml file [%s] failed: %s", absPath, err.Error())
	}

	file := &schemaFile{
		FilePath: absPath,
	}
	err = l.unmarshaller.Unmarshal(bytes, file)
	if err != nil {
		return fmt.Errorf("unmarshal yaml file [%s] failed: %s", absPath, err.Error())
	}

	schema, err := l.parser.parse(file)
	if err != nil {
		return fmt.Errorf("parse schema file [%s] failed: %s", absPath, err.Error())
	}

	files[absPath] = schema
	absDir := filepath.Dir(absPath)
	for _, importPath := range file.Imports {
		path := filepath.Join(absDir, importPath)
		err = l.load(path, files)
		if err != nil {
			return err
		}
	}

	return nil
}
