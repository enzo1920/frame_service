package models

import (
	"database/sql"
	"fmt"
	"log"

	"../frame"
	_ "github.com/lib/pq"
)

var db *sql.DB

func InitDB(cfg frame.Configuration) {

	var err error
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable", cfg.Database.Host, cfg.Database.Port, cfg.Database.DBuser, cfg.Database.Password, cfg.Database.DBname)
	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Panic(err)
	}

	if err = db.Ping(); err != nil {
		log.Panic(err)
	}
	log.Println("Successfully connected!")
}
