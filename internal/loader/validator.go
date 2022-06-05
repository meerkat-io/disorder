package loader

import (
	"regexp"

	"github.com/meerkat-lib/disorder/internal/schema"
)

type validator struct {
	variableName       *regexp.Regexp
	packageName        *regexp.Regexp
	simpleType         *regexp.Regexp
	qualifiedType      *regexp.Regexp
	simpleArrayType    *regexp.Regexp
	qualifiedArrayType *regexp.Regexp
	simpleMapType      *regexp.Regexp
	qualifiedMapType   *regexp.Regexp
}

func newValidator() *validator {
	return &validator{
		variableName:       regexp.MustCompile(`^[a-zA-Z_][a-zA-Z_0-9]*$`),
		packageName:        regexp.MustCompile(`^[a-zA-Z_][a-zA-Z_0-9]*(.[a-zA-Z_][a-zA-Z_0-9]*)*$`),
		simpleType:         regexp.MustCompile(`^[a-zA-Z_][a-zA-Z_0-9]*$`),
		qualifiedType:      regexp.MustCompile(`^[a-zA-Z_][a-zA-Z_0-9]*(.[a-zA-Z_][a-zA-Z_0-9]*)+$`),
		simpleArrayType:    regexp.MustCompile(`^array\[[a-zA-Z_][a-zA-Z_0-9]*\]$`),
		qualifiedArrayType: regexp.MustCompile(`^array\[[a-zA-Z_][a-zA-Z_0-9]*(.[a-zA-Z_][a-zA-Z_0-9]*)+\]$`),
		simpleMapType:      regexp.MustCompile(`^map\[[a-zA-Z_][a-zA-Z_0-9]*\]$`),
		qualifiedMapType:   regexp.MustCompile(`^map\[[a-zA-Z_][a-zA-Z_0-9]*(.[a-zA-Z_][a-zA-Z_0-9]*)+\]$`),
	}
}

func (v *validator) isPrimary(typ string) bool {
	_, ok := schema.PrimaryTypes[typ]
	return ok
}

func (v *validator) primaryType(typ string) schema.Type {
	if t, ok := schema.PrimaryTypes[typ]; ok {
		return t
	}
	return schema.TypeUndefined
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

func (v *validator) isSimpleType(typ string) bool {
	return v.simpleType.MatchString(typ)
}

func (v *validator) isQualifiedType(typ string) bool {
	return v.qualifiedType.MatchString(typ)
}

func (v *validator) isSimpleArrayType(typ string) bool {
	return v.simpleArrayType.MatchString(typ)
}

func (v *validator) isQualifiedArrayType(typ string) bool {
	return v.qualifiedArrayType.MatchString(typ)
}

func (v *validator) isSimpleMapType(typ string) bool {
	return v.simpleMapType.MatchString(typ)
}

func (v *validator) isQualifiedMapType(typ string) bool {
	return v.qualifiedMapType.MatchString(typ)
}
