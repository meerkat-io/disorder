package loader

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/meerkat-lib/disorder/schema"

	"gopkg.in/yaml.v3"
)

type yamlLoader struct {
	parser *parser
}

func NewYamlLoader() Loader {
	return &yamlLoader{
		parser: newParser(),
	}
}

func (l *yamlLoader) Load(filePath string) ([]*schema.File, error) {
	files := map[string]*schema.File{}
	err := l.load(filePath, files)
	if err != nil {
		return nil, err
	}

	result := []*schema.File{}
	for _, file := range files {
		result = append(result, file)
	}

	return result, nil
}

func (l *yamlLoader) load(filePath string, files map[string]*schema.File) error {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("yaml file not found: %s", err.Error())
	}

	if _, existing := files[absPath]; existing {
		return nil
	}

	bytes, err := ioutil.ReadFile(absPath)
	if err != nil {
		return fmt.Errorf("load yaml file [%s] failed: %s", absPath, err.Error())
	}

	file := &schemaFile{
		FilePath: absPath,
	}
	err = yaml.Unmarshal(bytes, &file)
	if err != nil {
		return fmt.Errorf("unmarshal yaml file [%s] failed: %s", absPath, err.Error())
	}

	schema, err := l.parser.parse(file)
	if err != nil {
		return fmt.Errorf("parse protocol file [%s] failed: %s", absPath, err.Error())
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
