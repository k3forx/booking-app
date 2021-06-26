package main

import (
	"encoding/gob"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/k3forx/booking-app/internal/config"
	"github.com/k3forx/booking-app/internal/driver"
	"github.com/k3forx/booking-app/internal/handlers"
	"github.com/k3forx/booking-app/internal/helpers"
	"github.com/k3forx/booking-app/internal/models"
	"github.com/k3forx/booking-app/internal/render"
)

const portNumber = ":8080"

var app config.AppConfig
var session *scs.SessionManager
var infoLog *log.Logger
var errorLog *log.Logger

// main is the main application function
func main() {
	db, err := run()
	if err != nil {
		log.Fatal(err)
	}
	defer db.SQL.Close()

	defer close(app.MailChan)
	app.InfoLog.Println("Starting mail listener")
	listenForMail()

	app.InfoLog.Printf("Starting application on port %s", portNumber)

	srv := &http.Server{
		Addr:    portNumber,
		Handler: routes(&app),
	}
	err = srv.ListenAndServe()
	log.Fatal(err)
}

func run() (*driver.DB, error) {
	// what am I going to put in the session
	gob.Register(models.Reservation{})
	gob.Register(models.User{})
	gob.Register(models.Room{})
	gob.Register(models.RoomRestriction{})
	gob.Register(map[string]int{})

	mailChan := make(chan models.MailData)
	app.MailChan = mailChan

	// Change this to true when in production
	app.InProduction = false

	infoLog = log.New(os.Stdout, "[INFO]\t", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
	app.InfoLog = infoLog

	errorLog = log.New(os.Stdout, "[ERROR]\t", log.Ldate|log.Ltime|log.Lshortfile)
	app.ErrorLog = errorLog

	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction

	app.Session = session

	app.InfoLog.Println("Connecting to database...")
	db, err := driver.ConnectSQL("host=localhost port=5432 dbname=booking user=root password=root")
	if err != nil {
		log.Fatal("Cannot connect to database! Dying...")
	}
	app.InfoLog.Println("Connected to database!")

	tc, err := render.CreateTemplateCache()
	if err != nil {
		app.ErrorLog.Printf("failed to create template cache: %s", err)
		return nil, err
	}

	app.TemplateCache = tc
	app.UseCache = false

	repo := handlers.NewRepo(&app, db)
	handlers.NewHandlers(repo)
	render.NewRenderer(&app)
	helpers.NewHelpers(&app)

	return db, nil
}
