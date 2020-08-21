package param

import (
	"fmt"
	"reflect"
	"strconv"
)

type Param interface {
	Set(string) error
}

func MayTakeValue(p Param) bool {
	_, ok := p.(boolParam)
	return !ok
}

func MustTakeValue(p Param) bool {
	_, ok1 := p.(boolParam)
	_, ok2 := p.(ptrParam)
	return !ok1 && !ok2
}

func New(v interface{}) (Param, error) {
	// If the input is already a Param, just return it immediately.
	if v, ok := v.(Param); ok {
		return v, nil
	}

	t := reflect.TypeOf(v)
	if t.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("v must be a pointer")
	}

	switch t := t.Elem(); t.Kind() {
	case reflect.Slice:
		// The user inputted a pointer to a slice. That means we should use
		// sliceParam. sliceParam will instantiate Param instances from the type
		// of elements of the given slice. We should thus immediately make sure
		// that the underlying type is Param-able.
		elem := reflect.New(t.Elem()).Interface()
		if _, err := newNoSliceOrPointer(elem); err != nil {
			return nil, fmt.Errorf("unsupported slice param type: %w", err)
		}

		return sliceParam{t.Elem(), reflect.ValueOf(v)}, nil
	case reflect.Ptr:
		// The user inputted a pointer to a pointer. That means we should use
		// ptrParam.
		//
		// See comment above for slices -- a similar logic holds here to make
		// sure that the pointed-at type is Param-able.
		elem := reflect.New(t.Elem()).Interface()
		if _, err := newNoSliceOrPointer(elem); err != nil {
			return nil, fmt.Errorf("unsupported pointer param type: %w", err)
		}

		return ptrParam{t.Elem(), reflect.ValueOf(v)}, nil
	default:
		return newNoSliceOrPointer(v)
	}
}

func newNoSliceOrPointer(v interface{}) (Param, error) {
	// If the input is already a Param, just return it immediately. We support
	// this both here and in New because it's valid to have a slice or pointer
	// to a custom type.
	if v, ok := v.(Param); ok {
		return v, nil
	}

	switch v := v.(type) {
	case *bool:
		return boolParam{v}, nil
	case *int:
		return intParam{v}, nil
	case *uint:
		return uintParam{v}, nil
	case *int8:
		return int8Param{v}, nil
	case *uint8:
		return uint8Param{v}, nil
	case *int16:
		return int16Param{v}, nil
	case *uint16:
		return uint16Param{v}, nil
	case *int32:
		return int32Param{v}, nil
	case *uint32:
		return uint32Param{v}, nil
	case *int64:
		return int64Param{v}, nil
	case *uint64:
		return uint64Param{v}, nil
	case *float32:
		return float32Param{v}, nil
	case *float64:
		return float64Param{v}, nil
	case *string:
		return stringParam{v}, nil
	default:
		return nil, fmt.Errorf("unsupported param type: %v", reflect.TypeOf(v).Elem())
	}
}

type sliceParam struct {
	t reflect.Type
	v reflect.Value
}

func (p sliceParam) Set(s string) error {
	// z := new(T)
	z := reflect.New(p.t)

	// elem := New(&z)
	elem, _ := New(z.Interface())

	// Update z's value via elem's Set
	if err := elem.Set(s); err != nil {
		return err
	}

	// *v = append(*v, *z)
	p.v.Elem().Set(reflect.Append(p.v.Elem(), z.Elem()))
	return nil
}

type ptrParam struct {
	t reflect.Type
	v reflect.Value
}

func (p ptrParam) Set(s string) error {
	// z := new(T)
	z := reflect.New(p.t)

	// elem := New(&z)
	elem, _ := New(z.Interface())

	// Update z's value via elem's Set
	if err := elem.Set(s); err != nil {
		return err
	}

	// *v = z
	p.v.Elem().Set(z)
	return nil
}

type boolParam struct {
	v *bool
}

func (p boolParam) Set(_ string) error {
	*p.v = true
	return nil
}

type intParam struct {
	v *int
}

func (p intParam) Set(s string) error {
	v, err := strconv.ParseInt(s, 0, 0)
	*p.v = int(v)
	return err
}

type uintParam struct {
	v *uint
}

func (p uintParam) Set(s string) error {
	v, err := strconv.ParseUint(s, 0, 0)
	*p.v = uint(v)
	return err
}

type int8Param struct {
	v *int8
}

func (p int8Param) Set(s string) error {
	v, err := strconv.ParseInt(s, 0, 8)
	*p.v = int8(v)
	return err
}

type uint8Param struct {
	v *uint8
}

func (p uint8Param) Set(s string) error {
	v, err := strconv.ParseUint(s, 0, 8)
	*p.v = uint8(v)
	return err
}

type int16Param struct {
	v *int16
}

func (p int16Param) Set(s string) error {
	v, err := strconv.ParseInt(s, 0, 16)
	*p.v = int16(v)
	return err
}

type uint16Param struct {
	v *uint16
}

func (p uint16Param) Set(s string) error {
	v, err := strconv.ParseUint(s, 0, 16)
	*p.v = uint16(v)
	return err
}

type int32Param struct {
	v *int32
}

func (p int32Param) Set(s string) error {
	v, err := strconv.ParseInt(s, 0, 32)
	*p.v = int32(v)
	return err
}

type uint32Param struct {
	v *uint32
}

func (p uint32Param) Set(s string) error {
	v, err := strconv.ParseUint(s, 0, 32)
	*p.v = uint32(v)
	return err
}

type int64Param struct {
	v *int64
}

func (p int64Param) Set(s string) error {
	v, err := strconv.ParseInt(s, 0, 64)
	*p.v = int64(v)
	return err
}

type uint64Param struct {
	v *uint64
}

func (p uint64Param) Set(s string) error {
	v, err := strconv.ParseUint(s, 0, 64)
	*p.v = uint64(v)
	return err
}

type float32Param struct {
	v *float32
}

func (p float32Param) Set(s string) error {
	v, err := strconv.ParseFloat(s, 32)
	*p.v = float32(v)
	return err
}

type float64Param struct {
	v *float64
}

func (p float64Param) Set(s string) error {
	v, err := strconv.ParseFloat(s, 64)
	*p.v = float64(v)
	return err
}

type stringParam struct {
	v *string
}

func (p stringParam) Set(s string) error {
	*p.v = s
	return nil
}
