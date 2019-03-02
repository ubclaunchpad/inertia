package out

import (
	// use bobheadxi/emoji until https://github.com/kyokomi/emoji/pull/32 gets
	// merged, for Stringer compatibility for the Colour class

	"fmt"
	"os"

	"github.com/bobheadxi/emoji"
	"github.com/fatih/color"
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

// Color converts a given string to the given colour
type Color struct {
	C *color.Color
	S string

	args []interface{}
}

// C creates a new colourable
func C(msg string, traits ...ColorTraits) *Color {
	var attrs = make([]color.Attribute, len(traits))
	for i, t := range traits {
		attrs[i] = color.Attribute(t)
	}
	return &Color{
		C: color.New(attrs...),
		S: msg,
	}
}

// With indicates that the C should Printf with given args
func (c *Color) With(args ...interface{}) *Color {
	c.args = args
	return c
}

// String lets us provide a custom stringifier
func (c Color) String() string {
	if WithColor() {
		c.C.EnableColor()
	} else {
		c.C.DisableColor()
	}
	if len(c.args) > 0 {
		return c.C.Sprintf(c.S, c.args...)
	}
	return c.C.Sprint(c.S)
}
