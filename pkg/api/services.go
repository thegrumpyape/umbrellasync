package api

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"golang.org/x/oauth2/clientcredentials"
)

// createHTTPClient creates and returns an HTTP client for the UmbrellaService.
func createHTTPClient(clientID, clientSecret, authURL string) (*http.Client, error) {
	clientConfig := clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     fmt.Sprintf(endpointCreateAuthorizationToken, authURL),
	}
	return clientConfig.Client(context.TODO()), nil
}

func NewUmbrellaService(hostname string, version string, clientID string, clientSecret string, logger *log.Logger) (UmbrellaService, error) {
	authUrl := fmt.Sprintf(authPath, hostname, version)
	deploymentsUrl := fmt.Sprintf(deployPath, hostname, version)
	adminUrl := fmt.Sprintf(adminPath, hostname, version)
	policiesUrl := fmt.Sprintf(policiesPath, hostname, version)
	reportsUrl := fmt.Sprintf(reportsPath, hostname, version)

	httpClient, err := createHTTPClient(clientID, clientSecret, authUrl)
	if err != nil {
		return UmbrellaService{}, fmt.Errorf("failed to create HTTP client: %w", err)
	}
	return UmbrellaService{
		authUrl:        authUrl,
		deploymentsUrl: deploymentsUrl,
		adminUrl:       adminUrl,
		policiesUrl:    policiesUrl,
		reportsUrl:     reportsUrl,
		client:         httpClient,
		logger:         logger,
	}, nil
}
