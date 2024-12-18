package main

import (
	"log"
	"net/http"

	"github.com/wlady3190/go-social/internal/store"
)

func (app *application) getUserFeedHandler(w http.ResponseWriter, r *http.Request) {
	//TODO pagination, filters

	fq := store.PaginatedFeedQuery{
		Limit:  20,
		Offset: 0,
		Sort:   "desc",
	}

	fq, err := fq.Parse(r)
	if err != nil {
		app.badRequestReponse(w, r, err)
		return
	}

	if err := Validate.Struct(fq); err != nil {
		app.badRequestReponse(w, r, err)
		return
	}

	ctx := r.Context()
	log.Printf("Handling /feed for userID = %d", int64(100))
	feed, err := app.store.Posts.GetUserFeed(ctx, int64(100), fq)

	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, feed); err != nil {
		app.internalServerError(w, r, err)
	}
}
