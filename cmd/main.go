package main

import (
	"context"
	"flag"
	"log"

	"github.com/RomanGolovanov/go-fhir-playground/internal/provider"
)

type Config struct {
	APIBaseURL   string
	ClientID     string
	ClientSecret string
	PatientID    string
	RefreshToken string
}

func main() {
	ctx := context.Background()

	config := Config{}
	flag.StringVar(&config.APIBaseURL, "api_base_url", "", "API base URL")
	flag.StringVar(&config.ClientID, "client_id", "", "Client ID")
	flag.StringVar(&config.ClientSecret, "client_secret", "", "Client Secret")

	flag.StringVar(&config.PatientID, "patient", "", "Patient's ID")
	flag.StringVar(&config.RefreshToken, "token", "", "Patient's refresh token")

	flag.Parse()

	client := provider.NewLighthouseProvider(config.APIBaseURL, config.ClientID, config.ClientSecret)

	cred, err := client.Authorize(ctx, config.PatientID, config.RefreshToken)
	if err != nil {
		log.Fatal(err)
	}

	patient, err := client.GetPatient(ctx, cred)
	if err != nil {
		log.Fatal("failed to retrieve patient: %w\n", err)
	}

	json, err := patient.MarshalJSON()
	if err != nil {
		log.Fatal("failed to serialize patient: %w\n", err)
	}

	log.Println(string(json))
}
