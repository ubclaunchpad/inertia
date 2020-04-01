package input

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPromptInteraction_GetBool(t *testing.T) {
	type fields struct {
		conf PromptConfig
		in   string
	}
	tests := []struct {
		name    string
		fields  fields
		want    bool
		wantErr bool
	}{
		{"y", fields{PromptConfig{}, "y\n"}, true, false},
		{"N", fields{PromptConfig{}, "N\n"}, false, false},

		{"disallowed empty", fields{PromptConfig{}, "\n"}, false, true},
		{"allowed empty", fields{PromptConfig{AllowEmpty: true}, "\n"}, false, false},
		{"disallowed invalid", fields{PromptConfig{}, "asdf\n"}, false, true},
		{"allowed invalid", fields{PromptConfig{AllowInvalid: true}, "asdf\n"}, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewPromptOnInput(strings.NewReader(tt.fields.in), &tt.fields.conf).
				Prompt("test prompt (y/N)").
				GetBool()
			if (err != nil) != tt.wantErr {
				t.Errorf("PromptInteraction.GetBool() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("PromptInteraction.GetBool() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPromptInteraction_GetString(t *testing.T) {
	type fields struct {
		conf PromptConfig
		in   string
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		{"arbitrary string", fields{PromptConfig{}, "hello\n"}, "hello", false},

		{"disallowed empty", fields{PromptConfig{}, "\n"}, "", true},
		{"allowed empty", fields{PromptConfig{AllowEmpty: true}, "\n"}, "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewPromptOnInput(strings.NewReader(tt.fields.in), &tt.fields.conf).
				Promptf("test prompt %s", "hello").
				GetString()
			if (err != nil) != tt.wantErr {
				t.Errorf("PromptInteraction.GetString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("PromptInteraction.GetString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPromptInteraction_PromptFromList(t *testing.T) {
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
				t.Errorf("PromptFromList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got != tt.want {
					t.Errorf("PromptFromList() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}
