package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"github.com/wlady3190/go-social/docs" //! Para generar documentación de swagger
	"go.uber.org/zap"

	"github.com/wlady3190/go-social/internal/store"
)

type application struct {
	config config
	store  store.Storage //! se pasa a main
	//* Logging estructurado con zap
	logger *zap.SugaredLogger //! Para el main

}

type config struct {
	addr   string
	db     dbConfig
	env    string //desarrollo, producción, etc.
	//* Viene del swagger
	apiURL string
	//* expiration 
	mail mailConfig

}

type mailConfig struct {
	exp time.Duration
}


type dbConfig struct {
	addr              string
	maxOpenConns      int
	maxIdleConnetions int
	maxIdleTime       string
} //! y para el main

// func (app *application) mount() *http.ServeMux { //! instalado chi
// func (app *application) mount() *chi.Mux {
func (app *application) mount() http.Handler {

	// mux := http.NewServeMux()

	// mux.HandleFunc("GET /v1/health", app.healthCheckHandler )
	r := chi.NewRouter()
	// A good base middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP) //! Ver documentcion sobre el uso con nginx
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(middleware.Timeout(60 * time.Second))

	// r.Get("/v1/health", app.healthCheckHandler)
	r.Route("/v1", func(r chi.Router) {
		r.Get("/health", app.healthCheckHandler)
		//! Swagger implementación
		docsURL := fmt.Sprintf("%s/swagger/doc.json", app.config.addr)

		r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL(docsURL)))

		r.Route("/posts", func(r chi.Router) {
			r.Post("/", app.createPostHandler)

			r.Route("/{postID}", func(r chi.Router) {
				r.Use(app.postsContextMiddleware)
				r.Get("/", app.getPostHandler)
				r.Delete("/", app.deletePostHandler)
				r.Patch("/", app.updatePostHandler)

				r.Route("/comments", func(r chi.Router) {
					r.Post("/", app.createCommentHandler)
				})

			},
			)
		})

		r.Route("/users", func(r chi.Router) {
			r.Route("/{userID}", func(r chi.Router) {
				r.Use(app.userContextMiddleware)
				r.Get("/", app.getUserHandler)
				r.Put("/follow", app.followUserHandler)
				r.Put("/unfollow", app.unfollowUserHandler)

			},
			)
			r.Group(func(r chi.Router) {
				r.Get("/feed", app.getUserFeedHandler)
			})
			//! Rutas públicas
			r.Route("/authentication", func(r chi.Router) {
				r.Post("/user", app.registerUserHandler)
			})
		})

	})

	return r
} //esto va  a main

// func (app *application) run(mux *http.ServeMux) error {
// func (app *application) run(mux *chi.Mux) error {
func (app *application) run(mux http.Handler) error {
	//! Swagger implementación
	docs.SwaggerInfo.Version = version
	docs.SwaggerInfo.Host = app.config.apiURL
	docs.SwaggerInfo.BasePath = "/v1"

	// mux := http.NewServeMux()
	srv := &http.Server{
		Addr:         app.config.addr,
		Handler:      mux,
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 10,
		IdleTimeout:  time.Minute,
	}
	app.logger.Infow("server has started ", "addr", app.config.addr, "env", app.config.env)
	return srv.ListenAndServe()

}
