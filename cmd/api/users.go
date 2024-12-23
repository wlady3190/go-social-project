package main

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/wlady3190/go-social/internal/store"
)

type userKey string

const userCtx userKey = "user"

// ActivateUser godoc
//
//	@Summary		Activates/Register a user
//	@Description	Activates/Register a user by invitation token
//	@Tags			users
//	@Produce		json
//	@Param			token	path		string	true	"Invitation token"
//	@Success		204		{string}	string	"User activated"
//	@Failure		404		{object}	error
//	@Failure		500		{object}	error
//	@Security		ApiKeyAuth
//	@Router			/users/activate/{token} [put]
func (app *application) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	
	token := chi.URLParam(r, "token")
	if err := app.store.Users.Activate(r.Context(), token); err != nil {
		switch err {
		case store.ErrNotFound:
			app.notFoundResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	if err := app.jsonResponse(w, http.StatusNoContent, ""); err != nil {
		app.internalServerError(w, r, err)
	}
}

// getUserHandler godoc
//
//	@Summary		Fetch user profile
//	@Description	Fecth a user profile by ID
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"User ID"
//	@Success		200	{object}	store.User
//	@Failure		400	{object}	error
//	@Failure		404	{object}	error
//	@Failure		500	{object}	error
//	@Security		ApiKeyAuth
//	@Router			/users/{id} [get]
func (app *application) getUserHandler(w http.ResponseWriter, r *http.Request) {
	// userID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)

	// if err != nil {
	// 	app.badRequestReponse(w, r, err)
	// 	return
	// }

	// ctx := r.Context()

	// user, err := app.store.Users.GetById(ctx, userID)
	// if err != nil {
	// 	switch {
	// 	case errors.Is(err, store.ErrNotFound):
	// 		app.notFoundResponse(w, r, err)
	// 		return
	// 	default:
	// 		app.internalServerError(w, r, err)
	// 		return
	// 	}
	// }

	user := getUserFromContext(r)

	if err := app.jsonResponse(w, http.StatusOK, user); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

type FollowUser struct {
	UserID int64 `json:"user_id"`
}

// FollowUser godoc
//
//	@Summary		Follow user
//	@Description	Follow a user profile by ID
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			userID	path		int	true	"User ID"
//	@Success		204		{string}	string "User followed"
//	@Failure		400		{object}	error  "Error"
//	@Failure		404		{object}	error "User Not found"
//	@Security		ApiKeyAuth
//	@Router			/users/{userID}/follow	[put]
func (app *application) followUserHandler(w http.ResponseWriter, r *http.Request) {

	followerUser := getUserFromContext(r)
	//TODO revert back to auht userID from ctx

	var payload FollowUser

	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestReponse(w, r, err)
		return
	}
	ctx := r.Context()

	//* payload esta auth
	if err := app.store.Followers.Follow(ctx, followerUser.ID, payload.UserID); err != nil {
		switch err {
		case store.ErrConflict:
			app.conflictResponse(w, r, err)
			return
		default:
			app.internalServerError(w, r, err)
			return
		}
	}

	if err := app.jsonResponse(w, http.StatusNoContent, nil); err != nil {
		//*arriba va nil xq no retorna nada
		app.internalServerError(w, r, err)
	}
}

// UnfollowUser godoc
//
//	@Summary		Unfollow a user
//	@Description	Unfollow a user by ID
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			userID	path		int		true	"User ID"
//	@Success		204		{string}	string	"User unfollowed"
//	@Failure		400		{object}	error	"User payload missing"
//	@Failure		404		{object}	error	"User not found"
//	@Security		ApiKeyAuth
//	@Router			/users/{userID}/unfollow	[put]
func (app *application) unfollowUserHandler(w http.ResponseWriter, r *http.Request) {

	unfollowerUser := getUserFromContext(r)

	// unfollowedID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	// if err != nil {
	// 	app.badRequestReponse(w, r, err)
	// 	return
	// }
	var payload FollowUser

	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestReponse(w, r, err)
		return
	}

	ctx := r.Context()
	if err := app.store.Followers.Unfollow(ctx, unfollowerUser.ID, payload.UserID); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusNoContent, nil); err != nil {
		//*arriba va nil xq no retorna nada
		app.internalServerError(w, r, err)
	}

}

func (app *application) userContextMiddleware(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		userID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)

		if err != nil {
			app.badRequestReponse(w, r, err)
			return
		}
		ctx := r.Context()

		user, err := app.store.Users.GetById(ctx, userID)
		if err != nil {
			switch {
			case errors.Is(err, store.ErrNotFound):
				app.notFoundResponse(w, r, err)
				return
			default:
				app.internalServerError(w, r, err)
				return
			}
		}

		ctx = context.WithValue(ctx, userCtx, user)
		next.ServeHTTP(w, r.WithContext(ctx))

	})
}

func getUserFromContext(r *http.Request) *store.User {
	user, _ := r.Context().Value(userCtx).(*store.User)
	return user
}
