package command

import (
	"reflect"

	"github.com/ucarion/cli/internal/tagparse"
)

type Command struct {
	Func                reflect.Value
	Config              reflect.Type
	Description         string
	ExtendedDescription string
	Flags               []Flag
	PosArgs             []PosArg
	Trailing            PosArg
}

type Flag struct {
	ShortName     string
	LongName      string
	Usage         string
	ExtendedUsage string
	ValueName     string
	FieldIndex    []int
}

type PosArg struct {
	Name       string
	FieldIndex []int
}

type ParentInfo struct {
	ChildName          string
	ParentType         reflect.Type
	ParentIndexInChild int
}

type description interface {
	Description() string
}

type extendedDescription interface {
	ExtendedDescription() string
}

const extendedUsagePrefix = "ExtendedUsage_"

func FromFunc(fn interface{}) (Command, ParentInfo, error) {
	return Command{}, ParentInfo{}, nil
}

func FromType(t reflect.Type) (Command, ParentInfo, error) {
	cmd := Command{Config: t}
	var pinfo ParentInfo

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)

		tag, err := tagparse.Parse(f.Tag)
		if err != nil {
			return Command{}, ParentInfo{}, err
		}

		if tag.Kind == tagparse.KindSubcmd {
			pinfo = ParentInfo{
				ChildName:          tag.CommandName,
				ParentType:         f.Type,
				ParentIndexInChild: i,
			}
		}
	}

	v := reflect.New(t).Interface()

	if v, ok := v.(description); ok {
		cmd.Description = v.Description()
	}

	if v, ok := v.(extendedDescription); ok {
		cmd.ExtendedDescription = v.ExtendedDescription()
	}

	return cmd, pinfo, addParams(&cmd, nil, t)
}

func addParams(cmd *Command, index []int, t reflect.Type) error {
	v := reflect.Zero(t)

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)

		if f.Anonymous {
			if err := addParams(cmd, append(index, i), f.Type); err != nil {
				return err
			}
		}

		tag, err := tagparse.Parse(f.Tag)
		if err != nil {
			return err
		}

		switch tag.Kind {
		case tagparse.KindFlag:
			var extendedUsage string
			if m, ok := t.MethodByName(extendedUsagePrefix + f.Name); ok {
				// Ensure the method has the right signature: it takes in a
				// receiver and no args, and returns just a string.
				if m.Type.NumIn() == 1 && m.Type.NumOut() == 1 && m.Type.Out(0) == reflect.TypeOf("") {
					m := v.MethodByName(extendedUsagePrefix + f.Name)
					extendedUsage = m.Call(nil)[0].Interface().(string)
				}
			}

			cmd.Flags = append(cmd.Flags, Flag{
				ShortName:     tag.ShortFlagName,
				LongName:      tag.LongFlagName,
				Usage:         tag.Usage,
				ExtendedUsage: extendedUsage,
				ValueName:     tag.FlagValueName,
				FieldIndex:    append(index, i),
			})
		case tagparse.KindPosArg:
			posArg := PosArg{
				Name:       tag.PosArgName,
				FieldIndex: append(index, i),
			}

			if tag.IsTrailing {
				cmd.Trailing = posArg
			} else {
				cmd.PosArgs = append(cmd.PosArgs, posArg)
			}
		}
	}

	return nil
}
