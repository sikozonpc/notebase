package user

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sikozonpc/notebase/auth"
	"github.com/sikozonpc/notebase/config"
	t "github.com/sikozonpc/notebase/types"
	u "github.com/sikozonpc/notebase/utils"
)

type Handler struct {
	store t.UserStore
}

func NewHandler(store t.UserStore) *Handler {
	return &Handler{store: store}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/users/{id}", u.MakeHTTPHandler(h.handleGetUser)).Methods("GET")
	router.HandleFunc("/login", u.MakeHTTPHandler(h.handleLogin)).Methods("POST")
	router.HandleFunc("/register", u.MakeHTTPHandler(h.handleRegister)).Methods("POST")
}

func (h *Handler) handleGetUser(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		return fmt.Errorf("method %s not allowed", r.Method)
	}

	vars := mux.Vars(r)
	id := vars["id"]

	user, err := h.store.GetUserByID(id)
	if err != nil {
		return err
	}

	return u.WriteJSON(w, http.StatusOK, user)
}

func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPost {
		return fmt.Errorf("method %s not allowed", r.Method)
	}

	payload := new(t.LoginRequest)
	if err := json.NewDecoder(r.Body).Decode(payload); err != nil {
		return err
	}

	user, err := h.store.GetUserByEmail(payload.Email)
	if err != nil {
		return err
	}

	if !auth.ComparePasswords(user.Password, []byte(payload.Password)) {
		return fmt.Errorf("invalid password or user does not exist")
	}

	token, err := createAndSetAuthCookie(user.ID, w)
	if err != nil {
		return err
	}

	return u.WriteJSON(w, http.StatusOK, token)
}

func (h *Handler) handleRegister(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPost {
		return fmt.Errorf("method %s not allowed", r.Method)
	}

	payload := new(t.RegisterRequest)
	if err := json.NewDecoder(r.Body).Decode(payload); err != nil {
		return err
	}

	hashedPassword, err := auth.HashPassword(payload.Password)
	if err != nil {
		return err
	}

	user := New(payload.FirstName, payload.LastName, payload.Email, hashedPassword)

	if err := h.store.CreateUser(*user); err != nil {
		return err
	}

	token, err := createAndSetAuthCookie(user.ID, w)
	if err != nil {
		return err
	}

	return u.WriteJSON(w, http.StatusOK, token)
}
func createAndSetAuthCookie(userID int, w http.ResponseWriter) (string, error) {
	secret := []byte(config.Envs.JWTSecret)
	token, err := auth.CreateJWT(secret, userID)
	if err != nil {
		return "", err
	}

	http.SetCookie(w, &http.Cookie{
		Name:  "Authorization",
		Value: token,
	})

	return token, nil
}
