package disorder_test

import (
	"testing"
	"time"

	"github.com/meerkat-io/disorder"
	"github.com/stretchr/testify/assert"
)

func TestPrimaryTypes(t *testing.T) {
	var b bool
	data, err := disorder.Marshal(true)
	assert.Nil(t, err)
	err = disorder.Unmarshal(data, &b)
	assert.Nil(t, err)
	assert.Equal(t, true, b)

	var i int32
	data, err = disorder.Marshal(int32(-123))
	assert.Nil(t, err)
	err = disorder.Unmarshal(data, &i)
	assert.Nil(t, err)
	assert.Equal(t, int32(-123), i)

	var l int64
	data, err = disorder.Marshal(int64(-456))
	assert.Nil(t, err)
	err = disorder.Unmarshal(data, &l)
	assert.Nil(t, err)
	assert.Equal(t, int64(-456), l)

	var f float32
	data, err = disorder.Marshal(float32(-3.14))
	assert.Nil(t, err)
	err = disorder.Unmarshal(data, &f)
	assert.Nil(t, err)
	assert.Equal(t, float32(-3.14), f)

	var d float64
	data, err = disorder.Marshal(float64(-6.28))
	assert.Nil(t, err)
	err = disorder.Unmarshal(data, &d)
	assert.Nil(t, err)
	assert.Equal(t, float64(-6.28), d)

	var bytes []byte
	data, err = disorder.Marshal([]byte{1, 2, 3})
	assert.Nil(t, err)
	err = disorder.Unmarshal(data, &bytes)
	assert.Nil(t, err)
	assert.Equal(t, []byte{1, 2, 3}, bytes)
}

func TestUnsupportedEncodeType(t *testing.T) {
	_, err := disorder.Marshal(int(123))
	assert.NotNil(t, err)

	_, err = disorder.Marshal(nil)
	assert.NotNil(t, err)
}

func TestUtilTypes(t *testing.T) {
	var s string
	data, err := disorder.Marshal("hello world")
	assert.Nil(t, err)
	err = disorder.Unmarshal(data, &s)
	assert.Nil(t, err)
	assert.Equal(t, "hello world", s)

	var now time.Time = time.UnixMilli(time.Now().UnixMilli())
	var result time.Time
	data, err = disorder.Marshal(&now)
	assert.Nil(t, err)
	err = disorder.Unmarshal(data, &result)
	assert.Nil(t, err)
	assert.Equal(t, result, now)

	var c Color
	data, err = disorder.Marshal(&ColorBlue)
	assert.Nil(t, err)
	err = disorder.Unmarshal(data, &c)
	assert.Nil(t, err)
	assert.Equal(t, ColorBlue, c)
}

func TestContainerTypes(t *testing.T) {

}
