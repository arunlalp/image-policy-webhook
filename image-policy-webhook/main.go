package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

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
	if err := json.Unmarshal(pod, &podSpec); err != nil {
		http.Error(w, "Invalid pod spec", http.StatusInternalServerError)
		return
	}

	containers := podSpec["spec"].(map[string]interface{})["containers"].([]interface{})
	for _, c := range containers {
		image := c.(map[string]interface{})["image"].(string)

		// Allow only docker.io/ or implicit DockerHub images
		if !(strings.HasPrefix(image, "docker.io/") || isImplicitDockerHubImage(image)) {
			allowed = false
			message = fmt.Sprintf("Only images from docker.io are allowed. Image %s is denied.", image)
			break
		}
	}

	admissionResponse := &admissionv1.AdmissionResponse{
		UID:     review.Request.UID,
		Allowed: allowed,
		Result: &metav1.Status{
			Message: message,
		},
	}

	responseReview := admissionv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "admission.k8s.io/v1",
			Kind:       "AdmissionReview",
		},
		Response: admissionResponse,
	}

	respBytes, err := json.Marshal(responseReview)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(respBytes)
}

func isImplicitDockerHubImage(image string) bool {
	return !strings.Contains(image, ".") && !strings.Contains(image, ":")
}

func main() {
	http.HandleFunc("/validate", serve)
	fmt.Println("Starting server on port 8443...")
	if err := http.ListenAndServeTLS(":8443", "/etc/webhook/certs/tls.crt", "/etc/webhook/certs/tls.key", nil); err != nil {
		panic(err)
	}
}
