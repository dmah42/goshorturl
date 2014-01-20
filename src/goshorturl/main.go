package main

import (
	"database/sql"
	"flag"
	"log"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

var (
	db             = flag.String("db", "", "the db")
	host	       = flag.String("host", "", "the db host")
	user           = flag.String("user", "", "the db user")
	pwd            = flag.String("pwd", "", "the db password")
	port	       = flag.String("port", "http", "the port to listen on")

	dbConn *sql.DB = nil
	selectStmt *sql.Stmt = nil
	updateStmt *sql.Stmt = nil
)

func Handler(w http.ResponseWriter, r *http.Request) {
	shorturl := mux.Vars(r)["shorturl"]

	// lookup shorturl
	var longurl string
	var access_count int
	err := selectStmt.QueryRow(shorturl).Scan(&longurl, &access_count)

	switch {
	case err == sql.ErrNoRows:
		log.Printf("%q not found", shorturl)
		http.NotFound(w, r)

	case err != nil:
		log.Fatal(err)

	default:
		log.Printf("%q -> %q", shorturl, longurl)
		access_count++
		access_time := time.Now().UTC()
		_, err := updateStmt.Exec(access_count, access_time, shorturl)
		if err != nil {
			log.Printf("Failed to update access information: %v\n", err)
		}
		http.Redirect(w, r, longurl, http.StatusFound)
	}
}

func main() {
	flag.Parse()

	c, err := sql.Open("mysql", *user+":"+*pwd+"@("+*host+")/"+*db)
	if err != nil {
		log.Fatal("Failed to open sql dbConnection")
	}
	dbConn = c
	defer dbConn.Close()

	selectLongURL, err := dbConn.Prepare("SELECT `long`,`access_count` FROM `url` WHERE short=?")
	if err != nil {
		log.Fatal("Failed to prepare selectStmt: ", err)
	}
	selectStmt = selectLongURL

	updateAccess, err := dbConn.Prepare("UPDATE `url` SET `access_count`=?,`accessed`=? WHERE short=?")
	if err != nil {
		log.Fatal("Failed to prepare updateStmt: ", err)
	}
	updateStmt = updateAccess

	r := mux.NewRouter()
	// TODO: dashboard with metrics
	// TODO: ui for adding short urls
	r.HandleFunc("/{shorturl}", Handler).Methods("GET")
	http.Handle("/", r)

	log.Fatal(http.ListenAndServe(":"+*port, nil))
}
