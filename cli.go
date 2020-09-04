package cli

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ucarion/cli/internal/autocompleter"
	"github.com/ucarion/cli/internal/cmdman"
	"github.com/ucarion/cli/internal/cmdtree"
	"github.com/ucarion/cli/internal/exectree"
)

const (
	envCompleteLine = "COMP_LINE"
	envCompleteArgc = "COMP_CWORD"

	envGenerateManDir = "UCARION_CLI_GENERATE_MAN"
)

func Run(ctx context.Context, funcs ...interface{}) {
	// If we fail to build a tree from the user's given functions, then we
	// should panic. Panicking early makes an experience similar to a
	// compilation error, where the program fails very early and with a message
	// meant for the program's developer, not its end user.
	tree, err := cmdtree.New(funcs)
	if err != nil {
		panic(err)
	}

	// Try to see if we are being called as Bash autocompleter.
	completeLine := os.Getenv(envCompleteLine)
	completeArgc := os.Getenv(envCompleteArgc)

	if completeLine != "" && completeArgc != "" {
		// We are being invoked as a Bash autocompleter.
		//
		// If we encounter errors during this procedure, there isn't much we can
		// do -- there is no interface by which an autocompleter can give back
		// an error. The best we can do is be silent.
		args := strings.Fields(completeLine)
		argc, err := strconv.Atoi(completeArgc)
		if err != nil {
			return
		}

		// With the suggestions in hand, output each of them as a separate line
		// to stdout.
		for _, s := range autocompleter.Autocomplete(tree, args[:argc]) {
			fmt.Println(s)
		}

		return
	}

	// Try to see if we are being called by the user for the purposes of
	// generating the command tree's man pages.
	if dir := os.Getenv(envGenerateManDir); dir != "" {
		for name, contents := range cmdman.Man(tree, os.Args[0]) {
			filename := filepath.Join(dir, name)

			// Write the file with mode 0666. This is the default with
			// os.Create, and corresponds to a file that's universally
			// read/write-able, but not executable.
			if err := ioutil.WriteFile(filename, []byte(contents), 0666); err != nil {
				// There isn't much we can usefully do in this situation. When
				// we're being invoked in order to generate man pages, we are
				// being invoked from the CLI. So we can't return an error
				// through any normal Go mechanism.
				//
				// The important thing is that we exit with a useful error
				// message and a non-zero exit code, so that whatever shell or
				// script that's invoking us can see that man page generation
				// did not succeed.
				panic(err)
			}
		}

		return
	}

	// Run the args against the user's command tree.
	if err := exectree.Exec(ctx, tree, os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
