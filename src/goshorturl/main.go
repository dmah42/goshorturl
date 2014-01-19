package main

import (
	"database/sql"
	"flag"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

var (
	dbConn *sql.DB = nil
	db             = flag.String("db", "", "the db")
	user           = flag.String("user", "", "the db user")
	pwd            = flag.String("pwd", "", "the db password")
	port	       = flag.String("port", "http", "the port to listen on")
)

func Handler(w http.ResponseWriter, r *http.Request) {
	shorturl := mux.Vars(r)["shorturl"]

	// lookup shorturl
	var longurl string
	var access_count int
	err := dbConn.QueryRow("select long,access_count from url where short=?", shorturl).Scan(&longurl, &access_count)

	// TODO: update accessed and access_count

	switch {
	case err == sql.ErrNoRows:
		http.NotFound(w, r)

	case err != nil:
		log.Fatal(err)

	default:
		access_count++
		http.Redirect(w, r, longurl, 200)
	}
}

func main() {
	flag.Parse()

	dbConn, err := sql.Open("mysql", *user+":"+*pwd+"@/"+*db)
	if err != nil {
		log.Fatal("Failed to open sql dbConnection")
	}
	defer dbConn.Close()

	r := mux.NewRouter()
	// TODO: dashboard with metrics
	r.HandleFunc("/{shorturl}", Handler).Methods("GET")
	http.Handle("/", r)

	log.Fatal(http.ListenAndServe(":"+*port, nil))
}
