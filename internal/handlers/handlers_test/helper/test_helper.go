package helper

import (
	"encoding/json"
	"io"
)

type SuccessResponse struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data,omitempty"`
}

type ErrorResponse struct {
	Success bool                   `json:"success"`
	Error   map[string]interface{} `json:"error"`
}

func ExtractData(responseBody io.Reader, target interface{}) error {
	var successResp SuccessResponse
	if err := json.NewDecoder(responseBody).Decode(&successResp); err != nil {
		return err
	}
	return json.Unmarshal(successResp.Data, target)
}
