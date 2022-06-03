package loader

import "regexp"

type validator struct {
	variableName       *regexp.Regexp
	packageName        *regexp.Regexp
	simpleType         *regexp.Regexp
	qualifiedType      *regexp.Regexp
	arraySimpleType    *regexp.Regexp
	arrayQualifiedType *regexp.Regexp
	mapSimpleType      *regexp.Regexp
	mapQualifiedType   *regexp.Regexp
}

func newValidator() *validator {
	return &validator{
		variableName:       regexp.MustCompile(`^[a-zA-Z_][a-zA-Z_0-9]*$`),
		packageName:        regexp.MustCompile(`^[a-zA-Z_][a-zA-Z_0-9]*(.[a-zA-Z_][a-zA-Z_0-9]*)*$`),
		simpleType:         regexp.MustCompile(`^[a-zA-Z_][a-zA-Z_0-9]*$`),
		qualifiedType:      regexp.MustCompile(`^[a-zA-Z_][a-zA-Z_0-9]*(.[a-zA-Z_][a-zA-Z_0-9]*)+$`),
		arraySimpleType:    regexp.MustCompile(`^array\[[a-zA-Z_][a-zA-Z_0-9]*\]$`),
		arrayQualifiedType: regexp.MustCompile(`^array\[[a-zA-Z_][a-zA-Z_0-9]*(.[a-zA-Z_][a-zA-Z_0-9]*)+\]$`),
		mapSimpleType:      regexp.MustCompile(`^map\[[a-zA-Z_][a-zA-Z_0-9]*\]$`),
		mapQualifiedType:   regexp.MustCompile(`^map\[[a-zA-Z_][a-zA-Z_0-9]*(.[a-zA-Z_][a-zA-Z_0-9]*)+\]$`),
	}
}

func (v *validator) validateEnumName(name string) bool {
	return v.variableName.MatchString(name)
}

func (v *validator) validateEnumValue(name string) bool {
	return v.variableName.MatchString(name)
}

func (v *validator) validateMessageName(name string) bool {
	return v.variableName.MatchString(name)
}

func (v *validator) validateFieldName(name string) bool {
	return v.variableName.MatchString(name)
}

func (v *validator) validatePackageName(pkg string) bool {
	return v.packageName.MatchString(pkg)
}
