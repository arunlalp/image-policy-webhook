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
	cert := flag.String("tls-cert", "/etc/webhook/certs/tls.crt", "cert file")
	key := flag.String("tls-key", "/etc/webhook/certs/tls.key", "key file")
	flag.Parse()

	http.HandleFunc("/validate", handleAdmissionReview)

	log.Println("✅ Starting webhook server on port 8443...")
	err := http.ListenAndServeTLS(":8443", *cert, *key, nil)
	if err != nil {
		log.Fatalf("❌ Server failed to start: %v", err)
	}
}

func handleAdmissionReview(w http.ResponseWriter, r *http.Request) {
	var review admissionv1.AdmissionReview

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Could not read request", http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(body, &review); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	allowed := true
	reason := ""

	if review.Request.Kind.Kind == "Pod" {
		var pod corev1.Pod
		if err := json.Unmarshal(review.Request.Object.Raw, &pod); err != nil {
			http.Error(w, "Invalid pod object", http.StatusBadRequest)
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
		response.Response.Result = &metav1.Status{Message: reason}
	}

	respBytes, _ := json.Marshal(response)
	w.Header().Set("Content-Type", "application/json")
	w.Write(respBytes)
}

func isDockerHubImage(image string) bool {
	if strings.Contains(image, "/") && strings.Contains(image, ".") {
		return strings.HasPrefix(image, "docker.io/")
	}
	return true
}
