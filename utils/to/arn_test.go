package to

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var InputStateMachine = `{
	"StartAt": "Start",
	"States": {
		"Start": {
			"Type": "Task",
			"Resource": "arn:aws:lambda:{{aws_region}}:{{aws_account}}:function:{{lambda_name}}",
			"Next": "WIN"
		},
		"WIN": {"Type": "Succeed"}
	}
}`

var DesiredStateMachine = `{
	"StartAt": "Start",
	"States": {
		"Start": {
			"Type": "Task",
			"Resource": "arn:aws:lambda:test-region:test-account:function:test-lambda",
			"Next": "WIN"
		},
		"WIN": {"Type": "Succeed"}
	}
}`

func Test_to_InterpolateArnVariables(t *testing.T) {
	resultStateMachine := InterpolateArnVariables(
		&InputStateMachine,
		Strp("test-region"),
		Strp("test-account"),
		Strp("test-lambda"),
	)
	assert.Equal(t, *resultStateMachine, DesiredStateMachine)
}
