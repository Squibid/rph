package utils

import (
	"encoding/json"
	"fmt"
)

type StringOrNumber string

func (s *StringOrNumber) UnmarshalJSON(b []byte) error {
	var asString string
	if err := json.Unmarshal(b, &asString); err == nil {
		*s = StringOrNumber(asString)
		return nil
	}

	var asNumber float64
	if err := json.Unmarshal(b, &asNumber); err == nil {
		*s = StringOrNumber(fmt.Sprintf("%.0f", asNumber))
		return nil
	}

	return fmt.Errorf("invalid frcYear format")
}
