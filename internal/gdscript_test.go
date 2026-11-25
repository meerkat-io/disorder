package internal_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/meerkat-io/disorder"
	"github.com/meerkat-io/disorder/internal/generator/gdscript"
	"github.com/meerkat-io/disorder/internal/loader"
	"github.com/meerkat-io/disorder/internal/test_data/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const gdscriptProject = "../lib/gdscript"

// TestGDScriptGeneratedCodeIsUpToDate regenerates the GDScript test classes
// and checks that the committed files match.
func TestGDScriptGeneratedCodeIsUpToDate(t *testing.T) {
	files, qualifiedPath, err := loader.NewLoader().Load("./test_data/schema.yaml")
	require.NoError(t, err)

	dir := t.TempDir()
	require.NoError(t, gdscript.NewGenerator().Generate(dir, files, qualifiedPath))

	for _, name := range []string{"test_data/test/schema.gd", "test_data/test/sub/sub_schema.gd"} {
		generated, err := os.ReadFile(filepath.Join(dir, name))
		require.NoError(t, err, name)
		committed, err := os.ReadFile(filepath.Join(gdscriptProject, "test/generated", name))
		require.NoError(t, err, name)
		assert.Equal(t, string(committed), string(generated), "%s is stale, run the generator again", name)
	}
}

func godotBinary() string {
	if bin := os.Getenv("GODOT"); bin != "" {
		return bin
	}
	for _, name := range []string{"godot", "godot4"} {
		if path, err := exec.LookPath(name); err == nil {
			return path
		}
	}
	return ""
}

// TestGDScriptInterop runs the Godot test suite, which decodes the data written
// here by Go and writes data back for Go to decode. It needs Godot 4 (set
// GODOT or put godot on PATH) and is skipped otherwise.
func TestGDScriptInterop(t *testing.T) {
	bin := godotBinary()
	if bin == "" {
		t.Skip("godot not found: set GODOT or add godot to PATH")
	}
	dir := t.TempDir()
	expected := newObject()
	binary, err := disorder.Marshal(expected)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "object.bin"), binary, 0o600))
	text, err := json.Marshal(expected)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "object.json"), text, 0o600))

	cmd := exec.Command(bin, "--headless", "--path", gdscriptProject, "--script", "test/run.gd", "--", dir)
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "godot tests failed:\n%s", output)
	require.Contains(t, string(output), " 0 failures", "godot output:\n%s", output)

	// GDScript cannot tell an unset enum, timestamp or container from an empty
	// one, so those fields are written even when empty. Compare the rest.
	normalize := func(o *test.Object) {
		o.EmptyEnum, o.EmptyTime, o.EmptyArray, o.EmptyMap = nil, nil, nil, nil
		if o.TimeField != nil { // same instant, different Location pointer
			utc := o.TimeField.UTC()
			o.TimeField = &utc
		}
	}
	normalize(expected)

	fromBinary, err := os.ReadFile(filepath.Join(dir, "from_gd.bin"))
	require.NoError(t, err)
	got := &test.Object{}
	require.NoError(t, disorder.Unmarshal(fromBinary, got))
	normalize(got)
	assert.Equal(t, expected, got, "binary written by GDScript")

	fromJSON, err := os.ReadFile(filepath.Join(dir, "from_gd.json"))
	require.NoError(t, err)
	got = &test.Object{}
	require.NoError(t, json.Unmarshal(fromJSON, got))
	normalize(got)
	assert.Equal(t, expected, got, "json written by GDScript")
}
