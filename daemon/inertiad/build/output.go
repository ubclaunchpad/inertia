package build

import (
	"fmt"
	"io"
)

func reportDeployInit(buildType, name string, out io.Writer) {
	fmt.Fprintf(out, "Building %s project %s...\n", buildType, name)
}

func reportProjectBuildBegin(name string, out io.Writer) {
	fmt.Fprintf(out, "Building project %s...\n", name)
}

func reportProjectBuildComplete(name string, out io.Writer) {
	fmt.Fprintf(out, "%s build successful\n", name)
}

func reportProjectContainerCreateBegin(name string, out io.Writer) {
	fmt.Fprintf(out, "Perparing %s container...\n", name)
}

func reportProjectContainerCreateComplete(name string, out io.Writer) {
	fmt.Fprintf(out, "%s container created\n", name)
}

func reportProjectStartup(name string, out io.Writer) {
	fmt.Fprintf(out, "Starting up %s...\n", name)
}
