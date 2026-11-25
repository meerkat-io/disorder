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

func writeFiles(t *testing.T, files map[string]string) string {
	dir := t.TempDir()
	for name, content := range files {
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600))
	}
	return dir
}

func TestLoaderRejectsInvalidNames(t *testing.T) {
	const header = "schema: disorder\nversion: v1\n"
	cases := map[string]string{
		"package with dash":  header + "package: a-b\nmessages:\n  m:\n    a: int\n",
		"type with dash":     header + "package: p\nmessages:\n  m:\n    a: x-y\n",
		"nested type dash":   header + "package: p\nmessages:\n  m:\n    a: array[x-y]\n",
		"field is a type":    header + "package: p\nmessages:\n  m:\n    int: int\n",
		"duplicated message": header + "package: p\nmessages:\n  m:\n    a: int\n  m:\n    a: int\n",
	}
	dir := writeFiles(t, nil)
	for name, content := range cases {
		path := filepath.Join(dir, strings.ReplaceAll(name, " ", "_")+".yaml")
		require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
		_, _, err := loader.NewLoader().Load(path)
		assert.Error(t, err, name)
	}
}

func TestLoaderRejectsDuplicateDefinitionsAcrossFiles(t *testing.T) {
	const body = "schema: disorder\nversion: v1\npackage: p\n"
	dir := writeFiles(t, map[string]string{
		"main.yaml":     body + "import:\n  - other.yaml\nmessages:\n  m:\n    a: int\nenums:\n  e: [x]\n",
		"other.yaml":    body + "messages:\n  m:\n    a: int\n",
		"enum.yaml":     body + "import:\n  - dup_enum.yaml\nenums:\n  e: [x]\n",
		"dup_enum.yaml": body + "enums:\n  e: [y]\n",
	})
	for _, name := range []string{"main.yaml", "enum.yaml"} {
		_, _, err := loader.NewLoader().Load(filepath.Join(dir, name))
		require.Error(t, err, name)
		assert.Contains(t, err.Error(), "duplicate")
	}
}

func TestLoaderAllowsSameNameInDifferentPackages(t *testing.T) {
	dir := writeFiles(t, map[string]string{
		"a.yaml": "schema: disorder\nversion: v1\npackage: a\nimport:\n  - b.yaml\nmessages:\n  m:\n    x: b.m\n",
		"b.yaml": "schema: disorder\nversion: v1\npackage: b\nmessages:\n  m:\n    x: int\n",
	})
	_, qualified, err := loader.NewLoader().Load(filepath.Join(dir, "a.yaml"))
	require.NoError(t, err)
	assert.Contains(t, qualified, "a.m")
	assert.Contains(t, qualified, "b.m")
}

func TestLoaderKeepsDeclarationOrder(t *testing.T) {
	dir := writeFiles(t, map[string]string{
		"s.yaml": "schema: disorder\nversion: v1\npackage: p\nmessages:\n  zeta:\n    z: int\n    a: int\n    m: int\n  alpha:\n    b: int\nenums:\n  e: [z, a, m]\n",
	})
	files, _, err := loader.NewLoader().Load(filepath.Join(dir, "s.yaml"))
	require.NoError(t, err)
	for _, file := range files {
		assert.Equal(t, "zeta", file.Messages[0].Name)
		assert.Equal(t, "alpha", file.Messages[1].Name)
		var fields []string
		for _, f := range file.Messages[0].Fields {
			fields = append(fields, f.Name)
		}
		assert.Equal(t, []string{"z", "a", "m"}, fields)
		assert.Equal(t, []string{"z", "a", "m"}, file.Enums[0].Values)
	}
}

func TestLoaderIsReusable(t *testing.T) {
	path := writeFiles(t, map[string]string{
		"s.yaml": "schema: disorder\nversion: v1\npackage: p\nmessages:\n  m:\n    a: int\n",
	})
	l := loader.NewLoader()
	for i := 0; i < 2; i++ {
		_, _, err := l.Load(filepath.Join(path, "s.yaml"))
		require.NoError(t, err, "load #%d", i)
	}
}
