package handlers

import (
	"errors"
	"goafweb"
	"goafweb/context"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type articleHandler struct {
	ArticlesService goafweb.ArticleService
}

func NewArticles(service goafweb.ArticleService) *articleHandler {
	return &articleHandler{
		ArticlesService: service,
	}
}

// View returns an article from the database.
// GET /article.
func (ah *articleHandler) View(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	article, err := ah.ArticlesService.GetByID(id)
	if err != nil {
		if errors.Is(err, goafweb.ErrNotFound) {
			writeJson(w, err, http.StatusNotFound)
			return
		}
		writeJson(w, err, http.StatusBadRequest)
		return
	}

	writeJson(w, article, http.StatusOK)
}

// Create reads request body for a new article and inserts to database.
// POST /article.
func (ah articleHandler) Create(w http.ResponseWriter, r *http.Request) {
	var article goafweb.Article
	if err := readJson(r, &article); err != nil {
		writeJson(w, err, http.StatusUnprocessableEntity)
		return
	}
	user := context.GetUser(r.Context())
	article.Author = user.ID
	if err := ah.ArticlesService.Create(&article); err != nil {
		writeJson(w, err, http.StatusBadRequest)
		return
	}
	writeJson(w, article, http.StatusCreated)

}

// Update reads request body for an article and updates the database.
// PUT /article.
func (ah *articleHandler) Update(w http.ResponseWriter, r *http.Request) {
	var article goafweb.Article

	if err := readJson(r, &article); err != nil {
		writeJson(w, err, http.StatusUnprocessableEntity)
		return
	}

	if err := ah.ArticlesService.Update(&article); err != nil {
		writeJson(w, err, http.StatusBadRequest)
		return
	}
	writeJson(w, article, http.StatusCreated)
}

// Delete reads request body for an articles and removes it from the database.
// DELETE /article.
func (ah *articleHandler) Delete(w http.ResponseWriter, r *http.Request) {
	var article *goafweb.Article

	if err := readJson(r, &article); err != nil {
		writeJson(w, err, http.StatusUnprocessableEntity)
		return
	}

	if err := ah.ArticlesService.Delete(article.ID); err != nil {
		writeJson(w, err, http.StatusBadRequest)
		return
	}
	writeJson(w, nil, http.StatusOK)

}
