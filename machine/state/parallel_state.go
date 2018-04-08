package state

import (
	"context"
	"fmt"

	"github.com/coinbase/step/utils/to"
)

type ParallelState struct {
	stateStr // Include Defaults

	Type    *string
	Comment *string `json:",omitempty"`
}

func (s *ParallelState) Execute(_ context.Context, input interface{}) (output interface{}, next *string, err error) {
	return input, nil, nil
}

func (s *ParallelState) Validate() error {
	s.SetType(to.Strp("Parallel"))

	if err := ValidateNameAndType(s); err != nil {
		return fmt.Errorf("%v %v", errorPrefix(s), err)
	}

	return nil
}

func (s *ParallelState) SetType(t *string) {
	s.Type = t
}

func (s *ParallelState) GetType() *string {
	return s.Type
}
