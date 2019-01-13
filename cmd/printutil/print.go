package printutil

import (
	"fmt"
	"os"
)

// Fatal is a wrapper around fmt.Print that exits with status 1
func Fatal(args ...interface{}) {
	fmt.Print(args...)
	println()
	os.Exit(1)
}

// Fatalf is a wrapper around fmt.Printf that exits with status 1
func Fatalf(format string, args ...interface{}) {
	fmt.Printf(format, args...)
	os.Exit(1)
}
