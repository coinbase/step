package machine

import (
	"testing"

	"github.com/coinbase/step/machine/state"
	"github.com/coinbase/step/utils/to"
	"github.com/stretchr/testify/assert"
)

func Test_Machine_Parser_FromJSON(t *testing.T) {
	json := []byte(`
  {
      "Comment": "Adds some coordinates to the input",
      "StartAt": "Coords",
      "States": {
        "Coords": {
          "Type": "Pass",
          "Result": {
            "x": 3.14,
            "y": 103.14159
          },
          "ResultPath": "$.coords",
          "End": true
        }
    }
  }`)

	_, err := FromJSON(json)

	assert.Equal(t, err, nil)
}

func Test_Parser_Expands_TaskFn(t *testing.T) {
	json := []byte(`
  {
      "StartAt": "A",
      "States": {
        "A": {
          "Type": "TaskFn",
          "Next": "B"
        },
        "B": {
          "Type": "TaskFn",
          "End": true
        }
    }
  }`)

	sm, err := FromJSON(json)
	assert.NoError(t, err)

	// Names and Types
	assert.Equal(t, len(sm.States), 4)
	assert.Equal(t, *sm.States["A"].GetType(), "Pass")
	assert.Equal(t, *sm.States[TaskFnName("A")].GetType(), "Task")

	assert.Equal(t, *sm.States["B"].GetType(), "Pass")
	assert.Equal(t, *sm.States[TaskFnName("B")].GetType(), "Task")

	a_state := sm.States["A"].(*state.PassState)
	atask_state := sm.States[TaskFnName("A")].(*state.TaskState)

	b_state := sm.States["B"].(*state.PassState)
	btask_state := sm.States[TaskFnName("B")].(*state.TaskState)

	// ORDER
	assert.Equal(t, *a_state.Next, TaskFnName("A"))
	assert.Equal(t, *atask_state.Next, "B")
	assert.Equal(t, *b_state.Next, TaskFnName("B"))
	assert.Equal(t, *btask_state.End, true)

	// Passing the right things
	assert.Equal(t, a_state.Result.(string), "_A_")
	path, err := to.PrettyJSON(a_state.ResultPath)
	assert.NoError(t, err)
	assert.Equal(t, path, `"$.Task"`)
}

func Test_Machine_Parser_FileNonexistantFile(t *testing.T) {
	_, err := ParseFile("../examples/non_existent_file.json")
	assert.Error(t, err)
}

func Test_Machine_Parser_OfBadStateType(t *testing.T) {
	_, err := ParseFile("../examples/bad_type.json")

	assert.Error(t, err)
	assert.Regexp(t, "Unknown State", err.Error())
}

func Test_Machine_Parser_OfBadPath(t *testing.T) {
	_, err := ParseFile("../examples/bad_path.json")

	assert.Error(t, err)
	assert.Regexp(t, "Bad JSON path", err.Error())
}

// Validation Tests

// BASIC TYPE TESTS

func Test_Machine_Parser_AllTypes(t *testing.T) {
	sm, err := ParseFile("../examples/all_types.json")
	assert.NoError(t, err)

	assert.NoError(t, sm.Validate())
}

func Test_Machine_Parser_BasicPass(t *testing.T) {
	sm, err := ParseFile("../examples/basic_pass.json")
	assert.NoError(t, err)
	assert.NoError(t, sm.Validate())
}

func Test_Machine_Parser_BasicChoice(t *testing.T) {
	sm, err := ParseFile("../examples/basic_choice.json")

	assert.Equal(t, err, nil)
	assert.NoError(t, sm.Validate())
}

func Test_Machine_Parser_TaskFn(t *testing.T) {
	sm, err := ParseFile("../examples/taskfn.json")

	assert.Equal(t, err, nil)
	assert.NoError(t, sm.Validate())
}
