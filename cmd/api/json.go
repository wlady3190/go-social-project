package main

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
)

var Validate *validator.Validate

// ! Funcion init ejecuta al arrancar la app
func init() {
	Validate = validator.New(validator.WithRequiredStructEnabled())

}

func writeJSON(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data) // Codificar la petición a JSON

}

func readJSON(w http.ResponseWriter, r *http.Request, data any) error {

	maxBytes := 1_048_578 //es el valor de 1 MB, es el tamaño maximo de la petición para evitar DDoS

	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	decoder := json.NewDecoder(r.Body) //Decodifica de la petición

	//No permitir campos desconocidos
	decoder.DisallowUnknownFields()

	return decoder.Decode(data)

}

func writeJSONError(w http.ResponseWriter, status int, message string) error {
	type envelope struct {
		Error string `json:"error"`
	}

	return writeJSON(w, status, &envelope{Error: message})
}

//* para mejorar la impresión de resultaoos y errores
func (app *application) jsonResponse(w http.ResponseWriter, status int, data any) error {
	type envelope struct {
		Data any `json:"data"`
	}
	return writeJSON(w, status, &envelope{Data: data})

}
