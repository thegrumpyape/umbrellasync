package umbrella

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/thegrumpyape/umbrellasync/pkg/configurationManager"
	"github.com/thegrumpyape/umbrellasync/pkg/logging"
)

type UmbrellaConnector struct {
	configurationManager configurationManager.ConfigurationManager
	log                  logging.Logger
	client               *UmbrellaClient
}

func New(client *UmbrellaClient, configurationManager configurationManager.ConfigurationManager, logger logging.Logger) (*UmbrellaConnector, error) {
	return &UmbrellaConnector{
		configurationManager: configurationManager,
		log:                  logger,
		client:               client,
	}, nil
}

// Destination List Methods

// Gets all destination lists using pagination
func (u *UmbrellaConnector) GetDestinationLists(limit int) ([]DestinationList, error) {
	page := 1
	allDestinationLists := []DestinationList{}

	// Pagination to get all Destination Lists
	for {
		params := map[string]string{
			"page":  strconv.Itoa(page),
			"limit": strconv.Itoa(limit),
		}
		res, err := u.client.Get("policies", "/destinationlists", nil, params)
		if err != nil {
			return nil, err
		}

		var destinationLists []DestinationList
		err = unmarshalUmbrellaResponse(res, &destinationLists)
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
func (u *UmbrellaConnector) GetDestinationList(id int) (DestinationList, error) {
	endpoint := fmt.Sprintf("/destinationlists/%d", id)
	body, err := u.client.Get("policies", endpoint, nil, nil)
	if err != nil {
		return DestinationList{}, err
	}

	var destinationList DestinationList
	err = unmarshalUmbrellaResponse(body, &destinationList)
	if err != nil {
		return DestinationList{}, err
	}

	return destinationList, nil
}

// Creates a new destination list
func (u *UmbrellaConnector) CreateDestinationList(access string, isGlobal bool, name string) (DestinationList, error) {
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

	res, err := u.client.Post("policies", "/destinationlists", headers, nil, bytes.NewBuffer(jsonData))
	if err != nil {
		return DestinationList{}, err
	}

	var destinationList DestinationList
	err = unmarshalUmbrellaResponse(res, &destinationList)
	if err != nil {
		return DestinationList{}, err
	}

	return destinationList, nil
}

// Updates a destination lists name
func (u *UmbrellaConnector) UpdateDestinationList(id int, name string) (DestinationList, error) {
	endpoint := fmt.Sprintf("/destinationlists/%d", id)

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

	res, err := u.client.Patch("policies", endpoint, headers, nil, bytes.NewBuffer(jsonData))
	if err != nil {
		return DestinationList{}, err
	}

	var destinationList DestinationList
	err = unmarshalUmbrellaResponse(res, &destinationList)
	if err != nil {
		return DestinationList{}, err
	}

	return destinationList, nil
}

func (u *UmbrellaConnector) DeleteDestinationList(id int) error {
	endpoint := fmt.Sprintf("/destinationlists/%d", id)

	_, err := u.client.Delete("policies", endpoint, nil, nil, nil)
	if err != nil {
		return err
	}

	return nil
}

// Destinations Methods

// Gets all destinations from a destination list
func (u *UmbrellaConnector) GetDestinations(id int, limit int) ([]Destination, error) {
	page := 1
	allDestinations := []Destination{}

	// Pagination to get all Destinations
	for {
		endpoint := fmt.Sprintf("/destinationlists/%d/destinations", id)
		params := map[string]string{
			"page":  strconv.Itoa(page),
			"limit": strconv.Itoa(limit),
		}

		u.log.Debug("Getting destinations ", limit*(page-1), "-", limit*page)
		res, err := u.client.Get("policies", endpoint, nil, params)
		if err != nil {
			return nil, err
		}

		var destinations []Destination
		err = unmarshalUmbrellaResponse(res, &destinations)
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
func (u *UmbrellaConnector) AddDestinations(destinationList DestinationList, destinationsToAdd []string, chunkSize int) (DestinationList, error) {
	destinationsToAdd, err := u.ValidateDestinationValues(destinationsToAdd)
	if err != nil {
		return DestinationList{}, err
	}

	u.log.Info("Adding ", len(destinationsToAdd), " destinations")

	for i := 0; i < len(destinationsToAdd); i += chunkSize {
		end := i + chunkSize
		if end > len(destinationsToAdd) {
			end = len(destinationsToAdd)
		}

		var addPayload []NewDestination
		for _, destination := range destinationsToAdd[i:end] {
			addPayload = append(addPayload, NewDestination{Destination: destination})
		}

		endpoint := fmt.Sprintf("/destinationlists/%d/destinations", destinationList.ID)
		jsonData, err := json.Marshal(addPayload)
		if err != nil {
			return destinationList, err
		}

		headers := map[string]string{
			"Content-Type": "application/json",
		}

		u.log.Debug("Adding destinations ", i, "-", i+chunkSize)
		res, err := u.client.Post("policies", endpoint, headers, nil, bytes.NewBuffer(jsonData))
		if err != nil {
			u.log.Warn("Error adding destinations ", i, "-", i+chunkSize)
			var umbrellaError UmbrellaResponseError
			err = unmarshalUmbrellaResponse(res, &umbrellaError)
			if err != nil {
				u.log.Error(err)
				continue
			}
			umbrellaMessage := fmt.Sprint(umbrellaError.Message)
			if strings.Contains(umbrellaMessage, "high_volume_list_domain") {
				highVolumeDomain := strings.Split(umbrellaMessage, "\\/")[2]
				u.log.Warn("Umbrella rejected ", highVolumeDomain, " as a high volume domain")
				u.log.Warn("Adding ", highVolumeDomain, " to ignore list")
				u.configurationManager.Append("highvolumedomains", highVolumeDomain)
			} else {
				u.log.Error(umbrellaMessage)
			}
			continue
		}

		err = unmarshalUmbrellaResponse(res, &destinationList)
		if err != nil {
			u.log.Error(err)
			continue
		}
	}

	return destinationList, nil
}

// Removes destinations from a destination list
func (u *UmbrellaConnector) DeleteDestinations(destinationList DestinationList, destinationsToRemove []string, existingDestinations []Destination, chunkSize int) (DestinationList, error) {
	destinationMap := mapDestinationIDs(existingDestinations) // Assuming this maps destinations to IDs

	u.log.Info("Removing ", len(destinationsToRemove), " destinations")

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

		endpoint := fmt.Sprintf("/destinationlists/%d/destinations/remove", destinationList.ID)
		jsonData, err := json.Marshal(removePayload)
		if err != nil {
			return destinationList, err
		}

		headers := map[string]string{
			"Content-Type": "application/json",
		}

		u.log.Debug("Removing destinations ", i, "-", i+chunkSize)
		res, err := u.client.Delete("policies", endpoint, headers, nil, bytes.NewBuffer(jsonData))
		if err != nil {
			u.log.Warn("Error removing destinations ", i, "-", i+chunkSize)
			u.log.Error(err)
			continue
		}

		err = unmarshalUmbrellaResponse(res, &destinationList)
		if err != nil {
			u.log.Error(err)
			continue
		}
	}

	return destinationList, nil
}

func unmarshalUmbrellaResponse(response UmbrellaResponse, v interface{}) error {
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

func (u *UmbrellaConnector) ValidateDestinationValues(destinations []string) ([]string, error) {
	var validURLs []string
	var highVolumeDomains []interface{}
	var ignoreCount int

	hvd := u.configurationManager.Get("highvolumedomains")
	if hvd != nil {
		var ok bool
		highVolumeDomains, ok = hvd.([]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid type for highvolumedomains")
		}
	}

	isHighVolumeDomain := func(host string, domains []interface{}) bool {
		for _, domain := range domains {
			pattern := `(^|\.)` + regexp.QuoteMeta(domain.(string)) + `$`
			matched, err := regexp.MatchString(pattern, host)
			if err != nil {
				return false
			}
			if matched {
				return true
			}
		}
		return false
	}

	u.log.Debug("Validating domains")

	for _, d := range destinations {
		dUrl, err := url.Parse(d)
		if err != nil {
			ignoreCount++
			u.log.Debug("Ignoring ", d)
			continue
		}

		host, _, err := net.SplitHostPort(dUrl.Host)
		if err != nil {
			host = dUrl.Host
		}
		if net.ParseIP(host) != nil {
			ignoreCount++
			u.log.Debug("Ignoring ", d)
			continue
		}

		if isHighVolumeDomain(host, highVolumeDomains) {
			ignoreCount++
			u.log.Debug("Ignoring ", d)
			continue
		}

		url := dUrl.Scheme + "://" + host + dUrl.Path
		if dUrl.RawQuery != "" {
			url = url + "?" + dUrl.RawQuery
		}
		validURLs = append(validURLs, url)
	}

	u.log.Info("Ignoring ", ignoreCount, " URLs")
	return validURLs, nil
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
