package machine

import (
	"testing"

	"github.com/coinbase/step/utils/to"
	"github.com/stretchr/testify/assert"
)

func Test_PassState_Defaults(t *testing.T) {
	state := parsePassState([]byte(`{ "Next": "Pass", "End": true}`), t)
	err := state.Validate()
	assert.Error(t, err)

	assert.Equal(t, *state.GetType(), "Pass")
	assert.Equal(t, errorPrefix(state), "PassState(TestState) Error:")

	assert.Regexp(t, "End and Next both defined", err.Error())
}

// Validations

func Test_PassState_EndNextBothDefined(t *testing.T) {
	state := parsePassState([]byte(`{ "Next": "Pass", "End": true}`), t)
	err := state.Validate()
	assert.Error(t, err)

	assert.Regexp(t, "End and Next both defined", err.Error())
}

func Test_PassState_EndNextBothUnDefined(t *testing.T) {
	state := parsePassState([]byte(`{}`), t)
	err := state.Validate()
	assert.Error(t, err)

	assert.Regexp(t, "End and Next both undefined", err.Error())
}

// Execution

func Test_PassState_ResultPath(t *testing.T) {
	state := parsePassState([]byte(`{ "Next": "Pass", "Result": "b", "ResultPath": "$.a"}`), t)
	testState(state, stateTestData{Output: map[string]interface{}{"a": "b"}}, t)
}

func Test_PassState_ResultPathOverrwite(t *testing.T) {
	state := parsePassState([]byte(`{ "Next": "Pass", "Result": "b", "ResultPath": "$.a"}`), t)
	testState(state, stateTestData{
		Input:  map[string]interface{}{"a": "c"},
		Output: map[string]interface{}{"a": "b"},
	}, t)
}

func Test_PassState_InputPath(t *testing.T) {
	state := parsePassState([]byte(`{"Next": "Pass",  "InputPath": "$.a"}`), t)

	deep := map[string]interface{}{"a": "b"}
	input := map[string]interface{}{"a": deep}

	testState(state, stateTestData{
		Input:  input,
		Output: deep,
	}, t)
}

func Test_PassState_OutputPath(t *testing.T) {
	state := parsePassState([]byte(`{ "Next": "Pass",  "OutputPath": "$.a"}`), t)

	deep := map[string]interface{}{"a": "b"}
	input := map[string]interface{}{"a": deep}

	testState(state, stateTestData{
		Input:  input,
		Output: deep,
	}, t)
}

// Bad Execution

func Test_PassState_BadInputPath(t *testing.T) {
	state := parsePassState([]byte(`{"Next": "Pass","InputPath": "$.a.b"}`), t)

	testState(state, stateTestData{
		Input: map[string]interface{}{"a": "b"},
		Error: to.Strp("Input Error"),
	}, t)
}

func Test_PassState_BadOutputPath(t *testing.T) {
	state := parsePassState([]byte(`{"Next": "Pass","OutputPath": "$.a.b"}`), t)

	testState(state, stateTestData{
		Input: map[string]interface{}{"a": "b"},
		Error: to.Strp("Output Error"),
	}, t)
}
