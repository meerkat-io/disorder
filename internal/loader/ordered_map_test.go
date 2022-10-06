package loader

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMarshal(t *testing.T) {
	ms := mapSlice{
		mapItem{Key: "abc", Value: 123},
		mapItem{Key: "def", Value: 456},
		mapItem{Key: "ghi", Value: 789},
	}

	b, err := json.Marshal(ms)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, string(b), "{\"abc\":123,\"def\":456,\"ghi\":789}")
}

func TestUnmarshal(t *testing.T) {
	ms := mapSlice{}
	if err := json.Unmarshal([]byte("{\"abc\":123,\"def\":456,\"ghi\":789}"), &ms); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, ms[0].Key, "abc")
	assert.Equal(t, ms[1].Key, "def")
	assert.Equal(t, ms[2].Key, "ghi")
	assert.Equal(t, ms[0].Value, 123.0)
	assert.Equal(t, ms[1].Value, 456.0)
	assert.Equal(t, ms[2].Value, 789.0)
}
