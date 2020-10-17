package main

// TODO: AUTOMIGRATE a root user into user table
import (
	"fmt"
	"goafweb"
	"goafweb/hash"
	"goafweb/mail"
	"goafweb/storage"
	"goafweb/validation"

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
func WithUsers(userPwPepper, hmacSecretKey string) serviceOpts {
	return func(services *Services) error {
		hmac := hash.NewHMAC(hmacSecretKey)
		udb := storage.NewUserDB(services.gorm)
		v := validation.NewUserValidator(udb, hmac, userPwPepper)
		pwrdb := storage.NewPwResetDB(services.gorm)
		pwrv := validation.NewPwResetValidator(pwrdb, hmac)
		us := storage.NewUserService(v, pwrv, userPwPepper)
		services.UserService = us
		return nil
	}
}

// Loads Articles service, allows articles functionality as defined by ArticleInterface
// Used for blogs/news sections
func WithArticles() serviceOpts {
	return func(services *Services) error {
		adb := storage.NewArticleDB(services.gorm)
		av := validation.NewArticleValidator(adb)
		services.ArticleService = av
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
