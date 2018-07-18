package state

import (
	"context"
	"testing"

	"github.com/coinbase/step/utils/to"
	"github.com/stretchr/testify/assert"
)

/////////
// TYPES
/////////

type TestError struct{}

func (t *TestError) Error() string {
	return "This is a Test Error"
}

type TestHandler func(context.Context, interface{}) (interface{}, error)

func countCalls(th TestHandler) (TestHandler, *int) {
	calls := 0
	return func(ctx context.Context, input interface{}) (interface{}, error) {
		calls++
		return th(ctx, input)
	}, &calls
}

func ThrowTestErrorHandler(_ context.Context, input interface{}) (interface{}, error) {
	return nil, &TestError{}
}

func ReturnMapTestHandler(_ context.Context, input interface{}) (interface{}, error) {
	return map[string]interface{}{"z": "y"}, nil
}

// Execution

func Test_TaskState_ValidateResource(t *testing.T) {
	state := parseTaskState([]byte(`{ "Next": "Pass"}`), t)
	assert.Error(t, state.Validate())
	state.Resource = to.Strp("resource")
	assert.NoError(t, state.Validate())
}

func Test_TaskState_Valid_ErrorEquals_StatesAll(t *testing.T) {
	state := parseTaskState([]byte(`{
		"Resource": "asd",
		"Next": "Pass",
		"Retry": [{ "ErrorEquals": ["States.ALL"] }]
	}`), t)

	assert.NoError(t, state.Validate())

	state = parseTaskState([]byte(`{
		"Resource": "asd",
		"Next": "Pass",
		"Retry": [{ "ErrorEquals": ["States.ALL", "NoMoreErrors"] }]
	}`), t)
	assert.Error(t, state.Validate())

	state = parseTaskState([]byte(`{
		"Resource": "asd",
		"Next": "Pass",
		"Retry": [{ "ErrorEquals": ["States.ALL"] }, { "ErrorEquals": ["NotLast"] }]
	}`), t)

	state = parseTaskState([]byte(`{
		"Resource": "asd",
		"Next": "Pass",
		"Retry": [{ "ErrorEquals": ["States.NotRealError"] }]
	}`), t)

	assert.Error(t, state.Validate())
}

func Test_TaskState_ResourceFunction(t *testing.T) {
	th, calls := countCalls(ReturnMapTestHandler)

	state := parseValidTaskState([]byte(`{ "Next": "Pass"}`), th, t)

	testState(state, stateTestData{
		Input:  map[string]interface{}{"a": "c"},
		Output: map[string]interface{}{"z": "y"},
	}, t)

	assert.Equal(t, 1, *calls)
}

func Test_TaskState_Catch_Works(t *testing.T) {
	state := parseValidTaskState([]byte(`{
		"Next": "Pass",
		"Catch": [{
			"ErrorEquals": ["TestError"],
			"Next": "Fail"
		}]
	}`), ThrowTestErrorHandler, t)

	testState(state, stateTestData{
		Input:  map[string]interface{}{"a": "c"},
		Output: map[string]interface{}{"Error": "TestError", "Cause": "This is a Test Error"},
		Next:   to.Strp("Fail"),
	}, t)
}

func Test_TaskState_Catch_Doesnt_Catch(t *testing.T) {
	state := parseValidTaskState([]byte(`{
		"Next": "Pass",
		"Catch": [{
			"ErrorEquals": ["NotTestError"],
			"Next": "Fail"
		}]
	}`), ThrowTestErrorHandler, t)

	testState(state, stateTestData{
		Input: map[string]interface{}{"a": "c"},
		Error: to.Strp("This is a Test Error"),
	}, t)
}

func Test_TaskState_Retry_Works(t *testing.T) {
	th, calls := countCalls(ThrowTestErrorHandler)

	state := parseValidTaskState([]byte(`{
		"Next": "Pass",
		"Retry": [{
			"ErrorEquals": ["TestError"],
			"MaxAttempts": 2
		}]
	}`), th, t)

	testState(state, stateTestData{
		Input: map[string]interface{}{"a": "c"},
		Next:  state.Name(),
	}, t)

	testState(state, stateTestData{
		Input: map[string]interface{}{"a": "c"},
		Next:  state.Name(),
	}, t)

	testState(state, stateTestData{
		Input: map[string]interface{}{"a": "c"},
		Error: to.Strp("This is a Test Error"),
	}, t)

	// 1 initial call, + 2 retries
	assert.Equal(t, 3, *calls)
}

func Test_TaskState_Catch_AND_Retry_Works(t *testing.T) {
	th, calls := countCalls(ThrowTestErrorHandler)

	state := parseValidTaskState([]byte(`{
		"Next": "Pass",
		"Retry": [{
			"ErrorEquals": ["TestError"],
			"MaxAttempts": 1
		}],
		"Catch": [{
			"ErrorEquals": ["TestError"],
			"Next": "Fail"
		}]
	}`), th, t)

	testState(state, stateTestData{
		Input: map[string]interface{}{"a": "c"},
		Next:  state.Name(),
	}, t)

	testState(state, stateTestData{
		Input: map[string]interface{}{"a": "c"},
		Next:  to.Strp("Fail"),
	}, t)

	assert.Equal(t, 2, *calls)
}

func Test_TaskState_Catch_AND_Retry_StateAll(t *testing.T) {
	th, calls := countCalls(ThrowTestErrorHandler)

	state := parseValidTaskState([]byte(`{
		"Next": "Pass",
		"Retry": [{
			"ErrorEquals": ["States.ALL"],
			"MaxAttempts": 1
		}],
		"Catch": [{
			"ErrorEquals": ["States.ALL"],
			"Next": "Fail"
		}]
	}`), th, t)

	testState(state, stateTestData{
		Input: map[string]interface{}{"a": "c"},
		Next:  state.Name(),
	}, t)

	testState(state, stateTestData{
		Input: map[string]interface{}{"a": "c"},
		Next:  to.Strp("Fail"),
	}, t)

	assert.Equal(t, 2, *calls)
}
