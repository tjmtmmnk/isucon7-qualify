package main

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/go-sql-driver/mysql"
	"io/ioutil"
	"log"
	"os" //nolint:gofmt
)

func main() {
	dbHost := os.Getenv("ISUBATA_DB_HOST")
	if dbHost == "" {
		dbHost = "127.0.0.1"
	}
	dbPort := os.Getenv("ISUBATA_DB_PORT")
	if dbPort == "" {
		dbPort = "13306"
	}
	dbUser := os.Getenv("ISUBATA_DB_USER")
	if dbUser == "" {
		dbUser = "root"
	}
	dbPassword := os.Getenv("ISUBATA_DB_PASSWORD")
	if dbPassword != "" {
		dbPassword = ":" + dbPassword
	}

	dsn := fmt.Sprintf("%s%s@tcp(%s:%s)/isubata?parseTime=true&loc=Local&charset=utf8mb4",
		dbUser, dbPassword, dbHost, dbPort)

	fmt.Println(dsn)
	db, err := sqlx.Connect("mysql", dsn)

	if err != nil {
		log.Fatalf(err.Error())
	}

	defer db.Close()

	rows, _ := db.Query("select `name`, `data` from image")

	defer rows.Close()

	var (
		name string
		data []byte
	)

	for rows.Next() {
		rows.Scan(&name, &data)
		fmt.Println(name)
		err := ioutil.WriteFile(name, data, 0666)
		if err != nil {
			log.Fatalf(err.Error())
		}
	}

}
