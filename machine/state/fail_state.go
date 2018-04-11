package state

import (
	"context"
	"fmt"

	"github.com/coinbase/step/utils/is"
	"github.com/coinbase/step/utils/to"
)

type FailState struct {
	stateStr // Include Defaults

	Type    *string
	Comment *string `json:",omitempty"`

	Error *string `json:",omitempty"`
	Cause *string `json:",omitempty"`
}

func (s *FailState) Execute(_ context.Context, input interface{}) (output interface{}, next *string, err error) {
	return errorOutput(s.Error, s.Cause), nil, fmt.Errorf("Fail")
}

func (s *FailState) Validate() error {
	s.SetType(to.Strp("Fail"))

	if err := ValidateNameAndType(s); err != nil {
		return fmt.Errorf("%v %v", errorPrefix(s), err)
	}

	if is.EmptyStr(s.Error) {
		return fmt.Errorf("%v %v", errorPrefix(s), "must contain Error")
	}

	return nil
}

func (s *FailState) SetType(t *string) {
	s.Type = t
}

func (s *FailState) GetType() *string {
	return s.Type
}
