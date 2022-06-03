package loader

import (
	"fmt"

	"github.com/meerkat-lib/disorder/internal/schema"
	"github.com/meerkat-lib/disorder/internal/utils"
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

	for name, enumValues := range file.Enums {
		if !p.validator.validateEnumName(name) {
			return nil, fmt.Errorf("invalid enum name: %s", name)
		}
		enumValuesMap := map[string]bool{}
		enumDefine := &schema.Enum{
			Name: file.Package + "." + name,
		}
		for _, enumValue := range enumValues {
			if _, existing := enumValuesMap[enumValue]; existing {
				return nil, fmt.Errorf("duplicated enum value [%s]", enumValue)
			}
			if !p.validator.validateEnumValue(enumValue) {
				return nil, fmt.Errorf("invalid enum value: %s", enumValue)
			}
			enumValuesMap[enumValue] = true
			enumDefine.Values = append(enumDefine.Values, enumValue)
		}
		schemaFile.Enums = append(schemaFile.Enums, enumDefine)
	}

	for name, msgFields := range file.Messages {
		if !p.validator.validateMessageName(name) {
			return nil, fmt.Errorf("invalid message name: %s", name)
		}
		fieldsMap := map[string]bool{}
		msgDefine := &schema.Message{
			Name: file.Package + "." + name,
		}
		for fieldName, fieldType := range msgFields {
			if _, existing := fieldsMap[fieldName]; existing {
				return nil, fmt.Errorf("duplicated field [%s]", fieldName)
			}
			if !p.validator.validateFieldName(fieldName) {
				return nil, fmt.Errorf("invalid field name: %s", fieldName)
			}
			fieldsMap[fieldName] = true
			field, err := p.parseField(fieldName, fieldType)
			if err != nil {
				return nil, fmt.Errorf("invalid field type: %s", err.Error())
			}
			msgDefine.Fields = append(msgDefine.Fields, field)
		}
		schemaFile.Messages = append(schemaFile.Messages, msgDefine)
	}

	utils.PrettyPrint(schemaFile)

	return schemaFile, nil
}

func (p *parser) parseField(name, typ string) (*schema.Field, error) {
	return &schema.Field{
		Name:    name,
		SubType: "sub_type",
	}, nil
}
