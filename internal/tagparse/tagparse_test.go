package tagparse_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ucarion/cli/internal/tagparse"
)

func TestParse(t *testing.T) {
	testCases := []struct {
		In  string
		Out tagparse.ParsedTag
		Err string
	}{
		{
			In:  `json:"foobar"`,
			Out: tagparse.ParsedTag{},
		},
		{
			In:  `cli:"a,b,c"`,
			Err: "too many options in cli tag: a,b,c",
		},
		{
			In:  `cli:"foo,subcmd"`,
			Out: tagparse.ParsedTag{Kind: tagparse.KindSubcmd, CommandName: "foo"},
		},
		{
			In:  `cli:"-foo,subcmd"`,
			Err: "invalid subcommand name: -foo",
		},
		{
			In:  `cli:"-f"`,
			Out: tagparse.ParsedTag{Kind: tagparse.KindFlag, ShortFlagName: "f"},
		},
		{
			In:  `cli:"-~"`,
			Err: "invalid short flag name: -~",
		},
		{
			In:  `cli:"-ff"`,
			Err: "invalid short flag name: -ff",
		},
		{
			In:  `cli:"--foo"`,
			Out: tagparse.ParsedTag{Kind: tagparse.KindFlag, LongFlagName: "foo"},
		},
		{
			In:  `cli:"--foo~"`,
			Err: "invalid long flag name: --foo~",
		},
		{
			In:  `cli:"-f,--foo"`,
			Out: tagparse.ParsedTag{Kind: tagparse.KindFlag, ShortFlagName: "f", LongFlagName: "foo"},
		},
		{
			In:  `cli:"--foo,-f"`,
			Out: tagparse.ParsedTag{Kind: tagparse.KindFlag, ShortFlagName: "f", LongFlagName: "foo"},
		},
		{
			In:  `cli:"-f,foo"`,
			Err: "invalid flag name: foo",
		},
		{
			In:  `cli:"-f,-g"`,
			Err: "flags can only have one short form: -g",
		},
		{
			In:  `cli:"--foo,--bar"`,
			Err: "flags can only have one long form: --bar",
		},
		{
			In:  `cli:"foo"`,
			Out: tagparse.ParsedTag{Kind: tagparse.KindPosArg, PosArgName: "foo"},
		},
		{
			In:  `cli:"foo..."`,
			Out: tagparse.ParsedTag{Kind: tagparse.KindPosArg, PosArgName: "foo", IsTrailing: true},
		},
		{
			In:  `cli:"foo~"`,
			Err: "invalid positional argument name: foo~",
		},
		{
			In:  `cli:"foo~..."`,
			Err: "invalid positional argument name: foo~...",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.In, func(t *testing.T) {
			out, err := tagparse.Parse(reflect.StructTag(tt.In))
			assert.Equal(t, tt.Out, out)

			if tt.Err != "" {
				assert.Equal(t, tt.Err, err.Error())
			}
		})
	}
}
