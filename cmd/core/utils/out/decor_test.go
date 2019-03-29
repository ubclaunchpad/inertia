package out

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrint(t *testing.T) {
	t.Run("enabled", func(t *testing.T) {
		Printf("hello world %s :rocket:", "yaoharry")
		Print("hello world :rocket: ?? ")
		Println("hello world :rocket:")
		assert.Contains(t,
			Sprintf("am I :rocket:"),
			"\U0001f680")
	})

	t.Run("disabled", func(t *testing.T) {
		os.Setenv(EnvEmojiToggle, "false")
		Printf("hello world %s :rocket:", "yaoharry")
		Print("hello world :rocket: ?? ")
		Println("hello world :rocket:")
		assert.NotContains(t,
			Sprintf("am I :rocket:"),
			"\U0001f680")
		// color should render still
		assert.Contains(t,
			Sprintf("am I %s?\n", C("coloured %s", CY, UL).With(":rocket:")),
			"[36;4m") // cyan
		os.Setenv(EnvEmojiToggle, "")
	})
}

func TestColor(t *testing.T) {
	t.Run("enabled", func(t *testing.T) {
		// emoji should render in color
		assert.Contains(t,
			Sprintf("am I %s?\n", C("coloured %s", RD, UL).With(":rocket:")),
			"\U0001f680")
		// color should render
		assert.Contains(t,
			Sprintf("am I %s?\n", C("coloured %s", CY, UL)),
			"[36;4m") // cyan
	})

	t.Run("disabled", func(t *testing.T) {
		os.Setenv(EnvColorToggle, "false")
		// emoji should render
		assert.Contains(t,
			Sprintf("am I %s?\n", C("coloured %s", RD, UL).With(":rocket:")),
			"\U0001f680")
		// color should not render
		assert.NotContains(t,
			Sprintf("am I %s?\n", C("coloured %s", CY, UL).With(":rocket:")),
			"[36;4m") // cyan
		os.Setenv(EnvColorToggle, "")
	})
}
