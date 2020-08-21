package param_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ucarion/cli/param"
)

func TestMayMustTakeValue(t *testing.T) {
	var v1 bool
	p1, err := param.New(&v1)
	assert.NoError(t, err)

	var v2 string
	p2, err := param.New(&v2)
	assert.NoError(t, err)

	var v3 *string
	p3, err := param.New(&v3)
	assert.NoError(t, err)

	assert.False(t, param.MayTakeValue(p1))
	assert.True(t, param.MayTakeValue(p2))
	assert.True(t, param.MayTakeValue(p3))

	assert.False(t, param.MustTakeValue(p1))
	assert.True(t, param.MustTakeValue(p2))
	assert.False(t, param.MustTakeValue(p3))
}

func TestNewNotPointer(t *testing.T) {
	var v bool
	_, err := param.New(v)
	assert.Equal(t, "v must be a pointer", err.Error())
}

func TestNewUnsupportedType(t *testing.T) {
	var v chan string
	_, err := param.New(&v)
	assert.Equal(t, "unsupported param type: chan string", err.Error())
}

func TestNewUnsupportedTypeSlice(t *testing.T) {
	var v []chan string
	_, err := param.New(&v)
	assert.Equal(t, "unsupported slice param type: unsupported param type: chan string", err.Error())
}

func TestNewUnsupportedTypePtr(t *testing.T) {
	var v *chan string
	_, err := param.New(&v)
	assert.Equal(t, "unsupported pointer param type: unsupported param type: chan string", err.Error())
}

func TestNewBool(t *testing.T) {
	var v bool
	p, err := param.New(&v)
	if assert.NoError(t, err) {
		p.Set("")
		assert.Equal(t, true, v)
	}
}

func TestNewInt(t *testing.T) {
	var v int
	p, err := param.New(&v)
	if assert.NoError(t, err) {
		p.Set("-10")
		assert.Equal(t, int(-10), v)

		p.Set("0xf")
		assert.Equal(t, int(15), v)
	}
}

func TestNewUint(t *testing.T) {
	var v uint
	p, err := param.New(&v)
	if assert.NoError(t, err) {
		p.Set("10")
		assert.Equal(t, uint(10), v)

		p.Set("0xf")
		assert.Equal(t, uint(15), v)
	}
}

func TestNewInt8(t *testing.T) {
	var v int8
	p, err := param.New(&v)
	if assert.NoError(t, err) {
		p.Set("-10")
		assert.Equal(t, int8(-10), v)

		p.Set("0xf")
		assert.Equal(t, int8(15), v)
	}
}

func TestNewUint8(t *testing.T) {
	var v uint8
	p, err := param.New(&v)
	if assert.NoError(t, err) {
		p.Set("10")
		assert.Equal(t, uint8(10), v)

		p.Set("0xf")
		assert.Equal(t, uint8(15), v)
	}
}

func TestNewInt16(t *testing.T) {
	var v int16
	p, err := param.New(&v)
	if assert.NoError(t, err) {
		p.Set("-10")
		assert.Equal(t, int16(-10), v)

		p.Set("0xf")
		assert.Equal(t, int16(15), v)
	}
}

func TestNewUint16(t *testing.T) {
	var v uint16
	p, err := param.New(&v)
	if assert.NoError(t, err) {
		p.Set("10")
		assert.Equal(t, uint16(10), v)

		p.Set("0xf")
		assert.Equal(t, uint16(15), v)
	}
}

func TestNewInt32(t *testing.T) {
	var v int32
	p, err := param.New(&v)
	if assert.NoError(t, err) {
		p.Set("-10")
		assert.Equal(t, int32(-10), v)

		p.Set("0xf")
		assert.Equal(t, int32(15), v)
	}
}

func TestNewUint32(t *testing.T) {
	var v uint32
	p, err := param.New(&v)
	if assert.NoError(t, err) {
		p.Set("10")
		assert.Equal(t, uint32(10), v)

		p.Set("0xf")
		assert.Equal(t, uint32(15), v)
	}
}

