package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/wlady3190/go-social/internal/store"
)

func (app *application) AuthTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			app.unathorizedErrorResponse(w, r, fmt.Errorf("authorization header is missing - authHeader"))
			return
		}
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			app.unathorizedErrorResponse(w, r, fmt.Errorf("authorization header is malformed - authtokenMiddleware"))
			return
		}
		token := parts[1]

		jwtToken, err := app.authenticator.ValidateToken(token)

		if err != nil {
			app.unathorizedErrorResponse(w, r, err)
			return
		}
		claims, _ := jwtToken.Claims.(jwt.MapClaims)

		userID, err := strconv.ParseInt(fmt.Sprintf("%.f", claims["sub"]), 10, 64)

		if err != nil {
			app.unathorizedErrorResponse(w, r, err)
			return
		}
		ctx := r.Context()

		//! implementando cache con redis

		// user, err := app.store.Users.GetById(ctx, userID)

		// if err != nil {
		// 	app.unathorizedErrorResponse(w, r, err)
		// 	return
		// }
		user, err := app.getUser(ctx, userID)
		if err != nil {
			app.unathorizedErrorResponse(w, r, err)
			return
		}

		ctx = context.WithValue(ctx, userCtx, user)
		next.ServeHTTP(w, r.WithContext(ctx))

	})

}

func (app *application) getUser(ctx context.Context, userID int64) (*store.User, error) {
	if !app.config.redisCfg.enabled {
		return app.store.Users.GetById(ctx, userID)
	}
	//! verificacion si esta en caché
	app.logger.Infow("cache hit", "key", "user", "id", userID)
	user, err := app.cacheStorage.Users.Get(ctx, userID)
	if err != nil {
		return nil, err
	}

	if user == nil {
		user, err = app.store.Users.GetById(ctx, userID)
		if err != nil {
			// app.unathorizedErrorResponse(w, r, err)
			// return
			return nil, err
		}
		if err := app.cacheStorage.Users.Set(ctx, user); err != nil{
			return nil, err
		}

	}
	return user, nil
}

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
				app.unathorizedBasicErrorResponse(w, r, fmt.Errorf("invalid credentials"))
				return
			}

			//check credentials

			next.ServeHTTP(w, r)
		})
	}

}

func (app *application) checkPostOwnership(requiredRole string, next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := getUserFromContext(r)
		post := getPostFromCtx(r)

		//* check if is ths post of the user
		if post.UserID == user.ID {
			next.ServeHTTP(w, r)
		}

		//* check role
		allowed, err := app.CheckRolePrecedence(r.Context(), user, requiredRole)
		if err != nil {
			app.internalServerError(w, r, err)
			return
		}
		if !allowed {
			app.forbiddenResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (app *application) CheckRolePrecedence(ctx context.Context, user *store.User, roleName string) (bool, error) {
	role, err := app.store.Roles.GetByName(ctx, roleName)

	if err != nil {
		return false, err
	}

	return user.Role.Level >= role.Level, nil

}


func (app *application) RateLimiterMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if app.config.rateLimiter.Enabled {
			if allow, retryAfter := app.rateLimiter.Allow(r.RemoteAddr); !allow {
				app.rateLimiterExceededResponse(w, r, retryAfter.String())
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}