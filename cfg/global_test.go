package cfg

import (
	"testing"
)

func TestNewInertiaConfig(t *testing.T) {
	var c = NewInertiaConfig()
	if c == nil {
		t.Error("unexpected nil val")
	}
}
