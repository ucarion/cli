package main

import (
	"context"
	"fmt"

	"github.com/ucarion/cli"
)

type args struct {
	Force   bool     `cli:"-f,--force"`
	Output  string   `cli:"-o,--output" value:"format"`
	N       int      `cli:"-n" value:"times"`
	RFC3339 bool     `cli:"--rfc3339"`
	Foo     string   `cli:"foo"`
	Bar     string   `cli:"bar"`
	Baz     []string `cli:"baz..."`
}

func (_ args) Description() string {
	return "dummy command with custom man page"
}

func (_ args) ExtendedDescription() string {
	return "This is just a program that shows you how to customize man pages."
}

func (_ args) ExtendedUsage_Force() string {
	return "Do the thing no matter what."
}

func (_ args) ExtendedUsage_Output() string {
	return "The format to output in."
}

func (_ args) ExtendedUsage_N() string {
	return "How many times to do the thing."
}

func (_ args) ExtendedUsage_RFC3339() string {
	return "Use RFC3339 timestamps."
}

func main() {
	cli.Run(context.Background(), func(ctx context.Context, args args) error {
		fmt.Println(args)
		return nil
	})
}
