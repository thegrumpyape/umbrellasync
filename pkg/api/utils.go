package api

import (
	"bytes"
	"encoding/json"
	"fmt"
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
