package disorder_test

import (
	"bytes"
	"io"
	"strings"
	"testing"
	"testing/iotest"
	"time"

	"github.com/meerkat-io/disorder"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type sample struct {
	Name  string            `disorder:"name"`
	Tags  []string          `disorder:"tags"`
	Attrs map[string]int64  `disorder:"attrs"`
	When  *time.Time        `disorder:"when"`
	Blob  []byte            `disorder:"blob"`
	Inner *sample           `disorder:"inner"`
	Any   interface{}       `disorder:"any"`
	Skip  string            `disorder:"-"`
	Maps  map[string][]bool `disorder:"maps"`
}

func newSample() *sample {
	when := time.UnixMilli(1700000000000)
	return &sample{
		Name:  "outer",
		Tags:  []string{"a", "b"},
		Attrs: map[string]int64{"x": 1, "y": -2},
		When:  &when,
		Blob:  []byte("blob"),
		Inner: &sample{Name: "inner"},
		Any:   int32(7),
		Skip:  "skipped",
		Maps:  map[string][]bool{"k": {true, false}},
	}
}

func TestStructRoundTrip(t *testing.T) {
	data, err := disorder.Marshal(newSample())
	require.NoError(t, err)

	out := &sample{}
	require.NoError(t, disorder.Unmarshal(data, out))
	assert.Equal(t, "outer", out.Name)
	assert.Equal(t, []string{"a", "b"}, out.Tags)
	assert.Equal(t, map[string]int64{"x": 1, "y": -2}, out.Attrs)
	assert.True(t, out.When.Equal(*newSample().When))
	assert.Equal(t, []byte("blob"), out.Blob)
	assert.Equal(t, "inner", out.Inner.Name)
	assert.Equal(t, int32(7), out.Any)
	assert.Empty(t, out.Skip)
	assert.Equal(t, map[string][]bool{"k": {true, false}}, out.Maps)
}

// Readers may return short reads; the decoder must keep reading.
func TestDecodeFromSlowReader(t *testing.T) {
	data, err := disorder.Marshal(newSample())
	require.NoError(t, err)

	out := &sample{}
	require.NoError(t, disorder.NewDecoder(iotest.OneByteReader(bytes.NewReader(data))).Decode(out))
	assert.Equal(t, "outer", out.Name)
	assert.Equal(t, "inner", out.Inner.Name)
}

func TestDecodeTruncatedData(t *testing.T) {
	data, err := disorder.Marshal(newSample())
	require.NoError(t, err)

	for i := 0; i < len(data); i++ {
		out := &sample{}
		assert.Error(t, disorder.Unmarshal(data[:i], out), "prefix of %d bytes", i)
	}
}

func TestDecodeTruncatedUnknownField(t *testing.T) {
	data, err := disorder.Marshal(newSample())
	require.NoError(t, err)

	type tiny struct {
		Name string `disorder:"name"`
	}
	assert.NoError(t, disorder.Unmarshal(data, &tiny{}))
	assert.Error(t, disorder.Unmarshal(data[:len(data)-5], &tiny{}))
}

func TestLargeBytes(t *testing.T) {
	big := bytes.Repeat([]byte{0xAB}, 1<<20)
	data, err := disorder.Marshal(big)
	require.NoError(t, err)

	var out []byte
	require.NoError(t, disorder.Unmarshal(data, &out))
	assert.Equal(t, big, out)

	// A length prefix larger than the payload must fail without allocating it.
	assert.Error(t, disorder.Unmarshal(data[:100], &out))
	lie := []byte{6, 0xFF, 0xFF, 0xFF, 0xFF, 1, 2, 3}
	assert.Error(t, disorder.Unmarshal(lie, &out))
}

func TestStringWithoutPrefixLimit(t *testing.T) {
	s := strings.Repeat("z", 100000)
	data, err := disorder.Marshal(s)
	require.NoError(t, err)
	var out string
	require.NoError(t, disorder.Unmarshal(data, &out))
	assert.Equal(t, s, out)
}

func TestEncodeErrors(t *testing.T) {
	_, err := disorder.Marshal(nil)
	assert.Error(t, err)

	var np *sample
	_, err = disorder.Marshal(np)
	assert.Error(t, err)

	_, err = disorder.Marshal(map[int]string{1: "a"})
	assert.Error(t, err, "non string map key")

	_, err = disorder.Marshal(map[string]string{"": "a"})
	assert.Error(t, err, "empty map key")

	_, err = disorder.Marshal(map[string]string{strings.Repeat("k", 256): "a"})
	assert.Error(t, err, "key longer than 255 bytes")

	_, err = disorder.Marshal(int(1))
	assert.Error(t, err, "plain int has no fixed width")

	_, err = disorder.Marshal(make(chan int))
	assert.Error(t, err)
}

func TestDecodeErrors(t *testing.T) {
	data, err := disorder.Marshal(int32(1))
	require.NoError(t, err)

	var s string
	assert.Error(t, disorder.Unmarshal(data, &s), "type mismatch")
	assert.Error(t, disorder.Unmarshal(data, s), "non pointer target")
	assert.Error(t, disorder.Unmarshal(data, nil))
	assert.Error(t, disorder.Unmarshal(nil, &s), "empty input")
	assert.Error(t, disorder.Unmarshal([]byte{99}, &s), "unknown tag")

	var l int64
	assert.Error(t, disorder.Unmarshal(data, &l), "int32 into int64")

	arr, err := disorder.Marshal([]int32{1})
	require.NoError(t, err)
	assert.Error(t, disorder.Unmarshal(arr, &l), "array into scalar")
	var obj struct{ A int32 }
	assert.Error(t, disorder.Unmarshal(arr, &obj), "array into struct")
}

func TestDecodeIntoGenericTypes(t *testing.T) {
	cases := []interface{}{
		true, int32(-1), int64(-2), float32(1.5), float64(-2.5), "str", []byte{1, 2},
		time.UnixMilli(1700000000000),
	}
	for _, c := range cases {
		data, err := disorder.Marshal(c)
		require.NoError(t, err)
		var out interface{}
		require.NoError(t, disorder.Unmarshal(data, &out))
		assert.Equal(t, c, out)
	}

	data, err := disorder.Marshal(map[string]interface{}{
		"list": []interface{}{int32(1), "two", map[string]interface{}{"three": true}},
	})
	require.NoError(t, err)
	var out interface{}
	require.NoError(t, disorder.Unmarshal(data, &out))
	assert.Equal(t, map[string]interface{}{
		"list": []interface{}{int32(1), "two", map[string]interface{}{"three": true}},
	}, out)
}

func TestEncodeIsDeterministic(t *testing.T) {
	m := map[string]int32{}
	for _, k := range []string{"q", "w", "e", "r", "t", "y", "u", "i", "o", "p"} {
		m[k] = int32(len(k))
	}
	first, err := disorder.Marshal(m)
	require.NoError(t, err)
	for i := 0; i < 20; i++ {
		again, err := disorder.Marshal(m)
		require.NoError(t, err)
		require.Equal(t, first, again)
	}
}

func TestEncoderDecoderStream(t *testing.T) {
	buf := &bytes.Buffer{}
	enc := disorder.NewEncoder(buf)
	require.NoError(t, enc.Encode(int32(1)))
	require.NoError(t, enc.Encode("two"))
	require.NoError(t, enc.Encode([]int32{3}))

	dec := disorder.NewDecoder(buf)
	var a int32
	var b string
	var c []int32
	require.NoError(t, dec.Decode(&a))
	require.NoError(t, dec.Decode(&b))
	require.NoError(t, dec.Decode(&c))
	assert.Equal(t, int32(1), a)
	assert.Equal(t, "two", b)
	assert.Equal(t, []int32{3}, c)
	assert.ErrorIs(t, dec.Decode(&a), io.EOF)
}

func TestDecodeEnumFromString(t *testing.T) {
	var c Color
	data, err := disorder.Marshal("green")
	require.NoError(t, err)
	require.NoError(t, disorder.Unmarshal(data, &c))
	assert.Equal(t, "green", c.value)

	data, err = disorder.Marshal("purple")
	require.NoError(t, err)
	assert.Error(t, disorder.Unmarshal(data, &c))

	// enum into plain string
	data, err = disorder.Marshal(&c)
	require.NoError(t, err)
	var s string
	require.NoError(t, disorder.Unmarshal(data, &s))
	assert.Equal(t, "green", s)
}
