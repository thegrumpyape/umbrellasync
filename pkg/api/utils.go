package api

import (
	"encoding/json"
	"strconv"
)

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
