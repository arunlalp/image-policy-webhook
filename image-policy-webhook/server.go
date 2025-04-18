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

func handleValidate(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "could not read request", http.StatusBadRequest)
		return
	}

	var review admissionv1.AdmissionReview
	err = json.Unmarshal(body, &review)
	if err != nil {
		http.Error(w, "could not decode admission review", http.StatusBadRequest)
		return
	}

	var pod corev1.Pod
	if err := json.Unmarshal(review.Request.Object.Raw, &pod); err != nil {
		http.Error(w, "could not unmarshal pod", http.StatusBadRequest)
		return
	}

	allowed := true
	var reason string

	for _, container := range pod.Spec.Containers {
		if !strings.HasPrefix(container.Image, "docker.io") && !isDockerHubShorthand(container.Image) {
			allowed = false
			reason = fmt.Sprintf("container image %s is not from docker.io", container.Image)
			break
		}
	}

	review.Response = &admissionv1.AdmissionResponse{
		UID:     review.Request.UID,
		Allowed: allowed,
	}

	if !allowed {
		review.Response.Result = &corev1.Status{
			Message: reason,
		}
	}

	respBytes, err := json.Marshal(review)
	if err != nil {
		http.Error(w, "could not encode response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(respBytes)
}

func isDockerHubShorthand(image string) bool {
	// Accepts images like nginx:latest or library/nginx
	if !strings.Contains(image, "/") {
		return true
	}
	if strings.HasPrefix(image, "library/") {
		return true
	}
	return false
}
