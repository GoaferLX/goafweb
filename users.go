package goafweb

import (
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type userService struct {
	UserDB
	pwResetDB PwResetDB
	PwPepper  string
}

func NewUserService(userDB UserDB, pwrDB PwResetDB, pwPepper string) UserService {
	return &userService{
		UserDB:    userDB,
		pwResetDB: pwrDB,
		PwPepper:  pwPepper,
	}
}

// Authenticate will match a users email/password to an exisiting database record and call Login() if details are correct.
// Will return a blanket error if email or password are incorrect.
func (us *userService) Authenticate(email, password string) (*User, error) {
	user, err := us.GetByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("Could not retreive user: %w", err)
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password+us.PwPepper))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return nil, ErrPWInvalid
		}
		return nil, fmt.Errorf("Could not authenticate: %v", err)
	}
	return user, nil
}

// InitiatePWReset will begin the process for an automated password reset.
// If email address provided is not in db provides an error.
// Otherwise creates a pwResetToken, stores in DB and emails it to the email address provided.
func (us *userService) InitiatePWReset(email string) (string, error) {
	user, err := us.GetByEmail(email)
	if err != nil {
		return "", fmt.Errorf("Could not retreive user: %w", err)
	}

	pwr := PwReset{
		UserID: user.ID,
	}
	if err := us.pwResetDB.Create(&pwr); err != nil {
		return "", fmt.Errorf("Unable to create reset token: %w", err)
	}
	return pwr.Token, nil
}

// CompletePWReset validates the token provided by the user and update the database User with a new user provided password.
// Tokens valid for 12 hours.
func (us *userService) CompletePWReset(token, newPw string) (*User, error) {
	pwr, err := us.pwResetDB.GetByToken(token)
	if err != nil {
		return nil, fmt.Errorf("Unble to retreive reset data: %w", err)
	}

	if time.Now().Sub(pwr.CreatedAt) > (12 * time.Hour) {
		return nil, errors.New("Token no longer valid")
	}
	user, err := us.GetByID(pwr.UserID)
	if err != nil {
		return nil, fmt.Errorf("Unable to reset password: %w", err)
	}
	user.Password = newPw
	if err = us.Update(user); err != nil {
		return nil, fmt.Errorf("Unable to reset password: %w", err)
	}
	us.pwResetDB.Delete(pwr.ID)
	return user, nil
}
