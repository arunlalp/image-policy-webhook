package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func main() {
	http.HandleFunc("/validate", handleAdmissionReview)
	fmt.Println("Starting server on :8443...")
	err := http.ListenAndServeTLS(":8443", "tls.crt", "tls.key", nil)
	if err != nil {
		panic(err)
	}
}

func handleAdmissionReview(w http.ResponseWriter, r *http.Request) {
	var review admissionv1.AdmissionReview
	body, _ := ioutil.ReadAll(r.Body)
	_ = json.Unmarshal(body, &review)

	allowed := true
	reason := ""

	// Only check if it's a pod
	if review.Request.Kind.Kind == "Pod" {
		var pod corev1.Pod
		_ = json.Unmarshal(review.Request.Object.Raw, &pod)

		for _, container := range pod.Spec.Containers {
			if !isDockerHubImage(container.Image) {
				allowed = false
				reason = fmt.Sprintf("Image %s is not from Docker Hub", container.Image)
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
		response.Response.Result = &corev1.Status{
			Message: reason,
		}
	}

	respBytes, _ := json.Marshal(response)
	w.Header().Set("Content-Type", "application/json")
	w.Write(respBytes)
}

func isDockerHubImage(image string) bool {
	// Accept if image starts with docker.io or has no registry prefix
	return image == "docker.io" || !containsRegistryPrefix(image)
}

func containsRegistryPrefix(image string) bool {
	return !(len(image) >= 11 && image[:11] == "docker.io/") && !startsWithNoRegistry(image)
}

func startsWithNoRegistry(image string) bool {
	// no domain prefix like "nginx", "library/nginx", or "user/image"
	return !containsDotOrColon(image)
}

func containsDotOrColon(image string) bool {
	for _, c := range image {
		if c == '.' || c == ':' {
			return true
		}
	}
	return false
}
