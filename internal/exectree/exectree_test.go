package exectree_test

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ucarion/cli/internal/cmdhelp"
	"github.com/ucarion/cli/internal/cmdtree"
	"github.com/ucarion/cli/internal/exectree"
)

func TestExec_Basic(t *testing.T) {
	type rootArgs struct{}

	called := false
	tree, err := cmdtree.New([]interface{}{
		func(_ context.Context, args rootArgs) error {
			called = true
			assert.Equal(t, rootArgs{}, args)
			return nil
		},
	})

	assert.NoError(t, err)
	assert.NoError(t, exectree.Exec(context.Background(), tree, []string{"a"}))
	assert.True(t, called)
}

func TestExec_FuncError(t *testing.T) {
	type rootArgs struct{}

	tree, err := cmdtree.New([]interface{}{
		func(_ context.Context, _ rootArgs) error {
			return errors.New("dummy err")
		},
	})

	assert.NoError(t, err)
	assert.Equal(t,
		"a: dummy err",
		exectree.Exec(context.Background(), tree, []string{"a"}).Error())
}

func TestExec_ShortHelp(t *testing.T) {
	type rootArgs struct{}

	called := false
	tree, err := cmdtree.New([]interface{}{
		func(_ context.Context, args rootArgs) error {
			called = true
			return nil
		},
	})

	initialHelpOut := exectree.HelpWriter
	var helpBuf bytes.Buffer
	exectree.HelpWriter = &helpBuf
	defer func() {
		exectree.HelpWriter = initialHelpOut
	}()

	assert.NoError(t, err)
	assert.NoError(t, exectree.Exec(context.Background(), tree, []string{"./cmd", "-h"}))
	assert.False(t, called)
	assert.Equal(t, cmdhelp.Help(tree, []string{"./cmd"}), helpBuf.String())
}

func TestExec_LongHelp(t *testing.T) {
	type rootArgs struct{}

	called := false
	tree, err := cmdtree.New([]interface{}{
		func(_ context.Context, args rootArgs) error {
			called = true
			return nil
		},
	})

	initialHelpOut := exectree.HelpWriter
	var helpBuf bytes.Buffer
	exectree.HelpWriter = &helpBuf
	defer func() {
		exectree.HelpWriter = initialHelpOut
	}()

	assert.NoError(t, err)
	assert.NoError(t, exectree.Exec(context.Background(), tree, []string{"./cmd", "--help"}))
	assert.False(t, called)
	assert.Equal(t, cmdhelp.Help(tree, []string{"./cmd"}), helpBuf.String())
}

func TestExec_SubcmdHelp(t *testing.T) {
	type rootArgs struct{}
	type subArgs struct {
		Root rootArgs `cli:"sub,subcmd"`
	}

	called := false
	tree, err := cmdtree.New([]interface{}{
		func(_ context.Context, args subArgs) error {
			called = true
			return nil
		},
	})

	initialHelpOut := exectree.HelpWriter
	var helpBuf bytes.Buffer
	exectree.HelpWriter = &helpBuf
	defer func() {
		exectree.HelpWriter = initialHelpOut
	}()

	assert.NoError(t, err)
	assert.NoError(t, exectree.Exec(context.Background(), tree, []string{"cmd", "sub", "--help"}))
	assert.False(t, called)
	assert.Equal(t, cmdhelp.Help(tree.Children["sub"].CommandTree, []string{"cmd", "sub"}), helpBuf.String())
}

func TestExec_SubcmdFuncError(t *testing.T) {
	type rootArgs struct{}
	type subArgs struct {
		RootArgs rootArgs `cli:"sub,subcmd"`
	}

	tree, err := cmdtree.New([]interface{}{
		func(_ context.Context, _ subArgs) error {
			return errors.New("dummy err")
		},
	})

	assert.NoError(t, err)
	assert.Equal(t,
		"a sub: dummy err",
		exectree.Exec(context.Background(), tree, []string{"a", "sub"}).Error())
}

