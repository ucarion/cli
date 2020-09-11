package main

import (
	"context"
	"fmt"

	"github.com/ucarion/cli"
)

type args struct {
	Force   bool   `cli:"-f,--force"`
	Output  string `cli:"-o,--output"`
	N       int    `cli:"-n"`
	RFC3339 bool   `cli:"--rfc3339"`
}

func main() {
	cli.Run(context.Background(), func(ctx context.Context, args args) error {
		fmt.Println(args)
		return nil
	})
}
