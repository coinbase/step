package state

import (
	"context"
	"fmt"

	"github.com/coinbase/step/jsonpath"
	"github.com/coinbase/step/utils/to"
)

type SucceedState struct {
	stateStr // Include Defaults

	Type    *string
	Comment *string `json:",omitempty"`

	InputPath  *jsonpath.Path `json:",omitempty"`
	OutputPath *jsonpath.Path `json:",omitempty"`
}

func (s *SucceedState) process(ctx context.Context, input interface{}) (interface{}, *string, error) {
	return input, nil, nil
}

func (s *SucceedState) Execute(ctx context.Context, input interface{}) (output interface{}, next *string, err error) {
	return processError(s,
		inputOutput(
			s.InputPath,
			s.OutputPath,
			s.process,
		),
	)(ctx, input)
}

func (s *SucceedState) Validate() error {
	s.SetType(to.Strp("Succeed"))

	if err := ValidateNameAndType(s); err != nil {
		return fmt.Errorf("%v %v", errorPrefix(s), err)
	}

	return nil
}

func (s *SucceedState) SetType(t *string) {
	s.Type = t
}

func (s *SucceedState) GetType() *string {
	return s.Type
}
