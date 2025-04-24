package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/Lokee86/serverProject/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

const pathRoot = "."
const port = ":8080"

// Create router and server
func createServer(apiCfg *apiConfig) *http.Server {
	router := http.NewServeMux()
	handler := http.StripPrefix("/app/", http.FileServer(http.Dir(pathRoot)))
	router.Handle("/app/", apiCfg.serverHitCounter(handler))
	router.Handle("/Assets/", handler)
	router.HandleFunc("GET /admin/metrics", apiCfg.metricsHandler)
	router.HandleFunc("GET /api/healthz", healthCheck)
	router.HandleFunc("POST /admin/reset", apiCfg.resetCounter)
	router.HandleFunc("POST /api/chirps", apiCfg.validateChirp)
	router.HandleFunc("GET /api/chirps", apiCfg.fetchChirps)
	router.HandleFunc("POST /api/users", apiCfg.createUserHandler)
	return &http.Server{
		Addr:    port,
		Handler: router,
	}
}

// EXECUTE MAIN FUNCTION
func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Printf("Error Loading Database: %v", err)
	}
	apiCfg := &apiConfig{}
	apiCfg.databaseQueries = database.New(db)
	server := createServer(apiCfg)
	apiCfg.platform = os.Getenv("PLATFORM")
	log.Printf("Server running on Port%v from %v", port, pathRoot)
	log.Fatal(server.ListenAndServe())
}
