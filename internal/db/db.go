package db

import (
	"context"
	"database/sql"
	"time"
)

// ! No se pasa la struc que se tienen en main xq los paquetes de internal no deben saber de los externos, hay que repetir los parametros
func New(add string, maxOpenConns, maxIdleConnetions int, maxIdleTime string) (*sql.DB, error) {
	db, err := sql.Open("postgres", add)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConnetions)
	
	duration, err := time.ParseDuration(maxIdleTime)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxIdleTime(time.Duration(duration))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}
	return db, nil //! Y va para el main

}
