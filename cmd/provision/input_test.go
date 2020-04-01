package provisioncmd

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_enterEC2CredentialsWalkthrough(t *testing.T) {
	tests := []struct {
		name    string
		wantID  string
		wantKey string
		wantErr bool
	}{
		{"bad ID", "", "asdf", true},
		{"bad key", "asdf", "", true},
		{"good", "asdf", "asdf", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in, err := ioutil.TempFile("", "")
			assert.NoError(t, err)
			defer in.Close()

			fmt.Fprintln(in, tt.wantID)
			fmt.Fprintln(in, tt.wantKey)

			_, err = in.Seek(0, io.SeekStart)
			assert.NoError(t, err)

			var old = os.Stdin
			os.Stdin = in
			defer func() { os.Stdin = old }()
			gotID, gotKey, err := enterEC2CredentialsWalkthrough()
			if (err != nil) != tt.wantErr {
				t.Errorf("enterEC2CredentialsWalkthrough() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if gotID != tt.wantID {
					t.Errorf("enterEC2CredentialsWalkthrough() gotId = %v, want %v", gotID, tt.wantID)
				}
				if gotKey != tt.wantKey {
					t.Errorf("enterEC2CredentialsWalkthrough() gotKey = %v, want %v", gotKey, tt.wantKey)
				}
			}
		})
	}
}
