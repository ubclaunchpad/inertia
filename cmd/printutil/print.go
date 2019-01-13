package printutil

import (
	"fmt"
	"os"
)

func Fatal(args ...interface{}) {
	fmt.Print(args...)
	println()
	os.Exit(1)
}

func Fatalf(format string, args ...interface{}) {
	fmt.Printf(format, args...)
	os.Exit(1)
}
