package restapi

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

func init() {
	registerPublicEndpoint("/api/v1/auth", authRouter())
}

func authRouter() http.Handler {
	r := chi.NewRouter()
	r.Post("/login", login)
	return r
}

type loginRequest struct {
	Token string `json:"token"`
}

type loginResponse struct {
	Token     string `json:"token"`
	ExpiresIn int    `json:"expiresIn"`
}

func login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, ErrBadRequest)
		return
	}

	// _token is package-level variable set in Start()
	if req.Token != _token {
		render.Status(r, http.StatusUnauthorized)
		render.JSON(w, r, ErrUnauthorized)
		return
	}

	render.JSON(w, r, struct {
		Success bool          `json:"success"`
		Message string        `json:"message"`
		Data    loginResponse `json:"data"`
	}{
		Success: true,
		Message: "Authentication successful",
		Data: loginResponse{
			Token:     req.Token,
			ExpiresIn: 0,
		},
	})
}
