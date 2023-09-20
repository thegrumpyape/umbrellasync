package umbrella

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"golang.org/x/oauth2/clientcredentials"
)

type UmbrellaService struct {
	authUrl        string
	deploymentsUrl string
	adminUrl       string
	policiesUrl    string
	reportsUrl     string
	client         *http.Client
	logger         *log.Logger
}

type Status struct {
	Code int    `json:"code"`
	Text string `json:"text"`
}

type Meta struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
}

func CreateClient(hostname string, version string, clientID string, clientSecret string, logger *log.Logger) (UmbrellaService, error) {
	tokenUrl := fmt.Sprintf("https://%s/auth/%s/token", hostname, version)
	policiesUrl := fmt.Sprintf("https://%s/policies/%s", hostname, version)

	httpClient, err := createHTTPClient(clientID, clientSecret, tokenUrl)
	if err != nil {
		return UmbrellaService{}, fmt.Errorf("failed to create HTTP client: %w", err)
	}
	return UmbrellaService{
		policiesUrl: policiesUrl,
		client:      httpClient,
		logger:      logger,
	}, nil
}

// createHTTPClient creates and returns an HTTP client for the UmbrellaService.
func createHTTPClient(clientID, clientSecret, tokenUrl string) (*http.Client, error) {
	clientConfig := clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     tokenUrl,
	}
	return clientConfig.Client(context.TODO()), nil
}

func CreateJSONPayload(data interface{}) (*bytes.Buffer, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("error marshalling payload: %w", err)
	}

	return bytes.NewBuffer(jsonData), nil
}

func RemoveAtIndex(s []string, index int) []string {
	return append(s[:index], s[index+1:]...)
}
