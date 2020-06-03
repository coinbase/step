// State Machine Parser
package machine

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/coinbase/step/utils/to"
)

// Takes a file, and a map of Task Function s
func ParseFile(file string) (*StateMachine, error) {
	raw, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	json_sm, err := FromJSON(raw)
	return json_sm, err
}

func FromJSON(raw []byte) (*StateMachine, error) {
	var sm StateMachine
	err := json.Unmarshal(raw, &sm)
	return &sm, err
}

func (sm *States) UnmarshalJSON(b []byte) error {
	// States
	var rawStates map[string]*json.RawMessage
	err := json.Unmarshal(b, &rawStates)

	if err != nil {
		return err
	}

	newStates := States{}
	for name, raw := range rawStates {
		states, err := unmarshallState(name, raw)
		if err != nil {
			return err
		}

		for _, s := range states {
			newStates[*s.Name()] = s
		}
	}

	*sm = newStates
	return nil
}

type stateType struct {
	Type string
}

func unmarshallState(name string, raw_json *json.RawMessage) ([]State, error) {
	var err error

	// extract type (safer than regex)
	var state_type stateType
	if err = json.Unmarshal(*raw_json, &state_type); err != nil {
		return nil, err
	}

	var newState State

	switch state_type.Type {
	case "Pass":
		var s PassState
		err = json.Unmarshal(*raw_json, &s)
		newState = &s
	case "Task":
		var s TaskState
		err = json.Unmarshal(*raw_json, &s)
		newState = &s
	case "Choice":
		var s ChoiceState
		err = json.Unmarshal(*raw_json, &s)
		newState = &s
	case "Wait":
		var s WaitState
		err = json.Unmarshal(*raw_json, &s)
		newState = &s
	case "Succeed":
		var s SucceedState
		err = json.Unmarshal(*raw_json, &s)
		newState = &s
	case "Fail":
		var s FailState
		err = json.Unmarshal(*raw_json, &s)
		newState = &s
	case "Parallel":
		var s ParallelState
		err = json.Unmarshal(*raw_json, &s)
		newState = &s
	case "Map":
		var s MapState
		err = json.Unmarshal(*raw_json, &s)
		newState = &s
	case "TaskFn":
		// This is a custom state that adds values to Task to be handled
		var s TaskState
		err = json.Unmarshal(*raw_json, &s)
		// This will inject the Task name into the input
		s.Parameters = map[string]interface{}{"Task": name, "Input.$": "$"}
		s.Type = to.Strp("Task")
		newState = &s
	default:
		err = fmt.Errorf("Unknown State %q", state_type.Type)
	}

	// End of loop return error
	if err != nil {
		return nil, err
	}

	// Set Name and Defaults
	newName := name
	newState.SetName(&newName) // Require New Variable Pointer

	return []State{newState}, nil
}
