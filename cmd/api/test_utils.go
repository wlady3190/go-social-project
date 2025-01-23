package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	// "github.com/go-chi/chi/v5"
	"github.com/wlady3190/go-social/internal/auth"
	"github.com/wlady3190/go-social/internal/ratelimiter"
	"github.com/wlady3190/go-social/internal/store"
	"github.com/wlady3190/go-social/internal/store/cache"
	"go.uber.org/zap"
)


func  newTestApplication(t *testing.T, cfg config) *application  {
	t.Helper()
	//*usado para test
	// logger := zap.NewNop().Sugar()
	logger := zap.Must(zap.NewProduction()).Sugar()

	mockStore := store.NewMockStore()
	mockCacheStorage := cache.NewMockStore()
	// y se crea en internal -> auth -> mocks
	testAuth := &auth.TestAuthenticator{}

		// Rate limiter
		rateLimiter := ratelimiter.NewFixedWindowLimiter(
			cfg.rateLimiter.RequestPerTimeFrame,
			cfg.rateLimiter.TimeFrame,
		)

	return &application{
		logger: logger,
		store: mockStore,
		cacheStorage: mockCacheStorage,
		authenticator: testAuth,
		config: cfg,
		rateLimiter: rateLimiter,

	}
}

func executeRequest(req *http.Request, mux http.Handler) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	return rr
}

func checkResponseCode( t *testing.T, expected, actual int)  {
	if expected != actual{
		t.Errorf("Expected response code %d. Got %d", expected, actual)
	}
}