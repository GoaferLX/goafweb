package handlers

import (
	"errors"
	"fmt"
	"goafweb"
	"goafweb/context"
	"goafweb/rand"
	"net/http"
)

type userHandler struct {
	UserService  goafweb.UserService
	EmailService goafweb.MailService
}

func NewUsers(us goafweb.UserService, emailer goafweb.MailService) *userHandler {
	return &userHandler{
		UserService:  us,
		EmailService: emailer,
	}
}

type resetPWForm struct {
	Email    string `schema:"email"`
	Password string `schema:"password"`
	Token    string `schema:"token"`
}
type loginForm struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// POST /signup
// Create processes a new user and adds to database if okay
// Reloads signup page and returns errors if now
func (uh *userHandler) Create(w http.ResponseWriter, r *http.Request) {
	var user goafweb.User

	if err := readJson(r, &user); err != nil {
		writeJson(w, err, http.StatusUnprocessableEntity)
		return
	}
	if err := uh.UserService.Create(&user); err != nil {
		writeJson(w, err, http.StatusBadRequest)
		return
	}
	writeJson(w, user, http.StatusCreated)

}

// Login processes a users login attempt, authenticates the User and logs them in or returns an error.
// It expects the authentication request to come via an Basic Authorization header.
// It processes the header and authenticates the user.
// It returns http.StatusUnauthorized if header is not set or authentication details are incorrect,
// otherwise returns http.StatusOK and a RememberToken.
// POST /login
func (uh *userHandler) Login(w http.ResponseWriter, r *http.Request) {
	email, password, ok := r.BasicAuth()
	if !ok {
		w.Header().Set("WWW-Authenticate", "Basic realm=\"Access to goafweb\"")
		writeJson(w, "Please provide authentication details", http.StatusUnauthorized)
		return
	}
	user, err := uh.UserService.Authenticate(email, password)
	if err != nil {
		// If user not found or password is invalid return a general authentication error so
		// validity of email is not exposed. Otherwise return error.
		if errors.Is(err, goafweb.ErrNotFound) || errors.Is(err, goafweb.ErrPWInvalid) {
			w.Header().Set("WWW-Authenticate", "Basic realm=\"Access to goafweb\"")
			writeJson(w, goafweb.ErrAuth, http.StatusUnauthorized)
			return
		}
		w.Header().Set("WWW-Authenticate", "Basic realm=\"Access to goafweb\"")
		writeJson(w, err, http.StatusUnauthorized)
		return
	}
	// Authentication okay - issue new rememberToken
	if err := uh.login(w, user); err != nil {
		writeJson(w, err, http.StatusInternalServerError)
		return
	}
	// TODO: Implement this as an oauth2 token?
	writeJson(w, user.RememberToken, http.StatusOK)
}

// login is a helper function to set tokens once the user has been authenticated.
// It issues the user with a RememberToken and stores the hashed token in the database.
// The token can then be issued to the user by the function that calls login().
func (uh *userHandler) login(w http.ResponseWriter, user *goafweb.User) error {
	if user.RememberToken == "" {
		token, err := rand.RememberToken()
		if err != nil {
			return err
		}
		user.RememberToken = token
		err = uh.UserService.Update(user)
		if err != nil {
			return fmt.Errorf("Could not login: %w", err)
		}
	}
	return nil
}

// Logout logs a user out, removing any existing sessions/tokens
// GET /logout
func (uh *userHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Set a new rememberToken and updates db.
	// This prevents user from being able to use any exisiting tokens as it will not match database RememberToken.
	token, _ := rand.RememberToken()
	user := context.GetUser(r.Context())

	user.RememberToken = token
	if err := uh.UserService.Update(user); err != nil {
		writeJson(w, err, http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// Forgot will initiate the process for resetting a users password.
// It will create a reset token which is stored in the database and emailed to the user.
// POST /forgot.
func (uh *userHandler) Forgot(w http.ResponseWriter, r *http.Request) {
	var email string
	if err := readJson(r, &email); err != nil {
		writeJson(w, err, http.StatusUnprocessableEntity)
		return
	}
	token, err := uh.UserService.InitiatePWReset(email)
	if err != nil {
		writeJson(w, err, http.StatusBadRequest)
		return
	}
	if err := uh.EmailService.ResetPw(email, token); err != nil {
		writeJson(w, err, http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)

}

// Reset will process a users password reset, processing the provided token and
// return any errors if they exist.
// POST /reset.
func (uh *userHandler) Reset(w http.ResponseWriter, r *http.Request) {

	var form resetPWForm
	if err := readJson(r, &form); err != nil {
		writeJson(w, err, http.StatusUnprocessableEntity)
		return
	}
	user, err := uh.UserService.CompletePWReset(form.Token, form.Password)
	if err != nil {
		writeJson(w, err, http.StatusBadRequest)
		return
	}
	uh.login(w, user)
	w.WriteHeader(http.StatusOK)
}
