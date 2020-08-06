package main

import (
	"context"
	"fmt"

	"github.com/ucarion/cli"
)

func main() {
	cli.Run(foo, baz, quux)
}

type root struct {
	_ struct{} `cli:"subcommand-tree"`
}

type fooArgs struct {
	_   root   `cli:"foo"`
	Foo string `cli:"--foo-arg"`
}

func foo(ctx context.Context, args fooArgs) error {
	fmt.Println("hello from foo", args.Foo)
	return nil
}

type bar struct {
	_ root `cli:"bar"`
}

type bazArgs struct {
	_   bar    `cli:"baz"`
	Baz string `cli:"--baz-arg"`
}

func baz(ctx context.Context, args bazArgs) error {
	fmt.Println("hello from baz", args.Baz)
	return nil
}

type quuxArgs struct {
	_    bar    `cli:"quux"`
	Quux string `cli:"--quux-arg"`
}

func quux(ctx context.Context, args quuxArgs) error {
	fmt.Println("hello from quux", args.Quux)
	return nil
}
