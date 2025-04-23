package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
)

func main() {
	http.HandleFunc("/validate", serveValidate)
	fmt.Println("🚀 Starting ImagePolicyWebhook server on port 8443...")
	log.Fatal(http.ListenAndServeTLS(":8443", "/etc/webhook/certs/tls.crt", "/etc/webhook/certs/tls.key", nil))
}

func serveValidate(w http.ResponseWriter, r *http.Request) {
	var admissionReview admissionv1.AdmissionReview
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Could not read request", http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(body, &admissionReview)
	if err != nil {
		http.Error(w, "Could not parse JSON", http.StatusBadRequest)
		return
	}

	raw := admissionReview.Request.Object.Raw
	var pod corev1.Pod
	if err := json.Unmarshal(raw, &pod); err != nil {
		http.Error(w, "Could not unmarshal pod", http.StatusBadRequest)
		return
	}

	allowed := true
	reason := ""

	for _, container := range pod.Spec.Containers {
		image := container.Image
		parts := strings.Split(image, "/")
		if len(parts) > 1 && strings.Contains(parts[0], ".") {
			allowed = false
			reason = fmt.Sprintf("Only Docker Hub images are allowed. Found: %s", image)
			break
		}
	}

	admissionReview.Response = &admissionv1.AdmissionResponse{
		UID:     admissionReview.Request.UID,
		Allowed: allowed,
	}

	if !allowed {
		admissionReview.Response.Result = &metav1.Status{
			Message: reason,
		}
	}

	respBytes, err := json.Marshal(admissionReview)
	if err != nil {
		http.Error(w, "Could not encode response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(respBytes)
}
