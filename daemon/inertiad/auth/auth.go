package auth

import (
	"io"
	"io/ioutil"
	"os"

	jwt "github.com/dgrijalva/jwt-go"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
)

var (
	// DaemonGithubKeyLocation is the default path of the generated deploy key
	DaemonGithubKeyLocation = os.Getenv("INERTIA_GH_KEY_PATH") //"/app/host/.ssh/id_rsa_inertia_deploy"
)

const (
	tokenInvalidErrorMsg = "Token invalid"
)

// GetAPIPrivateKey returns the private RSA key to authenticate HTTP
// requests sent to the daemon. For now, we simply use the GitHub
// deploy key. Retrieves from default DaemonGithubKeyLocation.
func GetAPIPrivateKey(t *jwt.Token) (interface{}, error) {
	return getAPIPrivateKeyFromPath(t, DaemonGithubKeyLocation)
}

func getAPIPrivateKeyFromPath(t *jwt.Token, path string) (interface{}, error) {
	pemFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	key, err := GetGithubKey(pemFile)
	if err != nil {
		return nil, err
	}
	return []byte(key.String()), nil
}

// GetGithubKey returns an ssh.AuthMethod from the given io.Reader
// for use with the go-git library
func GetGithubKey(pemFile io.Reader) (ssh.AuthMethod, error) {
	bytes, err := ioutil.ReadAll(pemFile)
	if err != nil {
		return nil, err
	}
	return ssh.NewPublicKeys("git", bytes, "")
}

// GenerateToken creates a JSON Web Token (JWT) for a client to use when
// sending HTTP requests to the daemon server.
func GenerateToken(key []byte) (string, error) {
	// No claims for now.
	return jwt.New(jwt.SigningMethodHS256).SignedString(key)
}
