package loader

import "github.com/BurntSushi/toml"

func NewTomlLoader() Loader {
	return newLoaderImpl(&tomlUnmarshaller{})
}

type tomlUnmarshaller struct {
}

func (*tomlUnmarshaller) unmarshal(data []byte, schema *proto) error {
	return toml.Unmarshal(data, schema)
}
