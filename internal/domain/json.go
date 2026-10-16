package domain

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// JSONObject 存 PostgreSQL jsonb，空值按 {} 处理。
type JSONObject []byte

func (j JSONObject) Value() (driver.Value, error) {
	if len(j) == 0 {
		return "{}", nil
	}
	if !json.Valid(j) {
		return nil, fmt.Errorf("invalid json")
	}
	return string(j), nil
}

func (j *JSONObject) Scan(value any) error {
	if value == nil {
		*j = JSONObject("{}")
		return nil
	}
	switch v := value.(type) {
	case []byte:
		if len(v) == 0 {
			*j = JSONObject("{}")
			return nil
		}
		*j = append(JSONObject(nil), v...)
	case string:
		if v == "" {
			*j = JSONObject("{}")
			return nil
		}
		*j = JSONObject(v)
	default:
		return fmt.Errorf("cannot scan %T into JSONObject", value)
	}
	return nil
}

func (j JSONObject) MarshalJSON() ([]byte, error) {
	if len(j) == 0 {
		return []byte("{}"), nil
	}
	return j, nil
}

func (j *JSONObject) UnmarshalJSON(b []byte) error {
	if len(b) == 0 || string(b) == "null" {
		*j = JSONObject("{}")
		return nil
	}
	*j = append(JSONObject(nil), b...)
	return nil
}
