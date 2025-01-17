package main

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"github.com/mashfeii/chirpy/internal/domain"
	"github.com/mashfeii/chirpy/internal/infrastructure/api"
	"github.com/mashfeii/chirpy/internal/infrastructure/database"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("error loading .env file: %s", err.Error())
	}

	DBUrl := os.Getenv("DB_URL")

	db, err := sql.Open("postgres", DBUrl)
	if err != nil {
		log.Fatalf("unable to connect to database: %s", err.Error())
	}

	queries := database.New(db)

	conf := domain.APIConfig{
		Database: queries,
		Platform: os.Getenv("PLATFORM"),
		Secret:   os.Getenv("SECRET"),
		Polka:    os.Getenv("POLKA_KEY"),
	}

	mux := http.NewServeMux()
	server := http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Millisecond,
	}

	fileHandler := http.StripPrefix("/app/", http.FileServer(http.Dir("./public")))

	mux.Handle("/app/", api.MiddlewareLog(conf.MiddlewareInc(fileHandler)))

	mux.HandleFunc("POST /admin/reset", conf.ResetHandler)

	mux.HandleFunc("POST /api/users", conf.CreateUserHandler)
	mux.HandleFunc("PUT /api/users", conf.UpdateUserHandler)
	mux.HandleFunc("POST /api/login", conf.LoginUserHandler)
	mux.HandleFunc("POST /api/refresh", conf.RefreshHandler)
	mux.HandleFunc("POST /api/revoke", conf.RevokeHandler)

	mux.HandleFunc("GET /api/chirps", conf.ShowChirpsHandler)
	mux.HandleFunc("GET /api/chirps/{chirp_id}", conf.ShowChirpHandler)
	mux.HandleFunc("POST /api/chirps", conf.CreateChirpsHandler)
	mux.HandleFunc("DELETE /api/chirps/{chirp_id}", conf.DeleteChirpHandler)

	mux.HandleFunc("POST /api/polka/webhooks", conf.PolkaWebhookHandler)

	err = server.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("failed to start server: %s", err.Error())
	}
}
