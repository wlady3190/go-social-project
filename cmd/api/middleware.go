package main

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
)

func (app *application) BasicAuthMiddleware() func(http.Handler) http.Handler {

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			//read auth header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				app.unathorizedBasicErrorResponse(w, r, fmt.Errorf("authorization header is missing"))
				return
			}
			//parse it -> get the base64
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Basic" {
				app.unathorizedBasicErrorResponse(w, r, fmt.Errorf("authorization header is malformed"))
				return
			}
			//decode

			decoded, err := base64.StdEncoding.DecodeString(parts[1])
			if err != nil {
				app.unathorizedBasicErrorResponse(w, r, err)
				return
			}

			username := app.config.auth.basic.user
			password := app.config.auth.basic.pass //! Y van a la APi

			creds := strings.SplitN(string(decoded), ":", 2) //* las credenciales está como: user:contraseña
			if len(creds) != 2 || creds[0] != username || creds[1] != password {
				// app.unathorizedErrorResponse(w, r, fmt.Errorf("invalid credentials"))
				app.unathorizedBasicErrorResponse(w,r,fmt.Errorf("invalid credentials"))
				return
			}


			//check credentials

			next.ServeHTTP(w, r)
		})
	}

}
