package main

import (
	"errors"

	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/wlady3190/go-social/internal/store"
)

type CreatePostPayload struct {
	Title   string   `json:"title" validate:"required,max=100"`
	Content string   `json:"content" validate:"required,max=1000"`
	Tags    []string `json:"tags"`
}

func (app *application) createPostHandler(w http.ResponseWriter, r *http.Request) {
	// var post store.Post
	var payload CreatePostPayload

	if error := readJSON(w, r, &payload); error != nil {
		// writeJSONError(w, http.StatusBadRequest, error.Error())
		app.badRequestReponse(w, r, error)
		return
	}
	//* Validación del payload

	if err := Validate.Struct(payload); err != nil {
		app.badRequestReponse(w, r, err)
		return
	}

	post := &store.Post{
		Title:   payload.Title,
		Content: payload.Content,
		Tags:    payload.Tags,
		UserID:  1,
	}

	// TODO Change after AUTH

	ctx := r.Context()

	if err := app.store.Posts.Create(ctx, post); err != nil {
		// writeJSONError(w, http.StatusInternalServerError, err.Error())
		app.internalServerError(w, r, err)

		return
	}

	if err := writeJSON(w, http.StatusCreated, post); err != nil {
		// writeJSONError(w, http.StatusInternalServerError, err.Error())
		app.internalServerError(w, r, err)

		return
	}
}

func (app *application) getPostHandler(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "postID")         //viene del path
	id, err := strconv.ParseInt(idParam, 10, 64) //base 10, int64
	if err != nil {
		// writeJSONError(w, http.StatusInternalServerError, err.Error())
		app.internalServerError(w, r, err)
		return
	}
	ctx := r.Context()

	post, err := app.store.Posts.GetById(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			// writeJSONError(w, http.StatusNotFound, err.Error())
			app.notFoundResponse(w, r, err)
		default:
			// writeJSONError(w, http.StatusInternalServerError, err.Error())
			app.internalServerError(w, r, err)

		}
		return
	}

	// * Consumiendo la función de comments

	comments, err := app.store.Comments.GetByPostID(ctx, id)

	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	post.Comments = comments

	if err := writeJSON(w, http.StatusOK, post); err != nil {
		// writeJSONError(w, http.StatusInternalServerError, err.Error())
		app.internalServerError(w, r, err)

		return
	}
}

func (app *application) deletePostHandler(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "postID")
	id, err := strconv.ParseInt(idParam, 10, 64)

	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	ctx := r.Context()

	if err := app.store.Posts.Delete(ctx, id); err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.notFoundResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return

	}
	w.WriteHeader(http.StatusNoContent) //retornar un estado sin cotnen

}
