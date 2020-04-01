package input

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_chooseFromListWalkthrough(t *testing.T) {
	type args struct {
		optionName string
		options    []string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"bad", args{"chickens", []string{"deep fried", "baked"}}, "", true},
		{"good", args{"chickens", []string{"deep fried", "baked"}}, "baked", false},
	}
	for _, tt := range tests {
		in, err := ioutil.TempFile("", "")
		assert.NoError(t, err)
		defer in.Close()

		fmt.Fprintln(in, tt.want)

		_, err = in.Seek(0, io.SeekStart)
		assert.NoError(t, err)

		var old = os.Stdin
		os.Stdin = in
		defer func() { os.Stdin = old }()
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewPrompt(nil).
				PromptFromList(tt.args.optionName, tt.args.options).
				GetString()
			if (err != nil) != tt.wantErr {
				t.Errorf("chooseFromListWalkthrough() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got != tt.want {
					t.Errorf("chooseFromListWalkthrough() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}
