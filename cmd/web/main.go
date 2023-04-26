package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"text/template"

	"github.com/saparaly/snippentbox/db"
	"github.com/saparaly/snippentbox/pkg/models/sqlite"
)

type application struct {
	errorLog      *log.Logger
	infoLog       *log.Logger
	posts         *sqlite.UserModel
	templateCache map[string]*template.Template
}

func main() {
	addr := flag.String("addr", ":8000", "HTTP network address")
	// dsn := flag.String("dsn", )
	flag.Parse()
	infoLog := log.New(os.Stdout, "INFO/t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "Error/t", log.Ldate|log.Ltime|log.Lshortfile)

	d, err := db.CreateDB()
	if err != nil {
		errorLog.Fatal(err)
	}

	if err := db.CreateTables(d); err != nil {
		fmt.Println(err)
		return
	}
	defer d.Close()

	app := &application{
		errorLog: errorLog,
		infoLog:  infoLog,
		posts:    &sqlite.UserModel{DB: d},
	}

	srv := &http.Server{
		Addr:     *addr,
		ErrorLog: errorLog,
		Handler:  app.routers(),
	}

	infoLog.Println("Listening on port http://localhost:8000/")
	err = srv.ListenAndServe()
	errorLog.Fatal(err)
}
