package out

import (
	"os"
)

// Fatal is a wrapper around fmt.Print that exits with status 1
func Fatal(args ...interface{}) {
	Println(args...)
	os.Exit(1)
}

// Fatalf is a wrapper around out.Printf that exits with status 1
func Fatalf(format string, args ...interface{}) {
	Printf(format, args...)
	os.Exit(1)
}
