// Quick prototyping, expect a lot of repeated codes
package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/codegangsta/negroni"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
)

type (
	DB struct {
		*sql.DB
	}

	Tx struct {
		*sql.Tx
	}

	RecordResponse struct {
		Records []*Record `json:"records"`
	}

	Record struct {
		Email string `json:"email"`
		IP    string `json:"ip"`
		Count int    `json:"count"`
		Date  string `json:"date"`
	}
)

var buildVersion string

// I know, I know, it's not DRY, will merge later
func NewDBConn(dsn string) (*DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &DB{db}, nil
}

func (db *DB) Begin() (*Tx, error) {
	tx, err := db.DB.Begin()
	if err != nil {
		return nil, err
	}
	return &Tx{tx}, nil
}

// GetRecords fetches da awesome records
func (db *DB) GetRecords() ([]*Record, error) {
	var records []*Record

	tx, err := db.Begin()
	if err != nil {
		return records, err
	}

	stmt, err := tx.Prepare(`SELECT records.email, records.ip, SUM(records.count), records.date
FROM records
GROUP BY records.email, records.ip, records.date
ORDER BY records.date,
SUM(count) DESC;`)
	if err != nil {
		return records, err
	}
	defer stmt.Close()

	rows, err := stmt.Query()
	if err != nil {
		tx.Rollback()
		return records, err
	}
	defer rows.Close()

	for rows.Next() {
		record := &Record{}
		if err := rows.Scan(&record.Email, &record.IP, &record.Count, &record.Date); err != nil {
			log.Printf(err.Error())
		} else {
			records = append(records, record)
		}
	}

	return records, err
}

func (db *DB) RecordsHandler(rw http.ResponseWriter, req *http.Request) {
	recordList, err := db.GetRecords()
	if err != nil {
		log.Print(err.Error())
	}

	resp := &RecordResponse{recordList}

	var respB []byte
	respB, err = json.Marshal(resp)
	if err != nil {
		log.Print(err.Error())
		respB = []byte("{\"records\": []}")
	}

	rw.Write(respB)
}

func main() {
	fmt.Printf("hyperion-dashboard%s\n", buildVersion)

	httpAddr := flag.String("http-address", "127.0.0.1:12300", "<addr>:<port> to listen on")
	dsn := flag.String("db", "", "Database source name")
	flag.Parse()

	dataSource := *dsn
	if dataSource == "" {
		if os.Getenv("HYPERION_DB") != "" {
			dataSource = os.Getenv("HYPERION_DB")
		}
	}

	if dataSource == "" {
		flag.Usage()
		log.Fatal("--db or HYPERION_DB not found")
	}

	db, err := NewDBConn(dataSource)
	if err != nil {
		log.Fatal(err.Error())
	}

	router := mux.NewRouter()
	router.HandleFunc("/", db.RecordsHandler)

	recovery := negroni.NewRecovery()
	logger := negroni.NewLogger()

	n := negroni.New(recovery, logger)
	n.UseHandler(router)
	n.Run(*httpAddr)
}
