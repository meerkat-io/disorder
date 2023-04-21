package loader_test

import (
	"testing"

	"github.com/meerkat-io/disorder/internal/loader"
	"github.com/meerkat-io/disorder/internal/schema"
	"github.com/stretchr/testify/assert"
)

func TestLoadJsonFile(t *testing.T) {
	loader := loader.NewJsonLoader()
	files, _, err := loader.Load("../test_data/json_schema.json")
	assert.Nil(t, err)

	var file *schema.File
	for _, f := range files {
		file = f
	}

	assert.Equal(t, "number", file.Messages[0].Name)
	assert.Equal(t, "value", file.Messages[0].Fields[0].Name)
	assert.Equal(t, schema.TypeInt, file.Messages[0].Fields[0].Type.Type)

	assert.Equal(t, "number_wrapper", file.Messages[1].Name)
	assert.Equal(t, "value", file.Messages[1].Fields[0].Name)
	assert.Equal(t, "number", file.Messages[1].Fields[0].Type.TypeRef)
}
