package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
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
}
type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func respondWithError(w http.ResponseWriter, code int, msg string, err error) {
	if err != nil {
		log.Println(err)
	}
	if code > 499 {
		log.Printf("Responding with 5XX error: %s", msg)
	}
	type errorResponse struct {
		Error string `json:"error"`
	}
	respondWithJSON(w, code, errorResponse{
		Error: msg,
	})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	dat, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(code)
	w.Write(dat)
}

func (cfg *apiConfig) handlerUsersCreate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email string `json:"email"`
	}
	type response struct {
		User
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	user, err := cfg.db.CreateUser(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create user", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, response{
		User: User{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email:     user.Email,
		},
	})
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

//func (cfg *apiConfig)

func (cfg *apiConfig) resetHandler(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)
	w.WriteHeader(http.StatusOK)

}

func Validate_Chirp_Endpoint(w http.ResponseWriter, r *http.Request) {
	type ChirpRequest struct {
		Body string `json:"body"`
	}
	type ErrorResponse struct {
		Error string `json:"error"`
	}
	type returnVals struct {
		CleanedBody string `json:"cleaned_body"`
	}

	decoder := json.NewDecoder(r.Body)
	Chirp := ChirpRequest{}
	err := decoder.Decode(&Chirp)
	if err != nil {
		errorResponse := ErrorResponse{
			Error: "Invalid JSON in request body",
		}
		jsonResponse, _ := json.Marshal(errorResponse)
		w.WriteHeader(400)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonResponse)
		return
	}

	if len(Chirp.Body) > 140 {
		errorResponse := ErrorResponse{
			Error: "Chirp is too long",
		}
		jsonResponse, err := json.Marshal(errorResponse)
		if err != nil {
			// Log the error, set the status code and return
		}
		w.WriteHeader(400)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonResponse)
		return

	} else {

		w.WriteHeader(200)
		w.Header().Set("Content-Type", "application/json")

		// Define the set of keywords to censor
		keywords := map[string]struct{}{
			"kerfuffle": {},
			"sharbert":  {},
			"fornax":    {}, // Add more keywords as needed
		}

		// Split the body into a slice
		words := strings.Fields(Chirp.Body)

		// Loop through the words and replace keywords with "****"
		for i, word := range words {
			if _, found := keywords[strings.ToLower(word)]; found {
				words[i] = "****"
			}
		}

		modifiedBody := strings.Join(words, " ")
		cleanedBody := returnVals{
			CleanedBody: modifiedBody,
		}
		jsonResponse, _ := json.Marshal(cleanedBody)

		// Write the modified response
		w.Write(jsonResponse)
		return

	}

}

func createServer(dbQueries *database.Queries) {
	apiC := apiConfig{
		db: dbQueries,
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

	mux.HandleFunc("GET /admin/metrics", apiC.metricsHandler) //metrics holder
	mux.HandleFunc("POST /admin/reset", apiC.resetHandler)
	mux.HandleFunc("POST /api/users", apiC.handlerUsersCreate)
	mux.HandleFunc("POST /api/validate_chirp", Validate_Chirp_Endpoint)

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
