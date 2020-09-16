package main

import (
	"context"
	"fmt"
	"net"

	"github.com/ucarion/cli"
)

type args struct {
	Foo net.IP `cli:"--foo"`
}

func main() {
	cli.Run(context.Background(), func(ctx context.Context, args args) error {
		fmt.Printf("%#v\n", args)
		return nil
	})
}
