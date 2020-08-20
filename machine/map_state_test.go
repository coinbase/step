package machine

import (
	"github.com/coinbase/step/utils/to"
	"github.com/stretchr/testify/assert"
	"testing"
)

/////////
// Helpers
/////////

func initialize_state_machine(state *StateMachine, t *testing.T) {
	state.StartAt = to.Strp("Start")
	state.States = States{}
	sm := parseTaskState([]byte(`{
		"Resource": "asd",
		"End": true,
		"Retry": [{ "ErrorEquals": ["States.ALL"] }]
	}`), t)
	state.States["start"] = sm

}

// Execution

func Test_MapState_ValidateResource(t *testing.T) {
	state := parseMapState([]byte(`{ "End": true}`), t)
	assert.Error(t, state.Validate())
	state.Iterator = &StateMachine{}
	assert.Error(t, state.Validate())
	initialize_state_machine(state.Iterator, t)
	assert.NoError(t, state.Validate())
}

func Test_MapState_SingleState(t *testing.T) {
	state := parseMapState([]byte(`{
      "Type": "Map",
      "ItemsPath": "$.shipped",
      "ResultPath": "$.output.data",
      "OutputPath": "$.output",
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
	outputResults := map[string]interface{}{}
	var res []map[string]interface{}
	res = append(res, map[string]interface{}{"key": "value"})
	res = append(res, map[string]interface{}{"key": "value"})
	res = append(res, map[string]interface{}{"key": "value"})

	outputResults["data"] = res
	testState(state, stateTestData{
		Input:  map[string]interface{}{"shipped": []interface{}{1, 2, 3}, "output": []interface{}{}},
		Output: outputResults,
	}, t)

	testState(state, stateTestData{
		Input: map[string]interface{}{},
		Error: to.Strp("GetSlice Error \"Not Found\""),
	}, t)
}

func Test_MapState_Catch(t *testing.T) {
	state := parseMapState([]byte(`{
      "Type": "Map",
      "ItemsPath": "$.shipped",
      "ResultPath": "$.output.data",
      "OutputPath": "$.output",

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

	// No Input path data. Should be caught
	testState(state, stateTestData{
		Input:  map[string]interface{}{},
		Output: map[string]interface{}{"Error": "errorString", "Cause": "GetSlice Error \"Not Found\""},
	}, t)

}

func Test_MapState_Integration(t *testing.T) {
	state := parseMapState([]byte(`{
      "Type": "Map",
      "ItemsPath": "$.shipped",
      "ResultPath": "$.output.data",
      "OutputPath": "$.output",
      "MaxConcurrency": 0,
      "Iterator": {
        "StartAt": "Validate",
        "States": {
          "Validate": {
            "Type": "Pass",
            "Next": "Task"
          },
          "Task" : {
			"Type": "TaskFn",
			"Resource": "arn:aws:lambda:{{aws_region}}:{{aws_account}}:function:{{lambda_name}}",
			"Result": {"key": "value"},
			"End": true
         }
        }
      },
      "End": true
    }`), t)

	// Default
	var task = state.Iterator.States["Task"].(*TaskState)
	task.SetTaskHandler(ReturnInputHandler)
	outputResults := map[string]interface{}{}
	var res []map[string]interface{}
	res = append(res, map[string]interface{}{"Task": "Task", "Input": float64(11)})
	res = append(res, map[string]interface{}{"Task": "Task", "Input": float64(12)})
	res = append(res, map[string]interface{}{"Task": "Task", "Input": float64(13)})

	outputResults["data"] = res
	testState(state, stateTestData{
		Input:  map[string]interface{}{"shipped": []interface{}{11, 12, 13}, "output": []interface{}{}},
		Output: outputResults,
	}, t)
}
