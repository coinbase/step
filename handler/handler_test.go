package handler

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/coinbase/step/utils/to"
	"github.com/stretchr/testify/assert"
)

type TestStruct struct {
	Message *string
}

func Test_Handler_Execution(t *testing.T) {
	called := false
	testHandler := func(_ context.Context, ts *TestStruct) (interface{}, error) {
		assert.Equal(t, ts.Message, to.Strp("mmss"))
		called = true
		return "asd", nil
	}

	tm := TaskHandlers{"Tester": testHandler}
	handle, err := CreateHandler(&tm)
	assert.NoError(t, err)

	var raw RawMessage
	err = json.Unmarshal([]byte(`{"Task": "Tester", "Input": {"Message": "mmss"}}`), &raw)
	assert.NoError(t, err)

	out, err := handle(nil, &raw)

	assert.NoError(t, err)
	assert.True(t, called)
	assert.Equal(t, out, "asd")
}

func Test_Handler_Execution_with_TaskHandler_and_NoTaskHandler(t *testing.T) {
	nthCalled := false
	thCalled := false

	noTaskHandler := func(_ context.Context, ts *TestStruct) (interface{}, error) {
		assert.Equal(t, ts.Message, to.Strp("mmss"))
		nthCalled = true
		return "nth", nil
	}

	taskHandler := func(_ context.Context, ts *TestStruct) (interface{}, error) {
		assert.Equal(t, ts.Message, to.Strp("mmss"))
		thCalled = true
		return "th", nil
	}

	tm := TaskHandlers{"": noTaskHandler, "Tester": taskHandler}

	handle, err := CreateHandler(&tm)
	assert.NoError(t, err)

	var rawNth RawMessage
	err = json.Unmarshal([]byte(`{"Message": "mmss"}`), &rawNth)
	assert.NoError(t, err)

	var rawTh RawMessage
	err = json.Unmarshal([]byte(`{"Task": "Tester", "Input": {"Message": "mmss"}}`), &rawTh)
	assert.NoError(t, err)

	outNth, err := handle(nil, &rawNth)
	assert.NoError(t, err)

	outTh, err := handle(nil, &rawTh)
	assert.NoError(t, err)

	assert.True(t, nthCalled)
	assert.True(t, thCalled)

	assert.Equal(t, outNth, "nth")
	assert.Equal(t, outTh, "th")
}

func Test_Handler_Failure(t *testing.T) {
	tm := TaskHandlers{}
	handle, err := CreateHandler(&tm)
	assert.NoError(t, err)

	_, err = handle(nil, &RawMessage{})
	assert.Error(t, err)

	_, err = handle(nil, &RawMessage{Task: to.Strp("Tester")})
	assert.Error(t, err)
}
