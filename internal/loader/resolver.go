package loader

import (
	"fmt"
	"strings"

	"github.com/meerkat-io/disorder/internal/schema"
)

type resolver struct {
	qualified map[string]string
	enums     map[string]bool
	messages  map[string]bool
}

func newResolver() *resolver {
	return &resolver{
		qualified: map[string]string{},
		enums:     map[string]bool{},
		messages:  map[string]bool{},
	}
}

func (r *resolver) qualifiedName(pkg, name string) string {
	if strings.Contains(name, ".") {
		//remote
		return name
	}
	//local
	return fmt.Sprintf("%s.%s", pkg, name)
}

func (r *resolver) resolve(files map[string]*schema.File) error {
	for _, file := range files {
		for _, enum := range file.Enums {
			if f, exists := r.qualified[enum.Name]; exists {
				return fmt.Errorf("duplicate enum define [%s] in %s and %s", enum.Name, f, file.FilePath)
			}
			qualified := r.qualifiedName(file.Package, enum.Name)
			r.qualified[qualified] = file.FilePath
			r.enums[qualified] = true
		}
		for _, message := range file.Messages {
			if f, exists := r.qualified[message.Name]; exists {
				return fmt.Errorf("duplicate message define [%s] in %s and %s", message.Name, f, file.FilePath)
			}
			qualified := r.qualifiedName(file.Package, message.Name)
			fmt.Println("qualified message: ", qualified)
			fmt.Println("file path: ", file.FilePath)
			r.qualified[qualified] = file.FilePath
			r.messages[qualified] = true
		}
		for _, service := range file.Services {
			if f, exists := r.qualified[service.Name]; exists {
				return fmt.Errorf("duplicate rpc define [%s] in %s and %s", service.Name, f, file.FilePath)
			}
			r.qualified[r.qualifiedName(file.Package, service.Name)] = file.FilePath
		}
	}

	for _, file := range files {
		for _, message := range file.Messages {
			for _, field := range message.Fields {
				if err := r.resolveType(file, field.Type); err != nil {
					return fmt.Errorf("resolve type in file [%s] failed: %s", file.FilePath, err.Error())
				}
			}
		}
		for _, service := range file.Services {
			for _, rpc := range service.Rpc {
				if err := r.resolveType(file, rpc.Input); err != nil {
					return fmt.Errorf("resolve rpc input type in file [%s] failed: %s", file.FilePath, err.Error())
				}
				if err := r.resolveType(file, rpc.Output); err != nil {
					return fmt.Errorf("resolve rpc output type in file [%s] failed: %s", file.FilePath, err.Error())
				}
			}
		}
	}
	return nil
}

func (r *resolver) resolveType(file *schema.File, info *schema.TypeInfo) error {
	if info.Type == schema.TypeUndefined {
		// object or enum
		r.resolveTypeRef(file, info)
		if info.Type == schema.TypeUndefined {
			return fmt.Errorf("undefine type \"%s\"", info.TypeRef)
		}
	} else if info.ElementType != nil {
		// array or map
		return r.resolveType(file, info.ElementType)
	}
	return nil
}

func (r *resolver) resolveTypeRef(file *schema.File, info *schema.TypeInfo) {
	info.Type = schema.TypeUndefined
	qualified := r.qualifiedName(file.Package, info.TypeRef)
	importPath, ok := r.qualified[qualified]
	if !ok {
		return
	}
	if importPath != file.FilePath {
		if _, ok := file.AbsImports[importPath]; !ok {
			return
		}
	}
	if r.isEnum(qualified) {
		info.Qualified = qualified
		info.Type = schema.TypeEnum
	} else if r.isObject(qualified) {
		info.Qualified = qualified
		info.Type = schema.TypeObject
	}
}

func (r *resolver) isEnum(typ string) bool {
	_, exists := r.enums[typ]
	return exists
}

func (r *resolver) isObject(typ string) bool {
	_, exists := r.messages[typ]
	return exists
}
