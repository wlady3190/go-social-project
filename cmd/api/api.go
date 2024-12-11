package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/wlady3190/go-social/internal/store"
)

type application struct {
	config config
	store  store.Storage //! se pasa a main

}

type config struct {
	addr string
	db   dbConfig

	env string //desarrollo, producción, etc.

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

		r.Route("/posts", func(r chi.Router) {
			r.Post("/", app.createPostHandler)

			r.Route("/{postID}", func(r chi.Router) {
				r.Use(app.postsContextMiddleware)
				r.Get("/", app.getPostHandler)
				r.Delete("/", app.deletePostHandler)
				r.Patch("/", app.updatePostHandler)
			},
			)
		})

	})

	return r
} //esto va  a main

// func (app *application) run(mux *http.ServeMux) error {
// func (app *application) run(mux *chi.Mux) error {
func (app *application) run(mux http.Handler) error {

	// mux := http.NewServeMux()
	srv := &http.Server{
		Addr:         app.config.addr,
		Handler:      mux,
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 10,
		IdleTimeout:  time.Minute,
	}
	log.Printf("server is in  %s", app.config.addr)
	return srv.ListenAndServe()

}