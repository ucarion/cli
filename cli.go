package cli

import (
	"fmt"
	"reflect"

	"github.com/ucarion/cli/internal/cmdtree"
)

// type command struct {
// 	Name string
// 	Func interface{}
// }

type configWithFunc struct {
	config reflect.Type
	fn     interface{}
}

func Run(cmds ...interface{}) {
	tree, err := cmdtree.FromFuncs(cmds)
	if err != nil {
		panic(err)
	}

	fmt.Println(tree)

	// configWithFuncs := []configWithFunc{}
	// for _, cmd := range cmds {
	// 	t := reflect.ValueOf(cmd).Type()
	// 	config := t.In(1)

	// 	configWithFuncs = append(configWithFuncs, configWithFunc{
	// 		config: config,
	// 		fn:     cmd,
	// 	})
	// }

	// // Use configWithFuncs as a queue (initialized with all the leaves of the
	// // command tree), and from this construct a mapping from parent types to
	// // their children config/fns.
	// configWithFnsByParent := map[reflect.Type][]configWithFunc{}
	// for len(configWithFuncs) != 0 {
	// 	var configWithFn configWithFunc
	// 	configWithFn, configWithFuncs = configWithFuncs[0], configWithFuncs[1:]

	// 	var parentType reflect.Type
	// 	if parent, ok := configWithFn.config.FieldByName("_"); ok {
	// 		parentType = parent.Type
	// 		configWithFnsByParent = append(configWithFnsByParent, configWithFunc{

	// 		})
	// 	}
	// }

	// configTypesQueue := []reflect.Type{}
	// for _, cmd := range cmds {
	// 	t := reflect.ValueOf(cmd).Type()
	// 	args := t.In(1)

	// 	configTypesQueue = append(configTypesQueue, args)
	// }

	// configTypes := map[reflect.Type]reflect.Type{}
	// for len(configTypesQueue) > 0 {
	// 	var t reflect.Type
	// 	t, configTypesQueue = configTypesQueue[0], configTypesQueue[1:]

	// 	if parent, ok := t.FieldByName("_"); ok {
	// 		configTypes[t] = parent.Type
	// 		configTypesQueue = append(configTypesQueue, parent.Type)
	// 	}
	// }

	// for _, cmd := range cmds {
	// 	t := reflect.ValueOf(cmd).Type()
	// 	args := t.In(1)

	// 	configTypes[args] = struct{}{}

	// 	// if parent, ok := args.FieldByName("_"); ok {
	// 	// 	config
	// 	// }
	// }

	// fmt.Println(configTypes)

	// // mapping from config types to their parents
	// typeParents := map[reflect.Type]reflect.Type{}

	// for _, cmd := range cmds {
	// 	t := reflect.ValueOf(cmd).Type()
	// 	args := t.In(1)

	// 	if parent, ok := args.FieldByName("_"); ok {
	// 		typeParents[args] = parent.Type
	// 	}
	// }

	// fmt.Println(typeParents)

	// types := map[reflect.Type]bool{}

	// for _, cmd := range cmds {
	// 	v := reflect.ValueOf(cmd)
	// 	t := v.Type()
	// 	args := t.In(1)
	// 	fmt.Println("args", args)
	// 	fmt.Println(types[args])

	// 	types[args] = true

	// 	for i := 0; i < args.NumField(); i++ {
	// 		field := args.Field(i)
	// 		fmt.Println(field.Tag)
	// 		if field.Anonymous {
	// 			fmt.Println("_", field)
	// 			fmt.Println(types[field.Type])
	// 		}
	// 	}
	// }
}
