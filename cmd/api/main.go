package main

import (
	"time"

	"github.com/wlady3190/go-social/internal/db"
	"github.com/wlady3190/go-social/internal/env"
	"github.com/wlady3190/go-social/internal/store"
	"go.uber.org/zap"
)

const version = "0.0.2"

//* COnfiguración para SWAGGER
//	@title			Social API
//	@version		1.0
//	@description	This a new test in go
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

//	@BasePath	/v1
//
//* Parámetros adicionales
//@securityDefinitions.apikey ApiKeyAuth
//@in			header
//@name			Authorization
//@description
//! De aquí a API

func main() {
	cfg := config{
		addr: env.GetString("ADDR", ":8080"),
		//* Para el swagger
		apiURL: env.GetString("EXTERNAL_URL", "localhost:8080"),
		db: dbConfig{
			addr:              env.GetString("DB_ADDR", "postgres://admin:adminpassword@localhost/social?sslmode=disable"),
			maxOpenConns:      env.GetInt("DB_MAX_OPEN_CONNS", 30),
			maxIdleConnetions: env.GetInt("DB_MAX_IDLE_CONNS", 30),
			maxIdleTime:       env.GetString("DB_MAX_IDLE_TIME", "15m"),
		}, //! A internal para crear db
		env: env.GetString("ENV", "development"),
		 //* Para la expiración del token de UserInvite
		 mail: mailConfig{
			exp: time.Hour*24*3, //3 dias
		 },
	}

	//* Logger
	logger := zap.Must(zap.NewProduction()).Sugar()
	defer logger.Sync() //! va a main -> aplication
	//* database
	db, err := db.New(cfg.db.addr,
		cfg.db.maxOpenConns,
		cfg.db.maxIdleConnetions,
		cfg.db.maxIdleTime)

	if err != nil {
		logger.Fatal(err)
	}

	defer db.Close()
	logger.Info("db connected")

	store := store.NewPostgresStorage(db) //! Y se pasa a la API

	app := &application{
		config: cfg,
		store:  store,
		logger: logger,
	}

	mux := app.mount()

	logger.Fatal(app.run(mux))

}
