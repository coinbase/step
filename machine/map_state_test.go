package machine

import (
	"github.com/coinbase/step/utils/to"
	"github.com/stretchr/testify/assert"
	"testing"
)

/////////
// Helpers
/////////

func initialize_state_machine( state *StateMachine , t *testing.T) {
	state.StartAt = to.Strp("Start")
	state.States = States{}
	sm := parseTaskState([]byte(`{
		"Resource": "asd",
		"Next": "Pass",
		"Retry": [{ "ErrorEquals": ["States.ALL"] }]
	}`), t)
	state.States["start"] = sm

}

// Execution

func Test_MapState_ValidateResource(t *testing.T) {
	state := parseMapState([]byte(`{ "Next": "Pass"}`), t)
	assert.Error(t, state.Validate())
	state.Iterator = &StateMachine{}
	assert.Error(t, state.Validate())
	initialize_state_machine(state.Iterator, t)
	assert.NoError(t, state.Validate())
}

func Test_MapState_SingleState(t *testing.T){
	state := parseMapState([]byte(`{
      "Type": "Map",
      "ItemsPath": "$.shipped",
      "ResultPath": "$.output",
      "MaxConcurrency": 0,
      "Iterator": {
        "StartAt": "Validate",
        "States": {
          "Validate": {
            "Type": "Pass",
            "Result": {"key": "value"},
            "End": true
          }
        }
      },
      "End": true
    }`), t)
	// Default
	 testState(state, stateTestData{
		Input: map[string]interface{}{"shipped": []interface{}{1,2,3}},
	}, t)

	testState(state, stateTestData{
		Input: map[string]interface{}{},
		Error: to.Strp("GetSlice Error \"Not Found\""),
	}, t)
}

func Test_MapState_Catch(t *testing.T){
	state := parseMapState([]byte(`{
      "Type": "Map",
      "ItemsPath": "$.shipped",
      "ResultPath": "$.output",
      "MaxConcurrency": 0,
      "Catch": [{
			"ErrorEquals": ["States.ALL"],
			"Next": "Fail"
       }],
      "Iterator": {
        "StartAt": "Validate",
        "States": {
          "Validate": {
            "Type": "Pass",
            "Result": {"key": "value"},
            "End": true
          }
        }
      },
      "End": true
    }`), t)
	// Default
	testState(state, stateTestData{
		Input: map[string]interface{}{"shipped": []interface{}{1,2,3}},
	}, t)

	// No Input path data. Should be caught
	testState(state, stateTestData{
		Input: map[string]interface{}{},
	}, t)

}