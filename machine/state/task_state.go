package state

import (
	"context"
	"fmt"

	"github.com/coinbase/step/handler"
	"github.com/coinbase/step/jsonpath"
	"github.com/coinbase/step/utils/to"
)

type TaskState struct {
	stateStr // Include Defaults

	Type    *string
	Comment *string `json:",omitempty"`

	InputPath  *jsonpath.Path `json:",omitempty"`
	OutputPath *jsonpath.Path `json:",omitempty"`
	ResultPath *jsonpath.Path `json:",omitempty"`
	Parameters interface{}    `json:",omitempty"`

	Resource *string `json:",omitempty"`

	Catch []*Catcher `json:",omitempty"`
	Retry []*Retrier `json:",omitempty"`

	// Maps a Lambda Handler Function
	TaskHandler interface{} `json:"-"`

	Next *string `json:",omitempty"`
	End  *bool   `json:",omitempty"`
}

func (s *TaskState) SetTaskHandler(reasourcefn interface{}) {
	s.TaskHandler = reasourcefn
}

func (s *TaskState) process(ctx context.Context, input interface{}) (interface{}, *string, error) {
	result, err := handler.CallHandlerFunction(s.TaskHandler, ctx, input)

	if err != nil {
		return nil, nil, err
	}

	result, err = to.FromJSON(result)

	if err != nil {
		return nil, nil, err
	}

	return result, nextState(s.Next, s.End), nil
}

// Input must include the Task name in $.Task
func (s *TaskState) Execute(ctx context.Context, input interface{}) (output interface{}, next *string, err error) {
	return processError(s,
		processCatcher(s.Catch,
			processRetrier(s.Name(), s.Retry,
				inputOutput(
					s.InputPath,
					s.OutputPath,
					withParams(
						s.Parameters,
						result(s.ResultPath, s.process),
					),
				),
			),
		),
	)(ctx, input)
}

func (s *TaskState) Validate() error {
	s.SetType(to.Strp("Task"))

	if err := ValidateNameAndType(s); err != nil {
		return fmt.Errorf("%v %v", errorPrefix(s), err)
	}

	if err := endValid(s.Next, s.End); err != nil {
		return fmt.Errorf("%v %v", errorPrefix(s), err)
	}

	if s.TaskHandler == nil && s.Resource == nil {
		return fmt.Errorf("%v Requires Resource", errorPrefix(s))
	}

	if s.TaskHandler != nil {
		if err := handler.ValidateHandler(s.TaskHandler); err != nil {
			return err
		}
	}

	if err := catchValid(s.Catch); err != nil {
		return err
	}

	if err := retryValid(s.Retry); err != nil {
		return err
	}

	return nil
}

func (s *TaskState) SetType(t *string) {
	s.Type = t
}

func (s *TaskState) GetType() *string {
	return s.Type
}
