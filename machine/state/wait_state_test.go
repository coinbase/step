package state

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_WaitState_XORofFields(t *testing.T) {
	state := parseWaitState([]byte(`
  {
    "Seconds": 10,
    "TimestampPath": "$.a.b",
    "Timestamp": "2006-01-02T15:04:05Z",
    "Next": "Public"
  }`), t)

	err := state.Validate()
	assert.Error(t, err)

	assert.Regexp(t, "Exactly One", err.Error())
}

func Test_WaitState_SecondsPath(t *testing.T) {
	state := parseWaitState([]byte(`
  {
    "SecondsPath": "$.path",
    "Next": "Public"
	}`), t)

	_, _, err := state.Execute(nil, map[string]interface{}{"path": 30})
	assert.NoError(t, err)

	_, _, err = state.Execute(nil, map[string]interface{}{})
	assert.Error(t, err)
}
