package main

import (
	"context"
	"fmt"

	"github.com/ucarion/cli"
)

type args struct {
	Names []string `cli:"--name"`
}

func main() {
	cli.Run(context.Background(), func(ctx context.Context, args args) error {
		fmt.Println(args)
		return nil
	})
}
