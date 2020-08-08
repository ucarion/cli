package cmd

import (
	"context"
	"reflect"
	"strings"
	"time"
)

type Command struct {
	Flags  []Flag
	Config reflect.Type
	Func   reflect.Value
}

type Flag struct {
	LongNames  []string
	ShortNames []rune
	ShortHelp  string
	Index      []int
	TakesValue bool
}

var typeBool = reflect.TypeOf(true)

func FromFunc(fn interface{}) (Command, error) {
	v := reflect.ValueOf(fn)
	config := v.Type().In(1)

	flags := []Flag{}
	for i := 0; i < config.NumField(); i++ {
		field := config.Field(i)

		cli := field.Tag.Get("cli")
		if cli == "" {
			continue
		}

		flag := Flag{
			LongNames:  []string{},
			ShortNames: []rune{},
			ShortHelp:  field.Tag.Get("usage"),
			Index:      []int{i}, // TODO nested fields
			TakesValue: field.Type != typeBool,
		}

		for _, part := range strings.Split(cli, ",") {
			if strings.HasPrefix(part, "--") {
				flag.LongNames = append(flag.LongNames, part[2:])
			} else if strings.HasPrefix(part, "-") {
				// TODO assert there's exactly one char after the dash
				flag.ShortNames = append(flag.ShortNames, rune(part[1]))
			}
		}

		flags = append(flags, flag)
	}

	return Command{Flags: flags, Config: config, Func: v}, nil
}

func (c *Command) Exec(ctx context.Context, argv []string) error {
	config := reflect.New(c.Config).Elem()

	var parsingFlag *Flag

	for _, arg := range argv {
		if parsingFlag == nil {
			// We should expect to parse a new flag
			equalParts := strings.SplitN(arg, "=", 2)
			flagIdentifier := equalParts[0]

			if strings.HasPrefix(flagIdentifier, "--") {
				var indicatedFlag Flag

				// This is a long flag name that's being indicated.
				flagName := flagIdentifier[2:]
				for _, flag := range c.Flags {
					for _, name := range flag.LongNames {
						if flagName == name {
							indicatedFlag = flag
						}
					}
				}

				if indicatedFlag.TakesValue {
					// The flag in question takes a value. Either this arg has the
					// value, or the next one does.
					if len(equalParts) == 2 {
						// We were given something like --foo=bar. Immediately set
						// the config's value from the value in arg.
						flagValue := equalParts[1]
						field := config.FieldByIndex(indicatedFlag.Index)

						switch v := field.Addr().Interface().(type) {
						case *string:
							*v = flagValue
						case *time.Duration:
							dur, err := time.ParseDuration(flagValue)
							if err != nil {
								return err
							}

							*v = dur
						}
					} else if len(equalParts) == 1 {
						// We were given something like --foo. The next arg will
						// have the flag's value.
						parsingFlag = &indicatedFlag
					}
				} else {
					// The flag does not take a value. That can only happen if the
					// relevant config field is of type bool. Because the flag is
					// present, we set that config field to true.
					config.FieldByIndex(indicatedFlag.Index).SetBool(true)
				}
			} else if strings.HasPrefix(flagIdentifier, "-") {
				// When a an arg starts with a single dash, then multiple flags
				// may be indicated. Each character in the arg is a flag, until
				// the flag is one that takes a value. Once we reach a flag that
				// takes a value, if there are remaining characters in the arg,
				// then those characters are the flag's value. Otherwise, the
				// next arg will have the value.
				//
				// In other words, if -a and -b don't take values but -c does,
				// then "-abcdef" is equivalent to "-a -b -c def".

			iterChars:
				for i := 1; i < len(flagIdentifier); i++ {
					flagName := rune(flagIdentifier[i])
					for _, flag := range c.Flags {
						for _, shortName := range flag.ShortNames {
							if shortName == flagName {
								if flag.TakesValue {
									if i == len(flagIdentifier)-1 {
										parsingFlag = &flag
										break iterChars
									}

									field := config.FieldByIndex(flag.Index)
									flagValue := flagIdentifier[i+1:]

									switch v := field.Addr().Interface().(type) {
									case *string:
										*v = flagValue
									case *time.Duration:
										dur, err := time.ParseDuration(flagValue)
										if err != nil {
											return err
										}

										*v = dur
									}
								} else {
									config.FieldByIndex(flag.Index).SetBool(true)
								}
							}
						}
					}
				}
			}
		} else {
			// We should just expect a value.
			field := config.FieldByIndex(parsingFlag.Index)

			switch v := field.Addr().Interface().(type) {
			case *string:
				*v = arg
			case *time.Duration:
				dur, err := time.ParseDuration(arg)
				if err != nil {
					return err
				}

				*v = dur
			}

			parsingFlag = nil
		}
	}

	res := c.Func.Call([]reflect.Value{reflect.ValueOf(ctx), config})

	resErr := res[0].Interface()
	if resErr == nil {
		return nil
	}

	return resErr.(error)
}
