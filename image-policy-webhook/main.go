package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

// Allowed image name
const ALLOWED = "nginx"

// ImageReview represents the structure of the incoming JSON
type ImageReview struct {
	Spec struct {
		Containers  []struct {
			Image string `json:"image"`
		} `json:"containers"`
		Annotations map[string]string `json:"annotations"`
	} `json:"spec"`
	Status map[string]interface{} `json:"status"`
}

func main() {
	fmt.Println("Starting webhook")

	// Set up HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleRequest)

	server := &http.Server{
		Addr:    ":443",
		Handler: mux,
	}

	// Configure TLS
	certFile := "/etc/ssl/certs/webhook-server.crt"
	keyFile := "/etc/ssl/private/webhook-server.key"
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		log.Fatalf("Failed to load certificates: %v", err)
	}

	server.TLSConfig = &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	// Start the server
	log.Fatal(server.ListenAndServeTLS("", ""))
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("I am an imagePolicyWebhook example!\nYou need to post a JSON object of kind ImageReview"))

	case http.MethodPost:
		handlePost(w, r)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handlePost(w http.ResponseWriter, r *http.Request) {
	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Parse JSON
	var review ImageReview
	if err := json.Unmarshal(body, &review); err != nil {
		http.Error(w, "Failed to parse JSON", http.StatusBadRequest)
		return
	}

	// Check if annotations exist
	if len(review.Spec.Annotations) > 0 {
		review.Status = map[string]interface{}{
			"allowed": true,
			"reason":  "You broke the glass",
		}
	} else {
		// Check images
		for _, container := range review.Spec.Containers {
			if ALLOWED == container.Image {
				review.Status = map[string]interface{}{
					"allowed": true,
				}
			} else {
				review.Status = map[string]interface{}{
					"allowed": false,
					"reason":  "Only nginx images are allowed",
				}
				break
			}
		}
	}

	// Prepare response
	response, err := json.Marshal(review)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(response)))
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}