package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func serve(w http.ResponseWriter, r *http.Request) {
	var review admissionv1.AdmissionReview
	if err := json.NewDecoder(r.Body).Decode(&review); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	allowed := true
	message := ""

	pod := review.Request.Object.Raw
	var podSpec map[string]interface{}
	json.Unmarshal(pod, &podSpec)

	containers := podSpec["spec"].(map[string]interface{})["containers"].([]interface{})
	for _, c := range containers {
		image := c.(map[string]interface{})["image"].(string)
		if len(image) < 10 || (image[:10] != "docker.io" && !isImplicitDockerHubImage(image)) {
			allowed = false
			message = fmt.Sprintf("Only images from docker.io are allowed. Image %s is denied.", image)
			break
		}
	}

	review.Response = &admissionv1.AdmissionResponse{
		UID:     review.Request.UID,
		Allowed: allowed,
		Result: &metav1.Status{
			Message: message,
		},
	}

	resp, _ := json.Marshal(review)
	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

// Handles cases like "nginx" or "library/nginx"
func isImplicitDockerHubImage(image string) bool {
	return !containsDotOrColon(image) // no registry prefix, means it's implicitly docker.io
}

func containsDotOrColon(s string) bool {
	for _, r := range s {
		if r == '.' || r == ':' {
			return true
		}
	}
	return false
}

func main() {
	http.HandleFunc("/validate", serve)
	fmt.Println("Starting server on port 8443...")
	http.ListenAndServeTLS(":8443", "/etc/webhook/certs/tls.crt", "/etc/webhook/certs/tls.key", nil)
}
