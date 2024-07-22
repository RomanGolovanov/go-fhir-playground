package provider

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/samply/golang-fhir-models/fhir-models/fhir"
	"golang.org/x/oauth2"
)

type LighthouseClient struct {
	clientID     string
	clientSecret string
	tokenURL     string
	apiURL       string
}

func NewLighthouseProvider(baseURL, clientID, clientSecret string) LighthouseClient {
	return LighthouseClient{
		clientID:     clientID,
		clientSecret: clientSecret,
		tokenURL:     fmt.Sprintf("%s/oauth2/health/v1/token", baseURL),
		apiURL:       fmt.Sprintf("%s/services/fhir/v0/r4", baseURL),
	}
}

func (c *LighthouseClient) Authorize(ctx context.Context, patientId, refreshToken string) (*ClientCredential, error) {
	config := &oauth2.Config{
		ClientID:     c.clientID,
		ClientSecret: c.clientSecret,
		Endpoint: oauth2.Endpoint{
			TokenURL: c.tokenURL,
		},
	}
	tokenSource := config.TokenSource(ctx, &oauth2.Token{
		RefreshToken: refreshToken,
	})

	token, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	credential := &ClientCredential{
		PatientID: patientId,
		Token:     token,
	}
	return credential, nil
}

func (c *LighthouseClient) GetPatient(ctx context.Context, credential *ClientCredential) (*fhir.Patient, error) {
	if credential.Token == nil {
		return nil, fmt.Errorf("failed to get patient %s: not authorized", credential.PatientID)
	}

	url := fmt.Sprintf("%s/Patient?_id=%s", c.apiURL, credential.PatientID)

	bundle, err := c.getBundle(ctx, url, credential)
	if err != nil {
		return nil, fmt.Errorf("failed to get resource bundle: %w", err)
	}

	patient, err := fhir.UnmarshalPatient(bundle.Entry[0].Resource)
	if err != nil {
		return nil, fmt.Errorf("failed to decode patient: %w", err)
	}

	return &patient, nil
}

func (c *LighthouseClient) getBundle(ctx context.Context, url string, credential *ClientCredential) (*fhir.Bundle, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+credential.Token.AccessToken)
	req.Header.Set("Accept", "application/fhir+json")

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-200 response: %s", resp.Status)
	}

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	bundle, err := fhir.UnmarshalBundle(bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to decode bundle: %w", err)
	}
	return &bundle, nil
}
