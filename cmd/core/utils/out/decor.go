package out

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/kyokomi/emoji"
)

const (
	// EnvEmojiToggle is the environment variable used to disable or enable emoji
	EnvEmojiToggle = "INERTIA_EMOJI"
	// EnvColorToggle is the environment variable used to disable or enable colors
	EnvColorToggle = "INERTIA_COLOR"
)

// WithColor checks if colouring should be enabled
func WithColor() bool {
	var toggle = os.Getenv(EnvColorToggle)
	return toggle == "" || toggle == "true" || toggle == "on"
}

// WithEmoji checks if emoji should be enabled
func WithEmoji() bool {
	var toggle = os.Getenv(EnvEmojiToggle)
	return toggle == "" || toggle == "true" || toggle == "on"
}

// Sprintf wraps formatters
func Sprintf(format string, args ...interface{}) string {
	if WithEmoji() {
		return emoji.Sprintf(format, args...)
	}
	return fmt.Sprintf(format, args...)
}

// Printf wraps formatters
func Printf(format string, args ...interface{}) {
	if WithEmoji() {
		emoji.Printf(format, args...)
	} else {
		fmt.Printf(format, args...)
	}
}

// Println wraps formatters
func Println(args ...interface{}) {
	if WithEmoji() {
		emoji.Println(args...)
	} else {
		fmt.Println(args...)
	}
}

// Print wraps formatters
func Print(args ...interface{}) {
	if WithEmoji() {
		emoji.Print(args...)
	} else {
		fmt.Print(args...)
	}
}

// ColorTraits denotes colour customizations
type ColorTraits color.Attribute

const (
	// RD = red
	RD ColorTraits = ColorTraits(color.FgRed)
	// CY = cyan
	CY ColorTraits = ColorTraits(color.FgCyan)
	// GR = green
	GR ColorTraits = ColorTraits(color.FgGreen)
	// YE = yellow
	YE ColorTraits = ColorTraits(color.FgYellow)

	// BO = bold
	BO ColorTraits = ColorTraits(color.Bold)
	// UL = underline
	UL ColorTraits = ColorTraits(color.Underline)
)

// Colorer wraps fatih/color.Color
type Colorer struct{ c *color.Color }

// NewColorer instantiates a new Colorer
func NewColorer(traits ...ColorTraits) *Colorer {
	var attrs = make([]color.Attribute, len(traits))
	for i, t := range traits {
		attrs[i] = color.Attribute(t)
	}
	var c = color.New(attrs...)
	if WithColor() {
		c.EnableColor()
	} else {
		c.DisableColor()
	}
	return &Colorer{c}
}

// S is a shortcut for Sprint
func (c *Colorer) S(args ...interface{}) string { return c.c.Sprint(args...) }

// Sf is a shortcut for Sprintf
func (c *Colorer) Sf(f string, args ...interface{}) string { return c.c.Sprintf(f, args...) }

// Colored converts a given string to the given colour
type Colored struct {
	c *Colorer
	s string

	args []interface{}
}

// C creates a new colourable
func C(msg string, traits ...ColorTraits) *Colored {
	return &Colored{
		c: NewColorer(traits...),
		s: msg,
	}
}

// With indicates that the C should Printf with given args
func (c *Colored) With(args ...interface{}) *Colored {
	c.args = args
	return c
}

// String lets us provide a custom stringifier
func (c Colored) String() string {
	if len(c.args) > 0 {
		return c.c.c.Sprintf(c.s, c.args...)
	}
	return c.c.c.Sprint(c.s)
}
