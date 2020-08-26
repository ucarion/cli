package command_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ucarion/cli/internal/command"
)

func TestFromType_Empty(t *testing.T) {
	type args struct{}

	cmd, pinfo, err := command.FromType(reflect.TypeOf(args{}))
	assert.NoError(t, err)
	assert.Equal(t, command.Command{
		Config: reflect.TypeOf(args{}),
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
		Trailing: command.PosArg{Name: "f", FieldIndex: []int{4, 1, 0}},
	}, cmd)
	assert.Equal(t, command.ParentInfo{}, pinfo)
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
		},
	}, cmd)
	assert.Equal(t, command.ParentInfo{}, pinfo)
}
