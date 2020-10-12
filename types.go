package goafweb

import (
	"time"
)

// User defines a single User as stored in the database.
// Used to model a user single user throughout the app and mirror in database.
type User struct {
	ID            int    `gorm:"primary_key;"`
	Name          string `gorm:"not_null;"`
	Email         string `gorm:"not_null;unique_index;" json:"email"`
	Password      string `gorm:"-" json:"-"`
	PasswordHash  string `gorm:"not_null;"`
	RememberToken string `gorm:"-"`
	RememberHash  string `gorm:"not_null;unique_index;"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     *time.Time
}

// UserService defines the API for interacting with a User.
type UserService interface {
	Authenticate(email, password string) (*User, error)
	UserDB
	InitiatePWReset(email string) (string, error)
	CompletePWReset(token, newPW string) (*User, error)
}

// UserDB defines all database interactions for a single user.
type UserDB interface {
	// Standard CRUD actions.
	// Read - Methods for querying a user.
	GetByID(id int) (*User, error)
	GetByEmail(email string) (*User, error)
	GetByRemember(token string) (*User, error)
	// Methods for altering a user.
	Create(user *User) error
	Update(user *User) error
}

// Article defines a single Article as stored in the database.
// Can be used to model a news article or short blog post.
type Article struct {
	ID        int
	Title     string `gorm:"not_null"`
	Content   string `gorm:"not_null"`
	Author    int    `gorm:"not_null"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

// Defines API for interacting with an Article.
type ArticleService interface {
	ArticleDB
}

// Defines all database interactions for a single article.
type ArticleDB interface {
	// Standard CRUD actions.
	// Read - Methods for querying an Article.
	GetArticleByID(id int) (*Article, error)
	GetArticlesByUser(id int) ([]Article, error)
	// Methods for altering an Article.
	Create(article *Article) error
	Update(article *Article) error
	Delete(id int) error
}

type MailService interface {
	ResetPw(toEmail, token string) error
}
