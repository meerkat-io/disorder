package strcase

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

func SnakeCase(src string) string {
	fields := splitToLowerFields(src)
	return strings.Join(fields, "_")
}

func ChainCase(src string) string {
	fields := splitToLowerFields(src)
	return strings.Join(fields, "-")
}

func CamelCase(src string) string {
	fields := splitToLowerFields(src)
	for i, f := range fields {
		if i != 0 {
			fields[i] = strings.Title(f)
		}
	}
	return strings.Join(fields, "")
}

func PascalCase(src string) string {
	fields := splitToLowerFields(src)
	for i, f := range fields {
		fields[i] = strings.Title(f)
	}
	return strings.Join(fields, "")
}

func splitToLowerFields(src string) []string {
	fields := make([]string, 0)
	for _, sf := range strings.Fields(src) {
		for _, su := range strings.Split(sf, "_") {
			for _, sh := range strings.Split(su, "-") {
				for _, sc := range splitCamelCase(sh) {
					fields = append(fields, strings.ToLower(sc))
				}
			}
		}
	}
	return fields
}

func splitCamelCase(src string) []string {
	if !utf8.ValidString(src) {
		return []string{src}
	}
	entries := []string{}
	var runes [][]rune
	lastClass := 0
	class := 0
	for _, r := range src {
		switch true {
		case unicode.IsLower(r):
			class = 1
		case unicode.IsUpper(r):
			class = 2
		case unicode.IsDigit(r):
			class = 3
		default:
			class = 4
		}
		if class == lastClass {
			runes[len(runes)-1] = append(runes[len(runes)-1], r)
		} else {
			runes = append(runes, []rune{r})
		}
		lastClass = class
	}
	// handle upper case -> lower case sequences, e.g.
	// "PDFL", "oader" -> "PDF", "Loader"
	for i := 0; i < len(runes)-1; i++ {
		if unicode.IsUpper(runes[i][0]) && unicode.IsLower(runes[i+1][0]) {
			runes[i+1] = append([]rune{runes[i][len(runes[i])-1]}, runes[i+1]...)
			runes[i] = runes[i][:len(runes[i])-1]
		}
	}
	// construct []string from results
	for _, s := range runes {
		if len(s) > 0 {
			entries = append(entries, string(s))
		}
	}
	return entries
}