func TestNewInt64(t *testing.T) {
	var v int64
	p, err := param.New(&v)
	if assert.NoError(t, err) {
		p.Set("-10")
		assert.Equal(t, int64(-10), v)

		p.Set("0xf")
		assert.Equal(t, int64(15), v)
	}
}

func TestNewUint64(t *testing.T) {
	var v uint64
	p, err := param.New(&v)
	if assert.NoError(t, err) {
		p.Set("10")
		assert.Equal(t, uint64(10), v)

		p.Set("0xf")
		assert.Equal(t, uint64(15), v)
	}
}

func TestNewFloat32(t *testing.T) {
	var v float32
	p, err := param.New(&v)
	if assert.NoError(t, err) {
		p.Set("-2.5")
		assert.Equal(t, float32(-2.5), v)
	}
}

func TestNewFloat64(t *testing.T) {
	var v float64
	p, err := param.New(&v)
	if assert.NoError(t, err) {
		p.Set("-2.5")
		assert.Equal(t, float64(-2.5), v)
	}
}

func TestNewString(t *testing.T) {
	var v string
	p, err := param.New(&v)
	if assert.NoError(t, err) {
		p.Set("foo")
		assert.Equal(t, "foo", v)
	}
}

func TestNewCustomParam(t *testing.T) {
	var v customParam
	p, err := param.New(v)
	assert.NoError(t, err)
	assert.Equal(t, v, p)
}

type customParam struct{}

func (p customParam) Set(_ string) error { return nil }

func TestNewBoolSlice(t *testing.T) {
	// A slice of bools is quite useless, yes. But is should work nevertheless.
	// This test is meant to demonstrate that the sliceParam handles arbitrary
	// sub-Param instances.

	var v []bool
	p, err := param.New(&v)
	if assert.NoError(t, err) {
		p.Set("")
		p.Set("")
		assert.Equal(t, []bool{true, true}, v)
	}
}

func TestNewStringSlice(t *testing.T) {
	var v []string
	p, err := param.New(&v)
	if assert.NoError(t, err) {
		p.Set("foo")
		p.Set("bar")
		assert.Equal(t, []string{"foo", "bar"}, v)
	}
}

func TestNewStringSliceWithExistingValues(t *testing.T) {
	v := []string{"x", "y", "z"}
	p, err := param.New(&v)
	if assert.NoError(t, err) {
		p.Set("foo")
		p.Set("bar")
		assert.Equal(t, []string{"x", "y", "z", "foo", "bar"}, v)
	}
}

func TestNewCustomParamSlice(t *testing.T) {
	var v []customParam
	p, err := param.New(&v)
	if assert.NoError(t, err) {
		p.Set("foo")
		p.Set("bar")
		assert.Equal(t, []customParam{{}, {}}, v)
	}
}

func TestNewStringPointer(t *testing.T) {
	var v *string
	p, err := param.New(&v)
	if assert.NoError(t, err) {
		p.Set("foo")

		foo := "foo"
		assert.Equal(t, &foo, v)
	}
}

func TestNewCustomParamPointer(t *testing.T) {
	var v *customParam
	p, err := param.New(&v)
	if assert.NoError(t, err) {
		p.Set("foo")
		assert.Equal(t, &customParam{}, v)
	}
}

func TestSetBadValue(t *testing.T) {
	var v int8
	p, err := param.New(&v)
	if assert.NoError(t, err) {
		err := p.Set("foo").Error()
		assert.Equal(t, "strconv.ParseInt: parsing \"foo\": invalid syntax", err)
	}
}

func TestSetBadValueSlice(t *testing.T) {
	var v []int8
	p, err := param.New(&v)
	if assert.NoError(t, err) {
		err := p.Set("foo").Error()
		assert.Equal(t, "strconv.ParseInt: parsing \"foo\": invalid syntax", err)
	}
}

func TestSetBadValuePtr(t *testing.T) {
	var v *int8
	p, err := param.New(&v)
	if assert.NoError(t, err) {
		err := p.Set("foo").Error()
		assert.Equal(t, "strconv.ParseInt: parsing \"foo\": invalid syntax", err)
	}
}
