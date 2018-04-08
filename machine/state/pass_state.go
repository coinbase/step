package state

import (
	"context"
	"fmt"

	"github.com/coinbase/step/jsonpath"
	"github.com/coinbase/step/utils/to"
)

type PassState struct {
	stateStr // Include Defaults

	Type    *string
	Comment *string `json:",omitempty"`

	InputPath  *jsonpath.Path `json:",omitempty"`
	OutputPath *jsonpath.Path `json:",omitempty"`
	ResultPath *jsonpath.Path `json:",omitempty"`

	Result interface{} `json:",omitempty"`

	Next *string `json:",omitempty"`
	End  *bool   `json:",omitempty"`
}

func (s *PassState) Execute(ctx context.Context, input interface{}) (output interface{}, next *string, err error) {
	return processError(s,
		inputOutput(
			s.InputPath,
			s.OutputPath,
			result(s.ResultPath, s.process),
		),
	)(ctx, input)
}

func (s *PassState) process(ctx context.Context, input interface{}) (output interface{}, next *string, err error) {
	return s.Result, nextState(s.Next, s.End), nil
}

func (s *PassState) Validate() error {
	s.SetType(to.Strp("Pass"))

	if err := ValidateNameAndType(s); err != nil {
		return fmt.Errorf("%v %v", errorPrefix(s), err)
	}

	// Next xor End
	if err := endValid(s.Next, s.End); err != nil {
		return fmt.Errorf("%v %v", errorPrefix(s), err)
	}

	return nil
}

func (s *PassState) SetType(t *string) {
	s.Type = t
}

func (s *PassState) GetType() *string {
	return s.Type
}
