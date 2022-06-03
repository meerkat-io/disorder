package loader

import "regexp"

type validator struct {
	variableName *regexp.Regexp
	packageName  *regexp.Regexp
}

func newValidator() *validator {
	return &validator{
		variableName: regexp.MustCompile(`^[a-zA-Z_][a-zA-Z_0-9]*$`),
		packageName:  regexp.MustCompile(`^[a-zA-Z_][a-zA-Z_0-9]*(.[a-zA-Z_][a-zA-Z_0-9]*)*$`),
	}
}

func (v *validator) validateEnumName(name string) bool {
	return v.variableName.MatchString(name)
}

func (v *validator) validateEnumValueName(name string) bool {
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
