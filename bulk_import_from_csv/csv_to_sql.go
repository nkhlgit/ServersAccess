package main

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func chkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	// read data from CSV file
	var srvId int
	var name string
	csvFile, err := os.Open("infra.csv")

	if err != nil {
		fmt.Println(err)
	}
	defer csvFile.Close()

	reader := csv.NewReader(csvFile)
	reader.FieldsPerRecord = -1
	csvData, err := reader.ReadAll()
	chkErr(err)
	err = os.Remove("dc.db")
	chkErr(err)

	db, _ := sql.Open("sqlite3", "./dc.db")
	// Adding any columns, add crate table with: CREATE> columns_name TEXT; INSERT > Coumns_name and ?, Exec value feild
	statement, _ := db.Prepare("CREATE TABLE IF NOT EXISTS servers (srvId INTEGER PRIMARY KEY, name TEXT," +
		"ip TEXT,hostname TEXT,osUser TEXT,osPassword TEXT,osPort TEXT,webPort TEXT,product TEXT,datacenter TEXT," +
		"webPrefix TEXT,webSuffix TEXT, fav TEXT, dateTimeCreated TEXT, dateTimeModified TEXT, dateTimeLastAccessed TEXT)")
	statement.Exec()
	statement, _ = db.Prepare("INSERT INTO servers (name,ip,hostname,osUser,osPassword,osPort," +
		"webPort,product,datacenter,webPrefix,webSuffix,fav, dateTimeCreated, dateTimeModified," +
		"dateTimeLastAccessed ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
	var timeNow string
	for _, each := range csvData[1:] {
		timeNow = time.Now().Format(time.RFC3339)
		statement.Exec(each[0], each[1], each[2], each[3], each[4], each[5], each[6],
			each[7], each[8], each[9], each[10], each[11], timeNow, timeNow, timeNow)

	}
	rows, err := db.Query("SELECT srvId,name FROM servers")
	chkErr(err)

	for rows.Next() {
		err = rows.Scan(&srvId, &name)
		chkErr(err)
		fmt.Println(srvId, " ", name)
	}
	db.Close()
}
