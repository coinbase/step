// State Machine Parser
package machine

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/coinbase/step/machine/state"
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

type basicMachine struct {
	Comment *string
	StartAt *string
}

// from http://gregtrowbridge.com/golang-json-serialization-with-interfaces/
func (sm *StateMachine) UnmarshalJSON(b []byte) error {
	// First, deserialize StateMachine fields
	var bsm basicMachine
	err := json.Unmarshal(b, &bsm)
	if err != nil {
		return err
	}
	sm.Comment = bsm.Comment
	sm.StartAt = bsm.StartAt

	// States
	var objMap map[string]*json.RawMessage
	err = json.Unmarshal(b, &objMap)
	if err != nil {
		return err
	}

	var initStates map[string]*json.RawMessage

	if val, ok := objMap["States"]; ok {
		err = json.Unmarshal(*val, &initStates)
		if err != nil {
			return err
		}

		sm.States = map[string]state.State{}
		err = unmarshallStates(sm, initStates)
		if err != nil {
			return err
		}
	}

	return nil
}

type stateType struct {
	Type string
}

func unmarshallStates(sm *StateMachine, initStates map[string]*json.RawMessage) error {
	for name, raw_json := range initStates {
		var err error

		// extract type (safer than regex)
		var state_type stateType
		if err = json.Unmarshal(*raw_json, &state_type); err != nil {
			return err
		}

		var newState state.State

		switch state_type.Type {
		case "Pass":
			var s state.PassState
			err = json.Unmarshal(*raw_json, &s)
			newState = &s
		case "Task":
			var s state.TaskState
			err = json.Unmarshal(*raw_json, &s)
			newState = &s
		case "Choice":
			var s state.ChoiceState
			err = json.Unmarshal(*raw_json, &s)
			newState = &s
		case "Wait":
			var s state.WaitState
			err = json.Unmarshal(*raw_json, &s)
			newState = &s
		case "Succeed":
			var s state.SucceedState
			err = json.Unmarshal(*raw_json, &s)
			newState = &s
		case "Fail":
			var s state.FailState
			err = json.Unmarshal(*raw_json, &s)
			newState = &s
		case "Parallel":
			var s state.ParallelState
			err = json.Unmarshal(*raw_json, &s)
			newState = &s
		default:
			err = fmt.Errorf("Unknown State %q", state_type.Type)
		}

		// End of loop return error
		if err != nil {
			return err
		}

		// Set Name and Defaults
		newName := name
		newState.SetName(&newName) // Require New Variable Pointer

		if err := sm.AddState(newState); err != nil {
			return err
		}
	}

	return nil
}
