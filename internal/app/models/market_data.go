package models

import "fmt"

type MarketData struct {
	Name      string                 `json:"name"`
	Timestamp int64                  `json:"timestamp"`
	Exchange  string                 `json:"exchange"`
	Data      map[string]interface{} `json:"data"`
}

func (m *MarketData) Validate() error {
	if m.Name == "" {
		return fmt.Errorf("name is required")
	}
	if m.Timestamp <= 0 {
		return fmt.Errorf("invalid timestamp")
	}
	return nil
}
