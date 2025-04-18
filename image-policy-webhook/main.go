package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type AdmissionReview struct {
	*admissionv1.AdmissionReview
}

type Pod struct {
	Spec struct {
		Containers []struct {
			Image string `json:"image"`
		} `json:"containers"`
	} `json:"spec"`
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/validate", validateImage).Methods("POST")

	// Start HTTPS server with TLS certificates
	log.Println("Starting server on :8080...")
	err := http.ListenAndServeTLS(":8080", "/app/certs/tls.crt", "/app/certs/tls.key", r)
	if err != nil {
		log.Fatal("ListenAndServeTLS: ", err)
	}
}

func validateImage(w http.ResponseWriter, r *http.Request) {
	allowedRegistries := []string{"registry.example.com", "docker.io"}

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Unable to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Parse admission review request
	var admissionReview AdmissionReview
	if err := json.Unmarshal(body, &admissionReview); err != nil {
		http.Error(w, "Unable to parse admission review", http.StatusBadRequest)
		return
	}

	// Extract pod object
	var pod Pod
	if err := json.Unmarshal(admissionReview.Request.Object.Raw, &pod); err != nil {
		http.Error(w, "Unable to parse pod object", http.StatusBadRequest)
		return
	}

	// Validate container images
	for _, container := range pod.Spec.Containers {
		parts := strings.Split(container.Image, "/")
		registry := parts[0]

		allowed := false
		for _, allowedRegistry := range allowedRegistries {
			if registry == allowedRegistry {
				allowed = true
				break
			}
		}

		if !allowed {
			response := admissionv1.AdmissionReview{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "admission.k8s.io/v1",
					Kind:       "AdmissionReview",
				},
				Response: &admissionv1.AdmissionResponse{
					UID:     admissionReview.Request.UID,
					Allowed: false,
					Result: &metav1.Status{
						Message: "Image " + container.Image + " is not allowed. Allowed registries: " + strings.Join(allowedRegistries, ", "),
					},
				},
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}
	}

	// If all images are valid
	response := admissionv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "admission.k8s.io/v1",
			Kind:       "AdmissionReview",
		},
		Response: &admissionv1.AdmissionResponse{
			UID:     admissionReview.Request.UID,
			Allowed: true,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}