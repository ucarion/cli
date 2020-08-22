package exectree_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/mock"
	"github.com/ucarion/cli/internal/cmdtree"
	"github.com/ucarion/cli/internal/exectree"
)

func TestExec(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		var mock mock.Mock
		defer mock.AssertExpectations(t)

		type args struct{}

		tree, err := cmdtree.New([]interface{}{
			func(ctx context.Context, args args) error {
				return mock.Called(ctx, args).Error(0)
			},
		})

		assert.NoError(t, err)

		ctx := context.TODO()
		mock.On("1", ctx, args{}).Return(nil)
		assert.Equal(t, nil, exectree.Exec(ctx, tree, []string{}))
	})

	t.Run("err from provided func", func(t *testing.T) {
		var mock mock.Mock
		defer mock.AssertExpectations(t)

		type args struct{}

		tree, err := cmdtree.New([]interface{}{
			func(ctx context.Context, args args) error {
				return mock.Called(ctx, args).Error(0)
			},
		})

		assert.NoError(t, err)

		ctx := context.TODO()
		err = errors.New("dummy round-trip error")
		mock.On("1", ctx, args{}).Return(err)
		assert.Equal(t, err, exectree.Exec(ctx, tree, []string{}))
	})

	t.Run("flags", func(t *testing.T) {
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
			Err error
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
				var mock mock.Mock
				defer mock.AssertExpectations(t)

				tree, err := cmdtree.New([]interface{}{
					func(ctx context.Context, args args) error {
						return mock.Called(ctx, args).Error(0)
					},
				})

				assert.NoError(t, err)

				ctx := context.TODO()
				mock.On("1", ctx, tt.Out).Return(nil)
				assert.Equal(t, tt.Err, exectree.Exec(ctx, tree, tt.In))
			})
		}
	})

	t.Run("unknown flags", func(t *testing.T) {
		type args struct{}

		tree, err := cmdtree.New([]interface{}{
			func(ctx context.Context, args args) error {
				return nil
			},
		})

		assert.NoError(t, err)

		assert.Equal(t, "unknown option: --foo",
			exectree.Exec(context.TODO(), tree, []string{"--foo"}).Error())
		assert.Equal(t, "unknown option: --foo",
			exectree.Exec(context.TODO(), tree, []string{"--foo=bar"}).Error())
		assert.Equal(t, "unknown option: -f",
			exectree.Exec(context.TODO(), tree, []string{"-f"}).Error())
	})

	t.Run("non-value-taking flag given value", func(t *testing.T) {
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
			exectree.Exec(context.TODO(), tree, []string{"--foo=bar"}).Error())
	})

	t.Run("value-taking flag not given value", func(t *testing.T) {
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
			exectree.Exec(context.TODO(), tree, []string{"--foo"}).Error())
		assert.Equal(t, "option -f requires a value",
			exectree.Exec(context.TODO(), tree, []string{"-f"}).Error())
	})

	t.Run("error setting param value", func(t *testing.T) {
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
			exectree.Exec(context.TODO(), tree, []string{"-aX"}).Error())
		assert.Equal(t, "-a: strconv.ParseInt: parsing \"X\": invalid syntax",
			exectree.Exec(context.TODO(), tree, []string{"-a", "X"}).Error())
		assert.Equal(t, "--alpha: strconv.ParseInt: parsing \"X\": invalid syntax",
			exectree.Exec(context.TODO(), tree, []string{"--alpha=X"}).Error())
		assert.Equal(t, "--alpha: strconv.ParseInt: parsing \"X\": invalid syntax",
			exectree.Exec(context.TODO(), tree, []string{"--alpha", "X"}).Error())
		assert.Equal(t, "-b: strconv.ParseInt: parsing \"X\": invalid syntax",
			exectree.Exec(context.TODO(), tree, []string{"-bX"}).Error())
		assert.Equal(t, "--charlie: dummy errParam err",
			exectree.Exec(context.TODO(), tree, []string{"--charlie"}).Error())
		assert.Equal(t, "z: strconv.ParseInt: parsing \"X\": invalid syntax",
			exectree.Exec(context.TODO(), tree, []string{"X"}).Error())
		assert.Equal(t, "z: strconv.ParseInt: parsing \"X\": invalid syntax",
			exectree.Exec(context.TODO(), tree, []string{"--", "X"}).Error())
	})

	t.Run("positional arguments", func(t *testing.T) {
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
			Err error
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
				var mock mock.Mock
				defer mock.AssertExpectations(t)

				tree, err := cmdtree.New([]interface{}{
					func(ctx context.Context, args args) error {
						return mock.Called(ctx, args).Error(0)
					},
				})

				assert.NoError(t, err)

				ctx := context.TODO()
				mock.On("1", ctx, tt.Out).Return(nil)
				assert.Equal(t, tt.Err, exectree.Exec(ctx, tree, tt.In))
			})
		}
	})

	t.Run("extra positional args", func(t *testing.T) {
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
		assert.Equal(t, "unexpected positional argument: c",
			exectree.Exec(context.TODO(), tree, []string{"a", "b", "c"}).Error())

		assert.NoError(t, err)
		assert.Equal(t, "unexpected positional argument: c",
			exectree.Exec(context.TODO(), tree, []string{"a", "b", "--", "c"}).Error())
	})

	t.Run("subcommands", func(t *testing.T) {
		type rootArgs struct {
			A bool   `cli:"-a,--alpha"`
			B string `cli:"-b,--bravo"`
		}

		type subArgs struct {
			RootArgs rootArgs `subcmd:"sub"`
			C        bool     `cli:"-c,--charlie"`
			D        string   `cli:"-d,--delta"`
			X        string   `cli:"x"`
			Y        string   `cli:"y"`
			Z        []string `cli:"...z"`
		}

		testCases := []struct {
			In  []string
			Out subArgs
			Err error
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
				var mock mock.Mock
				defer mock.AssertExpectations(t)

				tree, err := cmdtree.New([]interface{}{
					func(ctx context.Context, args subArgs) error {
						return mock.Called(ctx, args).Error(0)
					},
				})

				assert.NoError(t, err)

				ctx := context.TODO()
				mock.On("1", ctx, tt.Out).Return(nil)
				assert.Equal(t, tt.Err, exectree.Exec(ctx, tree, tt.In))
			})
		}
	})

	t.Run("gamut of valid types", func(t *testing.T) {
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
			Err error
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
				var mock mock.Mock
				defer mock.AssertExpectations(t)

				tree, err := cmdtree.New([]interface{}{
					func(ctx context.Context, args args) error {
						return mock.Called(ctx, args).Error(0)
					},
				})

				assert.NoError(t, err)

				ctx := context.TODO()
				mock.On("1", ctx, tt.Out).Return(nil)
				assert.Equal(t, tt.Err, exectree.Exec(ctx, tree, tt.In))
			})
		}
	})
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
