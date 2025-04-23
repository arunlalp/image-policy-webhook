package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

// ImageReview represents the structure of the admission request
type ImageReview struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Spec       struct {
			Containers  []Container       `json:"containers"`
			Annotations map[string]string `json:"annotations"`
	} `json:"spec"`
	Status struct {
			Allowed bool   `json:"allowed"`
			Reason  string `json:"reason,omitempty"`
	} `json:"status,omitempty"`
}

// Container defines the container definition in an admission request
type Container struct {
	Image string `json:"image"`
}

const (
	// AllowedRegistry defines the allowed container registry
	AllowedRegistry = "docker.io"
	// AllowedImage defines the allowed image name
	AllowedImage = "nginx"
)

func main() {
	// Setup logging
	log.Println("Starting imagePolicyWebhook admission controller...")

	// Define HTTP endpoints
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/validate", validateHandler)

	// Get certificate paths from environment or use defaults
	certFile := getEnvOrDefault("CERT_FILE", "./certs/webhook-server.crt")
	keyFile := getEnvOrDefault("KEY_FILE", "./keys/webhook-server.key")

	// Check if certificates exist
	if _, err := os.Stat(certFile); os.IsNotExist(err) {
			log.Printf("Certificate file %s does not exist", certFile)
			log.Println("You can generate certificates using ./generate-certs.sh")
			// For development/testing, we can optionally run without TLS
			devMode := getEnvOrDefault("DEV_MODE", "false")
			if devMode == "true" {
					port := getEnvOrDefault("PORT", "8080")
					serverAddr := fmt.Sprintf("0.0.0.0:%s", port)
					log.Printf("Running in DEV_MODE without TLS on %s", serverAddr)
					log.Fatal(http.ListenAndServe(serverAddr, nil))
					return
			}
	}

	// Setup HTTPS server
	port := getEnvOrDefault("PORT", "8443")
	serverAddr := fmt.Sprintf("0.0.0.0:%s", port)
	log.Printf("Server listening on %s", serverAddr)
	log.Fatal(http.ListenAndServeTLS(serverAddr, certFile, keyFile, nil))
}

// getEnvOrDefault returns environment variable value or default if not set
func getEnvOrDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
			return value
	}
	return defaultValue
}

// rootHandler handles the root path requests
func rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("I am an imagePolicyWebhook example!\nYou need to post a JSON object of kind ImageReview to /validate"))
}

// validateHandler handles webhook validation requests
func validateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte("Method not allowed. Only POST requests are supported."))
			return
	}

	// Read request body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
			log.Printf("Error reading body: %v", err)
			http.Error(w, "Error reading request body", http.StatusBadRequest)
			return
	}
	defer r.Body.Close()

	// Parse JSON request
	var review ImageReview
	if err := json.Unmarshal(body, &review); err != nil {
			log.Printf("Error unmarshaling JSON: %v", err)
			http.Error(w, "Error parsing JSON request", http.StatusBadRequest)
			return
	}

	// Check for "break-glass" annotations
	if annotations := review.Spec.Annotations; annotations != nil && len(annotations) > 0 {
			log.Println("Found annotations, applying break-glass policy")
			review.Status.Allowed = true
			review.Status.Reason = "You broke the glass"
			sendResponse(w, review)
			return
	}

	// Validate each container image
	allowed := true
	reason := ""

	for _, container := range review.Spec.Containers {
			image := container.Image
			log.Printf("Validating image: %s", image)

			// Check if image is from DockerHub and is nginx
			isAllowed := isNginxFromDockerHub(image)
			if isAllowed {
					log.Printf("Image %s is allowed", image)
			} else {
					log.Printf("Image %s is not allowed", image)
					allowed = false
					reason = "Only nginx images from DockerHub are allowed"
					break
			}
	}

	// Set the response status
	review.Status.Allowed = allowed
	if !allowed {
			review.Status.Reason = reason
	}

	sendResponse(w, review)
}

// isNginxFromDockerHub checks if an image is nginx from DockerHub
func isNginxFromDockerHub(image string) bool {
	// Common image formats:
	// - nginx
	// - docker.io/nginx
	// - docker.io/library/nginx
	// - nginx:latest
	// - docker.io/nginx:latest
	// - docker.io/library/nginx:latest

	// First, strip off the tag/digest if present
	imageParts := strings.Split(image, ":")
	baseImage := imageParts[0]
	
	// Check for the simplest case: just "nginx"
	if baseImage == AllowedImage {
			return true
	}
	
	// Check for docker.io/nginx
	if baseImage == fmt.Sprintf("%s/%s", AllowedRegistry, AllowedImage) {
			return true
	}
	
	// Check for docker.io/library/nginx
	if baseImage == fmt.Sprintf("%s/library/%s", AllowedRegistry, AllowedImage) {
			return true
	}
	
	return false
}

// sendResponse sends the JSON response back to the client
func sendResponse(w http.ResponseWriter, review ImageReview) {
	w.Header().Set("Content-Type", "application/json")
	responseJSON, err := json.Marshal(review)
	if err != nil {
			log.Printf("Error marshaling response: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(responseJSON)
}