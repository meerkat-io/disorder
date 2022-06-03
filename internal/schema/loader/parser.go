package loader

import (
	"fmt"

	"github.com/meerkat-lib/disorder/internal/schema"
)

type parser struct {
	validator *validator
}

func newParser() *parser {
	return &parser{
		validator: newValidator(),
	}
}

func (p *parser) qualifiedName(pkg, name string) string {
	return fmt.Sprintf("%s.%s", pkg, name)
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
			Name: p.qualifiedName(file.Package, name),
		}
		for _, enumValue := range enumValues {
			if _, exists := enumValuesMap[enumValue]; exists {
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
			Name: p.qualifiedName(file.Package, name),
		}
		for fieldName, fieldType := range msgFields {
			if _, exists := fieldsMap[fieldName]; exists {
				return nil, fmt.Errorf("duplicated field [%s]", fieldName)
			}
			if !p.validator.validateFieldName(fieldName) {
				return nil, fmt.Errorf("invalid field name: %s", fieldName)
			}
			fieldsMap[fieldName] = true
			field, err := p.parseField(file.Package, fieldName, fieldType)
			if err != nil {
				return nil, err
			}
			msgDefine.Fields = append(msgDefine.Fields, field)
		}
		schemaFile.Messages = append(schemaFile.Messages, msgDefine)
	}

	//TO-DO rpc

	return schemaFile, nil
}

func (p *parser) parseField(pkg, name, typ string) (*schema.Field, error) {
	info, err := p.parseType(pkg, typ)
	if err != nil {
		return nil, err
	}
	return &schema.Field{
		Name: name,
		Type: info,
	}, nil
}

func (p *parser) parseType(pkg, typ string) (t *schema.TypeInfo, err error) {
	t = &schema.TypeInfo{}
	if p.validator.isSimpleType(typ) {
		if p.validator.isPrimary(typ) {
			t.Type = p.validator.primaryType(typ)
			return
		} else {
			t.TypeRef = p.qualifiedName(pkg, typ)
			return
		}
	} else if p.validator.isQualifiedType(typ) {
		t.TypeRef = p.qualifiedName(pkg, typ)
		return
	} else if p.validator.isSimpleArrayType(typ) {
		t.Type = schema.TypeArray
		subType := typ[6 : len(typ)-1]
		if p.validator.isPrimary(subType) {
			t.SubType = p.validator.primaryType(subType)
			return
		} else {
			t.SubTypeRef = p.qualifiedName(pkg, subType)
			return
		}
	} else if p.validator.isQualifiedArrayType(typ) {
		t.Type = schema.TypeArray
		t.SubTypeRef = typ[6 : len(typ)-1]
		return
	} else if p.validator.isSimpleMapType(typ) {
		t.Type = schema.TypeMap
		subType := typ[4 : len(typ)-1]
		if p.validator.isPrimary(subType) {
			t.SubType = p.validator.primaryType(subType)
			return
		} else {
			t.SubTypeRef = p.qualifiedName(pkg, subType)
			return
		}
	} else if p.validator.isQualifiedMapType(typ) {
		t.Type = schema.TypeMap
		t.SubTypeRef = typ[4 : len(typ)-1]
		return
	}
	return nil, fmt.Errorf("invalid type %s", typ)
}
