package cmd_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/ucarion/cli/internal/cmd"
)

func TestExec(t *testing.T) {
	t.Run("flag parsing", func(t *testing.T) {
		type args struct {
			Bool     bool          `cli:"--bool,-b"`
			Bool2    bool          `cli:"-c"`
			String   string        `cli:"--string,--str,-s"`
			Duration time.Duration `cli:"--duration,-d"`
		}

		testCases := []struct {
			In  []string
			Out args
			Err error
		}{
			{
				In:  []string{},
				Out: args{},
			},
			{
				In:  []string{"--bool"},
				Out: args{Bool: true},
			},
			{
				In:  []string{"-b"},
				Out: args{Bool: true},
			},
			{
				In:  []string{"-c"},
				Out: args{Bool2: true},
			},
			{
				In:  []string{"-bc"},
				Out: args{Bool: true, Bool2: true},
			},
			{
				In:  []string{"--string=foo"},
				Out: args{String: "foo"},
			},
			{
				In:  []string{"--str=foo"},
				Out: args{String: "foo"},
			},
			{
				In:  []string{"-s=foo"},
				Out: args{String: "foo"},
			},
			{
				In:  []string{"-sb"},
				Out: args{String: "b"},
			},
			{
				In:  []string{"--string", "foo"},
				Out: args{String: "foo"},
			},
			{
				In:  []string{"--str", "foo"},
				Out: args{String: "foo"},
			},
			{
				In:  []string{"-s", "foo"},
				Out: args{String: "foo"},
			},
			{
				In:  []string{"--duration=1m"},
				Out: args{Duration: time.Minute},
			},
			{
				In:  []string{"--duration", "1m"},
				Out: args{Duration: time.Minute},
			},
			{
				In:  []string{"-d=1m"},
				Out: args{Duration: time.Minute},
			},
			{
				In:  []string{"-d", "1m"},
				Out: args{Duration: time.Minute},
			},
			{
				In:  []string{"-d1m"},
				Out: args{Duration: time.Minute},
			},
			{
				In:  []string{"-bcd1m"},
				Out: args{Bool: true, Bool2: true, Duration: time.Minute},
			},
			{
				In:  []string{"--string", "foo", "--bool", "--duration=1m"},
				Out: args{Bool: true, String: "foo", Duration: time.Minute},
			},
			{
				In:  []string{"-s=foo", "-b", "-d", "1m"},
				Out: args{Bool: true, String: "foo", Duration: time.Minute},
			},
		}

		for _, tt := range testCases {
			t.Run(fmt.Sprint(tt.In), func(t *testing.T) {
				c, err := cmd.FromFunc(func(ctx context.Context, args args) error {
					assert.Equal(t, context.Background(), ctx)
					assert.Equal(t, tt.Out, args)
					return nil
				})

				assert.NoError(t, err)
				assert.Equal(t, tt.Err, c.Exec(context.Background(), tt.In))
			})
		}
	})

	t.Run("basic", func(t *testing.T) {
		type args struct {
			Foo string `cli:"--foo"`
		}

		errF := errors.New("error from f")
		f := func(ctx context.Context, args args) error {
			assert.Equal(t, args.Foo, "bar")
			return errF
		}

		c, err := cmd.FromFunc(f)
		assert.NoError(t, err)

		err = c.Exec(context.Background(), []string{"--foo=bar"})
		assert.Equal(t, errF, err)
	})
}
