package validation

import (
	"errors"
	"fmt"
	"goafweb"
	"strings"
)

// articleValidator will be responsible for validation/normalizing an Article ready for
// database storage/retreival.
type articleValidator struct {
	goafweb.ArticleDB
}

// NewArticleValidator creates a new articleValidator.
// It must receive something that satisfies the ArticleDB interface to satisfy
// the next layer of the interface. As well as any other arguments required for
// validation.
func NewArticleValidator(articleDB goafweb.ArticleDB) *articleValidator {
	return &articleValidator{
		ArticleDB: articleDB,
	}
}

func (av *articleValidator) GetByID(id int) (*goafweb.Article, error) {
	article := &goafweb.Article{ID: id}
	if err := runArticleValFuncs(article, av.idGreaterThan0); err != nil {
		return nil, fmt.Errorf("Validation Error: %w", err)
	}
	return av.ArticleDB.GetByID(article.ID)
}

func (av *articleValidator) Create(article *goafweb.Article) error {
	if err := runArticleValFuncs(article,
		av.titleRequired,
		av.contentRequired,
		av.authorRequired,
	); err != nil {
		return fmt.Errorf("Validation Error: %w", err)
	}
	return av.ArticleDB.Create(article)
}
func (av *articleValidator) Update(article *goafweb.Article) error {
	if err := runArticleValFuncs(article, av.idGreaterThan0, av.authorRequired, av.titleRequired); err != nil {
		return fmt.Errorf("Validation Error: %w", err)
	}
	return av.ArticleDB.Update(article)
}

func (av *articleValidator) Delete(id int) error {
	article := goafweb.Article{ID: id}
	if err := runArticleValFuncs(&article, av.idGreaterThan0); err != nil {
		return fmt.Errorf("Validation Error: %w", err)
	}
	return av.ArticleDB.Delete(article.ID)
}

// articleValFunc is a uniform type for all validation functions on an Article.
// All validation functions will be of this type so they can be used as variadic
// arguments in other functions.
// These funtions will return a customized error message if the validation fails,
// or nil if everything is okay.
type articleValFunc func(article *goafweb.Article) error

func runArticleValFuncs(article *goafweb.Article, fns ...articleValFunc) error {
	for _, fn := range fns {
		if err := fn(article); err != nil {
			return err
		}
	}
	return nil
}

func (av *articleValidator) titleRequired(article *goafweb.Article) error {
	article.Title = strings.TrimSpace(article.Title)
	if article.Title == "" {
		return errors.New("Title is required")
	}
	return nil
}

func (av *articleValidator) contentRequired(article *goafweb.Article) error {
	article.Content = strings.TrimSpace(article.Content)
	if article.Content == "" {
		return errors.New("Content is required")
	}
	return nil
}

func (av *articleValidator) idGreaterThan0(article *goafweb.Article) error {
	if article.ID <= 0 {
		return errors.New("ID cannot be zero")
	}
	return nil
}

func (av *articleValidator) authorRequired(article *goafweb.Article) error {
	if article.Author <= 0 {
		return errors.New("Author not set")
	}
	return nil
}
