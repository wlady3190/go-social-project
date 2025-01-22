package main

import (
	"log"
	"net/http"

	"testing"
)

//go test
func  TestGetUser(t *testing.T)  {
	app := newTestApplication(t)

	mux := app.mount()
	//! testToken := "abc123" Primero hacer con esto para ver el token marformado
	testToken, err := app.authenticator.GenerateToken(nil)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("should not allow unauthenticated requests", func(t *testing.T) {
		//check for 401 code
		req, err := http.NewRequest(http.MethodGet, "/v1/users/1",nil)
		if err != nil {
			t.Fatal(err)
		}
		// rr := httptest.NewRecorder()
		// mux.ServeHTTP(rr, req)
		rr := executeRequest(req, mux)
		
		// if rr.Code != http.StatusUnauthorized{
		// 	t.Errorf("expected the responde code to be %d  and we got %d", http.StatusUnauthorized, rr.Code)
		// }

		checkResponseCode(t, http.StatusUnauthorized, rr.Code)

	})

	t.Run("should allow authenticated requests", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/v1/users/1",nil)
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Authorization", "Bearer " + testToken)

		rr := executeRequest(req, mux )

		checkResponseCode(t, http.StatusOK, rr.Code)
		log.Println(rr.Body)
	})
}