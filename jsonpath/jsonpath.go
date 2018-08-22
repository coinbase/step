// Simple Implementation of JSON Path for state machine
package jsonpath

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"
)

/*
The `data` must be from JSON Unmarshal, that way we can guarantee the types:

bool, for JSON booleans
float64, for JSON numbers
string, for JSON strings
[]interface{}, for JSON arrays
map[string]interface{}, for JSON objects
nil for JSON null

*/

var NOT_FOUND_ERROR = errors.New("Not Found")

type Path struct {
	path []string
}

// NewPath takes string returns JSONPath Object
func NewPath(path_string string) (*Path, error) {
	path := Path{}
	path_array, err := ParsePathString(path_string)
	path.path = path_array
	return &path, err
}

// UnmarshalJSON makes a path out of a json string
func (path *Path) UnmarshalJSON(b []byte) error {
	var path_string string
	err := json.Unmarshal(b, &path_string)

	if err != nil {
		return err
	}

	path_array, err := ParsePathString(path_string)

	if err != nil {
		return err
	}

	path.path = path_array
	return nil
}

// MarshalJSON converts path to json string
func (path *Path) MarshalJSON() ([]byte, error) {
	if len(path.path) == 0 {
		return json.Marshal("$")
	}
	return json.Marshal(path.String())
}

func (path *Path) String() string {
	return fmt.Sprintf("$.%v", strings.Join(path.path[:], "."))
}

// ParsePathString parses a path string
func ParsePathString(path_string string) ([]string, error) {
	// must start with $.<value> otherwise empty path
	if path_string == "" || path_string[0:1] != "$" {
		return nil, fmt.Errorf("Bad JSON path: must start with $")
	}

	if path_string == "$" {
		// Default is no path
		return []string{}, nil
	}

	if len(path_string) < 2 {
		// This handles the case for $. or $* which are invalid
		return nil, fmt.Errorf("Bad JSON path: cannot not be 2 characters")
	}

	head := path_string[2:len(path_string)]
	path_array := strings.Split(head, ".")

	// if path contains an "" error
	for _, p := range path_array {
		if p == "" {
			return nil, fmt.Errorf("Bad JSON path: has empty element")
		}
	}
	// Simple Path Builder
	return path_array, nil
}

// PUBLIC METHODS

// GetTime returns Time from Path
func (path *Path) GetTime(input interface{}) (*time.Time, error) {
	output_value, err := path.Get(input)

	if err != nil {
		return nil, fmt.Errorf("GetTime Error %q", err)
	}

	var output time.Time
	switch output_value.(type) {
	case string:
		output, err = time.Parse(time.RFC3339, output_value.(string))
		if err != nil {
			return nil, fmt.Errorf("GetTime Error: time error %q", err)
		}
	default:
		return nil, fmt.Errorf("GetTime Error: time must be string")
	}

	return &output, nil
}

// GetBool returns Bool from Path
func (path *Path) GetBool(input interface{}) (*bool, error) {
	output_value, err := path.Get(input)

	if err != nil {
		return nil, fmt.Errorf("GetBool Error %q", err)
	}

	var output bool
	switch output_value.(type) {
	case bool:
		output = output_value.(bool)
	default:
		return nil, fmt.Errorf("GetBool Error: must return bool")
	}

	return &output, nil
}

// GetNumber returns Number from Path
func (path *Path) GetNumber(input interface{}) (*float64, error) {
	output_value, err := path.Get(input)

	if err != nil {
		return nil, fmt.Errorf("GetFloat Error %q", err)
	}

	var output float64
	switch output_value.(type) {
	case float64:
		output = output_value.(float64)
	case int:
		output = float64(output_value.(int))
	default:
		return nil, fmt.Errorf("GetFloat Error: must return float")
	}

	return &output, nil
}

// GetString returns String from Path
func (path *Path) GetString(input interface{}) (*string, error) {
	output_value, err := path.Get(input)

	if err != nil {
		return nil, fmt.Errorf("GetString Error %q", err)
	}

	var output string
	switch output_value.(type) {
	case string:
		output = output_value.(string)
	default:
		return nil, fmt.Errorf("GetString Error: must return string")
	}

	return &output, nil
}

// GetMap returns Map from Path
func (path *Path) GetMap(input interface{}) (output map[string]interface{}, err error) {
	output_value, err := path.Get(input)

	if err != nil {
		return nil, fmt.Errorf("GetMap Error %q", err)
	}

	switch output_value.(type) {
	case map[string]interface{}:
		output = output_value.(map[string]interface{})
	default:
		return nil, fmt.Errorf("GetMap Error: must return map")
	}

	return output, nil
}

// Get returns interface from Path
func (path *Path) Get(input interface{}) (value interface{}, err error) {
	if path == nil {
		return input, nil // Default is $
	}
	return recursiveGet(input, path.path)
}

// Set sets a Value in a map with Path
func (path *Path) Set(input interface{}, value interface{}) (output map[string]interface{}, err error) {
	var set_path []string
	if path == nil {
		set_path = []string{} // default "$"
	} else {
		set_path = path.path
	}

	if len(set_path) == 0 {
		// The output is the value
		switch value.(type) {
		case map[string]interface{}:
			output = value.(map[string]interface{})
			return output, nil
		default:
			return nil, fmt.Errorf("Cannot Set value %q type %q in root JSON path $", value, reflect.TypeOf(value))
		}
	}
	return recursiveSet(input, value, set_path), nil
}

// PRIVATE METHODS

func recursiveSet(data interface{}, value interface{}, path []string) (output map[string]interface{}) {
	var data_map map[string]interface{}

	switch data.(type) {
	case map[string]interface{}:
		data_map = data.(map[string]interface{})
	default:
		// Overwrite current data with new map
		// this will work for nil as well
		data_map = make(map[string]interface{})
	}

	if len(path) == 1 {
		data_map[path[0]] = value
	} else {
		data_map[path[0]] = recursiveSet(data_map[path[0]], value, path[1:])
	}

	return data_map
}

func recursiveGet(data interface{}, path []string) (interface{}, error) {
	if len(path) == 0 {
		return data, nil
	}

	if data == nil {
		return nil, errors.New("Not Found")
	}

	switch data.(type) {
	case map[string]interface{}:
		value, ok := data.(map[string]interface{})[path[0]]

		if !ok {
			return data, NOT_FOUND_ERROR
		}

		return recursiveGet(value, path[1:])

	default:
		return data, NOT_FOUND_ERROR
	}
}
