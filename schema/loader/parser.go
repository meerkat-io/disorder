package loader

import (
	"fmt"

	"github.com/meerkat-lib/disorder/schema"
	"github.com/meerkat-lib/disorder/utils"
)

type parser struct {
	validator *validator
}

func newParser() *parser {
	return &parser{
		validator: newValidator(),
	}
}

func (p *parser) parse(file *schemaFile) (*schema.File, error) {
	if file.Package == "" {
		file.Package = schema.PackageGlobal
	}
	if !p.validator.validatePackageName(file.Package) {
		return nil, fmt.Errorf("invalid package name: %s", file.Package)
	}

	schemaFile := &schema.File{
		FilePath: file.FilePath,
	}

	for name, enums := range file.Enums {
		if !p.validator.validateEnumName(name) {
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
			if !p.validator.validateEnumValueName(enum) {
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
