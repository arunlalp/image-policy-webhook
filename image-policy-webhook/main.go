package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func imagePolicyHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Could not read request body", http.StatusBadRequest)
		log.Println("Error reading body:", err)
		return
	}

	var admissionReview admissionv1.AdmissionReview
	if err := json.Unmarshal(body, &admissionReview); err != nil {
		http.Error(w, "Could not parse AdmissionReview", http.StatusBadRequest)
		log.Println("Error unmarshaling AdmissionReview:", err)
		return
	}

	pod := corev1.Pod{}
	if err := json.Unmarshal(admissionReview.Request.Object.Raw, &pod); err != nil {
		http.Error(w, "Could not parse Pod object", http.StatusBadRequest)
		log.Println("Error unmarshaling Pod:", err)
		return
	}

	allowed := true
	message := "All container images are valid"

	for _, container := range pod.Spec.Containers {
		log.Printf("Checking image: %s\n", container.Image)
		if !strings.HasPrefix(container.Image, "techiescamp/") {
			allowed = false
			message = fmt.Sprintf("Image %s is not allowed. Only 'techiescamp/' images are allowed", container.Image)
			break
		}
	}

	admissionReview.Response = &admissionv1.AdmissionResponse{
		UID:     admissionReview.Request.UID,
		Allowed: allowed,
		Result: &metav1.Status{
			Message: message,
		},
	}

	respBytes, err := json.Marshal(admissionReview)
	if err != nil {
		http.Error(w, "Could not encode response", http.StatusInternalServerError)
		log.Println("Error marshaling response:", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(respBytes)
	log.Printf("Admission decision: %v - %s\n", allowed, message)
}

func main() {
	http.HandleFunc("/", imagePolicyHandler)
	log.Println("🚀 Webhook server started on port 443")
	err := http.ListenAndServeTLS(":443", "/certs/tls.crt", "/certs/tls.key", nil)
	if err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}
