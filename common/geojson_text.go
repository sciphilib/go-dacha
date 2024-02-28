package common

import (
	"encoding/json"
	"errors"
	"log"
)

type GeoJSONText struct {
	Data json.RawMessage
}

func (gt *GeoJSONText) Scan(value interface{}) error {
	byteValue, ok := value.([]byte)
	if !ok {
		return errors.New("Failed to assert geoJSON value to bytes")
	}

	gt.Data = json.RawMessage(byteValue)
	log.Printf("gt.Data =", string(gt.Data))
	return nil
}

func (gt GeoJSONText) MarshalJSON() ([]byte, error) {
	return gt.Data, nil
}

func (gt *GeoJSONText) UnmarshalJSON(data []byte) error {
	if data == nil {
		return errors.New("geoJSON value is nil")
	}
	gt.Data = data
	return nil
}
