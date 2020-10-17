package validation

import (
	"errors"
	"fmt"
	"goafweb"
	"goafweb/hash"
	"goafweb/rand"
	"regexp"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// userValidator will be responsible for validation/normalizing a User ready for
// database storage/retreival.
type userValidator struct {
	goafweb.UserDB
	hmac     hash.HMAC
	PwPepper string
}

// NewUserValidator creates a new userValidator.
// It must receive something that satisfies the UserDB interface to satisfy
// the next layer of the interface. As well as any other arguments required for
// validation.
func NewUserValidator(userDB goafweb.UserDB, hmac hash.HMAC, userPwPepper string) goafweb.UserDB {
	return &userValidator{
		UserDB:   userDB,
		hmac:     hmac,
		PwPepper: userPwPepper,
	}
}

func (uv *userValidator) GetByID(id int) (*goafweb.User, error) {
	user := &goafweb.User{ID: id}
	if err := runUserValFuncs(user, uv.isGreaterThan(0)); err != nil {
		return nil, fmt.Errorf("Validation Error: %w", err)
	}
	return uv.UserDB.GetByID(user.ID)
}

func (uv *userValidator) GetByEmail(email string) (*goafweb.User, error) {
	user := &goafweb.User{Email: email}
	if err := runUserValFuncs(user, uv.emailNormalize, uv.emailRequired, uv.emailFormat); err != nil {
		return nil, fmt.Errorf("Validation Error: %w", err)
	}
	return uv.UserDB.GetByEmail(user.Email)
}

// Returns an error if RememberToken is not set or is not found in database.
func (uv *userValidator) GetByRemember(token string) (*goafweb.User, error) {
	user := &goafweb.User{RememberToken: token}
	if err := runUserValFuncs(user, uv.rememberHashRequired); err != nil {
		return nil, fmt.Errorf("Validation Error: %w", err)
	}
	return uv.UserDB.GetByRemember(user.RememberHash)
}

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

// userValFunc is a uniform type for all validation functions on a User.
// All validation functions will be of this type so they can be used as variadic
// arguments in other functions.
// These funtions will return a customized error message if the validation fails,
// or nil if everything is okay.
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
