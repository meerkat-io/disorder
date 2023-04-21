package loader

import (
	"sort"
	"sync/atomic"
)

var index uint64

type mapItem struct {
	key   string
	value interface{}
	index uint64
}

type mapSlice []mapItem

func (m mapSlice) Len() int           { return len(m) }
func (m mapSlice) Less(i, j int) bool { return m[i].index < m[j].index }
func (m mapSlice) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }

func (m *mapSlice) UnmarshalYAML(unmarshal func(interface{}) error) error {
	values := map[string]mapItem{}
	if err := unmarshal(&values); err != nil {
		return err
	}
	for k, v := range values {
		*m = append(*m, mapItem{key: k, value: v.value, index: v.index})
	}
	sort.Sort(*m)
	return nil
}

func (p *mapItem) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var v interface{}
	if err := unmarshal(&v); err != nil {
		return err
	}
	p.value = v
	p.index = next()
	return nil
}

func next() uint64 {
	return atomic.AddUint64(&index, 1)
}
