package umbrella

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"golang.org/x/oauth2/clientcredentials"
)

func Create(hostname string, version string, clientID string, clientSecret string, logger *log.Logger) (UmbrellaService, error) {
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

// createHTTPClient creates and returns an HTTP client for the UmbrellaService.
func createHTTPClient(clientID, clientSecret, authURL string) (*http.Client, error) {
	clientConfig := clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     fmt.Sprintf(endpointCreateAuthorizationToken, authURL),
	}
	return clientConfig.Client(context.TODO()), nil
}

// Destination List Methods

// Gets all destination lists using pagination
func (u *UmbrellaService) GetDestinationLists(limit int) ([]DestinationList, error) {
	page := 1
	allDestinationLists := []DestinationList{}

	// Pagination to get all Destination Lists
	for {
		url := fmt.Sprintf(endpointGetDestinationLists, u.policiesUrl)
		params := map[string]string{
			"page":  strconv.Itoa(page),
			"limit": strconv.Itoa(limit),
		}
		res, err := u.get(url, nil, params)
		if err != nil {
			return nil, err
		}

		var destinationLists []DestinationList
		err = unmarshalResponseBody(res, &destinationLists)
		if err != nil {
			return nil, err
		}

		if len(destinationLists) == 0 {
			break
		}

		allDestinationLists = append(allDestinationLists, destinationLists...)

		meta := res.Meta

		if meta.Limit > meta.Total {
			break
		}
		page++
	}

	return allDestinationLists, nil
}

// Gets a single destination list
func (u *UmbrellaService) GetDestinationList(id int) (DestinationList, error) {
	url := fmt.Sprintf(endpointGetDestinationList, u.policiesUrl, id)
	body, err := u.get(url, nil, nil)
	if err != nil {
		return DestinationList{}, err
	}

	var destinationList DestinationList
	err = unmarshalResponseBody(body, &destinationList)
	if err != nil {
		return DestinationList{}, err
	}

	return destinationList, nil
}

// Creates a new destination list
func (u *UmbrellaService) CreateDestinationList(access string, isGlobal bool, name string) (DestinationList, error) {
	url := fmt.Sprintf(endpointCreateDestinationList, u.policiesUrl)

	payload := map[string]interface{}{
		"access":   access,
		"isGlobal": isGlobal,
		"name":     name,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return DestinationList{}, err
	}

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	res, err := u.post(url, headers, nil, bytes.NewBuffer(jsonData))
	if err != nil {
		return DestinationList{}, err
	}

	var destinationList DestinationList
	err = unmarshalResponseBody(res, &destinationList)
	if err != nil {
		return DestinationList{}, err
	}

	return destinationList, nil
}

// Updates a destination lists name
func (u *UmbrellaService) UpdateDestinationList(id int, name string) (DestinationList, error) {
	url := fmt.Sprintf(endpointUpdateDestinationList, u.policiesUrl, id)

	payload := map[string]interface{}{
		"name": name,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return DestinationList{}, err
	}

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	res, err := u.patch(url, headers, nil, bytes.NewBuffer(jsonData))
	if err != nil {
		return DestinationList{}, err
	}

	var destinationList DestinationList
	err = unmarshalResponseBody(res, &destinationList)
	if err != nil {
		return DestinationList{}, err
	}

	return destinationList, nil
}

func (u *UmbrellaService) DeleteDestinationList(id int) error {
	url := fmt.Sprintf(endpointDeleteDestinationList, u.policiesUrl, id)

	_, err := u.delete(url, nil, nil, nil)
	if err != nil {
		return err
	}

	return nil
}

// Destinations Methods

// Gets all destinations from a destination list
func (u *UmbrellaService) GetDestinations(id int, limit int) ([]Destination, error) {
	page := 1
	allDestinations := []Destination{}

	// Pagination to get all Destinations
	for {
		url := fmt.Sprintf(endpointGetDestinationsInDestinationList, u.policiesUrl, id)
		params := map[string]string{
			"page":  strconv.Itoa(page),
			"limit": strconv.Itoa(limit),
		}

		res, err := u.get(url, nil, params)
		if err != nil {
			return nil, err
		}

		var destinations []Destination
		err = unmarshalResponseBody(res, &destinations)
		if err != nil {
			return nil, err
		}

		if len(destinations) == 0 {
			break
		}

		allDestinations = append(allDestinations, destinations...)

		if res.Meta.Limit > res.Meta.Total {
			break
		}

		page++
	}

	return allDestinations, nil
}

