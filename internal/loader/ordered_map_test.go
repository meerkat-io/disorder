package loader

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMarshal(t *testing.T) {
	ms := mapSlice{
		mapItem{key: "abc", value: "123"},
		mapItem{key: "def", value: "456"},
		mapItem{key: "ghi", value: "789"},
	}

	for i := 0; i < 10; i++ {
		b, err := json.Marshal(ms)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, string(b), "{\"abc\":\"123\",\"def\":\"456\",\"ghi\":\"789\"}")
	}
}

func TestUnmarshal(t *testing.T) {
	for i := 0; i < 10; i++ {
		ms := mapSlice{}
		if err := json.Unmarshal([]byte("{\"abc\":\"123\",\"def\":\"456\",\"ghi\":\"789\"}"), &ms); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, ms[0].key, "abc")
		assert.Equal(t, ms[1].key, "def")
		assert.Equal(t, ms[2].key, "ghi")
		assert.Equal(t, ms[0].value, "123")
		assert.Equal(t, ms[1].value, "456")
		assert.Equal(t, ms[2].value, "789")
	}
}
