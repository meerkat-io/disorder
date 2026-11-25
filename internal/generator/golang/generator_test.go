package golang_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/meerkat-io/disorder/internal/generator/golang"
	"github.com/meerkat-io/disorder/internal/loader"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	header = "schema: disorder\nversion: v1\npackage: p\n"
	module = "example.com/gen"
)

// generate writes the schema files, generates Go code from main, and checks
// that the output builds as a standalone module. It returns the output dir.
func generate(t *testing.T, main string, schemas map[string]string) string {
	dir := t.TempDir()
	for name, content := range schemas {
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600))
	}
	files, qualified, err := loader.NewLoader().Load(filepath.Join(dir, main))
	require.NoError(t, err)

	out := filepath.Join(dir, "out")
	require.NoError(t, os.Mkdir(out, 0o700))
	require.NoError(t, golang.NewGoGenerator().Generate(out, files, qualified))

	require.NoError(t, os.WriteFile(filepath.Join(out, "go.mod"), []byte("module "+module+"\n\ngo 1.21\n"), 0o600))
	cmd := exec.Command("go", "build", "./...")
	cmd.Dir = out
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "generated code does not build:\n%s", output)
	return out
}

func generateOne(t *testing.T, schema string) string {
	out := generate(t, "s.yaml", map[string]string{"s.yaml": schema})
	source, err := os.ReadFile(filepath.Join(out, "p", "s.go"))
	require.NoError(t, err)
	return string(source)
}

func TestTimeImportedForNestedTimestamps(t *testing.T) {
	for _, field := range []string{"timestamp", "array[timestamp]", "map[array[timestamp]]"} {
		assert.Contains(t, generateOne(t, header+"messages:\n  m:\n    t: "+field+"\n"), `"time"`, field)
	}
	assert.NotContains(t, generateOne(t, header+"messages:\n  m:\n    a: int\n"), `"time"`)
}

func TestEnumImports(t *testing.T) {
	source := generateOne(t, header+"enums:\n  e: [x]\n")
	assert.Contains(t, source, `"fmt"`)
	assert.Contains(t, source, `"encoding/json"`)
}

func TestAllTypesBuild(t *testing.T) {
	generateOne(t, header+`messages:
  m:
    b: bool
    i: int
    l: long
    f: float
    d: double
    s: string
    y: bytes
    t: timestamp
    e: color
    o: other
    arr: array[other]
    nested: map[map[array[array[map[color]]]]]
  other:
    v: int
enums:
  color: [red, green]
`)
}

func TestCrossPackageBuild(t *testing.T) {
	out := generate(t, "a.yaml", map[string]string{
		"a.yaml": "schema: disorder\nversion: v1\npackage: a\nimport:\n  - b.yaml\noption:\n  go_package_prefix: " + module +
			"\nmessages:\n  m:\n    one: b.n\n    many: array[b.n]\n    color: b.c\n",
		"b.yaml": "schema: disorder\nversion: v1\npackage: b\noption:\n  go_package_prefix: " + module +
			"\nmessages:\n  n:\n    v: int\nenums:\n  c: [red]\n",
	})
	assert.FileExists(t, filepath.Join(out, "a", "a.go"))
	assert.FileExists(t, filepath.Join(out, "b", "b.go"))
}
