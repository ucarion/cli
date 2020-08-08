package main

import (
	"context"
	"fmt"
	"time"

	"github.com/ucarion/cli"
)

type args struct {
	Name string        `cli:"--name" usage:"Who to greet"`
	Wait time.Duration `cli:"--wait" usage:"How long to sleep before greeting"`
}

func main() {
	cli.Run(context.Background(), func(ctx context.Context, args args) error {
		time.Sleep(args.Wait)
		fmt.Println("hello", args.Name)

		return nil
	})
}
