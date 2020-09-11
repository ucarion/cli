package main

import (
	"context"
	"fmt"

	"github.com/ucarion/cli"
)

type args struct {
	FirstName string `cli:"--first-name"`
	LastName  string `cli:"--last-name"`
}

func main() {
	cli.Run(context.Background(), func(ctx context.Context, args args) error {
		fmt.Println("hello", args.FirstName, args.LastName)
		return nil
	})
}
