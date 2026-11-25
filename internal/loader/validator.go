package loader

import (
	"regexp"
	"strings"

	"github.com/meerkat-io/disorder/internal/schema"
)

var (
	identifier = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z_0-9]*$`)
	// dotted is a package name or a (possibly package-qualified) type name.
	dotted = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z_0-9]*(\.[a-zA-Z_][a-zA-Z_0-9]*)*$`)
)

// validName reports whether name is usable as an enum, enum value, message or
// field name: a plain identifier that is not a primary type name.
func validName(name string) bool {
	_, primary := schema.PrimaryTypes[name]
	return !primary && identifier.MatchString(name)
}

func validPackageName(pkg string) bool {
	return dotted.MatchString(pkg)
}

func isArrayType(typ string) bool {
	return strings.HasPrefix(typ, "array[") && strings.HasSuffix(typ, "]")
}

func isMapType(typ string) bool {
	return strings.HasPrefix(typ, "map[") && strings.HasSuffix(typ, "]")
}
