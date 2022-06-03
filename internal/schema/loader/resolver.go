package loader

import (
	"fmt"

	"github.com/meerkat-lib/disorder/internal/schema"
)

type resolver struct {
	enums    map[string]string
	messages map[string]string
}

func newResolver() *resolver {
	return &resolver{
		enums:    map[string]string{},
		messages: map[string]string{},
	}
}

func (r *resolver) resolve(files []*schema.File) error {
	for _, file := range files {
		for _, enum := range file.Enums {
			if f, exists := r.enums[enum.Name]; exists {
				return fmt.Errorf("duplicate enum define [%s] in %s and %s", enum.Name, f, file.FilePath)
			}
			r.enums[enum.Name] = file.FilePath
		}
		for _, message := range file.Messages {
			if f, exists := r.messages[message.Name]; exists {
				return fmt.Errorf("duplicate message define [%s] in %s and %s", message.Name, f, file.FilePath)
			}
			r.messages[message.Name] = file.FilePath
		}
		//TO-DO rpc
	}
	for _, file := range files {
		for _, message := range file.Messages {
			for _, field := range message.Fields {
				if err := r.resolveType(field.Type); err != nil {
					return fmt.Errorf("resolve type in file [%s] failed: %s", file.FilePath, err.Error())
				}
			}
		}
		//TO-DO rpc
	}
	return nil
}

func (r *resolver) resolveType(info *schema.TypeInfo) error {
	if info.Type == schema.TypeUndefined {
		if r.isEnum(info.TypeRef) {
			info.Type = schema.TypeEnum
		} else if r.isObject(info.TypeRef) {
			info.Type = schema.TypeObject
		} else {
			return fmt.Errorf("undefine type \"%s\"", info.TypeRef)
		}
	}
	if info.SubTypeRef != "" {
		if r.isEnum(info.SubTypeRef) {
			info.SubType = schema.TypeEnum
		} else if r.isObject(info.SubTypeRef) {
			info.SubType = schema.TypeObject
		} else {
			return fmt.Errorf("undefine type \"%s\"", info.SubTypeRef)
		}
	}
	return nil
}

func (r *resolver) isEnum(typ string) bool {
	_, exists := r.enums[typ]
	return exists
}

func (r *resolver) isObject(typ string) bool {
	_, exists := r.messages[typ]
	return exists
}
