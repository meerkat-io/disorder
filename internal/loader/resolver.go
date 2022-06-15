package loader

import (
	"fmt"
	"strings"

	"github.com/meerkat-io/disorder/internal/schema"
)

type resolver struct {
	names    map[string]string
	enums    map[string]bool
	messages map[string]bool
}

func newResolver() *resolver {
	return &resolver{
		names:    map[string]string{},
		enums:    map[string]bool{},
		messages: map[string]bool{},
	}
}

func (r *resolver) qualifiedName(pkg, name string) string {
	if strings.Contains(name, ".") {
		return name
	}
	return fmt.Sprintf("%s.%s", pkg, name)
}

func (r *resolver) resolve(files map[string]*schema.File) error {
	for _, file := range files {
		for _, enum := range file.Enums {
			if f, exists := r.names[enum.Name]; exists {
				return fmt.Errorf("duplicate enum define [%s] in %s and %s", enum.Name, f, file.FilePath)
			}
			qualified := r.qualifiedName(file.Package, enum.Name)
			r.names[qualified] = file.FilePath
			r.enums[qualified] = true
		}
		for _, message := range file.Messages {
			if f, exists := r.names[message.Name]; exists {
				return fmt.Errorf("duplicate message define [%s] in %s and %s", message.Name, f, file.FilePath)
			}
			qualified := r.qualifiedName(file.Package, message.Name)
			r.names[qualified] = file.FilePath
			r.messages[qualified] = true
		}
		for _, service := range file.Services {
			if f, exists := r.names[service.Name]; exists {
				return fmt.Errorf("duplicate rpc define [%s] in %s and %s", service.Name, f, file.FilePath)
			}
			r.names[r.qualifiedName(file.Package, service.Name)] = file.FilePath
		}
	}

	for _, file := range files {
		for _, message := range file.Messages {
			for _, field := range message.Fields {
				if err := r.resolveType(file.Package, field.Type); err != nil {
					return fmt.Errorf("resolve type in file [%s] failed: %s", file.FilePath, err.Error())
				}
				if field.Type.Type == schema.TypeTimestamp {
					file.HasTimestampDefine = true
				}
			}
		}
		for _, service := range file.Services {
			for _, rpc := range service.Rpc {
				if err := r.resolveType(file.Package, rpc.Input); err != nil {
					return fmt.Errorf("resolve rpc input type in file [%s] failed: %s", file.FilePath, err.Error())
				}
				if err := r.resolveType(file.Package, rpc.Output); err != nil {
					return fmt.Errorf("resolve rpc output type in file [%s] failed: %s", file.FilePath, err.Error())
				}
				if rpc.Input.Type == schema.TypeTimestamp {
					file.HasTimestampRpc = true
				}
				if rpc.Output.Type == schema.TypeTimestamp {
					file.HasTimestampRpc = true
				}
			}
		}
	}
	return nil
}

func (r *resolver) resolveType(pkg string, info *schema.TypeInfo) error {
	if info.Type == schema.TypeUndefined {
		qualified := r.qualifiedName(pkg, info.TypeRef)
		if r.isEnum(qualified) {
			info.Type = schema.TypeEnum
		} else if r.isObject(qualified) {
			info.Type = schema.TypeObject
		} else {
			return fmt.Errorf("undefine type \"%s\"", info.TypeRef)
		}
	} else if info.TypeRef != "" {
		qualified := r.qualifiedName(pkg, info.TypeRef)
		if r.isEnum(qualified) {
			info.SubType = schema.TypeEnum
		} else if r.isObject(qualified) {
			info.SubType = schema.TypeObject
		} else {
			return fmt.Errorf("undefine type \"%s\"", info.TypeRef)
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
