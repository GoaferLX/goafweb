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
	"regexp"
	"strings"
	"time"

	"goafweb/hash"
	"goafweb/rand"

	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
)

type userService struct {
	goafweb.UserDB
	pwResetService pwResetService
	PwPepper       string
}

type userValidator struct {
	goafweb.UserDB
	hmac     hash.HMAC
	PwPepper string
}

type userDB struct {
	gorm *gorm.DB
}

func NewUserService(db *gorm.DB, userPwPepper, hmacSecretKey string) goafweb.UserService {
	hmac := hash.NewHMAC(hmacSecretKey)
	return &userService{
		UserDB: &userValidator{
			UserDB: &userDB{
				gorm: db,
			},
			hmac:     hmac,
			PwPepper: userPwPepper,
		},
		PwPepper:       userPwPepper,
		pwResetService: newPWResetService(db, hmac),
	}
}

// checkErr is a helper function to check dependency errors and convert them to
// app scoped errors where appropriate.
func checkErr(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return goafweb.ErrNotFound
	}
	return fmt.Errorf("Database error: %v", err)
}

// GetByID retrieves a User from the DB using the unique ID for lookup.
// ID provided must be greater than 0.
func (uv *userValidator) GetByID(id int) (*goafweb.User, error) {
	user := &goafweb.User{ID: id}
	if err := runUserValFuncs(user, uv.isGreaterThan(0)); err != nil {
		return nil, fmt.Errorf("Validation Error: %w", err)
	}
	return uv.UserDB.GetByID(user.ID)
}

func (udb *userDB) GetByID(id int) (*goafweb.User, error) {
	var user goafweb.User
	err := checkErr(udb.gorm.First(&user, id).Error)
	return &user, err
}

// GetByEmail retrieves a User from the DB using their email address for lookup.
func (uv *userValidator) GetByEmail(email string) (*goafweb.User, error) {
	user := &goafweb.User{Email: email}
	if err := runUserValFuncs(user, uv.emailNormalize, uv.emailRequired, uv.emailFormat); err != nil {
		return nil, fmt.Errorf("Validation Error: %w", err)
	}
	return uv.UserDB.GetByEmail(user.Email)
}

func (udb *userDB) GetByEmail(email string) (*goafweb.User, error) {
	var user goafweb.User
	err := checkErr(udb.gorm.Where("email = ?", email).First(&user).Error)
	return &user, err
}

// GetByRemember retrieves a User from the DB using their RememberToken retreived from cookies for lookup.
// Returns an error if RememberToken is not set or is not found in database.
func (uv *userValidator) GetByRemember(token string) (*goafweb.User, error) {
	user := &goafweb.User{RememberToken: token}
	if err := runUserValFuncs(user, uv.rememberHashRequired); err != nil {
		return nil, fmt.Errorf("Validation Error: %w", err)
	}
	return uv.UserDB.GetByRemember(user.RememberHash)
}

func (udb *userDB) GetByRemember(token string) (*goafweb.User, error) {
	var user goafweb.User
	err := checkErr(udb.gorm.Where("remember_hash = ?", token).First(&user).Error)
	return &user, err
}

// Create adds a new User to the database.
// Returns the user once its been added.
// User password cleared from memory after storing - only hash is stored.
func (uv *userValidator) Create(user *goafweb.User) error {
	if err := runUserValFuncs(user,
		uv.emailNormalize,
		uv.emailRequired,
		uv.emailFormat,
		uv.emailIsAvail,
		uv.passwordRequired,
		uv.passwordMinLength,
		uv.passwordBcrypt,
		uv.passwordHashRequired,
		uv.setRememberToken,
		uv.rememberHashRequired,
	); err != nil {
		return fmt.Errorf("Validation Error: %w", err)
	}

	return uv.UserDB.Create(user)

}
func (udb *userDB) Create(user *goafweb.User) error {
	return checkErr(udb.gorm.Create(user).Error)
}

// Update updates an existing record in the database.
// Password is not required as user might not update password, but PasswordHash is required.
// If PasswordHash is not record being updated, it didn't come from our db and may be a fradulent request.
func (uv *userValidator) Update(user *goafweb.User) error {
	if err := runUserValFuncs(user,
		uv.emailNormalize,
		uv.emailRequired,
		uv.emailFormat,
		uv.passwordBcrypt,
		uv.passwordHashRequired,
		uv.rememberHashRequired,
	); err != nil {
		return fmt.Errorf("Validation Error: %w", err)
	}
	return uv.UserDB.Update(user)
}
func (udb *userDB) Update(user *goafweb.User) error {
	return checkErr(udb.gorm.Save(user).Error)
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

	pwr := pwReset{
		UserID: user.ID,
	}
	if err := us.pwResetService.Create(&pwr); err != nil {
		return "", fmt.Errorf("Unable to create reset token: %w", err)
	}
	return pwr.Token, nil
}

