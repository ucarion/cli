package exectree_test

import (
	"context"
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

	t.Run("flags", func(t *testing.T) {
		type args struct {
			A bool    `cli:"-a,--alpha"`
			B bool    `cli:"-b,--bravo"`
			C bool    `cli:"-c,--charlie"`
			X string  `cli:"-x,--x-ray"`
			Y string  `cli:"-y,--yankee"`
			Z *string `cli:"-z,--zulu"`
		}

		type testCase struct {
			In  []string
			Out args
			Err error
		}

		testCases := []testCase{
			// short bool flags
			testCase{
				In:  []string{"-a"},
				Out: args{A: true},
			},
			testCase{
				In:  []string{"-a", "-b", "-c"},
				Out: args{A: true, B: true, C: true},
			},
			testCase{
				In:  []string{"-abc"},
				Out: args{A: true, B: true, C: true},
			},
			testCase{
				In:  []string{"-ba", "-c"},
				Out: args{A: true, B: true, C: true},
			},

			// short string flags
			testCase{
				In:  []string{"-x", "foo"},
				Out: args{X: "foo"},
			},
			testCase{
				In:  []string{"-xfoo"},
				Out: args{X: "foo"},
			},
			testCase{
				In:  []string{"-abcxfoo"},
				Out: args{A: true, B: true, C: true, X: "foo"},
			},
			testCase{
				In:  []string{"-xabc"},
				Out: args{X: "abc"},
			},

			// short optionally-taking-value flags
			testCase{
				In:  []string{"-z"},
				Out: args{Z: strPointer("")},
			},
			testCase{
				In:  []string{"-zfoo"},
				Out: args{Z: strPointer("foo")},
			},

			// long bool flags
			testCase{
				In:  []string{"--alpha"},
				Out: args{A: true},
			},
			testCase{
				In:  []string{"--alpha", "--bravo", "--charlie"},
				Out: args{A: true, B: true, C: true},
			},

			// long string flags
			testCase{
				In:  []string{"--x-ray", "foo"},
				Out: args{X: "foo"},
			},
			testCase{
				In:  []string{"--x-ray=foo"},
				Out: args{X: "foo"},
			},
			testCase{
				In:  []string{"--x-ray=foo=bar"},
				Out: args{X: "foo=bar"},
			},

			// long optionally-taking-value flags
			testCase{
				In:  []string{"--zulu"},
				Out: args{Z: strPointer("")},
			},
			testCase{
				In:  []string{"--zulu=foo"},
				Out: args{Z: strPointer("foo")},
			},

			// mixed
			testCase{
				In:  []string{"--alpha", "-cxfoo", "-bz", "--yankee", "--bravo"},
				Out: args{A: true, B: true, C: true, X: "foo", Y: "--bravo", Z: strPointer("")},
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
