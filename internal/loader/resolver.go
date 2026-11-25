package loader

import (
	"fmt"
	"strings"

	"github.com/meerkat-io/disorder/internal/schema"
)

// resolver links type references to the enums and messages they name, and
// records which file defines each qualified name.
type resolver struct {
	qualified map[string]string // qualified name -> defining file path
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

// qualifiedName returns name unchanged if it already carries a package,
// otherwise prefixes it with pkg.
func qualifiedName(pkg, name string) string {
	if strings.Contains(name, ".") {
		return name
	}
	return pkg + "." + name
}

func (r *resolver) resolve(files map[string]*schema.File) error {
	for _, file := range files {
		for _, enum := range file.Enums {
			if err := r.define("enum", enum.Name, file, r.enums); err != nil {
				return err
			}
		}
		for _, message := range file.Messages {
			if err := r.define("message", message.Name, file, r.messages); err != nil {
				return err
			}
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
	}
	return nil
}

// define registers a name in kinds (the enum or message set) and rejects
// names already defined in the same package.
func (r *resolver) define(kind, name string, file *schema.File, kinds map[string]bool) error {
	qualified := qualifiedName(file.Package, name)
	if other, exists := r.qualified[qualified]; exists {
		return fmt.Errorf("duplicate %s define [%s] in %s and %s", kind, name, other, file.FilePath)
	}
	r.qualified[qualified] = file.FilePath
	kinds[qualified] = true
	return nil
}

func (r *resolver) resolveType(file *schema.File, info *schema.TypeInfo) error {
	switch {
	case info.Type == schema.TypeUndefined:
		if !r.resolveRef(file, info) {
			return fmt.Errorf("undefine type \"%s\"", info.TypeRef)
		}
	case info.ElementType != nil:
		return r.resolveType(file, info.ElementType)
	}
	return nil
}

// resolveRef sets the type of a message or enum reference. The target must be
// defined in the same file or in a file it imports.
func (r *resolver) resolveRef(file *schema.File, info *schema.TypeInfo) bool {
	qualified := qualifiedName(file.Package, info.TypeRef)
	path, ok := r.qualified[qualified]
	if !ok || (path != file.FilePath && !file.AbsImports[path]) {
		return false
	}
	switch {
	case r.enums[qualified]:
		info.Type = schema.TypeEnum
	case r.messages[qualified]:
		info.Type = schema.TypeObject
	default:
		return false
	}
	info.Qualified = qualified
	return true
}
