package storage

import (
	"goafweb"

	"github.com/jinzhu/gorm"
)

type pwResetDB struct {
	gorm *gorm.DB
}

// NewPwResetDB returns a new service that implements a gorm database connection
// that fulfils goafweb.PwResetDB interface.
func NewPwResetDB(db *gorm.DB) goafweb.PwResetDB {
	return &pwResetDB{
		gorm: db,
	}
}

// GetByToken will lookup a pwReset using the token provided by the a User.
func (pwrdb *pwResetDB) GetByToken(tokenHash string) (*goafweb.PwReset, error) {
	var pwr goafweb.PwReset
	err := checkErr(pwrdb.gorm.Where("token_hash = ?", tokenHash).First(&pwr).Error)
	if err != nil {
		return nil, err
	}
	return &pwr, nil
}

// Create will add a new pwReset to the database.
func (pwrdb *pwResetDB) Create(pwr *goafweb.PwReset) error {
	return checkErr(pwrdb.gorm.Create(pwr).Error)
}

// Delete will remove a pwReset entry from the database.
// Note: This is a soft delete, pwReset will have DeletedAt field updated to time.Now()
// making it invisible to normal queries, but will still retreival when needed.
func (pwrdb *pwResetDB) Delete(id int) error {
	pwr := goafweb.PwReset{ID: id}
	return checkErr(pwrdb.gorm.Delete(&pwr).Error)
}
