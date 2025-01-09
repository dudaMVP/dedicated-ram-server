package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"sync/atomic"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/mhishmeh/dedicated-ram-server/internal/database"
)

func ReadinessEndpoint(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	w.WriteHeader(http.StatusOK)

	message := "OK"
	w.Write([]byte(message))
}

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Increment the counter
		cfg.fileserverHits.Add(1)
		// Call the next handler
		next.ServeHTTP(w, r)
	})
}
func (cfg *apiConfig) metricsHandler(w http.ResponseWriter, r *http.Request) {
	// Get the current count from fileserverHits
	// Format it as "Hits: x"
	// Write it to the response
	count := cfg.fileserverHits.Load()

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, "<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>", count)

}

func (cfg *apiConfig) resetHandler(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Reset is only allowed in dev environment."))
		return
	}

	cfg.fileserverHits.Store(0)
	cfg.db.DeleteAllUsers(r.Context())
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits reset to 0 and database reset to initial state."))
}

func createServer(dbQueries *database.Queries) {
	platform := os.Getenv("PLATFORM")
	apiC := apiConfig{
		fileserverHits: atomic.Int32{},
		db:             dbQueries,
		platform:       platform,
	}
	mux := http.NewServeMux()
	server := &http.Server{
		Handler: mux,
		Addr:    ":8080",
	}

	// Create file server and strip the /app/ prefix
	fileServer := http.FileServer(http.Dir("."))
	handler := http.StripPrefix("/app/", fileServer)
	mux.Handle("/app/", apiC.middlewareMetricsInc(handler)) //middeware metrics

	mux.Handle("/assets", http.FileServer(http.Dir("./assets/logo.png")))
	mux.HandleFunc("GET /api/healthz", ReadinessEndpoint) //readinessEndpoint

	mux.HandleFunc("POST /api/chirps", apiC.handlerChirpsCreate) // create a chirp for a user

	mux.HandleFunc("GET /api/chirps", apiC.handlerChirpsRetrieve) // get all chrips
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiC.handlerGetSingleChirp)

	mux.HandleFunc("GET /admin/metrics", apiC.metricsHandler) //metrics holder
	mux.HandleFunc("POST /admin/reset", apiC.resetHandler)
	mux.HandleFunc("POST /api/users", apiC.handlerUsersCreate)

	mux.HandleFunc("POST /api/login", apiC.handlerLogin)
	//mux.HandleFunc("POST /api/validate_chirp", Validate_Chirp_Endpoint)

	addr := server.Addr
	fmt.Println("Starting server on", addr)
	if err := server.ListenAndServe(); err != nil {
		fmt.Println("Error starting server:", err)
	}
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("cannot start seerrvurrr")
	}
	dbQueries := database.New(db)
	createServer(dbQueries) // Pass dbQueries here

}
