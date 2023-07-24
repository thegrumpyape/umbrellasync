package umbrella

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
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
}

func NewUmbrellaService(hostname string, version string, clientID string, clientSecret string) UmbrellaService {
	baseUrl := "https://" + hostname
	authUrl := baseUrl + "/auth/" + version
	deploymentsUrl := baseUrl + "/deployments/" + version
	adminUrl := baseUrl + "/admin/" + version
	policiesUrl := baseUrl + "/policies/" + version
	reportsUrl := baseUrl + "/reports/" + version

	clientConfig := clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     authUrl + "/token",
	}
	httpClient := clientConfig.Client(context.TODO())
	return UmbrellaService{authUrl: authUrl, deploymentsUrl: deploymentsUrl, adminUrl: adminUrl, policiesUrl: policiesUrl, reportsUrl: reportsUrl, client: httpClient}
}

func (u *UmbrellaService) GetDestinationLists(limit int) ([]DestinationList, error) {
	page := 1
	allDestinationLists := []DestinationList{}

	for {
		var data ResponseStatusMetaDestinationLists
		url := fmt.Sprintf("%s/destinationlists?page=%d&limit=%d", u.policiesUrl, page, limit)
		body, err := u.get(url)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(body, &data)
		if err != nil {
			return nil, err
		}

		if len(data.Data) == 0 {
			break
		}
		allDestinationLists = append(allDestinationLists, data.Data...)

		if data.Meta.Limit > data.Meta.Total {
			break
		}
		page++
	}

	return allDestinationLists, nil
}

func (u *UmbrellaService) GetDestinationList(id int) (DestinationList, error) {
	var data ResponseStatusMetaDestiantionList
	url := fmt.Sprintf("%s/destinationlists/%d", u.policiesUrl, id)
	body, err := u.get(url)
	if err != nil {
		return DestinationList{}, err
	}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return DestinationList{}, err
	}
	return data.Data, nil
}

func (u *UmbrellaService) CreateDestinationList(access string, isGlobal bool, name string) (DestinationList, error) {
	var data DestinationList
	url := fmt.Sprintf("%s/destinationlists", u.policiesUrl)

	payload := map[string]interface{}{
		"access":   access,
		"isGlobal": isGlobal,
		"name":     name,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return DestinationList{}, err
	}

	body, err := u.post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return DestinationList{}, err
	}

	err = json.Unmarshal(body, &data)
	if err != nil {
		return DestinationList{}, err
	}

	return data, nil
}

func (u *UmbrellaService) UpdateDestinationList(id int, name string) (DestinationList, error) {
	var data ResponseStatusDestinationList
	url := fmt.Sprintf("%s/destinationlists/%d", u.policiesUrl, id)

	payload := map[string]interface{}{
		"name": name,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return DestinationList{}, err
	}

	body, err := u.patch(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return DestinationList{}, err
	}

	err = json.Unmarshal(body, &data)
	if err != nil {
		return DestinationList{}, err
	}
	return data.Data, nil
}

func (u *UmbrellaService) DeleteDestinationList(id int) error {
	url := fmt.Sprintf("%s/destinationlists/%d", u.policiesUrl, id)

	_, err := u.delete(url, "", nil)
	if err != nil {
		return err
	}

	return nil
}

func (u *UmbrellaService) GetDestinations(id int, limit int) ([]Destination, error) {
	page := 1
	allDestinations := []Destination{}

	for {
		var data ResponseStatusMetaDestinations

		url := fmt.Sprintf("%s/destinationlists/%d/destinations?page=%d&limit=%d", u.policiesUrl, id, page, limit)
		body, err := u.get(url)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(body, &data)
		if err != nil {
			return nil, err
		}

		if len(data.Data) == 0 {
			break
		}
		allDestinations = append(allDestinations, data.Data...)

		if data.Meta.Limit > data.Meta.Total {
			break
		}
		page++
	}

	return allDestinations, nil
}

func (u *UmbrellaService) AddDestinations(id int, destinations []NewDestination) (DestinationList, error) {
	var data ResponseStatusDestinationList
	url := fmt.Sprintf("%s/destinationlists/%d/destinations", u.policiesUrl, id)

	jsonData, err := json.Marshal(destinations)
	if err != nil {
		return DestinationList{}, err
	}

	body, err := u.post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return DestinationList{}, err
	}

	err = json.Unmarshal(body, &data)
	if err != nil {
		return DestinationList{}, err
	}

	return data.Data, nil
}

func (u *UmbrellaService) DeleteDestinations(id int, destinationIDs []int) (DestinationList, error) {
	var data ResponseStatusDestinationList
	url := fmt.Sprintf("%s/destinationlists/%d/destinations/remove", u.policiesUrl, id)

	jsonData, err := json.Marshal(destinationIDs)
	if err != nil {
		return DestinationList{}, err
	}

	body, err := u.delete(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return DestinationList{}, err
	}

	err = json.Unmarshal(body, &data)
	if err != nil {
		return DestinationList{}, err
	}

	return data.Data, nil
}

func (u *UmbrellaService) get(url string) ([]byte, error) {
	res, err := u.client.Get(url)
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func (u *UmbrellaService) post(url string, contentType string, data io.Reader) ([]byte, error) {
	res, err := u.client.Post(url, contentType, data)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return body, err
}

func (u *UmbrellaService) patch(url string, contentType string, data io.Reader) ([]byte, error) {
	req, err := http.NewRequest(http.MethodPatch, url, data)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", contentType)

	res, err := u.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (u *UmbrellaService) delete(url string, contentType string, data io.Reader) ([]byte, error) {
	req, err := http.NewRequest(http.MethodDelete, url, data)
	if err != nil {
		return nil, err
	}

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	res, err := u.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
