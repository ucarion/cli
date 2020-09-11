package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/ucarion/cli"
)

type args struct {
	Foo string `cli:"--foo"`
	Bar string `cli:"--bar"`
}

func (a args) Autocomplete_Bar() []string {
	if a.Foo == "" {
		return nil
	}

	return []string{strings.ToUpper(a.Foo), strings.ToLower(a.Foo)}
}

func main() {
	cli.Run(context.Background(), func(ctx context.Context, args args) error {
		fmt.Println(args)
		return nil
	})
}
