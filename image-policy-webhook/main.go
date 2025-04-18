package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func main() {
	http.HandleFunc("/validate", handleAdmissionReview)
	fmt.Println("Starting webhook server on port 8443...")
	err := http.ListenAndServeTLS(":8443", "/etc/webhook/certs/tls.crt", "/etc/webhook/certs/tls.key", nil)
	if err != nil {
		panic(err)
	}
}

func handleAdmissionReview(w http.ResponseWriter, r *http.Request) {
	var review admissionv1.AdmissionReview
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "could not read request", http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(body, &review)
	if err != nil {
		http.Error(w, "could not parse admission review", http.StatusBadRequest)
		return
	}

	allowed := true
	reason := ""

	if review.Request.Kind.Kind == "Pod" {
		var pod corev1.Pod
		err := json.Unmarshal(review.Request.Object.Raw, &pod)
		if err != nil {
			http.Error(w, "could not parse pod object", http.StatusBadRequest)
			return
		}

		for _, container := range pod.Spec.Containers {
			if !isDockerHubImage(container.Image) {
				allowed = false
				reason = fmt.Sprintf("Image %q is not from Docker Hub", container.Image)
				break
			}
		}
	}

	response := admissionv1.AdmissionReview{
		Response: &admissionv1.AdmissionResponse{
			UID:     review.Request.UID,
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
		http.Error(w, "could not encode response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(respBytes)
}

func isDockerHubImage(image string) bool {
	// If image contains a registry (like "quay.io/" or "gcr.io/"), deny it
	if strings.Contains(image, "://") {
		return false
	}

	if strings.HasPrefix(image, "docker.io/") || !strings.Contains(image, ".") {
		return true
	}

	return false
}
