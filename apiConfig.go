package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileServerHits atomic.Int32
}

// increment hit counter
func (a *apiConfig) serverHitCounter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.fileServerHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

// metrics handler
func (a *apiConfig) metricsHandler(response http.ResponseWriter, r *http.Request) {
	log.Printf("Hits: %v", a.fileServerHits.Load())
	response.WriteHeader(http.StatusOK)
	response.Header().Set("Content-Type", "text/html")
	response.Write(fmt.Appendf(nil,
		`<html>
	<body>
		<h1>Welcome, Chirpy Admin</h1>
		<p>Chirpy has been visited %d times!</p>
	</body>
</html>`, a.fileServerHits.Load()))
}

// reset counter
func (a *apiConfig) resetCounter(response http.ResponseWriter, r *http.Request) {
	a.fileServerHits.Store(0)
	response.WriteHeader(http.StatusOK)
	response.Write([]byte("Hits counter reset to 0"))
	log.Println("Hit counter reset to 0")

}
