package deployer

import (
	"github.com/coinbase/step/aws"
	"github.com/coinbase/step/machine"
)

// StateMachine returns the StateMachine for the deployer
func StateMachine() (*machine.StateMachine, error) {
	return machine.FromJSON([]byte(`{
    "Comment": "Step Function Deployer",
    "StartAt": "ValidateFn",
    "States": {
      "ValidateFn": {
        "Type": "Pass",
        "Result": "Validate",
        "ResultPath": "$.Task",
        "Next": "Validate"
      },
      "Validate": {
        "Type": "Task",
        "Comment": "Validate and Set Defaults",
        "Next": "LockFn",
        "Catch": [
          {
            "Comment": "Bad Release or Error GoTo end",
            "ErrorEquals": ["BadReleaseError", "UnmarshalError"],
            "ResultPath": "$.error",
            "Next": "FailureClean"
          }
        ]
      },
      "LockFn": {
        "Type": "Pass",
        "Result": "Lock",
        "ResultPath": "$.Task",
        "Next": "Lock"
      },
      "Lock": {
        "Type": "Task",
        "Comment": "Grab Lock",
        "Next": "ValidateResourcesFn",
        "Catch": [
          {
            "Comment": "Something else is deploying",
            "ErrorEquals": ["LockExistsError"],
            "ResultPath": "$.error",
            "Next": "FailureClean"
          },
          {
            "Comment": "Try Release Lock Then Fail",
            "ErrorEquals": ["LockError"],
            "ResultPath": "$.error",
            "Next": "ReleaseLockFailureFn"
          }
        ]
      },
      "ValidateResourcesFn": {
        "Type": "Pass",
        "Result": "ValidateResources",
        "ResultPath": "$.Task",
        "Next": "ValidateResources"
      },
      "ValidateResources": {
        "Type": "Task",
        "Comment": "ValidateResources",
        "Next": "DeployFn",
        "Catch": [
          {
            "Comment": "Try Release Lock Then Fail",
            "ErrorEquals": ["BadReleaseError"],
            "ResultPath": "$.error",
            "Next": "ReleaseLockFailureFn"
          }
        ]
      },
      "DeployFn": {
        "Type": "Pass",
        "Result": "Deploy",
        "ResultPath": "$.Task",
        "Next": "Deploy"
      },
      "Deploy": {
        "Type": "Task",
        "Comment": "Upload Step-Function and Lambda",
        "Next": "Success",
        "Catch": [
          {
            "Comment": "Unsure of State, Leave Lock and Fail",
            "ErrorEquals": ["DeploySFNError"],
            "ResultPath": "$.error",
            "Next": "ReleaseLockFailureFn"
          },
          {
            "Comment": "Unsure of State, Leave Lock and Fail",
            "ErrorEquals": ["DeployLambdaError"],
            "ResultPath": "$.error",
            "Next": "FailureDirty"
          }
        ]
      },
      "ReleaseLockFailureFn": {
        "Type": "Pass",
        "Result": "ReleaseLockFailure",
        "ResultPath": "$.Task",
        "Next": "ReleaseLockFailure"
      },
      "ReleaseLockFailure": {
        "Type": "Task",
        "Comment": "Release the Lock and Fail",
        "Next": "FailureClean",
        "Retry": [ {
          "Comment": "Keep trying to Release",
          "ErrorEquals": ["LockError"],
          "MaxAttempts": 3,
          "IntervalSeconds": 30
        }],
        "Catch": [{
          "ErrorEquals": ["LockError"],
          "ResultPath": "$.error",
          "Next": "FailureDirty"
        }]
      },
      "FailureClean": {
        "Comment": "Deploy Failed Cleanly",
        "Type": "Fail",
        "Error": "FailureClean"
      },
      "FailureDirty": {
        "Comment": "Deploy Failed, Resources left in Bad State, ALERT!",
        "Type": "Fail",
        "Error": "FailureDirty"
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

	awsc := aws.CreateAwsClients()
	AddStateMachineHandlers(state_machine, awsc)

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
