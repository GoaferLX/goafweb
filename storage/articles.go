package storage

import (
	"goafweb"

	"github.com/jinzhu/gorm"
)

type articleDB struct {
	gorm *gorm.DB
}

// NewArticleDB returns a new service that implements a gorm database connection
// that fulfils goafweb.ArticleDB interface.
func NewArticleDB(db *gorm.DB) *articleDB {
	return &articleDB{
		gorm: db,
	}
}

// GetByID will retreive an article from the database.
func (adb *articleDB) GetByID(id int) (*goafweb.Article, error) {
	var article goafweb.Article
	err := checkErr(adb.gorm.First(&article, id).Error)
	return &article, err
}

// Create will add a new article to the database.
func (adb *articleDB) Create(article *goafweb.Article) error {
	return checkErr(adb.gorm.Create(article).Error)
}

// Update will update an existing article in the database.
func (adb *articleDB) Update(article *goafweb.Article) error {
	return checkErr(adb.gorm.Save(article).Error)
}

// Delete will remove article from database.
// Note: This is a soft delete, article will have DeletedAt field updated to time.Now()
// making it invisible to normal queries, but will still retreival when needed.
func (adb *articleDB) Delete(id int) error {
	article := goafweb.Article{ID: id}
	return checkErr(adb.gorm.Delete(&article).Error)
}
