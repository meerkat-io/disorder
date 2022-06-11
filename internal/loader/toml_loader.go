package loader

import toml "github.com/pelletier/go-toml/v2"

func NewTomlLoader() Loader {
	return newLoaderImpl(&tomlUnmarshaller{})
}

type tomlUnmarshaller struct {
}

func (*tomlUnmarshaller) unmarshal(data []byte, schema *proto) error {
	return toml.Unmarshal(data, schema)
}
