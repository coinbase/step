package deployer

import (
	"github.com/coinbase/step/aws"
	"github.com/coinbase/step/machine"
)

// StateMachine returns the StateMachine for the deployer
func StateMachine() (*machine.StateMachine, error) {
	return machine.FromJSON([]byte(`{
    "Comment": "Step Function Deployer",
    "StartAt": "Validate",
    "States": {
      "Validate": {
        "Type": "TaskFn",
        "Comment": "Validate and Set Defaults",
        "Next": "Lock",
        "Catch": [
          {
            "Comment": "Bad Release or Error GoTo end",
            "ErrorEquals": ["BadReleaseError", "UnmarshalError", "PanicError"],
            "ResultPath": "$.error",
            "Next": "FailureClean"
          }
        ]
      },
      "Lock": {
        "Type": "TaskFn",
        "Comment": "Grab Lock",
        "Next": "ValidateResources",
        "Catch": [
          {
            "Comment": "Something else is deploying",
            "ErrorEquals": ["LockExistsError"],
            "ResultPath": "$.error",
            "Next": "FailureClean"
          },
          {
            "Comment": "Try Release Lock Then Fail",
            "ErrorEquals": ["LockError", "PanicError"],
            "ResultPath": "$.error",
            "Next": "ReleaseLockFailure"
          }
        ]
      },
      "ValidateResources": {
        "Type": "TaskFn",
        "Comment": "ValidateResources",
        "Next": "Deploy",
        "Catch": [
          {
            "Comment": "Try Release Lock Then Fail",
            "ErrorEquals": ["BadReleaseError", "PanicError"],
            "ResultPath": "$.error",
            "Next": "ReleaseLockFailure"
          }
        ]
      },
      "Deploy": {
        "Type": "TaskFn",
        "Comment": "Upload Step-Function and Lambda",
        "Next": "Success",
        "Catch": [
          {
            "Comment": "Unsure of State, Leave Lock and Fail",
            "ErrorEquals": ["DeploySFNError"],
            "ResultPath": "$.error",
            "Next": "ReleaseLockFailure"
          },
          {
            "Comment": "Unsure of State, Leave Lock and Fail",
            "ErrorEquals": ["DeployLambdaError", "PanicError"],
            "ResultPath": "$.error",
            "Next": "FailureDirty"
          }
        ]
      },
      "ReleaseLockFailure": {
        "Type": "TaskFn",
        "Comment": "Release the Lock and Fail",
        "Next": "FailureClean",
        "Retry": [ {
          "Comment": "Keep trying to Release",
          "ErrorEquals": ["LockError"],
          "MaxAttempts": 3,
          "IntervalSeconds": 30
        }],
        "Catch": [{
          "ErrorEquals": ["LockError", "PanicError"],
          "ResultPath": "$.error",
          "Next": "FailureDirty"
        }]
      },
      "FailureClean": {
        "Comment": "Deploy Failed Cleanly",
        "Type": "Fail",
        "Error": "NotifyError"
      },
      "FailureDirty": {
        "Comment": "Deploy Failed, Resources left in Bad State, ALERT!",
        "Type": "Fail",
        "Error": "AlertError"
      },
      "Success": {
        "Type": "Succeed"
      }
    }
  }`))
}

func StateMachineWithTaskHandlers() (*machine.StateMachine, error) {
	// ASSIGN THE TASK FUNCTIONS
	state_machine, err := StateMachine()
	if err != nil {
		return nil, err
	}

	AddStateMachineHandlers(state_machine, &aws.AwsClientsStr{})

	return state_machine, nil
}

func AddStateMachineHandlers(state_machine *machine.StateMachine, awsc aws.AwsClients) {
	state_machine.SetResourceFunction("Validate", ValidateHandler(awsc))
	state_machine.SetResourceFunction("Lock", LockHandler(awsc))
	state_machine.SetResourceFunction("ValidateResources", ValidateResourcesHandler(awsc))
	state_machine.SetResourceFunction("Deploy", DeployHandler(awsc))
	state_machine.SetResourceFunction("ReleaseLockFailure", ReleaseLockFailureHandler(awsc))
}

func StateMachineWithLambdaArn(lambdaArn *string) (*machine.StateMachine, error) {
	state_machine, err := StateMachine()
	if err != nil {
		return nil, err
	}
	state_machine.SetResource(lambdaArn) // Add Lambda Arn to Tasks
	return state_machine, nil
}
