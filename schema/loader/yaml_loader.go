package loader

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"

	"github.com/meerkat-lib/disorder/schema"
	"github.com/meerkat-lib/disorder/utils"

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
	validName *regexp.Regexp
}

func NewYamlLoader() Loader {
	return &yamlLoader{
		validName: regexp.MustCompile(`^[a-zA-Z_][a-zA-Z_0-9]*$`),
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

	file := &yamlFile{
		FilePath: absPath,
	}
	err = yaml.Unmarshal(bytes, &file)
	if err != nil {
		return fmt.Errorf("unmarshal yaml file [%s] failed: %s", absPath, err.Error())
	}

	schema, err := l.parse(file)
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

func (l *yamlLoader) parse(file *yamlFile) (*schema.File, error) {
	if file.Package == "" {
		file.Package = schema.PackageGlobal
	}

	schemaFile := &schema.File{
		FilePath: file.FilePath,
	}

	for name, enums := range file.Enums {
		if !l.validName.MatchString(name) {
			return nil, fmt.Errorf("invalid enum name: %s", name)
		}
		enumMap := map[string]bool{}
		enumDefine := &schema.Enum{
			Name: file.Package + "." + name,
		}
		for _, enum := range enums {
			if _, existing := enumMap[enum]; existing {
				return nil, fmt.Errorf("duplicated enum value [%s]", enum)
			}
			if !l.validName.MatchString(enum) {
				return nil, fmt.Errorf("invalid enum value: %s", enum)
			}
			enumMap[enum] = true
			enumDefine.Values = append(enumDefine.Values, enum)
		}
		schemaFile.Enums = append(schemaFile.Enums, enumDefine)
	}

	utils.PrettyPrint(schemaFile)

	return schemaFile, nil
}

func (l *yamlLoader) parseField(t string) (*schema.Field, error) {
	return nil, nil
}
