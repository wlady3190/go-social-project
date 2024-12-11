package main

import (
	"context"
	"errors"

	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/wlady3190/go-social/internal/store"
)

// * Para el uso en el middleware
type postKey string

const postCtx postKey = "post"

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

	if err := app.jsonResponse(w, http.StatusCreated, post); err != nil {
		// writeJSONError(w, http.StatusInternalServerError, err.Error())
		app.internalServerError(w, r, err)

		return
	}
}

func (app *application) getPostHandler(w http.ResponseWriter, r *http.Request) {

	// idParam := chi.URLParam(r, "postID")         //viene del path
	// id, err := strconv.ParseInt(idParam, 10, 64) //base 10, int64
	// if err != nil {
	// 	// writeJSONError(w, http.StatusInternalServerError, err.Error())
	// 	app.internalServerError(w, r, err)
	// 	return
	// }
	// ctx := r.Context()

	// post, err := app.store.Posts.GetById(ctx, id)
	// if err != nil {
	// 	switch {
	// 	case errors.Is(err, store.ErrNotFound):
	// 		// writeJSONError(w, http.StatusNotFound, err.Error())
	// 		app.notFoundResponse(w, r, err)
	// 	default:
	// 		// writeJSONError(w, http.StatusInternalServerError, err.Error())
	// 		app.internalServerError(w, r, err)

	// 	}
	// 	return
	// }

	// * Consumiendo la función de posts getPostFromCtx del middleware
	post := getPostFromCtx(r)

	comments, err := app.store.Comments.GetByPostID(r.Context(), post.ID)

	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	post.Comments = comments

	if err := app.jsonResponse(w, http.StatusOK, post); err != nil {
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

type UpdatePostPayload struct {
	Title   *string `json:"title" validate:"omitempty,max=100"`
	Content *string `json:"content" validate:"omitempty,max=1000"`
}

func (app *application) updatePostHandler(w http.ResponseWriter, r *http.Request) {
	post := getPostFromCtx(r)

	ctx := r.Context()

	var payload UpdatePostPayload
	//* UNMARSHALL
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestReponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestReponse(w, r, err)
		return
	}

	//! para evitar q si un campo en postman va vacio, este ponga vacio en la bdd
	if payload.Content != nil {
		post.Content = *payload.Content
	}

	if payload.Title != nil {
		post.Title = *payload.Title
	}

	if err := app.store.Posts.Update(ctx, post); err != nil {

		switch {
		case errors.Is(err, store.ErrNotFound):
			app.notFoundResponse(w, r, err)
		default:
			//409 por conflicto
			app.internalConflictResponse(w, r, err)
		}
		return

	}

	if err := app.jsonResponse(w, http.StatusOK, post); err != nil {
		app.internalServerError(w, r, err)
		return
	}

}

// ! Este middleware se consume en mount, en api
func (app *application) postsContextMiddleware(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		// ctx = context.WithValue(ctx, "post", post)
		ctx = context.WithValue(ctx, postCtx, post)

		//* creando un ctx con información del otro contexto. No se está modificado el original
		next.ServeHTTP(w, r.WithContext(ctx))

	})

}

func getPostFromCtx(r *http.Request) *store.Post {
	post, _ := r.Context().Value(postCtx).(*store.Post)
	return post
}
