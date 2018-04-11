package execution

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/service/sfn"
	"github.com/aws/aws-sdk-go/service/sfn/sfniface"
	"github.com/coinbase/step/utils/to"
)

type Execution struct {
	ExecutionArn *string
	StartDate    *time.Time
}

type ExecutionWaiter func(*ExecutionDetails, *StateDetails, error) error

func StartExecution(sfnc sfniface.SFNAPI, arn *string, name *string, input interface{}) (*Execution, error) {
	input_json, err := to.PrettyJSON(input)

	if err != nil {
		return nil, err
	}
	return StartExecutionRaw(sfnc, arn, name, to.Strp(string(input_json)))
}

func StartExecutionRaw(sfnc sfniface.SFNAPI, arn *string, name *string, input_json *string) (*Execution, error) {
	out, err := sfnc.StartExecution(&sfn.StartExecutionInput{
		Input:           input_json,
		StateMachineArn: arn,
		Name:            name,
	})

	if err != nil {
		return nil, err
	}

	return &Execution{ExecutionArn: out.ExecutionArn, StartDate: out.StartDate}, nil
}
func FindExecution(sfnc sfniface.SFNAPI, arn *string, name_prefix string) (*Execution, error) {
	// TODO search through pages for first match
	out, err := sfnc.ListExecutions(&sfn.ListExecutionsInput{
		MaxResults:      to.Int64p(100),
		StatusFilter:    to.Strp("RUNNING"),
		StateMachineArn: arn,
	})

	if err != nil {
		return nil, err
	}

	for _, exec := range out.Executions {
		name := *exec.Name
		if len(name) < len(name_prefix) {
			continue
		}

		if name[0:len(name_prefix)] == name_prefix {
			return &Execution{ExecutionArn: exec.ExecutionArn, StartDate: exec.StartDate}, nil
		}
	}

	return nil, nil
}

type ExecutionDetails struct {
	*sfn.DescribeExecutionOutput
}

type StateDetails struct {
	LastStateName *string
	LastTaskName  *string
	LastOutput    *string
	Timestamp     *time.Time
}

func (e *Execution) getDetails(sfnc sfniface.SFNAPI) (*ExecutionDetails, *StateDetails, error) {
	exec_out, err := sfnc.DescribeExecution(&sfn.DescribeExecutionInput{
		ExecutionArn: e.ExecutionArn,
	})

	if err != nil {
		return nil, nil, err
	}

	history_out, err := sfnc.GetExecutionHistory(&sfn.GetExecutionHistoryInput{
		ExecutionArn: e.ExecutionArn,
		ReverseOrder: to.Boolp(true),
		MaxResults:   to.Int64p(20), // Enough to Get the Most Recent State Output
	})

	if err != nil {
		return nil, nil, err
	}

	sd := StateDetails{}

	// We reverse look for last State Existed Event with Output.
	// So even on Failure we can see the final details of Failure
	for _, he := range history_out.Events {
		if he.Timestamp == nil {
			sd.Timestamp = he.Timestamp
		}

		if he.StateEnteredEventDetails != nil {
			if sd.LastStateName == nil {
				sd.LastStateName = he.StateEnteredEventDetails.Name
			}
		}

		if he.StateExitedEventDetails != nil {
			if sd.LastStateName == nil {
				sd.LastStateName = he.StateExitedEventDetails.Name
			}

			if sd.LastOutput == nil {
				sd.LastOutput = he.StateExitedEventDetails.Output
			}

			if sd.LastTaskName == nil && *he.Type == "TaskStateExited" {
				sd.LastTaskName = he.StateExitedEventDetails.Name
			}
		}
	}

	return &ExecutionDetails{exec_out}, &sd, nil
}

// WaitForExecution allows another application to wait for the execution to finish
// and process output as it comes in for usability
func (e *Execution) WaitForExecution(sfnc sfniface.SFNAPI, sleep int, fn ExecutionWaiter) {
	for {
		exec, state, err := e.getDetails(sfnc)

		err = fn(exec, state, err)

		if err != nil {
			fmt.Println(err.Error())
			return
		}

		if *exec.Status != "RUNNING" {
			return // Exit out of loop if execution has finished
		}

		time.Sleep(time.Duration(int64(sleep)) * time.Second)
	}
}
