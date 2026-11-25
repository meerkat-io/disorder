package loader

import (
	"fmt"

	"github.com/meerkat-io/disorder/internal/schema"
)

type parser struct {
	validator *validator
}

func newParser() *parser {
	return &parser{
		validator: newValidator(),
	}
}

func (p *parser) parse(proto *proto) (*schema.File, error) {
	if proto.Package == "" {
		return nil, fmt.Errorf("package name is required")
	}
	if !p.validator.validatePackageName(proto.Package) {
		return nil, fmt.Errorf("invalid package name: %s", proto.Package)
	}
	file := &schema.File{
		FilePath: proto.FilePath,
		Package:  proto.Package,
		Imports:  proto.Imports,
		Options:  proto.Options,
	}

	for _, e := range proto.Enums {
		if !p.validator.validateEnumName(e.key) {
			return nil, fmt.Errorf("invalid enum name: %s", e.key)
		}
		if e.value == nil {
			continue
		}
		valuesSet := map[string]bool{}
		enum := &schema.Enum{
			Name: e.key,
		}
		if _, ok := e.value.([]interface{}); !ok {
			return nil, fmt.Errorf("expect string list for enum \"%s\"", e.key)
		}
		for _, data := range e.value.([]interface{}) {
			value, ok := data.(string)
			if !ok {
				return nil, fmt.Errorf("expect string value for enum %s", e.key)
			}
			if _, exists := valuesSet[value]; exists {
				return nil, fmt.Errorf("duplicated enum value [%s]", value)
			}
			if !p.validator.validateEnumValue(value) {
				return nil, fmt.Errorf("invalid enum value: %s", value)
			}
			valuesSet[value] = true
			enum.Values = append(enum.Values, value)
		}
		if len(enum.Values) == 0 {
			return nil, fmt.Errorf("empty enum define: %s", e.key)
		}
		file.Enums = append(file.Enums, enum)
	}

	for _, m := range proto.Messages {
		if !p.validator.validateMessageName(m.key) {
			return nil, fmt.Errorf("invalid message name: %s", m.key)
		}
		if m.value == nil {
			continue
		}
		fieldsSet := map[string]bool{}
		message := &schema.Message{
			Name: m.key,
		}
		for _, f := range m.value {
			if f.value == nil {
				continue
			}
			if _, exists := fieldsSet[f.key]; exists {
				return nil, fmt.Errorf("duplicated field [%s]", f.key)
			}
			if !p.validator.validateFieldName(f.key) {
				return nil, fmt.Errorf("invalid field name: %s", f.key)
			}
			fieldsSet[f.key] = true
			if _, ok := f.value.(string); !ok {
				return nil, fmt.Errorf("expect string for field type of \"%s\"", f.key)
			}
			field, err := p.parseField(proto.Package, f.key, f.value.(string))
			if err != nil {
				return nil, err
			}
			message.Fields = append(message.Fields, field)
		}
		if len(message.Fields) == 0 {
			return nil, fmt.Errorf("empty message define: %s", m.key)
		}
		file.Messages = append(file.Messages, message)
	}

	return file, nil
}

func (p *parser) parseField(pkg, name, typ string) (*schema.Field, error) {
	info, err := p.parseType(pkg, typ)
	if err != nil {
		return nil, fmt.Errorf("field [%s] error: %s", name, err.Error())
	}
	return &schema.Field{
		Name: name,
		Type: info,
	}, nil
}

func (p *parser) parseType(pkg, typ string) (t *schema.TypeInfo, err error) {
	if typ == "" {
		err = fmt.Errorf("empty type")
		return
	}
	t = &schema.TypeInfo{}
	if p.validator.isSingularType(typ) {
		if p.validator.isPrimary(typ) {
			t.Type = p.validator.primaryType(typ)
			return
		} else {
			t.TypeRef = typ
			return
		}
	} else if p.validator.isArrayType(typ) {
		t.Type = schema.TypeArray
		elementType := typ[6 : len(typ)-1]
		t.ElementType, err = p.parseType(pkg, elementType)
		return
	} else if p.validator.isMapType(typ) {
		t.Type = schema.TypeMap
		elementType := typ[4 : len(typ)-1]
		t.ElementType, err = p.parseType(pkg, elementType)
		return
	}
	return nil, fmt.Errorf("invalid type %s", typ)
}
