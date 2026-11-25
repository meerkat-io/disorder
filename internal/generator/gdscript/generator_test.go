package gdscript_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/meerkat-io/disorder/internal/generator/gdscript"
	"github.com/meerkat-io/disorder/internal/loader"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const header = "schema: disorder\nversion: v1\npackage: p\n"

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

// generate writes the schemas, generates GDScript from main and, if Godot is
// available, checks that every generated script parses. It returns the output
// dir.
func generate(t *testing.T, main string, schemas map[string]string) (string, error) {
	dir := t.TempDir()
	for name, content := range schemas {
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600))
	}
	files, qualified, err := loader.NewLoader().Load(filepath.Join(dir, main))
	require.NoError(t, err)

	out := filepath.Join(dir, "out")
	require.NoError(t, os.Mkdir(out, 0o700))
	if err := gdscript.NewGenerator().Generate(out, files, qualified); err != nil {
		return "", err
	}
	if bin := godotBinary(); bin != "" {
		require.NoError(t, os.WriteFile(filepath.Join(out, "project.godot"), []byte("config_version=5\n"), 0o600))
		var scripts []string
		require.NoError(t, filepath.WalkDir(out, func(path string, d os.DirEntry, err error) error {
			if err == nil && filepath.Ext(path) == ".gd" {
				scripts = append(scripts, path)
			}
			return err
		}))
		require.NotEmpty(t, scripts)
		for _, script := range scripts {
			cmd := exec.Command(bin, "--headless", "--path", out, "--check-only", "--script", script)
			output, err := cmd.CombinedOutput()
			require.NoError(t, err, "%s does not parse:\n%s", script, output)
			require.NotContains(t, string(output), "ERROR", "%s does not parse:\n%s", script, output)
		}
	}
	return out, nil
}

func generateOne(t *testing.T, schema string) string {
	out, err := generate(t, "s.yaml", map[string]string{"s.yaml": schema})
	require.NoError(t, err)
	source, err := os.ReadFile(filepath.Join(out, "p", "s.gd"))
	require.NoError(t, err)
	return string(source)
}

func TestTypes(t *testing.T) {
	source := generateOne(t, header+`messages:
  m:
    b: bool
    i: int
    l: long
    f: float
    d: double
    s: string
    y: bytes
    t: timestamp
    e: tone
    o: other
    arr: array[other]
    enums: array[tone]
    nested: array[array[int]]
    map: map[map[tone]]
  other:
    v: int
enums:
  tone: [low, High_Pitch]
`)
	for _, want := range []string{
		"var b: bool = false",
		"var i: int = 0",
		"var l: int = 0",
		"var f: float = 0.0",
		"var d: float = 0.0",
		`var s: String = ""`,
		"var y: PackedByteArray = PackedByteArray()",
		"var t: int = 0  # unix time in milliseconds",
		"var e: Tone = Tone.LOW",
		"var o: Other = null",
		"var arr: Array[Other] = []",
		"var enums: Array[int] = []",
		"var nested: Array = []",
		"var map: Dictionary = {}",
		"enum Tone {",
		"HIGH_PITCH,",
		`const TONE_VALUES = ["low", "High_Pitch"]`,
		`["l", "long"]`,
		`["e", ["enum", TONE_VALUES]]`,
		`["map", ["map", ["map", ["enum", TONE_VALUES]]]]`,
	} {
		assert.Contains(t, source, want)
	}
}

func TestClassPrefix(t *testing.T) {
	source := generateOne(t, header+"option:\n  gdscript_class_prefix: Dis\nmessages:\n  object:\n    c: color\nenums:\n  color: [red]\n")
	assert.Contains(t, source, "class DisObject extends RefCounted:")
	assert.Contains(t, source, "enum DisColor {")
	assert.Contains(t, source, "const DIS_COLOR_VALUES")
}

func TestBuiltinNameClashIsRejected(t *testing.T) {
	_, err := generate(t, "s.yaml", map[string]string{"s.yaml": header + "messages:\n  object:\n    a: int\n"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "gdscript_class_prefix")

	_, err = generate(t, "s.yaml", map[string]string{"s.yaml": header + "enums:\n  color: [red]\n"})
	require.Error(t, err)
}

func TestEnumValueClashIsRejected(t *testing.T) {
	_, err := generate(t, "s.yaml", map[string]string{"s.yaml": header + "enums:\n  e: [Red, red]\n"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "clash")
}

func TestCrossPackageImports(t *testing.T) {
	out, err := generate(t, "a.yaml", map[string]string{
		"a.yaml": "schema: disorder\nversion: v1\npackage: x.a\nimport:\n  - b.yaml\nmessages:\n  m:\n    one: y.b.n\n    many: array[y.b.n]\n    c: y.b.c\n",
		"b.yaml": "schema: disorder\nversion: v1\npackage: y.b\nmessages:\n  n:\n    v: int\nenums:\n  c: [red]\n",
	})
	require.NoError(t, err)
	source, err := os.ReadFile(filepath.Join(out, "x", "a", "a.gd"))
	require.NoError(t, err)
	assert.Contains(t, string(source), `const _YBB = preload("../../y/b/b.gd")`)
	assert.Contains(t, string(source), "var one: _YBB.N = null")
	assert.Contains(t, string(source), "var many: Array[_YBB.N] = []")
	assert.Contains(t, string(source), "var c: _YBB.C = _YBB.C.RED")
	assert.Contains(t, string(source), `["enum", _YBB.C_VALUES]`)
}
