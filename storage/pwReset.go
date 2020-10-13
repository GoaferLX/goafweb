package storage

import (
	"errors"
	"fmt"
	"time"

	"goafweb/hash"
	"goafweb/rand"

	"github.com/jinzhu/gorm"
)

type pwReset struct {
	ID        int
	UserID    int    `gorm:"not null"`
	Token     string `gorm:"-"`
	TokenHash string `gorm:"not null;unique_index"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

type pwResetService interface {
	GetByToken(token string) (*pwReset, error)
	Create(pwr *pwReset) error
	Delete(id int) error
}
type pwResetValidator struct {
	pwResetService
	hmac hash.HMAC
}

type pwResetDB struct {
	gorm *gorm.DB
}

func newPWResetService(db *gorm.DB, hmac hash.HMAC) pwResetService {
	return &pwResetValidator{
		pwResetService: &pwResetDB{
			gorm: db,
		},
		hmac: hmac,
	}
}
func (pwrv *pwResetValidator) GetByToken(token string) (*pwReset, error) {
	pwr := &pwReset{Token: token}
	if err := runPWResetValFuncs(pwr, pwrv.tokenHashRequired); err != nil {
		return nil, fmt.Errorf("Validation Error: %w", err)
	}
	return pwrv.pwResetService.GetByToken(pwr.TokenHash)

}
func (pwrdb *pwResetDB) GetByToken(tokenHash string) (*pwReset, error) {
	var pwr pwReset
	err := checkErr(pwrdb.gorm.Where("token_hash = ?", tokenHash).First(&pwr).Error)
	if err != nil {
		return nil, err
	}
	return &pwr, nil
}

func (pwrv *pwResetValidator) Create(pwr *pwReset) error {
	token, err := rand.RememberToken()
	if err != nil {
		return fmt.Errorf("Unable to create reset token: %w", err)
	}
	pwr.Token = token
	if err := runPWResetValFuncs(pwr, pwrv.idRequired, pwrv.tokenHashRequired); err != nil {
		return fmt.Errorf("Validation Error: %w", err)
	}
	return pwrv.pwResetService.Create(pwr)
}

func (pwrdb *pwResetDB) Create(pwr *pwReset) error {
	return checkErr(pwrdb.gorm.Create(pwr).Error)
}

func (pwrv *pwResetValidator) Delete(id int) error {
	if id <= 0 {
		return errors.New("Invalid ID")
	}
	return pwrv.pwResetService.Delete(id)
}

func (pwrdb *pwResetDB) Delete(id int) error {
	pwr := pwReset{ID: id}
	return checkErr(pwrdb.gorm.Delete(&pwr).Error)
}

func runPWResetValFuncs(pwr *pwReset, fns ...pwrValFunc) error {
	for _, fn := range fns {
		if err := fn(pwr); err != nil {
			return err
		}
	}
	return nil
}

type pwrValFunc func(pwr *pwReset) error

func (pwrv *pwResetValidator) idRequired(pwr *pwReset) error {
	if pwr.UserID <= 0 {
		return errors.New("ID Invalid")
	}
	return nil
}
func (pwrv *pwResetValidator) tokenHashRequired(pwr *pwReset) error {
	if pwr.TokenHash == "" {
		if pwr.Token != "" {
			pwr.TokenHash = pwrv.hmac.Hash(pwr.Token)
			return nil
		}
		return errors.New("Remember Hash is required")
	}
	return nil
}
