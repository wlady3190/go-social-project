package main

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	// "fmt"
	"net/http"

	"github.com/google/uuid"
	// "github.com/wlady3190/go-social/internal/mailer"
	"github.com/wlady3190/go-social/internal/store"
)

type RegisterUserPayload struct {
	Username string `json:"username" validate:"required,max=100"`
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=3,max=72"`
}

type UserWithToken struct {
	*store.User
	Token string `json:"token"`
}

// registerUser godoc
//
//	@Summary		Registers a user
//	@Description	Registers a user
//	@Tags			authentication
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		RegisterUserPayload	true	"User credentials"
//	@Success		201		{object}	UserWithToken		"User registered"
//	@Failure		400		{object}	error
//	@Failure		500		{object}	error
//	@Router			/authentication/user [post]
func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {

	var payload RegisterUserPayload

	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestReponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestReponse(w, r, err)
		return
	}
	user := &store.User{
		Username: payload.Username,
		Email:    payload.Email,
	}

	//* hash password
	if err := user.Password.Set(payload.Password); err != nil {
		app.internalServerError(w, r, err)
		return
	}
	//* store user
	ctx := r.Context()

	plainToken := uuid.New().String() //! este token se envia por correo para la activación

	//*store

	hash := sha256.Sum256([]byte(plainToken))

	hashToken := hex.EncodeToString(hash[:])

	// err := app.store.Users.CreateAndInvite(ctx, user, hashToken, app.config.mail.exp)
	err := app.store.Users.CreateAndInvite(ctx, user, hashToken, time.Hour*24)

	//! Se crean migraciones de invitación
	if err != nil {
		switch err {
		case store.ErrDuplicateEmail:
			app.badRequestReponse(w, r, err)
		case store.ErrDuplicateUsername:
			app.badRequestReponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return

	}
	UserWithToken := UserWithToken{
		User:  user,
		Token: plainToken,
	}

	// isProdEnv := app.config.env == "production"

	// activationURL := fmt.Sprintf("%s/confirm/%s",app.config.frontendURL, plainToken)
	// vars := struct {
	// 	Username      string
	// 	ActivationURL string
	// }{
	// 	Username: user.Username,
	// 	ActivationURL: activationURL,
	// }
	//! Enviadno correos
	// status, err := app.mailer.Send(mailer.UserWelcomeTemplate, user.Username, user.Email, vars, !isProdEnv)
	// if err != nil {
	// 	app.logger.Errorw("error sending welcome email", "error", err)
	// 	//! *****************SAGA PATTTERN ******************************
	// 	//! rollback user creation if email fails SAGA
	// 	if err := app.store.Users.Delete(ctx,user.ID); err != nil {
	// 		app.logger.Errorw("error deleting user", "error", err)
	// 	}

	// 	app.internalServerError(w, r, err)
	// 	return
	// }
	// app.logger.Infow("Email sent", "status code: ", status)

	if err := app.jsonResponse(w, http.StatusCreated, UserWithToken); err != nil {
		app.internalServerError(w, r, err)
	}

}

// createTokenHandler godoc
//
//	@Summary		Creates a token
//	@Description	Creates a token for a user
//	@Tags			authentication
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		CreateUserTokenPayload	true	"User credentials"
//	@Success		200		{string}	string					"Token"
//	@Failure		400		{object}	error
//	@Failure		401		{object}	error
//	@Failure		500		{object}	error
//	@Router			/authentication/token [post]

func (app *application) createTokenHandler(w http.ResponseWriter, r *http.Request) {

}