package loader

import (
	"fmt"

	"github.com/meerkat-io/disorder/internal/schema"
)

var (
	undefined = &schema.TypeInfo{}
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

	for name, values := range proto.Enums {
		if !p.validator.validateEnumName(name) {
			return nil, fmt.Errorf("invalid enum name: %s", name)
		}
		if values == nil {
			continue
		}
		valuesSet := map[string]bool{}
		enum := &schema.Enum{
			Name: name,
		}
		for _, value := range values {
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
			return nil, fmt.Errorf("empty enum define: %s", name)
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
			if _, exists := fieldsSet[f.key]; exists {
				return nil, fmt.Errorf("duplicated field [%s]", f.key)
			}
			if !p.validator.validateFieldName(f.key) {
				return nil, fmt.Errorf("invalid field name: %s", f.key)
			}
			fieldsSet[f.key] = true
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

	for name, rpcs := range proto.Services {
		if !p.validator.validateServiceName(name) {
			return nil, fmt.Errorf("invalid service name: %s", name)
		}
		if rpcs == nil {
			continue
		}
		rpcsSet := map[string]bool{}
		service := &schema.Service{
			Name: name,
		}
		for rpcName, rpcDefine := range rpcs {
			if rpcDefine == nil {
				continue
			}
			if _, exists := rpcsSet[rpcName]; exists {
				return nil, fmt.Errorf("duplicated rpc [%s]", rpcName)
			}
			if !p.validator.validateRpcName(rpcName) {
				return nil, fmt.Errorf("invalid rpc name: %s", rpcName)
			}
			rpcsSet[rpcName] = true
			rpc, err := p.parseRpc(proto.Package, rpcName, rpcDefine)
			if err != nil {
				return nil, err
			}
			service.Rpc = append(service.Rpc, rpc)
		}
		if len(service.Rpc) == 0 {
			return nil, fmt.Errorf("empty service define: %s", name)
		}
		file.Services = append(file.Services, service)
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

func (p *parser) parseRpc(pkg, name string, rpc *rpc) (*schema.Rpc, error) {
	var err error
	r := &schema.Rpc{
		Name: name,
	}
	if rpc.Input == "" {
		r.Input = undefined
	} else {
		r.Input, err = p.parseType(pkg, rpc.Input)
		if err != nil {
			return nil, fmt.Errorf("rpc [%s] input type error: %s", name, err.Error())
		}
	}
	if rpc.Output == "" {
		r.Output = undefined
	} else {
		r.Output, err = p.parseType(pkg, rpc.Output)
		if err != nil {
			return nil, fmt.Errorf("rpc [%s] output type error: %s", name, err.Error())
		}
	}
	return r, nil
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