func TestExec_Flags(t *testing.T) {
	type args struct {
		A        bool     `cli:"-a,--alpha"`
		B        bool     `cli:"-b,--bravo"`
		C        bool     `cli:"-c,--charlie"`
		X        string   `cli:"-x,--x-ray"`
		Y        string   `cli:"-y,--yankee"`
		Z        *string  `cli:"-z,--zulu"`
		Trailing []string `cli:"...trailing"`
	}

	testCases := []struct {
		In  []string
		Out args
	}{
		// short bool flags
		{
			In:  []string{"-a"},
			Out: args{A: true},
		},
		{
			In:  []string{"-a", "-b", "-c"},
			Out: args{A: true, B: true, C: true},
		},
		{
			In:  []string{"-abc"},
			Out: args{A: true, B: true, C: true},
		},
		{
			In:  []string{"-ba", "-c"},
			Out: args{A: true, B: true, C: true},
		},

		// short string flags
		{
			In:  []string{"-x", "foo"},
			Out: args{X: "foo"},
		},
		{
			In:  []string{"-xfoo"},
			Out: args{X: "foo"},
		},
		{
			In:  []string{"-abcxfoo"},
			Out: args{A: true, B: true, C: true, X: "foo"},
		},
		{
			In:  []string{"-xabc"},
			Out: args{X: "abc"},
		},

		// short optionally-taking-value flags
		{
			In:  []string{"-z"},
			Out: args{Z: strPointer("")},
		},
		{
			In:  []string{"-zfoo"},
			Out: args{Z: strPointer("foo")},
		},

		// long bool flags
		{
			In:  []string{"--alpha"},
			Out: args{A: true},
		},
		{
			In:  []string{"--alpha", "--bravo", "--charlie"},
			Out: args{A: true, B: true, C: true},
		},

		// long string flags
		{
			In:  []string{"--x-ray", "foo"},
			Out: args{X: "foo"},
		},
		{
			In:  []string{"--x-ray=foo"},
			Out: args{X: "foo"},
		},
		{
			In:  []string{"--x-ray=foo=bar"},
			Out: args{X: "foo=bar"},
		},

		// long optionally-taking-value flags
		{
			In:  []string{"--zulu"},
			Out: args{Z: strPointer("")},
		},
		{
			In:  []string{"--zulu=foo"},
			Out: args{Z: strPointer("foo")},
		},

		// mixed
		{
			In:  []string{"--alpha", "-cxfoo", "-bz", "--yankee", "--bravo"},
			Out: args{A: true, B: true, C: true, X: "foo", Y: "--bravo", Z: strPointer("")},
		},

		// a flag taking a value of - or a value ending in -
		{
			In:  []string{"--alpha", "-cxfoo-", "-bz", "--yankee", "-"},
			Out: args{A: true, B: true, C: true, X: "foo-", Y: "-", Z: strPointer("")},
		},

		// using -- to end flags
		{
			In:  []string{"--alpha", "-cxfoo", "--", "-bz", "--yankee", "--bravo"},
			Out: args{A: true, C: true, X: "foo", Trailing: []string{"-bz", "--yankee", "--bravo"}},
		},
	}

	for _, tt := range testCases {
		t.Run(strings.Join(tt.In, " "), func(t *testing.T) {
			tree, err := cmdtree.New([]interface{}{
				func(ctx context.Context, args args) error {
					assert.Equal(t, tt.Out, args)
					return nil
				},
			})

			assert.NoError(t, err)

			// To save on typing, the entries of testCases don't have the
			// initial argv[0]. We'll add that in here.
			args := append([]string{"cmd"}, tt.In...)
			assert.NoError(t, exectree.Exec(context.Background(), tree, args))
		})
	}
}

func TestExec_UnknownFlags(t *testing.T) {
	type args struct{}

	tree, err := cmdtree.New([]interface{}{
		func(_ context.Context, _ args) error {
			return nil
		},
	})

	assert.NoError(t, err)

	assert.Equal(t, "unknown option: --foo",
		exectree.Exec(context.Background(), tree, []string{"cmd", "--foo"}).Error())
	assert.Equal(t, "unknown option: --foo",
		exectree.Exec(context.Background(), tree, []string{"cmd", "--foo=bar"}).Error())
	assert.Equal(t, "unknown option: -f",
		exectree.Exec(context.Background(), tree, []string{"cmd", "-f"}).Error())
}

func TestExec_ValueOnBoolFlag(t *testing.T) {
	type args struct {
		Foo bool `cli:"--foo"`
	}

	tree, err := cmdtree.New([]interface{}{
		func(ctx context.Context, args args) error {
			return nil
		},
	})

	assert.NoError(t, err)
	assert.Equal(t, "option --foo takes no value",
		exectree.Exec(context.TODO(), tree, []string{"cmd", "--foo=bar"}).Error())
}

