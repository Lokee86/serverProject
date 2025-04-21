package main

import (
	"encoding/json"
	"log"
	"net/http"
)

const pathRoot = "."
const port = ":8080"

type chirp struct {
	Body string `json:"body"`
}

// Create router and server
func createServer(apiCfg *apiConfig) *http.Server {
	router := http.NewServeMux()
	handler := http.StripPrefix("/app/", http.FileServer(http.Dir(pathRoot)))
	router.Handle("/app/", apiCfg.serverHitCounter(handler))
	router.HandleFunc("GET /admin/metrics", apiCfg.metricsHandler)
	router.HandleFunc("GET /api/healthz", healthCheck)
	router.HandleFunc("POST /admin/reset", apiCfg.resetCounter)
	router.HandleFunc("POST /api/validate_chirp", validateChirp)
	return &http.Server{
		Addr:    port,
		Handler: router,
	}
}

// EXECUTE MAIN FUNCTION
func main() {
	apiCfg := &apiConfig{}
	server := createServer(apiCfg)
	log.Printf("Server running on Port%v from %v", port, pathRoot)
	log.Fatal(server.ListenAndServe())
}

func validateChirp(response http.ResponseWriter, r *http.Request) {
	type validChirp struct {
		Valid bool `json:"valid"`
	}

	type invalidChirp struct {
		Error string `json:"error"`
	}

	decoder := json.NewDecoder(r.Body)
	checkedChirp := chirp{}
	err := decoder.Decode(&checkedChirp)
	if err != nil {
		log.Printf("Internal Server Error: %v", err)
		response.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(checkedChirp.Body) <= 140 {
		chirp := validChirp{Valid: true}
		jsonData, err := json.Marshal(chirp)
		if err != nil {
			log.Printf("Internal Server Error: %v", err)
			response.WriteHeader(http.StatusInternalServerError)
			return
		}
		response.WriteHeader(http.StatusOK)
		response.Write(jsonData)
	}

	if len(checkedChirp.Body) > 140 {
		chirp := invalidChirp{Error: "Chirp is too long"}
		jsonData, err := json.Marshal(chirp)
		if err != nil {
			log.Printf("Internal Server Error: %v", err)
			response.WriteHeader(http.StatusInternalServerError)
			return
		}
		response.WriteHeader(http.StatusBadRequest)
		response.Write(jsonData)

	}
	log.Println("Chirp validated")

}
