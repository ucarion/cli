package main

import (
	"context"
	"fmt"
	"net/url"

	"github.com/ucarion/cli"
)

type args struct {
	Foo urlParam `cli:"--foo"`
}

type urlParam struct {
	Value url.URL
}

func (p *urlParam) Set(s string) error {
	u, err := url.Parse(s)
	if err != nil {
		return err
	}

	p.Value = *u
	return nil
}

func main() {
	cli.Run(context.Background(), func(ctx context.Context, args args) error {
		fmt.Println(args)
		return nil
	})
}
