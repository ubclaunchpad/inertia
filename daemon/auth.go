package daemon

import (
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
)

var (
	// DefaultPort defines the standard daemon port
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

// authorized is a function decorator for authorizing RESTful
// daemon requests. It wraps handler functions and ensures the
// request is authorized. Returns a function
func authorized(handler http.HandlerFunc, keyLookup func(*jwt.Token) (interface{}, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Collect the token from the header.
		bearerString := r.Header.Get("Authorization")

		// Split out the actual token from the header.
		splitToken := strings.Split(bearerString, "Bearer ")
		if len(splitToken) < 2 {
			http.Error(w, malformedAuthStringErrorMsg, http.StatusForbidden)
			return
		}
		tokenString := splitToken[1]

		// Parse takes the token string and a function for looking up the key.
		token, err := jwt.Parse(tokenString, keyLookup)
		if err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}

		// Verify the claims (none for now) and token.
		if _, ok := token.Claims.(jwt.MapClaims); !ok || !token.Valid {
			http.Error(w, tokenInvalidErrorMsg, http.StatusForbidden)
			return
		}

		// We're authorized, run the handler.
		handler(w, r)
	}
}
