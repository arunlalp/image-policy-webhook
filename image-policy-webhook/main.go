package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
)

func main() {
	http.HandleFunc("/validate", handleValidation)
	fmt.Println("🚀 Starting webhook server on :8443...")
	err := http.ListenAndServeTLS(":8443", "/etc/webhook/certs/tls.crt", "/etc/webhook/certs/tls.key", nil)
	if err != nil {
		panic(err)
	}
}

func handleValidation(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "could not read request", http.StatusBadRequest)
		return
	}

	var admissionReview admissionv1.AdmissionReview
	err = json.Unmarshal(body, &admissionReview)
	if err != nil {
		http.Error(w, "could not decode admission review", http.StatusBadRequest)
		return
	}

	reviewResponse := admissionv1.AdmissionResponse{
		UID:     admissionReview.Request.UID,
		Allowed: true,
	}

	var pod corev1.Pod
	if err := json.Unmarshal(admissionReview.Request.Object.Raw, &pod); err == nil {
		for _, container := range pod.Spec.Containers {
			if !isDockerHubImage(container.Image) {
				reviewResponse.Allowed = false
				reviewResponse.Result = &corev1.Status{
					Message: fmt.Sprintf("Image %s is not from Docker Hub", container.Image),
				}
				break
			}
		}
	}

	admissionReview.Response = &reviewResponse
	resp, err := json.Marshal(admissionReview)
	if err != nil {
		http.Error(w, "could not encode response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

func isDockerHubImage(image string) bool {
	// Allow only docker.io or images without a registry (like nginx:latest)
	if strings.HasPrefix(image, "docker.io/") || !strings.Contains(image, ".") {
		return true
	}
	return false
}
