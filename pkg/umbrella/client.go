package umbrella

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/thegrumpyape/umbrellasync/pkg/configurationManager"
	"github.com/thegrumpyape/umbrellasync/pkg/logging"
	"golang.org/x/oauth2/clientcredentials"
)

type UmbrellaClient struct {
	client   *http.Client
	hostname string
	version  string
	log      logging.Logger
}

func CreateUmbrellaClient(configurationManager configurationManager.ConfigurationManager, logger logging.Logger) *UmbrellaClient {
	hostname := configurationManager.Get("apihostname").(string)
	version := configurationManager.Get("apiversion").(string)

	tokenUrl := fmt.Sprintf("https://%s/auth/%s/token", hostname, version)
	key := configurationManager.Get("key").(string)
	secret := configurationManager.Get("secret").(string)
	clientConfig := clientcredentials.Config{
		ClientID:     key,
		ClientSecret: secret,
		TokenURL:     tokenUrl,
	}
	httpClient := clientConfig.Client(context.TODO())

	return &UmbrellaClient{client: httpClient, hostname: hostname, version: version, log: logger}
}

func (u *UmbrellaClient) Get(scope string, endpoint string, headers map[string]string, params map[string]string) (UmbrellaResponse, error) {
	url := u.generateUrl(scope, endpoint)
	res, err := u.Request("GET", url, headers, params, nil)
	return res, err
}

func (u *UmbrellaClient) Post(scope string, endpoint string, headers map[string]string, params map[string]string, data io.Reader) (UmbrellaResponse, error) {
	url := u.generateUrl(scope, endpoint)
	res, err := u.Request("POST", url, headers, params, data)
	return res, err
}

func (u *UmbrellaClient) Patch(scope string, endpoint string, headers map[string]string, params map[string]string, data io.Reader) (UmbrellaResponse, error) {
	url := u.generateUrl(scope, endpoint)
	res, err := u.Request("PATCH", url, headers, params, data)
	return res, err
}

func (u *UmbrellaClient) Delete(scope string, endpoint string, headers map[string]string, params map[string]string, data io.Reader) (UmbrellaResponse, error) {
	url := u.generateUrl(scope, endpoint)
	res, err := u.Request("DELETE", url, headers, params, data)
	return res, err
}

func (u *UmbrellaClient) Request(method string, url string, headers map[string]string, params map[string]string, data io.Reader) (UmbrellaResponse, error) {
	var umbrellaResponse UmbrellaResponse

	// Create request
	req, err := http.NewRequest(method, url, data)
	if err != nil {
		return UmbrellaResponse{}, fmt.Errorf("error creating new request: %w", err)
	}

	// Adds provided headers to request
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	// Adds provided parameters to request
	query := req.URL.Query()
	for k, v := range params {
		query.Add(k, v)
	}
	req.URL.RawQuery = query.Encode()

	// Adding a timeout for the request
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	req = req.WithContext(ctx)

	// Sending the request
	// fmt.Printf(req.Method + " " + req.Proto + " " + req.URL.Scheme + "://" + req.URL.Host + req.URL.Path + "?" + req.URL.RawQuery + "\n")
	resp, err := u.client.Do(req)
	if err != nil {
		return UmbrellaResponse{}, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	// Reads response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return UmbrellaResponse{}, fmt.Errorf("error reading response body: %w", err)
	}

	// Checks if HTTP Error occurred
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusBadRequest {
		return UmbrellaResponse{}, fmt.Errorf("non-OK HTTP status: %s: %s", resp.Status, string(body))
	}

	// Unmarshals json response into UmbrellaResponse model
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if _, ok := result["status"]; ok {
		json.Unmarshal(body, &umbrellaResponse)
	} else if _, ok := result["id"]; ok {
		status := Status{Code: resp.StatusCode, Text: resp.Status}
		data := (*json.RawMessage)(&body)
		umbrellaResponse = UmbrellaResponse{Status: status, Data: data}
	} else {
		return UmbrellaResponse{}, fmt.Errorf("Unable to unmarshal HTTP response: %v", body)
	}

	if err != nil {
		return UmbrellaResponse{}, fmt.Errorf("error unmarshalling response: %w", err)
	}

	if umbrellaResponse.Status.Code != 200 {
		return umbrellaResponse, fmt.Errorf("non-OK UmbrellaResponse Status: %d %s %s", umbrellaResponse.Status.Code, umbrellaResponse.Status.Text, umbrellaResponse.Data)
	}

	return umbrellaResponse, nil
}

func (u *UmbrellaClient) generateUrl(scope string, endpoint string) string {
	scope = strings.TrimPrefix(strings.TrimSuffix(scope, "/"), "/")
	endpoint = strings.TrimPrefix(strings.TrimSuffix(endpoint, "/"), "/")
	url := fmt.Sprintf("https://%s/%s/%s/%s", u.hostname, scope, u.version, endpoint)
	return url
}
