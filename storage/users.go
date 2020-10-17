/*
Package storage handles any and all actions for persisiting data.
It can be used for database/file storage as appropriate, as long as the implementation
adheres to the types defined in goafweb.
*/
package storage

import (
	"errors"
	"fmt"
	"goafweb"
	"time"

	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
)

type userDB struct {
	gorm *gorm.DB
}

func NewUserDB(db *gorm.DB) goafweb.UserDB {
	return &userDB{
		gorm: db,
	}
}

// checkErr is a helper function to check dependency errors and convert them to
// app scoped errors where appropriate.
func checkErr(err error) error {
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return goafweb.ErrNotFound
		}
		return fmt.Errorf("Database error: %v", err)
	}
	return nil
}

// GetByID retrieves a User from the DB using the unique ID for lookup.
// ID provided must be greater than 0.
func (udb *userDB) GetByID(id int) (*goafweb.User, error) {
	var user goafweb.User
	err := checkErr(udb.gorm.First(&user, id).Error)
	return &user, err
}

// GetByEmail retrieves a User from the DB using their email address for lookup.
func (udb *userDB) GetByEmail(email string) (*goafweb.User, error) {
	var user goafweb.User
	err := checkErr(udb.gorm.Where("email = ?", email).First(&user).Error)
	return &user, err
}

// GetByRemember retrieves a User from the DB using their RememberToken retreived from cookies for lookup.
func (udb *userDB) GetByRemember(token string) (*goafweb.User, error) {
	var user goafweb.User
	err := checkErr(udb.gorm.Where("remember_hash = ?", token).First(&user).Error)
	return &user, err
}

// Create adds a new User to the database.
// Returns the user once its been added.
func (udb *userDB) Create(user *goafweb.User) error {
	return checkErr(udb.gorm.Create(user).Error)
}

// Update updates an existing record in the database.
func (udb *userDB) Update(user *goafweb.User) error {
	return checkErr(udb.gorm.Save(user).Error)
}

// TODO: All code below here should be relocated as not directly to do with storage.
type userService struct {
	goafweb.UserDB
	pwResetDB goafweb.PwResetDB
	PwPepper  string
}

func NewUserService(userDB goafweb.UserDB, pwrDB goafweb.PwResetDB, pwPepper string) goafweb.UserService {
	return &userService{
		UserDB:    userDB,
		pwResetDB: pwrDB,
		PwPepper:  pwPepper,
	}
}

// Authenticate will match a users email/password to an exisiting database record and call Login() if details are correct.
// Will return a blanket error if email or password are incorrect.
func (us *userService) Authenticate(email, password string) (*goafweb.User, error) {
	user, err := us.GetByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("Could not retreive user: %w", err)
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password+us.PwPepper))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return nil, goafweb.ErrPWInvalid
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

	pwr := goafweb.PwReset{
		UserID: user.ID,
	}
	if err := us.pwResetDB.Create(&pwr); err != nil {
		return "", fmt.Errorf("Unable to create reset token: %w", err)
	}
	return pwr.Token, nil
}

// CompletePWReset validates the token provided by the user and update the database User with a new user provided password.
// Tokens valid for 12 hours.
func (us *userService) CompletePWReset(token, newPw string) (*goafweb.User, error) {
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