func TestExec_NoValueForStringFlag(t *testing.T) {
	type args struct {
		Foo string `cli:"--foo,-f"`
	}

	tree, err := cmdtree.New([]interface{}{
		func(ctx context.Context, args args) error {
			return nil
		},
	})

	assert.NoError(t, err)

	assert.Equal(t, "option --foo requires a value",
		exectree.Exec(context.TODO(), tree, []string{"cmd", "--foo"}).Error())
	assert.Equal(t, "option -f requires a value",
		exectree.Exec(context.TODO(), tree, []string{"cmd", "-f"}).Error())
}

func TestExec_ErrSettingParamValue(t *testing.T) {
	type args struct {
		A int       `cli:"--alpha,-a"`
		B *int      `cli:"-b"`
		C *errParam `cli:"--charlie"`
		Z int       `cli:"z"`
	}

	tree, err := cmdtree.New([]interface{}{
		func(ctx context.Context, args args) error {
			return nil
		},
	})

	assert.NoError(t, err)

	assert.Equal(t, "-a: strconv.ParseInt: parsing \"X\": invalid syntax",
		exectree.Exec(context.TODO(), tree, []string{"cmd", "-aX"}).Error())
	assert.Equal(t, "-a: strconv.ParseInt: parsing \"X\": invalid syntax",
		exectree.Exec(context.TODO(), tree, []string{"cmd", "-a", "X"}).Error())
	assert.Equal(t, "--alpha: strconv.ParseInt: parsing \"X\": invalid syntax",
		exectree.Exec(context.TODO(), tree, []string{"cmd", "--alpha=X"}).Error())
	assert.Equal(t, "--alpha: strconv.ParseInt: parsing \"X\": invalid syntax",
		exectree.Exec(context.TODO(), tree, []string{"cmd", "--alpha", "X"}).Error())
	assert.Equal(t, "-b: strconv.ParseInt: parsing \"X\": invalid syntax",
		exectree.Exec(context.TODO(), tree, []string{"cmd", "-bX"}).Error())
	assert.Equal(t, "--charlie: dummy errParam err",
		exectree.Exec(context.TODO(), tree, []string{"cmd", "--charlie"}).Error())
	assert.Equal(t, "z: strconv.ParseInt: parsing \"X\": invalid syntax",
		exectree.Exec(context.TODO(), tree, []string{"cmd", "X"}).Error())
	assert.Equal(t, "z: strconv.ParseInt: parsing \"X\": invalid syntax",
		exectree.Exec(context.TODO(), tree, []string{"cmd", "--", "X"}).Error())
}

func TestExec_PosArgs(t *testing.T) {
	type args struct {
		A bool     `cli:"-a,--alpha"`
		B string   `cli:"-b,--bravo"`
		X string   `cli:"x"`
		Y string   `cli:"y"`
		Z []string `cli:"...z"`
	}

	testCases := []struct {
		In  []string
		Out args
	}{
		{
			In:  []string{"a", "b"},
			Out: args{X: "a", Y: "b"},
		},
		{
			In:  []string{"a", "b", "c"},
			Out: args{X: "a", Y: "b", Z: []string{"c"}},
		},
		{
			In:  []string{"a", "b", "c", "d", "e"},
			Out: args{X: "a", Y: "b", Z: []string{"c", "d", "e"}},
		},
		{
			In:  []string{"--", "a", "b", "c", "d", "e"},
			Out: args{X: "a", Y: "b", Z: []string{"c", "d", "e"}},
		},
		{
			In:  []string{"a", "b", "c", "--", "d", "e"},
			Out: args{X: "a", Y: "b", Z: []string{"c", "d", "e"}},
		},
		{
			In:  []string{"a", "b", "c", "d", "e", "--"},
			Out: args{X: "a", Y: "b", Z: []string{"c", "d", "e"}},
		},
		{
			In:  []string{"--", "--", "a", "b", "c", "d", "e"},
			Out: args{X: "--", Y: "a", Z: []string{"b", "c", "d", "e"}},
		},
	}

	for _, tt := range testCases {
		t.Run(strings.Join(tt.In, " "), func(t *testing.T) {
			tree, err := cmdtree.New([]interface{}{
				func(ctx context.Context, args args) error {
					assert.Equal(t, tt.Out, args)
					return nil
				},
			})

			assert.NoError(t, err)

			// To save on typing, the entries of testCases don't have the
			// initial argv[0]. We'll add that in here.
			args := append([]string{"cmd"}, tt.In...)
			assert.NoError(t, exectree.Exec(context.Background(), tree, args))
		})
	}

}

