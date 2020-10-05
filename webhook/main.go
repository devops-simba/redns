package main

import (
	"net/http"

	admissionApi "k8s.io/api/admission/v1"
)

type VerifyDnsCRD struct {
}

func (this *VerifyDnsCRD) HandleAdmission(
	action string,
	req *http.Request,
	ar *admissionApi.AdmissionReview) (*admissionApi.AdmissionResponse, error) {
	//
}
