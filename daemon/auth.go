package daemon

import (
	"errors"
	"io/ioutil"
	"os"

	jwt "github.com/dgrijalva/jwt-go"
)

var (
	// DefaultPort defines the standard daemon port
	// TODO: Reference daemon pkg for this information?
	// We only want the package dependencies to go in one
	// direction, so best to think about how to do this.
	// Clearly cannot ask for this information over HTTP.
	DefaultPort = "8081"

	daemonGithubKeyLocation = "/app/host/.ssh/id_rsa_inertia_deploy"
)

// GetAPIPrivateKey returns the private RSA key to authenticate HTTP
// requests sent to the daemon. For now, we simply use the GitHub
// deploy key.
func GetAPIPrivateKey(*jwt.Token) (interface{}, error) {
	pemFile, err := os.Open(daemonGithubKeyLocation)
	if err != nil {
		return nil, err
	}
	key, err := getGithubKey(pemFile)
	if err != nil {
		return nil, err
	}
	return []byte(key.String()), nil
}

// GenerateToken creates a JSON Web Token (JWT) for a client to use when
// sending HTTP requests to the daemon server.
func GenerateToken(key []byte) (string, error) {
	// No claims for now.
	return jwt.New(jwt.SigningMethodHS256).SignedString(key)
}

// gitAuthFailedErr attaches the daemon key in the error message
func gitAuthFailedErr(keyloc string) error {
	bytes, err := ioutil.ReadFile(keyloc + ".pub")
	if err != nil {
		bytes = []byte(err.Error() + "\nError reading key - try running 'inertia [REMOTE] init' again: ")
	}
	return errors.New("Access to project repository rejected; did you forget to add\nInertia's deploy key to your repository settings?\n" + string(bytes[:]))
}
