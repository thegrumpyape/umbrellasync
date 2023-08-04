package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

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

// Adds destinations to a destiantion list
func (u *UmbrellaService) AddDestinations(id int, destinations []NewDestination) (DestinationList, error) {
	url := fmt.Sprintf(endpointAddDestinationsToDestinationList, u.policiesUrl, id)

	jsonData, err := json.Marshal(destinations)
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

// Removes destiantions from a destination list
func (u *UmbrellaService) DeleteDestinations(id int, destinationIDs []int) (DestinationList, error) {
	url := fmt.Sprintf(endpointDeleteDestinationsFromDestinationList, u.policiesUrl, id)

	jsonData, err := json.Marshal(destinationIDs)
	if err != nil {
		return DestinationList{}, err
	}

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	res, err := u.delete(url, headers, nil, bytes.NewBuffer(jsonData))
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

	return umbrellaResponse, nil
}
