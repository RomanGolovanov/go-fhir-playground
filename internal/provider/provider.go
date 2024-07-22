package provider

import (
	"context"

	"github.com/samply/golang-fhir-models/fhir-models/fhir"
	"golang.org/x/oauth2"
)

type FhirProvider interface {
	Authorize(ctx context.Context, patientId, refreshToken string) (*ClientCredential, error)
	GetPatient(ctx context.Context, credential *ClientCredential) (*fhir.Patient, error)
}

type ClientCredential struct {
	PatientID string
	Token     *oauth2.Token
}
