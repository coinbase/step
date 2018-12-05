package machine

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/coinbase/step/utils/to"
	"github.com/stretchr/testify/assert"
)

func loadFixture(file string, t *testing.T) *StateMachine {
	example_machine, err := ParseFile(file)
	assert.NoError(t, err)
	return example_machine
}

func execute(json []byte, input interface{}, t *testing.T) (map[string]interface{}, error) {
	example_machine, err := FromJSON(json)
	assert.NoError(t, err)
	example_machine.SetDefaultHandler()

	exec, err := example_machine.Execute(input)

	return exec.Output, err
}

func executeFixture(file string, input map[string]interface{}, t *testing.T) map[string]interface{} {
	example_machine := loadFixture(file, t)

	exec, err := example_machine.Execute(input)

	assert.NoError(t, err)

	return exec.Output
}

//////
// TESTS
//////

func Test_Machine_EmptyStateMachinePassExample(t *testing.T) {
	_, err := execute([]byte(EmptyStateMachine), make(map[string]interface{}), t)
	assert.NoError(t, err)
}

func Test_Machine_SimplePassExample_With_Execute(t *testing.T) {
	json := []byte(`
  {
      "StartAt": "start",
      "States": {
        "start": {
          "Type": "Pass",
          "Result": "b",
          "ResultPath": "$.a",
          "End": true
        }
    }
  }`)

	output, err := execute(json, make(map[string]interface{}), t)
	assert.NoError(t, err)
	assert.Equal(t, output["a"], "b")

	output, err = execute(json, "{}", t)
	assert.NoError(t, err)
	assert.Equal(t, output["a"], "b")

	output, err = execute(json, to.Strp("{}"), t)
	assert.NoError(t, err)
	assert.Equal(t, output["a"], "b")
}

func Test_Machine_ErrorUnknownState(t *testing.T) {
	example_machine := loadFixture("../examples/bad_unknown_state.json", t)
	_, err := example_machine.Execute(make(map[string]interface{}))

	assert.Error(t, err)
	assert.Regexp(t, "Unknown State", err.Error())
}

func Test_Machine_MarshallAllTypes(t *testing.T) {
	file := "../examples/all_types.json"
	sm, err := ParseFile(file)
	assert.NoError(t, err)

	sm.SetDefaultHandler()
	assert.NoError(t, sm.Validate())

	marshalled_json, err := json.Marshal(sm)
	assert.NoError(t, err)

	raw_json, err := ioutil.ReadFile(file)
	assert.NoError(t, err)

	assert.JSONEq(t, string(raw_json), string(marshalled_json))
}
