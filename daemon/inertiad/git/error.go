package git

import (
	"errors"
	"io/ioutil"

	"github.com/ubclaunchpad/inertia/daemon/inertiad/crypto"
)

// AuthFailedErr attaches the daemon key in the error message
func AuthFailedErr(path ...string) error {
	keyLoc := crypto.DaemonGithubKeyLocation
	if len(path) > 0 {
		keyLoc = path[0]
	}
	bytes, err := ioutil.ReadFile(keyLoc + ".pub")
	if err != nil {
		bytes = []byte(err.Error() + "\nError reading key - try running 'inertia [remote] init' again: ")
	}
	return errors.New("Access to project repository rejected; did you forget to add\nInertia's deploy key to your repository settings?\n" + string(bytes))
}