// CompletePWReset validates the token provided by the user and update the database User with a new user provided password.
// Tokens valid for 12 hours.
func (us *userService) CompletePWReset(token, newPw string) (*goafweb.User, error) {
	pwr, err := us.pwResetService.GetByToken(token)
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
	us.pwResetService.Delete(pwr.ID)
	return user, nil
}

// Uniform type for all validation functions on a User
type userValFunc func(user *goafweb.User) error

func runUserValFuncs(user *goafweb.User, fns ...userValFunc) error {
	for _, fn := range fns {
		if err := fn(user); err != nil {
			return err
		}
	}
	return nil
}

func (uc *userValidator) emailNormalize(user *goafweb.User) error {
	user.Email = strings.TrimSpace(user.Email)
	user.Email = strings.ToLower(user.Email)
	return nil
}
func (uv *userValidator) emailRequired(user *goafweb.User) error {
	if user.Email == "" {
		return errors.New("Email address is required")
	}
	return nil
}

func (uv *userValidator) emailFormat(user *goafweb.User) error {
	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,16}$`)
	if !emailRegex.MatchString(user.Email) {
		return errors.New("Email is not a valid format")
	}
	return nil
}

func (uv *userValidator) emailIsAvail(user *goafweb.User) error {
	_, err := uv.GetByEmail(user.Email)
	if err != nil {
		// If ErrRecordNotFound then email address is available
		if errors.Is(err, goafweb.ErrNotFound) {
			return nil
		}
		// Any other errors suggests db error and gets returned.
		return err
	}
	// No errors mean the address was found and is unavilable.
	return errors.New("That email address is already taken")
}
func (uv *userValidator) setRememberToken(user *goafweb.User) error {
	if user.RememberToken != "" {
		return nil
	}
	token, err := rand.RememberToken()
	if err != nil {
		return fmt.Errorf("Unable to generate token: %w", err)
	}
	user.RememberToken = token
	return nil
}

// rememberHashRequired will set a new RememberHash if field is empty but the User has a Token.
// If RememberHash is empty and no Token is available returns an error.
func (uv *userValidator) rememberHashRequired(user *goafweb.User) error {
	if user.RememberToken != "" {
		user.RememberHash = uv.hmac.Hash(user.RememberToken)
	}
	if user.RememberHash == "" {
		return errors.New("remember hash required")
	}
	return nil

}

func (uv *userValidator) isGreaterThan(n int) userValFunc {
	return userValFunc(func(user *goafweb.User) error {
		if user.ID <= n {
			return errors.New("Invalid ID")
		}
		return nil
	})
}

func (uv *userValidator) passwordRequired(user *goafweb.User) error {
	if strings.ToLower(user.Password) == "" {
		return errors.New("Password is required")
	}
	return nil
}

func (uv *userValidator) passwordMinLength(user *goafweb.User) error {
	if len(user.Password) < 8 {
		return errors.New("Password is too short")
	}
	return nil
}

// TODO: func pwFormat() -  regex for more "secure" passwords

// passwordBcrypt will hash a Password.
// When updating a user with Update(), password may not be required i.e. updating email address but Hash is required as they must be logged in to do so.
// Therefore passwordRequired is not run in Update().
// If Password is being updated and is not set/is blank it will not be Hashed and passwordHashRequired will fail.
func (uv *userValidator) passwordBcrypt(user *goafweb.User) error {
	// Will not set a new PasswordHash if the password field is blank in case user is not updating password.
	if user.Password == "" {
		return nil
	}
	pwhash, err := bcrypt.GenerateFromPassword([]byte(user.Password+uv.PwPepper), bcrypt.DefaultCost)
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return goafweb.ErrPWInvalid
		}
		return fmt.Errorf("Could not hash password: %v", err)
	}
	user.PasswordHash = string(pwhash)
	user.Password = ""
	return nil
}

// PasswordHash checks that a Hash exists.
// If user is logged in they will already have a RememberHash.
func (uv *userValidator) passwordHashRequired(user *goafweb.User) error {
	if user.PasswordHash == "" {
		return errors.New("Password is required")
	}
	return nil
}
