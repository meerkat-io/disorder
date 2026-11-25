package loader_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/meerkat-io/disorder/internal/loader"
	"github.com/meerkat-io/disorder/internal/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoader(t *testing.T) {
	loader := loader.NewLoader()
	files, _, err := loader.Load("../test_data/sub_schema.yaml")
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

func TestLoaderRejectsServices(t *testing.T) {
	path := filepath.Join(t.TempDir(), "old.yaml")
	content := "schema: disorder\nversion: v1\npackage: old\nmessages:\n  m:\n    a: int\nservices:\n  s:\n    r:\n      input: m\n      output: m\n"
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))

	_, _, err := loader.NewLoader().Load(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "services")
}

func TestLoaderErrors(t *testing.T) {
	dir := t.TempDir()
	write := func(name, content string) string {
		path := filepath.Join(dir, name)
		require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
		return path
	}
	cases := map[string]string{
		"missing package":  "schema: disorder\nversion: v1\nmessages:\n  m:\n    a: int\n",
		"unknown type":     "schema: disorder\nversion: v1\npackage: p\nmessages:\n  m:\n    a: nothing\n",
		"wrong schema":     "schema: other\nversion: v1\npackage: p\n",
		"duplicated field": "schema: disorder\nversion: v1\npackage: p\nmessages:\n  m:\n    a: int\n    a: long\n",
		"empty enum":       "schema: disorder\nversion: v1\npackage: p\nenums:\n  e: []\n",
	}
	for name, content := range cases {
		_, _, err := loader.NewLoader().Load(write(strings.ReplaceAll(name, " ", "_")+".yaml", content))
		assert.Error(t, err, name)
	}
	_, _, err := loader.NewLoader().Load(filepath.Join(dir, "missing.yaml"))
	assert.Error(t, err)
}