func TestExec_ExtraPosArgs(t *testing.T) {
	type args struct {
		A bool   `cli:"-a,--alpha"`
		B string `cli:"-b,--bravo"`
		X string `cli:"x"`
		Y string `cli:"y"`
	}

	tree, err := cmdtree.New([]interface{}{
		func(ctx context.Context, args args) error {
			return nil
		},
	})

	assert.NoError(t, err)
	assert.Equal(t, "unexpected argument: c",
		exectree.Exec(context.TODO(), tree, []string{"cmd", "a", "b", "c"}).Error())

	assert.NoError(t, err)
	assert.Equal(t, "unexpected argument: c",
		exectree.Exec(context.TODO(), tree, []string{"cmd", "a", "b", "--", "c"}).Error())
}

func TestExec_MissingPosArgs(t *testing.T) {
	type args struct {
		A bool   `cli:"-a,--alpha"`
		B string `cli:"-b,--bravo"`
		X string `cli:"x"`
		Y string `cli:"y"`
	}

	tree, err := cmdtree.New([]interface{}{
		func(ctx context.Context, args args) error {
			return nil
		},
	})

	assert.NoError(t, err)
	assert.Equal(t, "argument x requires a value",
		exectree.Exec(context.TODO(), tree, []string{"cmd"}).Error())
	assert.Equal(t, "argument y requires a value",
		exectree.Exec(context.TODO(), tree, []string{"cmd", "a"}).Error())
}

func TestExec_Subcmds(t *testing.T) {
	type rootArgs struct {
		A bool   `cli:"-a,--alpha"`
		B string `cli:"-b,--bravo"`
	}

	type subArgs struct {
		RootArgs rootArgs `cli:"sub,subcmd"`
		C        bool     `cli:"-c,--charlie"`
		D        string   `cli:"-d,--delta"`
		X        string   `cli:"x"`
		Y        string   `cli:"y"`
		Z        []string `cli:"...z"`
	}

	testCases := []struct {
		In  []string
		Out subArgs
	}{
		{
			In:  []string{"sub", "a", "b"},
			Out: subArgs{X: "a", Y: "b"},
		},
		{
			In:  []string{"sub", "a", "b", "c", "d", "e"},
			Out: subArgs{X: "a", Y: "b", Z: []string{"c", "d", "e"}},
		},
		{
			In:  []string{"--alpha", "--bravo=foo", "sub", "--charlie", "--delta=bar", "a", "b"},
			Out: subArgs{RootArgs: rootArgs{A: true, B: "foo"}, C: true, D: "bar", X: "a", Y: "b"},
		},
	}

	for _, tt := range testCases {
		t.Run(strings.Join(tt.In, " "), func(t *testing.T) {
			tree, err := cmdtree.New([]interface{}{
				func(ctx context.Context, args subArgs) error {
					assert.Equal(t, tt.Out, args)
					return nil
				},
			})

			assert.NoError(t, err)

			// To save on typing, the entries of testCases don't have the
			// initial argv[0]. We'll add that in here.
			args := append([]string{"cmd"}, tt.In...)
			assert.NoError(t, exectree.Exec(context.Background(), tree, args))
		})
	}
}

func TestExec_UnknownSubcmd(t *testing.T) {
	type rootArgs struct{}
	type subArgs struct {
		Root rootArgs `cli:"sub,subcmd"`
	}

	tree, err := cmdtree.New([]interface{}{
		func(_ context.Context, _ subArgs) error { return nil },
	})

	assert.NoError(t, err)
	assert.Equal(t,
		"unknown sub-command: foo",
		exectree.Exec(context.Background(), tree, []string{"cmd", "foo"}).Error())
}

