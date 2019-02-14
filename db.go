package main

import (
	"database/sql"
	_ "github.com/lib/pq"
	"log"
)

func sendToDb(msg string) error {

	connStr := "user=postgres password=postgres dbname=productdb sslmode=disable"
	//port 5432
	// ip 172.19.0.3
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		addMsg := "sql.Open error open connection:" + connStr
		return failOnError(err, addMsg)
	}
	defer func() {
		err = db.Close()
		failOnError(err, "Error db.Close()")
	}()

	result := db.QueryRow("insert into queue (msg) values ($1)", msg)
	err = result.Scan(&msg)
	if err == sql.ErrNoRows {
		log.Println("Placeholders case: NOT FOUND")
		return nil
	} else {
		addMsg := "Placeholders id:" + msg
		return failOnError(err, addMsg)
	}

}

func failOnError(err error, msg string) error {
	if err != nil {
		log.Println(msg, ":", err)
		return err
	}
	return nil
}
