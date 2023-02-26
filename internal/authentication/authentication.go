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
type EmptyResponse struct{}
type ErrorResponse struct {
	Error string `json:"error"`
}

type TokenResponse struct {
	Token string `json:"token"`
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

// @Summary Authenticate a user
// @Description Authenticate a user with a username and password
// @Tags Authentication
// @Accept json
// @Produce json
// @Param username body string true "Username"
// @Param password body string true "Password"
// @Success 200 {object} TokenResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /login [post]
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
	err = json.NewEncoder(w).Encode(TokenResponse{
		Token: tokenString,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// @Summary Refresh an authentication token
// @Description Refresh an existing token with a new one
// @Tags Authentication
// @Produce json
// @Param token body string true "Token"
// @Success 200 {object} TokenResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /refresh [post]
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
	err = json.NewEncoder(w).Encode(TokenResponse{
		Token: tokenString,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// @Summary Validate a token
// @Description Validate a JWT token
// @Tags Authentication
// @Produce json
// @Param token body string true "Token"
// @Success 200 {object} EmptyResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /validate [post]
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
	err := json.NewEncoder(w).Encode(EmptyResponse{})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// @Summary Register a user
// @Description Register a new user with a username and password
// @Tags Authentication
// @Accept json
// @Produce json
// @Param username body string true "Username"
// @Param password body string true "Password"
// @Success 201 {object} EmptyResponse
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /register [post]
func (api *API) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	type RegisterRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
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

// @Summary Change Password
// @Description Change a user's password
// @Tags Authentication
// @Produce json
// @Param old_password body string true "Old Password"
// @Param new_password body string true "New Password"
// @Security BearerAuth
// @Success 200 {object} EmptyResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /change-password [put]
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
