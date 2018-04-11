package to

import (
	"bytes"
	"encoding/json"
)

// FromJSON Map Converts a string of JSON or a Struct into a map[string]interface{}
func FromJSON(input interface{}) (interface{}, error) {
	str, err := PrettyJSON(input)
	if err != nil {
		return nil, err
	}

	var v interface{}
	if err := json.Unmarshal([]byte(str), &v); err != nil {
		return nil, err
	}

	return v, nil
}

// Takes a string, *string, or struct and returns []byte (json marshal)
func AByte(input interface{}) ([]byte, error) {
	switch input.(type) {
	case nil:
		return []byte(""), nil
	case string:
		return []byte(input.(string)), nil
	case *string:
		str := input.(*string)
		if str == nil {
			return []byte(""), nil
		}
		return []byte(*str), nil
	case []byte:
		return input.([]byte), nil
	case *[]byte:
		by := input.(*[]byte)
		if by == nil {
			return []byte(""), nil
		}
		return *by, nil
	default:
		return json.Marshal(input)
	}
}

// PrettyJSON takes a string or a struct and returns it as PrettyJSON
func PrettyJSON(input interface{}) (string, error) {
	raw, err := AByte(input)
	if err != nil {
		return "", err
	}

	var json_str interface{}
	if err := json.Unmarshal(raw, &json_str); err != nil {
		return string(raw), nil
	}

	by, err := json.MarshalIndent(json_str, "", " ")
	return string(by), err
}

// PrettyJSONStr takes a string or a struct and returns it as PrettyJSON, no error
func PrettyJSONStr(input interface{}) string {
	str, _ := PrettyJSON(input)
	return str
}

func CompactJSON(input interface{}) (string, error) {
	raw, err := AByte(input)
	if err != nil {
		return "", err
	}

	b := bytes.NewBuffer(nil)
	err = json.Compact(b, raw)
	if err != nil {
		return "", err
	}

	return string(b.Bytes()), nil
}

func CompactJSONStr(input interface{}) string {
	str, _ := CompactJSON(input)
	return str
}
