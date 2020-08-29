package command_test

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ucarion/cli/internal/command"
)

var helpFlag = command.Flag{
	IsHelp:        true,
	ShortName:     "h",
	LongName:      "help",
	Usage:         "display this help and exit",
	ExtendedUsage: "Display help message and exit.",
}

func TestFromFunc(t *testing.T) {
	type args struct{}

	called := false
	cmd, pinfo, err := command.FromFunc(func(_ context.Context, _ args) error {
		called = true
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, command.ParentInfo{}, pinfo)

	cmd.Func.Call([]reflect.Value{reflect.ValueOf(context.TODO()), reflect.ValueOf(args{})})
	assert.True(t, called)
}

func TestFromFunc_BadSignature(t *testing.T) {
	type args struct{}

	testCases := []interface{}{
		nil,
		"foo",
		func() {},
		func(_ context.Context) error { return nil },
		func(_ context.Context, _ string) error { return nil },
		func(_ string, _ args) error { return nil },
		func(_ context.Context, _ args, _ string) error { return nil },
		func(_ context.Context, _ args) {},
		func(_ context.Context, _ args) string { return "" },
	}

	for _, tt := range testCases {
		t.Run(fmt.Sprintf("%v", tt), func(t *testing.T) {
			_, _, err := command.FromFunc(tt)
			assert.True(t, strings.HasPrefix(
				err.Error(),
				"command funcs must have type: func(context.Context, T) error, got: "))
		})
	}
}

func TestFromType_Empty(t *testing.T) {
	type args struct{}

	cmd, pinfo, err := command.FromType(reflect.TypeOf(args{}))
	assert.NoError(t, err)
	assert.Equal(t, command.Command{
		Config: reflect.TypeOf(args{}),
		Flags:  []command.Flag{helpFlag},
	}, cmd)
	assert.Equal(t, command.ParentInfo{}, pinfo)
}

func TestFromType_ExistingShortHelp(t *testing.T) {
	type args struct {
		H string `cli:"-h"`
	}

	help := helpFlag
	help.ShortName = ""

	cmd, pinfo, err := command.FromType(reflect.TypeOf(args{}))
	assert.NoError(t, err)
	assert.Equal(t, command.Command{
		Config: reflect.TypeOf(args{}),
		Flags: []command.Flag{
			command.Flag{ShortName: "h", FieldIndex: []int{0}},
			help,
		},
	}, cmd)
	assert.Equal(t, command.ParentInfo{}, pinfo)
}

func TestFromType_ExistingLongHelp(t *testing.T) {
	type args struct {
		H string `cli:"--help"`
	}

	help := helpFlag
	help.LongName = ""

	cmd, pinfo, err := command.FromType(reflect.TypeOf(args{}))
	assert.NoError(t, err)
	assert.Equal(t, command.Command{
		Config: reflect.TypeOf(args{}),
		Flags: []command.Flag{
			command.Flag{LongName: "help", FieldIndex: []int{0}},
			help,
		},
	}, cmd)
	assert.Equal(t, command.ParentInfo{}, pinfo)
}

func TestFromType_ExistingShortAndLongHelp(t *testing.T) {
	type args struct {
		H1 string `cli:"--help"`
		H2 string `cli:"-h"`
	}

	cmd, pinfo, err := command.FromType(reflect.TypeOf(args{}))
	assert.NoError(t, err)
	assert.Equal(t, command.Command{
		Config: reflect.TypeOf(args{}),
		Flags: []command.Flag{
			command.Flag{LongName: "help", FieldIndex: []int{0}},
			command.Flag{ShortName: "h", FieldIndex: []int{1}},
		},
	}, cmd)
	assert.Equal(t, command.ParentInfo{}, pinfo)
}

func TestFromType_Subcmd(t *testing.T) {
	type parentArgs struct{}

	type args struct {
		X      string
		Parent parentArgs `cli:"foo,subcmd"`
	}

	cmd, pinfo, err := command.FromType(reflect.TypeOf(args{}))

	assert.NoError(t, err)

	assert.Equal(t, command.Command{
		Config: reflect.TypeOf(args{}),
		Flags:  []command.Flag{helpFlag},
	}, cmd)

	assert.Equal(t, command.ParentInfo{
		ChildName:          "foo",
		ParentType:         reflect.TypeOf(parentArgs{}),
		ParentIndexInChild: 1,
	}, pinfo)
}

func TestFromType_Flags(t *testing.T) {
	type args struct {
		A string `cli:"-a"`
		B string `cli:"--bravo"`
		C string
		D string `cli:"-d,--delta"`
	}

	cmd, pinfo, err := command.FromType(reflect.TypeOf(args{}))
	assert.NoError(t, err)
	assert.Equal(t, command.Command{
		Config: reflect.TypeOf(args{}),
		Flags: []command.Flag{
			command.Flag{ShortName: "a", FieldIndex: []int{0}},
			command.Flag{LongName: "bravo", FieldIndex: []int{1}},
			command.Flag{ShortName: "d", LongName: "delta", FieldIndex: []int{3}},
			helpFlag,
		},
	}, cmd)
	assert.Equal(t, command.ParentInfo{}, pinfo)
}

func TestFromType_EmbeddedFlags(t *testing.T) {
	type embed2 struct {
		F string `cli:"-f"`
	}

	type embed struct {
		E string `cli:"-e"`
		embed2
		G string `cli:"-g"`
	}

	type args struct {
		A string `cli:"-a"`
		B string `cli:"--bravo"`
		C string
		D string `cli:"-d,--delta"`
		embed
	}

	cmd, pinfo, err := command.FromType(reflect.TypeOf(args{}))
	assert.NoError(t, err)
	assert.Equal(t, command.Command{
		Config: reflect.TypeOf(args{}),
		Flags: []command.Flag{
			command.Flag{ShortName: "a", FieldIndex: []int{0}},
			command.Flag{LongName: "bravo", FieldIndex: []int{1}},
			command.Flag{ShortName: "d", LongName: "delta", FieldIndex: []int{3}},
			command.Flag{ShortName: "e", FieldIndex: []int{4, 0}},
			command.Flag{ShortName: "f", FieldIndex: []int{4, 1, 0}},
			command.Flag{ShortName: "g", FieldIndex: []int{4, 2}},
			helpFlag,
		},
	}, cmd)
	assert.Equal(t, command.ParentInfo{}, pinfo)
}

func TestFromType_PosArgs(t *testing.T) {
	type args struct {
		A string `cli:"a"`
		B string `cli:"...b"`
		C string
		D string `cli:"d"`
	}

	cmd, pinfo, err := command.FromType(reflect.TypeOf(args{}))
	assert.NoError(t, err)
	assert.Equal(t, command.Command{
		Config: reflect.TypeOf(args{}),
		PosArgs: []command.PosArg{
			command.PosArg{Name: "a", FieldIndex: []int{0}},
			command.PosArg{Name: "d", FieldIndex: []int{3}},
		},
		Flags:    []command.Flag{helpFlag},
		Trailing: command.PosArg{Name: "b", FieldIndex: []int{1}},
	}, cmd)
	assert.Equal(t, command.ParentInfo{}, pinfo)
}

func TestFromType_EmbeddedPosArgs(t *testing.T) {
	type embed2 struct {
		F string `cli:"...f"`
	}

	type embed struct {
		E string `cli:"e"`
		embed2
		G string `cli:"g"`
	}

	type args struct {
		A string `cli:"a"`
		B string `cli:"b"`
		C string
		D string `cli:"d"`
		embed
	}

	cmd, pinfo, err := command.FromType(reflect.TypeOf(args{}))
	assert.NoError(t, err)
	assert.Equal(t, command.Command{
		Config: reflect.TypeOf(args{}),
		PosArgs: []command.PosArg{
			command.PosArg{Name: "a", FieldIndex: []int{0}},
			command.PosArg{Name: "b", FieldIndex: []int{1}},
			command.PosArg{Name: "d", FieldIndex: []int{3}},
			command.PosArg{Name: "e", FieldIndex: []int{4, 0}},
			command.PosArg{Name: "g", FieldIndex: []int{4, 2}},
		},
		Flags:    []command.Flag{helpFlag},
		Trailing: command.PosArg{Name: "f", FieldIndex: []int{4, 1, 0}},
	}, cmd)
	assert.Equal(t, command.ParentInfo{}, pinfo)
}

func TestFromType_GamutOfTypes(t *testing.T) {
	type args struct {
		Bool          bool          `cli:"--bool"`
		Int           int           `cli:"--int"`
		Int8          int8          `cli:"--int8"`
		Int16         int16         `cli:"--int16"`
		Int32         int32         `cli:"--int32"`
		Int64         int64         `cli:"--int64"`
		Uint          uint          `cli:"--uint"`
		Uint8         uint8         `cli:"--uint8"`
		Uint16        uint16        `cli:"--uint16"`
		Uint32        uint32        `cli:"--uint32"`
		Uint64        uint64        `cli:"--uint64"`
		Float32       float32       `cli:"--float32"`
		Float64       float64       `cli:"--float64"`
		String        string        `cli:"--string"`
		Value         customValue   `cli:"--value"`
		StringArr     []string      `cli:"--string-arr"`
		ValueArr      []customValue `cli:"--value-arr"`
		OptionalValue *customValue  `cli:"--opt-value"`
		OptionalStr   *string       `cli:"--opt-string"`
	}

	_, _, err := command.FromType(reflect.TypeOf(args{}))
	assert.NoError(t, err)
}

type customValue struct{}

func (c customValue) Set(_ string) error {
	return nil
}

func TestFromType_UsageAndValueTag(t *testing.T) {
	type args struct {
		A string `cli:"-a" usage:"xxx" value:"yyy"`
	}

	cmd, pinfo, err := command.FromType(reflect.TypeOf(args{}))
	assert.NoError(t, err)
	assert.Equal(t, command.Command{
		Config: reflect.TypeOf(args{}),
		Flags: []command.Flag{
			command.Flag{ShortName: "a", Usage: "xxx", ValueName: "yyy", FieldIndex: []int{0}},
			helpFlag,
		},
	}, cmd)
	assert.Equal(t, command.ParentInfo{}, pinfo)
}

type argsWithMethods struct {
	A string `cli:"-a"`
	B string `cli:"-b"`
}

func (a argsWithMethods) Description() string {
	return "foo"
}

func (a argsWithMethods) ExtendedDescription() string {
	return "bar"
}

func (a argsWithMethods) ExtendedUsage_A() string {
	return "baz"
}

func TestFromType_Methods(t *testing.T) {
	cmd, pinfo, err := command.FromType(reflect.TypeOf(argsWithMethods{}))
	assert.NoError(t, err)
	assert.Equal(t, command.Command{
		Config:              reflect.TypeOf(argsWithMethods{}),
		Description:         "foo",
		ExtendedDescription: "bar",
		Flags: []command.Flag{
			command.Flag{ShortName: "a", ExtendedUsage: "baz", FieldIndex: []int{0}},
			command.Flag{ShortName: "b", FieldIndex: []int{1}},
			helpFlag,
		},
	}, cmd)
	assert.Equal(t, command.ParentInfo{}, pinfo)
}

func TestFromType_BadTag(t *testing.T) {
	type args struct {
		X string `cli:"_"`
	}

	_, _, err := command.FromType(reflect.TypeOf(args{}))

	assert.Equal(t,
		"invalid positional argument name: _",
		err.Error())
}

func TestFromType_BadEmbedTag(t *testing.T) {
	type embed struct {
		X string `cli:"_"`
	}

	type args struct {
		embed
	}

	_, _, err := command.FromType(reflect.TypeOf(args{}))

	assert.Equal(t,
		"invalid positional argument name: _",
		err.Error())
}

func TestFromType_StringPtrPtr(t *testing.T) {
	type args struct {
		X **string `cli:"-x"`
	}

	_, _, err := command.FromType(reflect.TypeOf(args{}))

	assert.Equal(t,
		"X: unsupported pointer param type: unsupported param type: *string",
		err.Error())
}

func TestFromType_Chan(t *testing.T) {
	type args struct {
		X chan bool `cli:"-x"`
	}

	_, _, err := command.FromType(reflect.TypeOf(args{}))

	assert.Equal(t,
		"X: unsupported param type: chan bool",
		err.Error())
}

func TestFromType_Uintptr(t *testing.T) {
	type args struct {
		X uintptr `cli:"-x"`
	}

	_, _, err := command.FromType(reflect.TypeOf(args{}))

	assert.Equal(t,
		"X: unsupported param type: uintptr",
		err.Error())
}

func TestFromType_Complex64(t *testing.T) {
	type args struct {
		X complex64 `cli:"-x"`
	}

	_, _, err := command.FromType(reflect.TypeOf(args{}))

	assert.Equal(t,
		"X: unsupported param type: complex64",
		err.Error())
}

func TestFromType_Func(t *testing.T) {
	type args struct {
		X func() `cli:"-x"`
	}

	_, _, err := command.FromType(reflect.TypeOf(args{}))

	assert.Equal(t,
		"X: unsupported param type: func()",
		err.Error())
}

func TestFromType_Map(t *testing.T) {
	type args struct {
		X map[bool]bool `cli:"-x"`
	}

	_, _, err := command.FromType(reflect.TypeOf(args{}))

	assert.Equal(t,
		"X: unsupported param type: map[bool]bool",
		err.Error())
}

func TestFromType_Interface(t *testing.T) {
	type args struct {
		X interface{} `cli:"-x"`
	}

	_, _, err := command.FromType(reflect.TypeOf(args{}))

	assert.Equal(t,
		"X: unsupported param type: interface {}",
		err.Error())
}

func TestFromType_Struct(t *testing.T) {
	type args struct {
		X struct{} `cli:"-x"`
	}

	_, _, err := command.FromType(reflect.TypeOf(args{}))

	assert.Equal(t,
		"X: unsupported param type: struct {}",
		err.Error())
}

func TestFromType_BadPosArgType(t *testing.T) {
	type args struct {
		X **string `cli:"x"`
	}

	_, _, err := command.FromType(reflect.TypeOf(args{}))

	assert.Equal(t,
		"X: unsupported pointer param type: unsupported param type: *string",
		err.Error())
}

func TestFromType_BadEmbedType(t *testing.T) {
	type embed struct {
		X **string `cli:"-x"`
	}

	type args struct {
		embed
	}

	_, _, err := command.FromType(reflect.TypeOf(args{}))

	assert.Equal(t,
		"X: unsupported pointer param type: unsupported param type: *string",
		err.Error())
}
