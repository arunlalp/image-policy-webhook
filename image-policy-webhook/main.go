package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Config holds the webhook configuration
type Config struct {
	TLSCertFile string
	TLSKeyFile  string
	Port        int
	AllowedRegistries []string
}

// AdmissionReviewResponse is a helper to build the response
type AdmissionReviewResponse struct {
	Allowed bool   `json:"allowed"`
	Message string `json:"message,omitempty"`
}

func main() {
	// Configuration - in a real app you'd get these from env vars or flags
	config := Config{
		TLSCertFile: "/etc/webhook/certs/tls.crt",
		TLSKeyFile:  "/etc/webhook/certs/tls.key",
		Port:        8443,
		AllowedRegistries: []string{
			"docker.io",
			"gcr.io",
			"quay.io",
		},
	}

	http.HandleFunc("/validate", validateHandler(config))
	http.HandleFunc("/healthz", healthzHandler)

	log.Printf("Starting server on port %d...", config.Port)
	err := http.ListenAndServeTLS(
		fmt.Sprintf(":%d", config.Port),
		config.TLSCertFile,
		config.TLSKeyFile,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func validateHandler(config Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
			return
		}

		var admissionReview admissionv1.AdmissionReview
		if err := json.NewDecoder(r.Body).Decode(&admissionReview); err != nil {
			http.Error(w, fmt.Sprintf("Error decoding request: %v", err), http.StatusBadRequest)
			return
		}

		// Default response (deny by default)
		response := AdmissionReviewResponse{
			Allowed: false,
			Message: "No containers found to validate",
		}

		// Get the pod object from the admission request
		var pod corev1.Pod
		if err := json.Unmarshal(admissionReview.Request.Object.Raw, &pod); err != nil {
			response.Message = fmt.Sprintf("Error unmarshaling pod: %v", err)
			sendResponse(w, admissionReview, response)
			return
		}

		// Validate all containers in the pod
		var invalidImages []string
		for _, container := range pod.Spec.Containers {
			if !isImageAllowed(container.Image, config.AllowedRegistries) {
				invalidImages = append(invalidImages, container.Image)
			}
		}

		if len(invalidImages) > 0 {
			response.Message = fmt.Sprintf("The following images are from unauthorized registries: %v", invalidImages)
		} else {
			response.Allowed = true
			response.Message = "All images are from allowed registries"
		}

		sendResponse(w, admissionReview, response)
	}
}

func isImageAllowed(image string, allowedRegistries []string) bool {
	parts := strings.Split(image, "/")
	if len(parts) < 2 {
		// Image like "nginx" defaults to docker.io
		return contains(allowedRegistries, "docker.io")
	}

	registry := parts[0]
	return contains(allowedRegistries, registry)
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func sendResponse(w http.ResponseWriter, admissionReview admissionv1.AdmissionReview, response AdmissionReviewResponse) {
	admissionResponse := &admissionv1.AdmissionResponse{
		UID:     admissionReview.Request.UID,
		Allowed: response.Allowed,
		Result: &metav1.Status{
			Message: response.Message,
		},
	}

	admissionReview.Response = admissionResponse
	respBytes, err := json.Marshal(admissionReview)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error marshaling response: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(respBytes)
}

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}