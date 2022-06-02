package loader

import "github.com/meerkat-lib/disorder/schema"

type Loader interface {
	Load(file string) ([]*schema.File, error)
}
