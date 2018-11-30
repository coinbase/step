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

	tm := TaskFunctions{"Tester": testHandler}
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

func Test_Handler_Failure(t *testing.T) {
	tm := TaskFunctions{}
	handle, err := CreateHandler(&tm)
	assert.NoError(t, err)

	_, err = handle(nil, &RawMessage{})
	assert.Error(t, err)

	_, err = handle(nil, &RawMessage{Task: to.Strp("Tester")})
	assert.Error(t, err)
}
