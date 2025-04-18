package main

import (
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
)

type AdmissionReview struct {
	Request  *admissionv1.AdmissionRequest  `json:"request,omitempty"`
	Response *admissionv1.AdmissionResponse `json:"response,omitempty"`
}

type PodSpec struct {
	Spec corev1.PodSpec `json:"spec"`
}
