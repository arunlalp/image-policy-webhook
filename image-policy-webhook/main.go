package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func main() {
	certFile := flag.String("tls-cert", "/etc/webhook/certs/tls.crt", "TLS certificate file")
	keyFile := flag.String("tls-key", "/etc/webhook/certs/tls.key", "TLS key file")
	flag.Parse()

	http.HandleFunc("/validate", handleAdmissionReview)

	log.Println("✅ Webhook server started on port 8443...")
	err := http.ListenAndServeTLS(":8443", *certFile, *keyFile, nil)
	if err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}

func handleAdmissionReview(w http.ResponseWriter, r *http.Request) {
	var admissionReview admissionv1.AdmissionReview

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Could not read request body", http.StatusBadRequest)
		log.Println("❌ Error reading request:", err)
		return
	}

	if err := json.Unmarshal(body, &admissionReview); err != nil {
		http.Error(w, "Could not parse admission review", http.StatusBadRequest)
		log.Println("❌ Error unmarshaling admission review:", err)
		return
	}

	review := admissionReview.Request
	log.Printf("🔍 Reviewing resource: %s/%s", review.Kind.Kind, review.Name)

	allowed := true
	reason := ""

	if review.Kind.Kind == "Pod" {
		var pod corev1.Pod
		if err := json.Unmarshal(review.Object.Raw, &pod); err != nil {
			http.Error(w, "Could not parse pod object", http.StatusBadRequest)
			log.Println("❌ Error parsing pod object:", err)
			return
		}

		for _, container := range pod.Spec.Containers {
			log.Printf("🔎 Checking image: %s", container.Image)
			if !isDockerHubImage(container.Image) {
				allowed = false
				reason = fmt.Sprintf("Image %q is not from Docker Hub", container.Image)
				log.Println("❌", reason)
				break
			}
		}
	}

	response := admissionv1.AdmissionReview{
		TypeMeta: admissionReview.TypeMeta,
		Response: &admissionv1.AdmissionResponse{
			UID:     review.UID,
			Allowed: allowed,
		},
	}

	if !allowed {
		response.Response.Result = &metav1.Status{
			Message: reason,
		}
	}

	respBytes, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Could not marshal response", http.StatusInternalServerError)
		log.Println("❌ Error marshaling response:", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(respBytes)
	log.Println("✅ AdmissionReview response sent")
}

func isDockerHubImage(image string) bool {
	// If the image contains a domain (registry), reject it unless it's docker.io
	if strings.Contains(image, "/") && strings.Contains(image, ".") {
		return strings.HasPrefix(image, "docker.io/")
	}
	return true // Accept images like 'nginx' or 'busybox'
}
