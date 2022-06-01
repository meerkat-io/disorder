package loader

import (
	"io/ioutil"

	"github.com/meerkat-lib/disorder/schema"

	"gopkg.in/yaml.v3"
)

type yamlRpc struct {
	Input  string `yaml:"input"`
	Output string `yaml:"output"`
}

type yamlFile struct {
	FilePath string `yaml:"-"`

	Package  string                       `yaml:"package"`
	Imports  []string                     `yaml:"import"`
	Enums    map[string][]string          `yaml:"enums"`
	Messages map[string]map[string]string `yaml:"messages"`
	Services map[string]*yamlRpc          `yaml:"services"`
}

type yamlLoader struct {
}

func NewYamlLoader() Loader {
	return &yamlLoader{}
}

func (l *yamlLoader) Load(filePath string) ([]*schema.SchemaFile, error) {
	bytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	file := &yamlFile{
		FilePath: filePath,
	}
	err = yaml.Unmarshal(bytes, &file)
	if err != nil {
		return nil, err
	}

	schemaFile, err := l.parse(file)
	if err != nil {
		return nil, err
	}

	schemaFiles := []*schema.SchemaFile{schemaFile}
	for _, importPath := range file.Imports {
		//TO-DO resolve file path
		subSchemaFiles, err := l.Load(importPath)
		if err != nil {
			return nil, err
		}
		schemaFiles = append(schemaFiles, subSchemaFiles...)
	}

	return schemaFiles, nil
}

func (l *yamlLoader) parse(file *yamlFile) (*schema.SchemaFile, error) {
	if file.Package == "" {
		file.Package = schema.PackageGlobal
	}

	schema := &schema.SchemaFile{
		FilePath: file.FilePath,
	}

	return schema, nil
}
