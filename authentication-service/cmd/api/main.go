package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	data "authentication/data"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
)

const webPort = "80"

var counts int64

type Config struct {
	Db     *sql.DB
	Models data.Models
}

func main() {

	log.Println("Starting auth service")

	conn := connectToDb()
	if conn == nil {
		log.Panic("couldnt connect to db")
	}

	app := Config{
		Db:     conn,
		Models: data.New(conn),
	}
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	err := srv.ListenAndServe()
	if err != nil {
		log.Panic(err)
	}
}

func openDB(dsn string) (*sql.DB, error) {

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func connectToDb() *sql.DB {
	dsn := os.Getenv("DSN")
	for {
		conn, err := openDB(dsn)
		if err != nil {
			log.Println("postgres not ready")
			counts++
		} else {
			log.Println("postgres connected")
			return conn
		}

		if counts > 10 {
			return nil
		}

		log.Println("backing off for two seconds")
		time.Sleep(2 * time.Second)

	}

}
