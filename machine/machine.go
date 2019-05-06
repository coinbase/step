// State Machine implementation
package machine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/coinbase/step/handler"
	"github.com/coinbase/step/machine/state"
	"github.com/coinbase/step/utils/is"
	"github.com/coinbase/step/utils/to"
)

func DefaultHandler(_ context.Context, input interface{}) (interface{}, error) {
	return map[string]string{}, nil
}

// EmptyStateMachine is a small Valid StateMachine
var EmptyStateMachine = `{
  "StartAt": "WIN",
  "States": { "WIN": {"Type": "Succeed"}}
}`

// IMPLEMENTATION

// States is the collection of states
type States map[string]state.State

// StateMachine the core struct for the machine
type StateMachine struct {
	Comment *string `json:",omitempty"`

	StartAt *string

	States States
}

// Global Methods
func Validate(sm_json *string) error {
	state_machine, err := FromJSON([]byte(*sm_json))
	if err != nil {
		return err
	}

	if err := state_machine.Validate(); err != nil {
		return err
	}

	return nil
}

func (sm *StateMachine) FindTask(name string) (*state.TaskState, error) {
	task, ok := sm.Tasks()[name]

	if !ok {
		return nil, fmt.Errorf("Handler Error: Cannot Find Task %v", name)
	}

	return task, nil
}

func (sm *StateMachine) Tasks() map[string]*state.TaskState {
	tasks := map[string]*state.TaskState{}
	for name, s := range sm.States {
		switch s.(type) {
		case *state.TaskState:
			tasks[name] = s.(*state.TaskState)
		}
	}
	return tasks
}

func (sm *StateMachine) SetResource(lambda_arn *string) {
	for _, task := range sm.Tasks() {
		if task.Resource == nil {
			task.Resource = lambda_arn
		}
	}
}

func (sm *StateMachine) SetDefaultHandler() {
	for _, task := range sm.Tasks() {
		task.SetTaskHandler(DefaultHandler)
	}
}

func (sm *StateMachine) SetTaskFnHandlers(tfs *handler.TaskHandlers) error {
	taskHandlers, err := handler.CreateHandler(tfs)
	if err != nil {
		return err
	}

	for name, _ := range *tfs {
		if err := sm.SetTaskHandler(name, taskHandlers); err != nil {
			return err
		}
	}

	return nil
}

func (sm *StateMachine) SetTaskHandler(task_name string, resource_fn interface{}) error {
	task, err := sm.FindTask(task_name)
	if err != nil {
		return err
	}

	task.SetTaskHandler(resource_fn)
	return nil
}

func (sm *StateMachine) Validate() error {
	if is.EmptyStr(sm.StartAt) {
		return errors.New("State Machine requires StartAt")
	}

	if sm.States == nil {
		return errors.New("State Machine must have States")
	}

	if len(sm.States) == 0 {
		return errors.New("State Machine must have States")
	}

	state_errors := []string{}

	for _, state := range sm.States {
		err := state.Validate()
		if err != nil {
			state_errors = append(state_errors, err.Error())
		}
	}

	if len(state_errors) != 0 {
		return fmt.Errorf("State Errors %q", state_errors)
	}

	// TODO: validate all states are reachable
	return nil
}

func (sm *StateMachine) DefaultLambdaContext(lambda_name string) context.Context {
	return lambdacontext.NewContext(context.Background(), &lambdacontext.LambdaContext{
		InvokedFunctionArn: fmt.Sprintf("arn:aws:lambda:us-east-1:000000000000:function:%v", lambda_name),
	})
}

func processInput(input interface{}) (interface{}, error) {
	// Make
	switch input.(type) {
	case string:
		var json_input map[string]interface{}
		if err := json.Unmarshal([]byte(input.(string)), &json_input); err != nil {
			return nil, err
		}
		return json_input, nil
	case *string:
		var json_input map[string]interface{}
		if err := json.Unmarshal([]byte(*(input.(*string))), &json_input); err != nil {
			return nil, err
		}
		return json_input, nil
	}

	// Converts the input interface into map[string]interface{}
	return to.FromJSON(input)
}

func (sm *StateMachine) Execute(input interface{}) (*Execution, error) {
	if err := sm.Validate(); err != nil {
		return nil, err
	}

	input, err := processInput(input)
	if err != nil {
		return nil, err
	}

	// Start Execution (records the history, inputs, outputs...)
	exec := &Execution{}
	exec.Start()

	// Execute Start State
	output, err := sm.stateLoop(exec, sm.StartAt, input)

	// Set Final Output
	exec.SetOutput(output, err)

	if err != nil {
		exec.Failed()
	} else {
		exec.Succeeded()
	}

	return exec, err
}

func (sm *StateMachine) stateLoop(exec *Execution, next *string, input interface{}) (output interface{}, err error) {
	// Flat loop instead of recursion to better implement timeouts
	for {
		s, ok := sm.States[*next]

		if !ok {
			return nil, fmt.Errorf("Unknown State: %v", *next)
		}

		if len(exec.ExecutionHistory) > 25000 {
			return nil, fmt.Errorf("State Overflow")
		}

		exec.EnteredEvent(s, input)

		output, next, err = s.Execute(sm.DefaultLambdaContext(*s.Name()), input)

		if *s.GetType() != "Fail" {
			// Failure States Dont exit.
			exec.SetLastOutput(output, err)
			exec.ExitedEvent(s, output)
		}

		// If Error return error
		if err != nil {
			return output, err
		}

		// If next is nil then END
		if next == nil {
			return output, nil
		}

		input = output
	}
}
