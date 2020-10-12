package storage

import (
	"errors"
	"goafweb"
	"strings"

	"github.com/jinzhu/gorm"
)

type articleService struct {
	goafweb.ArticleDB
}

type articleValidator struct {
	goafweb.ArticleDB
}

type articleDB struct {
	gorm *gorm.DB
}

func NewArticleService(db *gorm.DB) goafweb.ArticleService {
	return &articleService{
		ArticleDB: &articleValidator{
			ArticleDB: &articleDB{
				gorm: db,
			},
		},
	}
}

func (av *articleValidator) Create(article *goafweb.Article) error {
	if err := runArticleValFuncs(article,
		av.titleRequired,
		av.contentRequired,
		av.authorRequired,
	); err != nil {
		return err
	}
	return av.ArticleDB.Create(article)
}

func (adb *articleDB) Create(article *goafweb.Article) error {
	return adb.gorm.Create(article).Error
}

func (av *articleValidator) GetArticleByID(id int) (*goafweb.Article, error) {
	article := &goafweb.Article{ID: id}
	if err := runArticleValFuncs(article, av.idGreaterThan0); err != nil {
		return nil, err
	}
	return av.ArticleDB.GetArticleByID(article.ID)
}

func (adb *articleDB) GetArticleByID(id int) (*goafweb.Article, error) {
	var article goafweb.Article
	err := adb.gorm.First(&article, id).Error
	return &article, err
}

func (adb *articleDB) GetArticlesByUser(authorID int) ([]goafweb.Article, error) {
	var articles []goafweb.Article

	results := adb.gorm.Where("author = ?", authorID)
	if err := results.Find(&articles).Error; err != nil {
		return nil, err
	}
	return articles, nil
}
func (av *articleValidator) Update(article *goafweb.Article) error {
	if err := runArticleValFuncs(article, av.idGreaterThan0, av.authorRequired, av.titleRequired); err != nil {
		return err
	}
	return av.ArticleDB.Update(article)
}
func (adb *articleDB) Update(article *goafweb.Article) error {
	return adb.gorm.Save(article).Error
}

func (av *articleValidator) Delete(id int) error {
	article := goafweb.Article{ID: id}
	if err := runArticleValFuncs(&article, av.idGreaterThan0); err != nil {
		return err
	}
	return av.ArticleDB.Delete(article.ID)
}

func (adb *articleDB) Delete(id int) error {
	article := goafweb.Article{ID: id}
	return adb.gorm.Delete(&article).Error
}

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
