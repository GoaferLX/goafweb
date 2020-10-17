package validation

import (
	"errors"
	"fmt"
	"goafweb"
	"goafweb/hash"
	"goafweb/rand"
)

// pwResetValidator will be responsible for validation/normalizing a pwReset ready for
// database storage/retreival.
type pwResetValidator struct {
	goafweb.PwResetDB
	hmac hash.HMAC
}

// NewPwResetValidator creates a new pwResetValidator.
// It must receive something that satisfies the PwResetDB interface to satisfy
// the next layer of the interface. As well as any other arguments required for
// validation.
func NewPwResetValidator(pwrDB goafweb.PwResetDB, hmac hash.HMAC) goafweb.PwResetDB {
	return &pwResetValidator{
		PwResetDB: pwrDB,
		hmac:      hmac,
	}
}

func (pwrv *pwResetValidator) GetByToken(token string) (*goafweb.PwReset, error) {
	pwr := &goafweb.PwReset{Token: token}
	if err := runPWResetValFuncs(pwr, pwrv.tokenHashRequired); err != nil {
		return nil, fmt.Errorf("Validation Error: %w", err)
	}
	return pwrv.PwResetDB.GetByToken(pwr.TokenHash)
}

func (pwrv *pwResetValidator) Create(pwr *goafweb.PwReset) error {
	token, err := rand.RememberToken()
	if err != nil {
		return fmt.Errorf("Unable to create reset token: %w", err)
	}
	pwr.Token = token
	if err := runPWResetValFuncs(pwr, pwrv.idRequired, pwrv.tokenHashRequired); err != nil {
		return fmt.Errorf("Validation Error: %w", err)
	}
	return pwrv.PwResetDB.Create(pwr)
}

func (pwrv *pwResetValidator) Delete(id int) error {
	if id <= 0 {
		return errors.New("Invalid ID")
	}
	return pwrv.PwResetDB.Delete(id)
}

// pwResetValFunc is a uniform type for all validation functions on a pwReset.
// All validation functions will be of this type so they can be used as variadic
// arguments in other functions.
// These funtions will return a customized error message if the validation fails,
// or nil if everything is okay.
type pwResetValFunc func(pwr *goafweb.PwReset) error

func runPWResetValFuncs(pwr *goafweb.PwReset, fns ...pwResetValFunc) error {
	for _, fn := range fns {
		if err := fn(pwr); err != nil {
			return err
		}
	}
	return nil
}

func (pwrv *pwResetValidator) idRequired(pwr *goafweb.PwReset) error {
	if pwr.UserID <= 0 {
		return errors.New("ID Invalid")
	}
	return nil
}

func (pwrv *pwResetValidator) tokenHashRequired(pwr *goafweb.PwReset) error {
	if pwr.TokenHash == "" {
		if pwr.Token != "" {
			pwr.TokenHash = pwrv.hmac.Hash(pwr.Token)
			return nil
		}
		return errors.New("Remember Hash is required")
	}
	return nil
}
