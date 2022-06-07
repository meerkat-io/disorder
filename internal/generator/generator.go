package generator

import (
	"github.com/meerkat-lib/disorder/internal/schema"
)

type Generator interface {
	Generate(dir string, files map[string]*schema.File) error
}

/*
var (
	types = map[string]int{
		"bool":   1,
		"sbyte":  1,
		"byte":   1,
		"short":  2,
		"ushort": 2,
		"int":    4,
		"uint":   4,
		"long":   8,
		"ulong":  8,
		"float":  4,
		"double": 8,
		"string": 0,
	}
)
*/
