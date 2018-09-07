package state

import (
	"encoding/json"
	"testing"

	"github.com/coinbase/step/utils/to"
	"github.com/stretchr/testify/assert"
)

type stateTestData struct {
	Input  map[string]interface{}
	Output map[string]interface{}
	Error  *string
	Next   *string
}

func testState(state State, std stateTestData, t *testing.T) {
	// Make sure the execution is on Valid State
	err := state.Validate()
	assert.NoError(t, err)

	// default empty input
	if std.Input == nil {
		std.Input = map[string]interface{}{}
	}

	output, next, err := state.Execute(nil, std.Input)

	// expecting error?
	if std.Error != nil {
		assert.Error(t, err)
		assert.Regexp(t, *std.Error, err.Error())
	} else if err != nil {
		assert.NoError(t, err)
	}

	if std.Output != nil {
		assert.Equal(t, std.Output, output)
	}

	if std.Next != nil {
		assert.Equal(t, *std.Next, *next)
	}
}

func parseChoiceState(b []byte, t *testing.T) *ChoiceState {
	var p ChoiceState
	err := json.Unmarshal(b, &p)
	assert.NoError(t, err)
	p.SetName(to.Strp("TestState"))
	p.SetType(to.Strp("Choice"))
	return &p
}

func parsePassState(b []byte, t *testing.T) *PassState {
	var p PassState
	err := json.Unmarshal(b, &p)
	assert.NoError(t, err)
	p.SetName(to.Strp("TestState"))
	p.SetType(to.Strp("Pass"))
	return &p
}

func parseWaitState(b []byte, t *testing.T) *WaitState {
	var p WaitState
	err := json.Unmarshal(b, &p)
	assert.NoError(t, err)
	p.SetName(to.Strp("TestState"))
	p.SetType(to.Strp("Wait"))
	return &p
}

func parseTaskState(b []byte, t *testing.T) *TaskState {
	var p TaskState
	err := json.Unmarshal(b, &p)
	assert.NoError(t, err)
	p.SetName(to.Strp("TestState"))
	p.SetType(to.Strp("Task"))
	return &p
}

func parseValidTaskState(b []byte, handler interface{}, t *testing.T) *TaskState {
	state := parseTaskState(b, t)
	state.SetResourceFunction(handler)
	assert.NoError(t, state.Validate())
	return state
}
