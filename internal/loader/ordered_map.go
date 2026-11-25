package loader

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// Go maps lose the declaration order of yaml keys, and the generated code
// should follow the schema, so these types decode a yaml mapping into an
// ordered slice.

type mapItem struct {
	key   string
	value interface{}
}

// mapSlice is a yaml mapping whose values are arbitrary.
type mapSlice []mapItem

type mapSliceItem struct {
	key   string
	value mapSlice
}

// mapMatrix is a yaml mapping whose values are mapSlices.
type mapMatrix []mapSliceItem

func (m *mapSlice) UnmarshalYAML(node *yaml.Node) error {
	return decodeMapping(node, func(key string, value *yaml.Node) error {
		var v interface{}
		if err := value.Decode(&v); err != nil {
			return err
		}
		*m = append(*m, mapItem{key: key, value: v})
		return nil
	})
}

func (m *mapMatrix) UnmarshalYAML(node *yaml.Node) error {
	return decodeMapping(node, func(key string, value *yaml.Node) error {
		var v mapSlice
		if err := value.Decode(&v); err != nil {
			return err
		}
		*m = append(*m, mapSliceItem{key: key, value: v})
		return nil
	})
}

// decodeMapping calls add for each key/value of a yaml mapping in document
// order and rejects duplicated keys.
func decodeMapping(node *yaml.Node, add func(key string, value *yaml.Node) error) error {
	if node.Kind != yaml.MappingNode {
		return fmt.Errorf("line %d: expect a mapping", node.Line)
	}
	seen := map[string]bool{}
	for i := 0; i+1 < len(node.Content); i += 2 {
		key := node.Content[i]
		if seen[key.Value] {
			return fmt.Errorf("line %d: duplicated key \"%s\"", key.Line, key.Value)
		}
		seen[key.Value] = true
		if err := add(key.Value, node.Content[i+1]); err != nil {
			return err
		}
	}
	return nil
}