// Add destinations to a destination list
func (u *UmbrellaService) AddDestinations(destinationList DestinationList, destinationsToAdd []string, chunkSize int) (DestinationList, error) {
	destinationsToAdd = ValidateDestinationValues(destinationsToAdd)

	for i := 0; i < len(destinationsToAdd); i += chunkSize {
		end := i + chunkSize
		if end > len(destinationsToAdd) {
			end = len(destinationsToAdd)
		}

		var addPayload []NewDestination
		for _, destination := range destinationsToAdd[i:end] {
			addPayload = append(addPayload, NewDestination{Destination: destination})
		}

		url := fmt.Sprintf(endpointAddDestinationsToDestinationList, u.policiesUrl, destinationList.ID)
		jsonData, err := json.Marshal(addPayload)
		if err != nil {
			return destinationList, err
		}

		headers := map[string]string{
			"Content-Type": "application/json",
		}

		res, err := u.post(url, headers, nil, bytes.NewBuffer(jsonData))
		if err != nil {
			return destinationList, err
		}

		err = unmarshalResponseBody(res, &destinationList)
		if err != nil {
			return destinationList, err
		}
	}

	return destinationList, nil
}

// Removes destinations from a destination list
func (u *UmbrellaService) DeleteDestinations(destinationList DestinationList, destinationsToRemove []string, existingDestinations []Destination, chunkSize int) (DestinationList, error) {
	destinationMap := mapDestinationIDs(existingDestinations) // Assuming this maps destinations to IDs

	for i := 0; i < len(destinationsToRemove); i += chunkSize {
		end := i + chunkSize
		if end > len(destinationsToRemove) {
			end = len(destinationsToRemove)
		}

		var removePayload []int
		for _, destination := range destinationsToRemove[i:end] {
			if id, ok := destinationMap[destination]; ok {
				removePayload = append(removePayload, id)
			}
		}

		url := fmt.Sprintf(endpointDeleteDestinationsFromDestinationList, u.policiesUrl, destinationList.ID)
		jsonData, err := json.Marshal(removePayload)
		if err != nil {
			return destinationList, err
		}

		headers := map[string]string{
			"Content-Type": "application/json",
		}

		res, err := u.delete(url, headers, nil, bytes.NewBuffer(jsonData))
		if err != nil {
			return destinationList, err
		}

		err = unmarshalResponseBody(res, &destinationList)
		if err != nil {
			return destinationList, err
		}
	}

	return destinationList, nil
}

// HTTP Methods

func (u *UmbrellaService) get(url string, headers map[string]string, params map[string]string) (UmbrellaResponse, error) {
	res, err := u.request("GET", url, headers, params, nil)
	return res, err
}

func (u *UmbrellaService) post(url string, headers map[string]string, params map[string]string, data io.Reader) (UmbrellaResponse, error) {
	res, err := u.request("POST", url, headers, params, data)
	return res, err
}

func (u *UmbrellaService) patch(url string, headers map[string]string, params map[string]string, data io.Reader) (UmbrellaResponse, error) {
	res, err := u.request("PATCH", url, headers, params, data)
	return res, err
}

func (u *UmbrellaService) delete(url string, headers map[string]string, params map[string]string, data io.Reader) (UmbrellaResponse, error) {
	res, err := u.request("DELETE", url, headers, params, data)
	return res, err
}

func (u *UmbrellaService) request(method string, url string, headers map[string]string, params map[string]string, data io.Reader) (UmbrellaResponse, error) {
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
	u.logger.Printf(req.Method + " " + req.Proto + " " + req.URL.Scheme + "://" + req.URL.Host + "?" + req.URL.RawQuery)
	resp, err := u.client.Do(req)
	if err != nil {
		return UmbrellaResponse{}, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	// Checks if HTTP Error occurred
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusBadRequest {
		return UmbrellaResponse{}, fmt.Errorf("non-OK HTTP status: %s", resp.Status)
	}

	// Reads response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return UmbrellaResponse{}, fmt.Errorf("error reading response body: %w", err)
	}

	// Unmarshals json response into UmbrellaResponse model
	err = json.Unmarshal(body, &umbrellaResponse)
	if err != nil {
		return UmbrellaResponse{}, fmt.Errorf("error unmarshalling response: %w", err)
	}

	if umbrellaResponse.Status.Code != 200 {
		return umbrellaResponse, fmt.Errorf("non-OK HTTP Status: %d %s %s", umbrellaResponse.Status.Code, umbrellaResponse.Status.Text, umbrellaResponse.Data)
	}

	return umbrellaResponse, nil
}

func unmarshalResponseBody(response UmbrellaResponse, v interface{}) error {
	err := json.Unmarshal(*response.Data, v)
	if err != nil {
		return err
	}

	return nil
}

func mapDestinationIDs(destinations []Destination) map[string]int {
	destinationMap := make(map[string]int)
	for _, destination := range destinations {
		id, _ := strconv.Atoi(destination.ID)
		destinationMap[destination.Destination] = id
	}
	return destinationMap
}

func ValidateDestinationValues(destinations []string) []string {
	for i, d := range destinations {
		dUrl, err := url.Parse(d)
		if err != nil {
			log.Fatal(err)
		}

		if net.ParseIP(dUrl.Host) != nil {
			RemoveAtIndex(destinations, i)
			fmt.Println("Removed", dUrl.Host, "from list")
		}
	}
	return destinations
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
