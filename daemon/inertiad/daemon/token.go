package daemon

import (
	"net/http"

	"github.com/go-chi/render"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/crypto"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/res"
)

// tokenHandler generates a new token
func tokenHandler(w http.ResponseWriter, r *http.Request) {
	keyBytes, err := crypto.GetAPIPrivateKey(nil)
	if err != nil {
		render.Render(w, r, res.ErrInternalServer(r, "failed to get signing key", err))
		return
	}

	token, err := crypto.GenerateMasterToken(keyBytes.([]byte))
	if err != nil {
		render.Render(w, r, res.ErrInternalServer(r, "failed to generate token", err))
		return
	}

	render.Render(w, r, res.Message(r, "token generated", http.StatusOK,
		"token", token))
}
