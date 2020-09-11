package main

import (
	"context"
	"fmt"

	"github.com/ucarion/cli"
)

type args struct {
	Foo string   `cli:"foo"`
	Bar string   `cli:"bar"`
	Baz []string `cli:"baz..."`
}

func main() {
	cli.Run(context.Background(), func(ctx context.Context, args args) error {
		fmt.Println(args)
		return nil
	})
}
