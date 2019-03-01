package out

import (
	"testing"

	"github.com/fatih/color"
)

func TestPrint(t *testing.T) {
	Printf("hello world %s :rocket:", "yaoharry")
	Print("hello world :rocket: ?? ")
	Println("hello world :rocket:")
	// uncomment the following to see the emoji output
	// t.Fail()
}

func TestColor(t *testing.T) {
	Printf("am I %s?\n", Color{C: color.New(color.FgRed), S: "coloured"})
	Printf(":rocket: am I %s?\n", C("coloured :rocket:", RD, UL))
	Printf(":rocket: am I %s?\n", C("coloured %s", RD, UL).With(":rocket:"))
	// uncomment the following to see the coloured output
	t.Fail()
}
