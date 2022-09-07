package loader

import (
	"regexp"
	"strings"

	"github.com/meerkat-io/disorder/internal/schema"
)

type validator struct {
	variableName *regexp.Regexp
	packageName  *regexp.Regexp
	singularType *regexp.Regexp
}

func newValidator() *validator {
	return &validator{
		variableName: regexp.MustCompile(`^[a-zA-Z_][a-zA-Z_0-9]*$`),
		packageName:  regexp.MustCompile(`^[a-zA-Z_][a-zA-Z_0-9]*(.[a-zA-Z_][a-zA-Z_0-9]*)*$`),
		singularType: regexp.MustCompile(`^[a-zA-Z_][a-zA-Z_0-9]*(.[a-zA-Z_][a-zA-Z_0-9]*)*$`),
	}
}

func (v *validator) isPrimary(typ string) bool {
	_, ok := schema.PrimaryTypes[typ]
	return ok
}

func (v *validator) primaryType(typ string) schema.Type {
	return schema.PrimaryTypes[typ]
}

func (v *validator) validateEnumName(name string) bool {
	if v.isPrimary(name) {
		return false
	}
	return v.variableName.MatchString(name)
}

func (v *validator) validateEnumValue(name string) bool {
	return v.variableName.MatchString(name)
}

func (v *validator) validateMessageName(name string) bool {
	if v.isPrimary(name) {
		return false
	}
	return v.variableName.MatchString(name)
}

func (v *validator) validateFieldName(name string) bool {
	if v.isPrimary(name) {
		return false
	}
	return v.variableName.MatchString(name)
}

func (v *validator) validateServiceName(name string) bool {
	if v.isPrimary(name) {
		return false
	}
	return v.variableName.MatchString(name)
}

func (v *validator) validateRpcName(name string) bool {
	if v.isPrimary(name) {
		return false
	}
	return v.variableName.MatchString(name)
}

func (v *validator) validatePackageName(pkg string) bool {
	return v.packageName.MatchString(pkg)
}

func (v *validator) isSingularType(typ string) bool {
	return v.singularType.MatchString(typ)
}

func (v *validator) isArrayType(typ string) bool {
	return strings.HasPrefix(typ, "array[") && strings.HasSuffix(typ, "]")
}

func (v *validator) isMapType(typ string) bool {
	return strings.HasPrefix(typ, "map[") && strings.HasSuffix(typ, "]")
}
