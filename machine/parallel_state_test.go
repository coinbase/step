package machine

import (
	"encoding/json"
	"github.com/coinbase/step/utils/to"
	"github.com/stretchr/testify/assert"
	"testing"
)

// Tests on Non Valid Json

func Test_ParallelState_NonValid_State_No_Next_Or_End(t *testing.T) {
	state := parseParallelTaskState([]byte(`{
          "Type": "Parallel",
          "Branches": [
            {
              "StartAt": "Branch_A_Start",
              "States": {
                "Branch_A_Start": {
                "Type": "TaskFn",
                "Resource": "host1:machine1:func1",
                "Next": "Branch_A_Done"
                },
                "Branch_A_Done": {
                    "Type": "Succeed"
                }
              }
            },
            {
              "StartAt": "Branch_B_Start",
              "States": {
                "Branch_B_Start": {
                "Type": "TaskFn",
                "Resource": "host1:machine2:func2",
                "End": true
                }
              }
            }
          ]
        }`), t)

	err := state.Validate()
	assert.Error(t, err)
	assert.EqualErrorf(t, err, "parallel state must have either \"Next\" or \"End\" property", "Checking that Validate fails with no \"End\" or \"Next\" property")
}

func Test_ParallelState_NonValid_Branches_Empty(t *testing.T) {
	state := parseParallelTaskState([]byte(`{
          "Type": "Parallel",
          "End": true,
          "Branches": [
          ]
        }`), t)

	err := state.Validate()
	assert.Error(t, err)
	assert.EqualErrorf(t, err, "branches can't be nil or empty (len = 0)", "Checking that Validate fails with empty array")
}

func Test_ParallelState_NonValid_Branches_Unreachable_Next(t *testing.T) {
	state := parseParallelTaskState([]byte(`{
          "Type": "Parallel",
          "End": true,
          "Branches": [
            {
              "StartAt": "Branch_A_Start",
              "States": {
                "Branch_A_Start": {
                "Type": "TaskFn",
                "Resource": "host1:machine1:func1",
                "Next": "Branch_A_Done"
                }
              }
            }
          ]
        }`), t)

	err := state.Validate()
	assert.Error(t, err)
	assert.EqualErrorf(t, err, "state \"Branch_A_Start\" next state \"Branch_A_Done\" is unreachable",
		          "Checking that Validate fails when branches task next is unreachable")
}

func Test_ParallelState_Valid(t *testing.T) {
	state := parseParallelTaskState([]byte(`{
          "Type": "Parallel",
          "Next": "Parallel_Block_Done",
          "Branches": [
            {
              "StartAt": "Branch_A_Start",
              "States": {
                "Branch_A_Start": {
                "Type": "TaskFn",
                "Resource": "host1:machine1:func1",
                "Next": "Branch_A_Next_One"
                },
                "Branch_A_Next_One": {
                "Type": "TaskFn",
                "Resource": "host1:machine1:func2",
                "Next": "Branch_A_Done"
                },
                "Branch_A_Done": {
                    "Type": "Succeed"
                }
              }
            },
            {
              "StartAt": "Branch_B_Start",
              "States": {
                "Branch_B_Start": {
                "Type": "TaskFn",
                "Resource": "host1:machine2:func2",
                "End": true
                }
              }
            }
          ]
        }`), t)
	err := state.Validate()
	assert.Nil(t, err)
}

// Tests on state execution

func Test_Parallel_State_Execution(t *testing.T) {

}

func Test_Parallel_State_Execution_With_Subsequent_Input_Data_Usage(t *testing.T) {

}

// Private functions

func parseParallelTaskState(b []byte, t *testing.T) *ParallelState {
	var p ParallelState
	err := json.Unmarshal(b, &p)
	assert.NoError(t, err)
	p.SetName(to.Strp("TestState"))
	p.SetType(to.Strp("Parallel"))
	return &p
}