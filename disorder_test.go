package disorder_test

import (
	"math"
	"testing"

	"github.com/meerkat-io/disorder"
	"github.com/stretchr/testify/assert"
)

/*
	tagBool   tag = 1
	tagInt    tag = 2
	tagLong   tag = 3
	tagFloat  tag = 4
	tagDouble tag = 5
	tagBytes  tag = 6
*/

func TestPrimaryTypes(t *testing.T) {
	var result interface{}

	var b bool = true
	data, err := disorder.Marshal(b)
	assert.Nil(t, err)
	disorder.Unmarshal(data, &result)
	assert.Equal(t, result, b)

	var i int32 = -123
	data, err = disorder.Marshal(i)
	assert.Nil(t, err)
	disorder.Unmarshal(data, &result)
	assert.Equal(t, result, i)

	var l int64 = -456
	data, err = disorder.Marshal(l)
	assert.Nil(t, err)
	disorder.Unmarshal(data, &result)
	assert.Equal(t, result, l)

	var f float32 = -3.14
	data, err = disorder.Marshal(f)
	assert.Nil(t, err)
	disorder.Unmarshal(data, &result)
	assert.True(t, math.Abs(float64(result.(float32))-float64(f)) <= 1e-9)

	var d float64 = -6.28
	data, err = disorder.Marshal(d)
	assert.Nil(t, err)
	disorder.Unmarshal(data, &result)
	assert.Equal(t, result, d)
	assert.True(t, math.Abs(result.(float64)-d) <= 1e-9)
}

func TestUnsupportedType(t *testing.T) {
	_, err := disorder.Marshal(123)
	assert.NotNil(t, err)
}
