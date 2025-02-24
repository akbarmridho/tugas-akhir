package utility

import (
	"encoding/json"
	"github.com/perimeterx/marshmallow"
)

func MapToStruct(m map[string]interface{}, v interface{}) error {
	// Convert the map to JSON
	jsonBytes, err := json.Marshal(m)
	if err != nil {
		return err
	}

	// Unmarshal JSON into the struct
	return json.Unmarshal(jsonBytes, v)
}

func InterfaceToStruct(m interface{}, v interface{}) error {
	// Convert the map to JSON
	jsonBytes, err := json.Marshal(m)
	if err != nil {
		return err
	}

	// Unmarshal JSON into the struct
	// use marshmallow to allow extra fields
	// if needed, you can use result variable from unused _ variable below to access the additional fields
	_, err = marshmallow.Unmarshal(jsonBytes, v)
	return err
}
