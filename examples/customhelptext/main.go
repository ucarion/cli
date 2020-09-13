package main

import (
	"context"
	"fmt"

	"github.com/ucarion/cli"
)

type args struct {
	Human   bool     `cli:"-h" usage:"show human-readable output"`
	Force   bool     `cli:"-f,--force" usage:"do the thing no matter what"`
	Output  string   `cli:"-o,--output" value:"format" usage:"the format to output in"`
	N       int      `cli:"-n" value:"times" usage:"how many times to do the thing"`
	RFC3339 bool     `cli:"--rfc3339" usage:"use rfc3339 timestamps"`
	Foo     string   `cli:"foo"`
	Bar     string   `cli:"bar"`
	Baz     []string `cli:"baz..."`
}

func (_ args) ExtendedDescription() string {
	return "This is just a program that shows you how to customize help text."
}

func main() {
	cli.Run(context.Background(), func(ctx context.Context, args args) error {
		fmt.Printf("%#v\n", args)
		return nil
	})
}
