package handlers

import (
	"fmt"
	"goafweb"
	"goafweb/context"
	"goafweb/rand"
	"net/http"
	"time"
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
// POST /login
// TODO: Authorization: Basic header
func (uh *userHandler) Login(w http.ResponseWriter, r *http.Request) {
	var form *loginForm
	if err := readJson(r, &form); err != nil {
		writeJson(w, err, http.StatusUnprocessableEntity)
		return
	}
	user, err := uh.UserService.Authenticate(form.Email, form.Password)
	if err != nil {
		writeJson(w, err, http.StatusBadRequest)
		return
	}
	if err := uh.login(w, user); err != nil {
		writeJson(w, err, http.StatusInternalServerError)
		return
	}

	writeJson(w, user, http.StatusOK)
}

// login is a helper function to issue the user with a RememberToken via a http.Cookie and store the hashed token in the database.
// TODO: implement JTW
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

	rememberCookie := http.Cookie{
		Name:     "rememberToken",
		Value:    user.RememberToken,
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		//	Secure:   true,
	}
	http.SetCookie(w, &rememberCookie)
	return nil
}

// Logout logs a user out, removing any existing sessions/tokens
// GET /logout
// TODO: Update for JWT.
func (uh *userHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Expire old cookie
	cookie := &http.Cookie{
		Name:     "rememberToken",
		Value:    "",
		Path:     "/",
		Expires:  time.Now(),
		HttpOnly: true,
	}
	// Set expired cookie to log user out.
	http.SetCookie(w, cookie)

	// Set new rememberToken and update db.
	// If for any reason old cookie persists, it will not match database RememberToken.

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