func TestExec_NonExecableCommand(t *testing.T) {
	type rootArgs struct{}
	type subArgs struct {
		Root rootArgs `cli:"sub,subcmd"`
	}

	tree, err := cmdtree.New([]interface{}{
		func(_ context.Context, _ subArgs) error { return nil },
	})

	initialHelpOut := exectree.HelpWriter
	var helpBuf bytes.Buffer
	exectree.HelpWriter = &helpBuf
	defer func() {
		exectree.HelpWriter = initialHelpOut
	}()

	assert.NoError(t, err)
	assert.NoError(t, exectree.Exec(context.Background(), tree, []string{"cmd"}))
	assert.Equal(t, cmdhelp.Help(tree, []string{"cmd"}), helpBuf.String())
}

func TestExec_GamutOfTypes(t *testing.T) {
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

	testCases := []struct {
		In  []string
		Out args
	}{
		{
			In:  []string{"--bool"},
			Out: args{Bool: true},
		},
		{
			In:  []string{"--int=1"},
			Out: args{Int: 1},
		},
		{
			In:  []string{"--int=-1"},
			Out: args{Int: -1},
		},
		{
			In:  []string{"--int", "-1"},
			Out: args{Int: -1},
		},
		{
			In:  []string{"--int=0b111"},
			Out: args{Int: 7},
		},
		{
			In:  []string{"--int=0o10"},
			Out: args{Int: 8},
		},
		{
			In:  []string{"--int=0xf"},
			Out: args{Int: 15},
		},
		{
			In:  []string{"--int8=10"},
			Out: args{Int8: 10},
		},
		{
			In:  []string{"--int16=10"},
			Out: args{Int16: 10},
		},
		{
			In:  []string{"--int32=10"},
			Out: args{Int32: 10},
		},
		{
			In:  []string{"--int64=10"},
			Out: args{Int64: 10},
		},
		{
			In:  []string{"--uint=10"},
			Out: args{Uint: 10},
		},
		{
			In:  []string{"--uint8=10"},
			Out: args{Uint8: 10},
		},
		{
			In:  []string{"--uint16=10"},
			Out: args{Uint16: 10},
		},
		{
			In:  []string{"--uint32=10"},
			Out: args{Uint32: 10},
		},
		{
			In:  []string{"--uint64=10"},
			Out: args{Uint64: 10},
		},
		{
			In:  []string{"--float32=2.5"},
			Out: args{Float32: 2.5},
		},
		{
			In:  []string{"--float64=2.5"},
			Out: args{Float64: 2.5},
		},
		{
			In:  []string{"--string=xxx"},
			Out: args{String: "xxx"},
		},
		{
			In:  []string{"--value=xxx"},
			Out: args{Value: customValue{"xxx"}},
		},
		{
			In:  []string{"--string-arr=xxx", "--string-arr=yyy"},
			Out: args{StringArr: []string{"xxx", "yyy"}},
		},
		{
			In:  []string{"--value-arr=xxx", "--value-arr=yyy"},
			Out: args{ValueArr: []customValue{{"xxx"}, {"yyy"}}},
		},
		{
			In:  []string{"--opt-string=xxx"},
			Out: args{OptionalStr: strPointer("xxx")},
		},
		{
			In:  []string{"--opt-string"},
			Out: args{OptionalStr: strPointer("")},
		},
		{
			In:  []string{"--opt-value=xxx"},
			Out: args{OptionalValue: &customValue{"xxx"}},
		},
		{
			In:  []string{"--opt-value"},
			Out: args{OptionalValue: &customValue{}},
		},
	}

	for _, tt := range testCases {
		t.Run(strings.Join(tt.In, " "), func(t *testing.T) {
			tree, err := cmdtree.New([]interface{}{
				func(ctx context.Context, args args) error {
					assert.Equal(t, tt.Out, args)
					return nil
				},
			})

			assert.NoError(t, err)

			// To save on typing, the entries of testCases don't have the
			// initial argv[0]. We'll add that in here.
			args := append([]string{"cmd"}, tt.In...)
			assert.NoError(t, exectree.Exec(context.Background(), tree, args))
		})
	}
}

func strPointer(s string) *string {
	return &s
}

// customValue is an example dummy implementation of Value. It just sets its
// inner value to that of the passed string.
type customValue struct {
	Value string
}

func (c *customValue) Set(s string) error {
	c.Value = s
	return nil
}

type errParam struct{}

func (p errParam) Set(_ string) error {
	return errors.New("dummy errParam err")
}
