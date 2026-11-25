package generator

import (
	"github.com/meerkat-io/disorder/internal/schema"
)

// Generator turns loaded schema files into source code in dir.
type Generator interface {
	Generate(dir string, files map[string]*schema.File, qualified map[string]string) error
}
