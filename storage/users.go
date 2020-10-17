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

	"github.com/jinzhu/gorm"
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
