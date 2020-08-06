package cmd

import (
	"reflect"
)

type Command struct {
	Fn         interface{}
	Name       string
	ConfigType reflect.Type
	Flags      []Flag
}

type Flag struct {
	LongNames  []string
	ShortNames []string
	Usage      string
}

func FromFunc(fn interface{}) (Command, reflect.Type, error) {
	t := reflect.ValueOf(fn).Type()
	configType := t.In(1)

	cmd, parentType, err := FromConfig(configType)
	if err != nil {
		return Command{}, nil, err
	}

	cmd.Fn = fn
	return cmd, parentType, nil
}

func FromConfig(t reflect.Type) (Command, reflect.Type, error) {
	var name string
	var parentType reflect.Type
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		cli := field.Tag.Get("cli")

		if field.Anonymous {
			// An anonymous field that's marked with the cli tag indicates the
			// parent config type and the name of this subcommand.
			if cli != "" {
				name = cli
				parentType = field.Type
			}
		}

		// Strictly ignore (non-anonymous) fields that are not tagged with cli.
		if cli == "" {
			continue
		}
	}

	return Command{
		Name:       name,
		ConfigType: t,
		Flags:      []Flag{},
	}, parentType, nil
}
