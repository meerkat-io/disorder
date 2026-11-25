package internal_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/meerkat-io/disorder"
	"github.com/meerkat-io/disorder/internal/generator/golang"
	"github.com/meerkat-io/disorder/internal/loader"
	"github.com/meerkat-io/disorder/internal/test_data/test"
	"github.com/meerkat-io/disorder/internal/test_data/test/sub"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGeneratedCodeIsUpToDate regenerates the test types from the schema and
// checks that the committed files match, so templates and fixtures cannot drift.
func TestGeneratedCodeIsUpToDate(t *testing.T) {
	files, qualifiedPath, err := loader.NewLoader().Load("./test_data/schema.yaml")
	require.NoError(t, err)

	dir := t.TempDir()
	require.NoError(t, golang.NewGoGenerator().Generate(dir, files, qualifiedPath))

	for _, name := range []string{"test_data/test/schema.go", "test_data/test/sub/sub_schema.go"} {
		generated, err := os.ReadFile(filepath.Join(dir, name))
		require.NoError(t, err, name)
		committed, err := os.ReadFile(name)
		require.NoError(t, err, name)
		assert.Equal(t, string(committed), string(generated), "%s is stale, run the generator again", name)
	}
}

func newObject() *test.Object {
	timestamp := time.UnixMilli(1700000000123)
	color := test.ColorBlue
	return &test.Object{
		BoolField:   true,
		IntField:    123,
		LongField:   1 << 40,
		FloatField:  1.5,
		DoubleField: -2.25,
		StringField: "foo",
		BytesFields: []byte{0, 1, 2, 255},
		EnumField:   &color,
		TimeField:   &timestamp,
		ObjField:    &sub.NumberWrapper{Value: &sub.Number{Value: 789}},
		IntArray:    []int32{1, 2, 3},
		IntMap:      map[string]int32{"4": 4, "5": 5, "6": 6},
		ObjArray:    []*sub.NumberWrapper{{Value: &sub.Number{Value: 1}}, {Value: &sub.Number{Value: 2}}},
		ObjMap:      map[string]*sub.NumberWrapper{"789": {Value: &sub.Number{Value: 789}}},
		Nested: map[string]map[string][][]map[string]*test.Color{
			"key0": {"key1": {{{"key2": &color}}}},
		},
	}
}

func TestBinaryRoundTrip(t *testing.T) {
	object0 := newObject()
	data, err := disorder.Marshal(object0)
	require.NoError(t, err)

	object1 := &test.Object{}
	require.NoError(t, disorder.Unmarshal(data, object1))
	assert.True(t, object0.TimeField.Equal(*object1.TimeField))
	object1.TimeField = object0.TimeField
	assert.Equal(t, object0, object1)
}

func TestJSONRoundTrip(t *testing.T) {
	object0 := newObject()
	data, err := json.Marshal(object0)
	require.NoError(t, err)

	object1 := &test.Object{}
	require.NoError(t, json.Unmarshal(data, object1))
	assert.True(t, object0.TimeField.Equal(*object1.TimeField))
	object1.TimeField = object0.TimeField
	assert.Equal(t, object0, object1)
}

// The same value must survive JSON -> binary -> JSON, so JSON can be the
// authoring format and binary the release format.
func TestJSONToBinaryToJSON(t *testing.T) {
	json0, err := json.Marshal(newObject())
	require.NoError(t, err)

	object0 := &test.Object{}
	require.NoError(t, json.Unmarshal(json0, object0))
	binary, err := disorder.Marshal(object0)
	require.NoError(t, err)

	object1 := &test.Object{}
	require.NoError(t, disorder.Unmarshal(binary, object1))
	json1, err := json.Marshal(object1)
	require.NoError(t, err)
	assert.JSONEq(t, string(json0), string(json1))
}

func TestBytesAreBase64InJSON(t *testing.T) {
	data, err := json.Marshal(&test.Object{BytesFields: []byte{0, 1, 2, 255}})
	require.NoError(t, err)

	var fields map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &fields))
	assert.Equal(t, "AAEC/w==", fields["bytes_fields"])

	object := &test.Object{}
	require.NoError(t, json.Unmarshal([]byte(`{"bytes_fields":"AAEC/w=="}`), object))
	assert.Equal(t, []byte{0, 1, 2, 255}, object.BytesFields)
}

func TestBinaryIsSmallerThanJSON(t *testing.T) {
	object := newObject()
	binary, err := disorder.Marshal(object)
	require.NoError(t, err)
	text, err := json.Marshal(object)
	require.NoError(t, err)
	assert.Less(t, len(binary), len(text))
}

func TestBinaryEncodingIsDeterministic(t *testing.T) {
	data0, err := disorder.Marshal(newObject())
	require.NoError(t, err)
	for i := 0; i < 20; i++ {
		data1, err := disorder.Marshal(newObject())
		require.NoError(t, err)
		require.Equal(t, data0, data1)
	}
}

func TestDecodeToInterface(t *testing.T) {
	data, err := disorder.Marshal(newObject())
	require.NoError(t, err)

	var generic interface{}
	require.NoError(t, disorder.Unmarshal(data, &generic))

	fields := generic.(map[string]interface{})
	assert.Equal(t, int32(123), fields["int_field"])
	assert.Equal(t, "blue", fields["enum_field"])
	assert.Equal(t, []byte{0, 1, 2, 255}, fields["bytes_fields"])

	// A generic value can be re-encoded and decoded into the typed struct.
	again, err := disorder.Marshal(generic)
	require.NoError(t, err)
	object0 := newObject()
	object1 := &test.Object{}
	require.NoError(t, disorder.Unmarshal(again, object1))
	assert.True(t, object0.TimeField.Equal(*object1.TimeField))
	object1.TimeField = object0.TimeField
	assert.Equal(t, object0, object1)
}

func TestZeroContainers(t *testing.T) {
	data, err := disorder.Marshal(&test.Zero{ZeroArray: []int32{}, ZeroMap: map[string]int32{}})
	require.NoError(t, err)

	zero := &test.Zero{}
	require.NoError(t, disorder.Unmarshal(data, zero))
	assert.NotNil(t, zero.ZeroArray)
	assert.NotNil(t, zero.ZeroMap)
	assert.Empty(t, zero.ZeroArray)
	assert.Empty(t, zero.ZeroMap)
}

type miniObject struct {
	IntField int32 `disorder:"int_field"`
}

func TestUnknownFieldsAreSkipped(t *testing.T) {
	data, err := disorder.Marshal(newObject())
	require.NoError(t, err)

	mini := &miniObject{}
	decoder := disorder.NewDecoder(bytesReader(data))
	require.NoError(t, decoder.Decode(mini))
	assert.Equal(t, int32(123), mini.IntField)
	assert.NotEmpty(t, decoder.Warnings())
}

func TestEnumValidation(t *testing.T) {
	// JSON
	var color test.Color
	assert.NoError(t, json.Unmarshal([]byte(`"green"`), &color))
	assert.Equal(t, test.ColorGreen, color)
	assert.Error(t, json.Unmarshal([]byte(`"purple"`), &color))
	assert.Error(t, json.Unmarshal([]byte(`1`), &color))
	_, err := json.Marshal(test.Color("purple"))
	assert.Error(t, err)

	// binary
	bad := test.Color("purple")
	_, err = disorder.Marshal(&bad)
	assert.Error(t, err)

	data, err := disorder.Marshal(&color)
	require.NoError(t, err)
	data[2] = 'x' // corrupt the enum text: "xreen"
	assert.Error(t, disorder.Unmarshal(data, &color))
}
