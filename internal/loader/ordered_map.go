package loader

import (
	"bytes"
	"encoding/json"
	"fmt"
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

func (m *mapSlice) UnmarshalJSON(b []byte) error {
	values := map[string]mapItem{}
	if err := json.Unmarshal(b, &values); err != nil {
		return err
	}
	for k, v := range values {
		*m = append(*m, mapItem{key: k, value: v.value, index: v.index})
	}
	sort.Sort(*m)
	return nil
}

func (p *mapItem) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	p.value = v
	p.index = next()
	return nil
}

func (m mapSlice) MarshalJSON() ([]byte, error) {
	buf := &bytes.Buffer{}
	buf.Write([]byte{'{'})
	for i, item := range m {
		if i != 0 {
			buf.Write([]byte{','})
		}
		b, err := json.Marshal(item.value)
		if err != nil {
			return nil, err
		}
		buf.WriteString(fmt.Sprintf("%q:", fmt.Sprint(item.key)))
		buf.Write(b)
	}
	buf.Write([]byte{'}'})
	return buf.Bytes(), nil
}

func next() uint64 {
	return atomic.AddUint64(&index, 1)
}
