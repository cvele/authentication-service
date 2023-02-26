package authentication

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/rs/zerolog/log"

	"github.com/cvele/authentication-service/internal/config"
	"github.com/cvele/authentication-service/internal/crypt"
	"github.com/cvele/authentication-service/internal/db"
	"github.com/cvele/authentication-service/internal/token"
)

type API struct {
	cfg *config.Config
	db  db.DB
}

func NewAPI(cfg *config.Config, db db.DB) (*API, error) {
	return &API{
		cfg: cfg,
		db:  db,
	}, nil
}

func (api *API) AuthenticateUser(username string, password string) (bool, string, error) {
	// Fetch the user's record from the database
	user, err := api.db.GetUserByUsername(username)
	if err != nil {
		return false, "", err
	}

	if crypt.CheckPasswordHash(password, user.Password) {
		return true, user.ID, nil
	}
	// Check the password against the stored hash
	return false, "", nil
}

func (api *API) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	authenticated, userID, err := api.AuthenticateUser(data.Username, data.Password)
	if err != nil || !authenticated {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	tokenString, err := token.New(userID, api.cfg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(struct {
		Token string `json:"token"`
	}{
		Token: tokenString,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (api *API) RefreshHandler(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Token string `json:"token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID, err := token.Validate(data.Token, api.cfg)
	if err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	tokenString, err := token.New(userID, api.cfg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(struct {
		Token string `json:"token"`
	}{
		Token: tokenString,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (api *API) ValidateHandler(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Token string `json:"token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if _, err := token.Validate(data.Token, api.cfg); err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(struct{}{})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

type RegisterRequestBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterResponseBody struct {
	ID string `json:"id"`
}
type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (api *API) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	// Read the request body.
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	// Check if the request body is empty.
	if len(body) == 0 {
		http.Error(w, "Request body is empty", http.StatusBadRequest)
		return
	}

	// Decode the request body.
	var req RegisterRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Error decoding request body", http.StatusBadRequest)
		return
	}

	existingUser, err := api.db.GetUserByUsername(req.Username)
	if err != nil {
		log.Error().Err(err).Msg("Error checking for existing user")
		http.Error(w, "Error checking for existing user", http.StatusInternalServerError)
		return
	}
	if existingUser != nil {
		http.Error(w, "User already exists", http.StatusConflict)
		return
	}

	// Hash the user's password before storing it in the database.
	hashedPassword, err := crypt.HashPassword(req.Password)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}

	// Hash the user's password before storing it in the database.
	id, err := api.db.NewUUID()
	if err != nil {
		http.Error(w, "Error generating uuid", http.StatusInternalServerError)
		return
	}
	// Insert the new user into the database.
	err = api.db.InsertUser(id, req.Username, string(hashedPassword))

	if err != nil {
		log.Error().Err(err).Msg("Error creating user")
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (api *API) ChangePasswordHandler(w http.ResponseWriter, r *http.Request) {
	// Read the request body.
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	// Check if the request body is empty.
	if len(body) == 0 {
		http.Error(w, "Request body is empty", http.StatusBadRequest)
		return
	}

	// Decode the request body.
	var data struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		http.Error(w, "Error decoding request body", http.StatusBadRequest)
		return
	}

	// Get the user ID from the token.
	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		http.Error(w, "Authorization header is missing", http.StatusBadRequest)
		return
	}
	userID, err := token.Validate(tokenString, api.cfg)
	if err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	// Get the user's record from the database.
	user, err := api.db.GetUserByID(userID)
	if err != nil {
		log.Error().Err(err).Msg("Error fetching user")
		http.Error(w, "Error fetching user", http.StatusInternalServerError)
		return
	}
	if user == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Check the old password against the stored hash.
	if !crypt.CheckPasswordHash(data.OldPassword, user.Password) {
		http.Error(w, "invalid old password", http.StatusUnauthorized)
		return
	}

	// Hash the new password.
	hashedPassword, err := crypt.HashPassword(data.NewPassword)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}

	// Update the user's password in the database.
	if err := api.db.UpdatePassword(userID, string(hashedPassword)); err != nil {
		log.Error().Err(err).Msg("Error updating password")
		http.Error(w, "Error updating password", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
