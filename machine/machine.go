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

type StateMachine struct {
	Comment *string `json:",omitempty"`

	StartAt *string

	States map[string]state.State

	executionHistory []HistoryEvent
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

// Struct Methods
func (sm *StateMachine) ExecuteJSON(input *string) (*string, error) {
	var json_input map[string]interface{}
	if err := json.Unmarshal([]byte(*input), &json_input); err != nil {
		return nil, err
	}

	output, err := sm.Execute(json_input)

	if err != nil {
		return nil, err
	}

	output_json, err := to.PrettyJSON(output)

	if err != nil {
		return nil, err
	}
	str := string(output_json)
	return &str, nil
}

func (sm *StateMachine) ExecuteToMap(input interface{}) (map[string]interface{}, error) {
	output, err := sm.Execute(input)
	switch output.(type) {
	case map[string]interface{}:
		return output.(map[string]interface{}), err
	}
	return nil, err
}

func (sm *StateMachine) FindTask(name string) (*state.TaskState, error) {
	tasks := sm.Tasks()
	task, ok := tasks[TaskFnName(name)]
	if !ok {
		task, ok = sm.Tasks()[name]
		if !ok {
			return nil, fmt.Errorf("Handler Error: Cannot Find Task %v or %v", name, TaskFnName(name))
		}
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

func (sm *StateMachine) TaskFunctions() *handler.TaskFunctions {
	tm := handler.TaskFunctions{}
	for name, task := range sm.Tasks() {
		tm[name] = task.ResourceFunction
	}
	return &tm
}

func (sm *StateMachine) AddState(s state.State) error {
	_, ok := sm.States[*s.Name()]

	if ok {
		return fmt.Errorf("Already added the state")
	}

	sm.States[*s.Name()] = s

	return nil
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
		task.SetResourceFunction(DefaultHandler)
	}
}

func (sm *StateMachine) SetResourceFunction(task_name string, resource_fn interface{}) error {
	task, err := sm.FindTask(task_name)
	if err != nil {
		return err
	}

	task.SetResourceFunction(resource_fn)
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

	// TODO: all Task States must have PassState assigning its Name
	return nil
}

func validateInput(s state.State, input interface{}) error {
	if *s.GetType() != "Task" {
		return nil
	}

	switch input.(type) {
	case map[string]interface{}:
		m := input.(map[string]interface{})
		task, ok := m["Task"]
		if !ok {
			return &handler.TaskError{fmt.Sprintf("$.Task input is nil"), s.Name(), nil}
		}

		task_name := task.(string)
		taskfn_name := TaskFnName(task_name)

		if task_name == *s.Name() || taskfn_name == *s.Name() {
			return nil
		}
		fmt.Println(input)
		return &handler.TaskError{fmt.Sprintf("$.Task input doesn't equal %v or %v", task_name, taskfn_name), s.Name(), nil}
	}
	return &handler.TaskError{"Input wrong type", s.Name(), nil}
}

func (sm *StateMachine) DefaultLambdaContext(lambda_name string) context.Context {
	return lambdacontext.NewContext(context.Background(), &lambdacontext.LambdaContext{
		InvokedFunctionArn: fmt.Sprintf("arn:aws:lambda:us-east-1:000000000:function:%v", lambda_name),
	})
}

func (sm *StateMachine) Execute(input interface{}) (interface{}, error) {
	sm.executionHistory = startExecution()
	// Start Execution

	err := sm.Validate()

	if err != nil {
		return input, err
	}

	sanitizedInput, err := to.FromJSON(input)

	if err != nil {
		return input, err
	}

	// Execute Start State
	output, err := sm.stateLoop(sm.StartAt, sanitizedInput)
	if err != nil {
		sm.executionHistory = append(sm.executionHistory, createEvent("ExecutionFailed"))
	} else {
		sm.executionHistory = append(sm.executionHistory, createEvent("ExecutionSucceeded"))
	}

	return output, err
}

func (sm *StateMachine) stateLoop(next *string, input interface{}) (output interface{}, err error) {
	// Flat loop instead of recursion to better implement timeouts
	for {
		s, ok := sm.States[*next]

		if !ok {
			return output, fmt.Errorf("Unknown State: %v", *next)
		}

		if err := validateInput(s, input); err != nil {
			return output, err
		}

		sm.executionHistory = append(sm.executionHistory, createEnteredEvent(s, input))

		output, next, err = s.Execute(sm.DefaultLambdaContext(*s.Name()), input)

		if *s.GetType() != "Fail" {
			// Failure States Dont exit.
			sm.executionHistory = append(sm.executionHistory, createExitedEvent(s, output))
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
