package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/kelseyhightower/envconfig"
	_ "github.com/lib/pq"
	"github.com/mrbennbenn/pick6/database"
	"github.com/mrbennbenn/pick6/handlers"
	"github.com/mrbennbenn/pick6/middleware"
)

type Config struct {
	DatabaseURL  string `envconfig:"DATABASE_URL" required:"true"`
	Port         string `envconfig:"PORT" default:"8080"`
	SecureCookie bool   `envconfig:"SECURE_COOKIE" default:"true"`
	BaseURL      string `envconfig:"BASE_URL" default:"http://localhost:8080"`
}

func main() {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatal(err.Error())
	}

	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}

	log.Println("successfully connected to database")

	queries := database.New(db)
	logger := log.New(os.Stdout, "", log.LstdFlags)

	// Create chi router
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Timeout(20 * time.Second))

	// Serve static files with caching headers
	fileServer := http.FileServer(http.Dir("./static"))
	r.Handle("/static/*", middleware.CacheControl(http.StripPrefix("/static/", fileServer)))

	// API routes (public - no authentication required)
	r.Route("/api", func(r chi.Router) {
		apiHandler := &handlers.API{
			Queries: queries,
			Log:     logger,
			BaseURL: cfg.BaseURL,
		}

		// RESTful API for broadcast graphics
		r.Get("/events/{eventID}", apiHandler.GetEvent)
		r.Get("/events/{eventID}/questions", apiHandler.GetQuestions)
		r.Get("/events/{eventID}/questions/{questionID}", apiHandler.GetQuestion)
	})

	r.Route("/{slug}", func(r chi.Router) {
		uiHandler := &handlers.UI{
			Queries: queries,
			Log:     logger,
		}

		sessionMiddleware := &middleware.Session{
			SecureCookie: cfg.SecureCookie,
			Log:          logger,
			Queries:      queries,
		}
		r.Use(sessionMiddleware.ServeHTTP)

		r.Get("/", uiHandler.RedirectToFirst)
		r.Get("/question/{order}", uiHandler.ShowQuestion)
		r.Post("/question/{order}", uiHandler.SubmitAnswer)
		r.Get("/submit-info", uiHandler.ShowInfoForm)
		r.Post("/submit-info", uiHandler.SubmitInfoForm)
		r.Get("/end", uiHandler.ShowEnd)
	})

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("serving on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("server stopping: %v", err)
	}
}
