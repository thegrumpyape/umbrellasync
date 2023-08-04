package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/url"
)

func unmarshalResponseBody(response UmbrellaResponse, v interface{}) error {
	err := json.Unmarshal(*response.Data, v)
	if err != nil {
		return err
	}

	return nil
}

func createJSONPayload(data interface{}) (*bytes.Buffer, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("error marshalling payload: %w", err)
	}

	return bytes.NewBuffer(jsonData), nil
}

func ValidateDestinationValues(destinations []string) []string {
	for i, d := range destinations {
		dUrl, err := url.Parse(d)
		if err != nil {
			log.Fatal(err)
		}

		if net.ParseIP(dUrl.Host) != nil {
			removeAtIndex(destinations, i)
			fmt.Println("Removed", dUrl.Host, "from list")
		}
	}
	return destinations
}

func removeAtIndex(s []string, index int) []string {
	return append(s[:index], s[index+1:]...)
}
