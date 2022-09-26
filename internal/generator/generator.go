package generator

import (
	"github.com/meerkat-io/disorder/internal/schema"
)

type Generator interface {
	Generate(dir string, files map[string]*schema.File, qualified map[string]string) error
}
