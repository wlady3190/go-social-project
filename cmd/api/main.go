package main

import (
	"log"

	"github.com/wlady3190/go-social/internal/db"
	"github.com/wlady3190/go-social/internal/env"
	"github.com/wlady3190/go-social/internal/store"
)

const version = "0.0.1"

func main() {
	cfg := config{
		addr: env.GetString("ADDR", ":8080"),
		db: dbConfig{
			addr:              env.GetString("DB_ADDR", "postgres://admin:adminpassword@localhost/social?sslmode=disable"),
			maxOpenConns:      env.GetInt("DB_MAX_OPEN_CONNS", 30),
			maxIdleConnetions: env.GetInt("DB_MAX_IDLE_CONNS", 30),
			maxIdleTime:       env.GetString("DB_MAX_IDLE_TIME", "15m"),
		}, //! A internal para crear db
		env: env.GetString("ENV", "development"),
	}

	db, err := db.New(cfg.db.addr,
		cfg.db.maxOpenConns,
		cfg.db.maxIdleConnetions,
		cfg.db.maxIdleTime)

	if err != nil {
		log.Panic(err)
	}


	defer db.Close()
	log.Println("db connected")




	store := store.NewPostgresStorage(db) //! Y se pasa a la API

	app := &application{
		config: cfg,
		store:  store,
	}

	mux := app.mount()

	log.Fatal(app.run(mux))

}
