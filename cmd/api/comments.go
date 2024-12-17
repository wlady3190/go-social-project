package main

import (
	"github.com/wlady3190/go-social/internal/store"
	"net/http"
)

type CreateCommentPayload struct {
	Content string `json:"content" validate:"required,max=1000"`
}

func (app *application) createCommentHandler(w http.ResponseWriter, r *http.Request) {

	post := getPostFromCtx(r)
	
	var payload CreateCommentPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestReponse(w, r, err)
		return
	}

	//*  payload
	if err := Validate.Struct(payload); err != nil {
		app.badRequestReponse(w, r, err)
		return
	}
	ctx := r.Context()

	comment := &store.Comment{
		Content: payload.Content,
		UserID:  1,
		PostID:  post.ID,
	}

	if err := app.store.Comments.Create(ctx, comment); err != nil {
		app.internalServerError(w, r, err)
		return
	}

}
