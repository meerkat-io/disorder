package loader

import (
	"fmt"
	"strings"

	"github.com/meerkat-io/disorder/internal/schema"
)

// parse converts a decoded yaml document into a schema file, validating names
// and types. References to other messages and enums are resolved later.
func parse(p *proto) (*schema.File, error) {
	if p.Package == "" {
		return nil, fmt.Errorf("package name is required")
	}
	if !validPackageName(p.Package) {
		return nil, fmt.Errorf("invalid package name: %s", p.Package)
	}
	file := &schema.File{
		FilePath: p.FilePath,
		Package:  p.Package,
		Imports:  p.Imports,
		Options:  p.Options,
	}

	for _, item := range p.Enums {
		if !validName(item.key) {
			return nil, fmt.Errorf("invalid enum name: %s", item.key)
		}
		if item.value == nil {
			continue
		}
		enum, err := parseEnum(item)
		if err != nil {
			return nil, err
		}
		file.Enums = append(file.Enums, enum)
	}

	for _, item := range p.Messages {
		if !validName(item.key) {
			return nil, fmt.Errorf("invalid message name: %s", item.key)
		}
		if item.value == nil {
			continue
		}
		message, err := parseMessage(item)
		if err != nil {
			return nil, err
		}
		file.Messages = append(file.Messages, message)
	}
	return file, nil
}

func parseEnum(item mapItem) (*schema.Enum, error) {
	list, ok := item.value.([]interface{})
	if !ok {
		return nil, fmt.Errorf("expect string list for enum \"%s\"", item.key)
	}
	enum := &schema.Enum{Name: item.key}
	seen := map[string]bool{}
	for _, data := range list {
		value, ok := data.(string)
		if !ok {
			return nil, fmt.Errorf("expect string value for enum %s", item.key)
		}
		if seen[value] {
			return nil, fmt.Errorf("duplicated enum value [%s]", value)
		}
		if !validName(value) {
			return nil, fmt.Errorf("invalid enum value: %s", value)
		}
		seen[value] = true
		enum.Values = append(enum.Values, value)
	}
	if len(enum.Values) == 0 {
		return nil, fmt.Errorf("empty enum define: %s", item.key)
	}
	return enum, nil
}

func parseMessage(item mapSliceItem) (*schema.Message, error) {
	message := &schema.Message{Name: item.key}
	seen := map[string]bool{}
	for _, f := range item.value {
		if f.value == nil {
			continue
		}
		if seen[f.key] {
			return nil, fmt.Errorf("duplicated field [%s]", f.key)
		}
		if !validName(f.key) {
			return nil, fmt.Errorf("invalid field name: %s", f.key)
		}
		seen[f.key] = true
		typ, ok := f.value.(string)
		if !ok {
			return nil, fmt.Errorf("expect string for field type of \"%s\"", f.key)
		}
		info, err := parseType(typ)
		if err != nil {
			return nil, fmt.Errorf("field [%s] error: %s", f.key, err.Error())
		}
		message.Fields = append(message.Fields, &schema.Field{Name: f.key, Type: info})
	}
	if len(message.Fields) == 0 {
		return nil, fmt.Errorf("empty message define: %s", item.key)
	}
	return message, nil
}

// parseType parses a field type: a primary type, a reference to a message or
// enum, or array[T] / map[T] of another type.
func parseType(typ string) (*schema.TypeInfo, error) {
	switch {
	case typ == "":
		return nil, fmt.Errorf("empty type")
	case dotted.MatchString(typ):
		if primary, ok := schema.PrimaryTypes[typ]; ok {
			return &schema.TypeInfo{Type: primary}, nil
		}
		return &schema.TypeInfo{TypeRef: typ}, nil
	case isArrayType(typ):
		return parseContainer(schema.TypeArray, strings.TrimSuffix(strings.TrimPrefix(typ, "array["), "]"))
	case isMapType(typ):
		return parseContainer(schema.TypeMap, strings.TrimSuffix(strings.TrimPrefix(typ, "map["), "]"))
	}
	return nil, fmt.Errorf("invalid type %s", typ)
}

func parseContainer(container schema.Type, element string) (*schema.TypeInfo, error) {
	elementType, err := parseType(element)
	if err != nil {
		return nil, err
	}
	return &schema.TypeInfo{Type: container, ElementType: elementType}, nil
}
