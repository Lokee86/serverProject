package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
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
	decoder := json.NewDecoder(r.Body)
	checkedChirp := chirp{}
	err := decoder.Decode(&checkedChirp)
	if err != nil {
		log.Printf("Internal Server Error: %v", err)
		response.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(checkedChirp.Body) > 140 {
		errorResponse(response, http.StatusBadRequest, "Chirp is too long")
		return
	}
	log.Println("Chirp validated")

	if len(checkedChirp.Body) <= 140 {
		checkProfanity(checkedChirp, response)
	}

}

func checkProfanity(chirp chirp, response http.ResponseWriter) {
	type cleanedChirp struct {
		CleanedBody string `json:"cleaned_body"`
	}

	profaneWords := []string{"kerfuffle", "sharbert", "fornax"}
	listToCheck := strings.Split(chirp.Body, " ")
	for i, word := range listToCheck {
		for _, badWord := range profaneWords {
			if strings.ToLower(word) == badWord {
				listToCheck[i] = "****"
			}
		}
	}
	newChirpBody := strings.Join(listToCheck, " ")
	newChirp := cleanedChirp{}
	newChirp.CleanedBody = newChirpBody
	log.Println("Profanity cleaned from chirp")
	jsonResponse(response, http.StatusOK, newChirp)
}
