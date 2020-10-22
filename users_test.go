package goafweb

import (
	"errors"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

type mockDB struct {
	users []*User
}

func (m *mockDB) GetByID(id int) (*User, error) {
	for _, user := range m.users {
		if user.ID == id {
			return user, nil
		}
	}
	return nil, ErrNotFound
}
func (m *mockDB) GetByEmail(email string) (*User, error) {
	for _, user := range m.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, ErrNotFound
}
func (m *mockDB) GetByRemember(token string) (*User, error) {
	for _, user := range m.users {
		if user.RememberToken == token {
			return user, nil
		}
	}
	return nil, ErrNotFound
}

func (m *mockDB) Create(user *User) error {
	m.users = append(m.users, user)
	return nil
}
func (*mockDB) Update(user *User) error {
	return nil
}

func TestAuthenticate(t *testing.T) {
	mockDB := &mockDB{}
	us := NewUserService(mockDB, nil, "pwPepper")

	testUser := &User{Email: "test@test.com", Password: "test"}
	pwhash, _ := bcrypt.GenerateFromPassword([]byte(testUser.Password+us.PwPepper), bcrypt.DefaultCost)
	testUser.PasswordHash = string(pwhash)
	mockDB.Create(testUser)

	tests := map[string]struct {
		want error
		user *User
	}{
		"Authentication OK": {want: nil, user: &User{Email: "test@test.com", Password: "test"}},
		"Invalid email":     {want: ErrNotFound, user: &User{Email: "test@test.cm", Password: "test"}},
		"Invalid pass":      {want: ErrPWInvalid, user: &User{Email: "test@test.com", Password: "test1"}},
	}
	for i, tc := range tests {
		pwhash, _ := bcrypt.GenerateFromPassword([]byte(tc.user.Password+us.PwPepper), bcrypt.DefaultCost)
		tc.user.PasswordHash = string(pwhash)
		tests[i] = tc
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := us.Authenticate(tc.user.Email, tc.user.Password)
			if !errors.Is(err, tc.want) {
				t.Errorf("Got %v, wanted %v", err, tc.want)
			}
		})
	}
}
