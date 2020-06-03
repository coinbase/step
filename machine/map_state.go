package machine

import (
	"context"
	"fmt"
	"github.com/coinbase/step/jsonpath"
	"github.com/coinbase/step/utils/to"
)

type MapState struct {
	stateStr // Include Defaults

	Type    *string
	Comment *string `json:",omitempty"`

	Iterator   *StateMachine
	ItemsPath  *jsonpath.Path `json:",omitempty"`
	Parameters interface{}    `json:",omitempty"`

	MaxConcurrency *float64 `json:",omitempty"`

	InputPath  *jsonpath.Path `json:",omitempty"`
	OutputPath *jsonpath.Path `json:",omitempty"`
	ResultPath *jsonpath.Path `json:",omitempty"`

	Catch []*Catcher `json:",omitempty"`
	Retry []*Retrier `json:",omitempty"`

	Next *string `json:",omitempty"`
	End  *bool   `json:",omitempty"`
}

func (s *MapState) process(ctx context.Context, input interface{}) (interface{}, *string, error) {
	output, err := s.ItemsPath.GetSlice(input)
	if err != nil {
		return input, nextState(s.Next, s.End), err
	}
	var res []map[string]interface{}

	for _, item := range output {
		execution, err := s.Iterator.Execute(item)
		if err != nil {
			return input, nextState(s.Next, s.End), err
		}
		res = append(res, execution.Output)
	}

	return res, nextState(s.Next, s.End), nil
}

func (s *MapState) Execute(ctx context.Context, input interface{}) (output interface{}, next *string, err error) {
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

func (s *MapState) Validate() error {
	s.SetType(to.Strp("Map"))

	if err := ValidateNameAndType(s); err != nil {
		return fmt.Errorf("%v %v", errorPrefix(s), err)
	}

	if err := endValid(s.Next, s.End); err != nil {
		return fmt.Errorf("%v %v", errorPrefix(s), err)
	}

	if s.Iterator == nil {
		return fmt.Errorf("%v Requires Iterator", errorPrefix(s))
	}

	if err := s.Iterator.Validate(); err != nil {
		return fmt.Errorf("%v %v", errorPrefix(s), err)
	}
	return nil
}

func (s *MapState) SetType(t *string) {
	s.Type = t
}

func (s *MapState) GetType() *string {
	return s.Type
}
