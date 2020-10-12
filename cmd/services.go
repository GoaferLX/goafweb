package main

// TODO: AUTOMIGRATE a root user into user table
import (
	"fmt"
	"goafweb"
	"goafweb/mail"
	"goafweb/storage"

	"github.com/jinzhu/gorm"
)

type Services struct {
	gorm           *gorm.DB
	UserService    goafweb.UserService
	ArticleService goafweb.ArticleService
	MailService    goafweb.MailService
}
type serviceOpts func(*Services) error

func NewServices(opts ...serviceOpts) (*Services, error) {
	services := Services{}
	for _, opt := range opts {
		if err := opt(&services); err != nil {
			return nil, err
		}
	}
	return &services, nil
}

// Connect to database using GORM package
// Used by other services
func WithGorm(dialect, dsn string, prod bool) serviceOpts {
	return func(services *Services) error {
		db, err := gorm.Open(dialect, dsn)
		if err != nil {
			return fmt.Errorf("Could not establish a database connection: %w", err)
		}
		// LogMode should be off in production
		db.LogMode(!prod)
		services.gorm = db
		return nil
	}
}

// Loads user service, allows user functionality as defined by UserInterface//
func WithUsers(pepper, hmackey string) serviceOpts {
	return func(services *Services) error {
		services.UserService = storage.NewUserService(services.gorm, pepper, hmackey)
		return nil
	}
}

// Loads Articles service, allows articles functionality as defined by ArticleInterface
// Used for blogs/news sections
func WithArticles() serviceOpts {
	return func(services *Services) error {
		services.ArticleService = storage.NewArticleService(services.gorm)
		return nil
	}
}
func WithMail(domain, apiKey string) serviceOpts {
	return func(services *Services) error {
		services.MailService = mail.NewMailService(domain, apiKey)
		return nil
	}
}

// a Wrapper for gorms AutoMigrate function
func (s *Services) AutoMigrate() error {
	return s.gorm.AutoMigrate(&goafweb.User{}, &goafweb.Article{}).Error
}
