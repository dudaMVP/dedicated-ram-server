package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

func ReadinessEndpoint(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	w.WriteHeader(http.StatusOK)

	message := "OK"
	w.Write([]byte(message))
}

type apiConfig struct {
	fileserverHits atomic.Int32
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
	fmt.Fprintf(w, "Hits: %d", count)

}

func (cfg *apiConfig) resetHandler(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)
	w.WriteHeader(http.StatusOK)

}

func createServer() {
	apiC := apiConfig{}
	mux := http.NewServeMux()
	server := &http.Server{
		Handler: mux,
		Addr:    ":8080",
	}
	mux.HandleFunc("/healthz", ReadinessEndpoint)

	// Create file server and strip the /app/ prefix
	fileServer := http.FileServer(http.Dir("."))
	handler := http.StripPrefix("/app/", fileServer)
	mux.Handle("/app/", apiC.middlewareMetricsInc(handler))

	mux.Handle("/assets", http.FileServer(http.Dir("./assets/logo.png")))

	mux.HandleFunc("/metrics", apiC.metricsHandler)
	mux.HandleFunc("/reset", apiC.resetHandler)

	addr := server.Addr
	fmt.Println("Starting server on", addr)
	if err := server.ListenAndServe(); err != nil {
		fmt.Println("Error starting server:", err)
	}
}

func main() {

	createServer()
}
