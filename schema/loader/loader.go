package loader

import "github.com/meerkat-lib/disorder/schema"

type Loader interface {
	Load(file string) ([]*schema.File, error)
}

type rpc struct {
	Input  string `yaml:"input", json:"input"`
	Output string `yaml:"output", json:"output"`
}

type schemaFile struct {
	FilePath string `yaml:"-", json:"-"`

	Package  string                       `yaml:"package", json:"package"`
	Imports  []string                     `yaml:"import", json:"import"`
	Enums    map[string][]string          `yaml:"enums", json:"enums"`
	Messages map[string]map[string]string `yaml:"messages", json:"messages"`
	Services map[string]*rpc              `yaml:"services", json:"services"`
}
