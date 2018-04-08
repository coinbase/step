package state

import (
	"context"
	"fmt"
	"time"

	"github.com/coinbase/step/jsonpath"
	"github.com/coinbase/step/utils/to"
)

type WaitState struct {
	stateStr // Include Defaults

	Type    *string
	Comment *string `json:",omitempty"`

	InputPath  *jsonpath.Path `json:",omitempty"`
	OutputPath *jsonpath.Path `json:",omitempty"`

	Seconds       *float64       `json:",omitempty"`
	Timestamp     *time.Time     `json:",omitempty"`
	TimestampPath *jsonpath.Path `json:",omitempty"`

	Next *string `json:",omitempty"`
	End  *bool   `json:",omitempty"`
}

func (s *WaitState) process(ctx context.Context, input interface{}) (interface{}, *string, error) {

	if s.Seconds != nil {
		// Run at 100X speed
		time.Sleep(time.Duration(int64(*s.Seconds*10)) * time.Millisecond)
	} else {
		// Sleep for 50 milliseconds
		time.Sleep(50 * time.Millisecond)
	}
	return input, nextState(s.Next, s.End), nil
}

func (s *WaitState) Execute(ctx context.Context, input interface{}) (output interface{}, next *string, err error) {
	return processError(s,
		inputOutput(
			s.InputPath,
			s.OutputPath,
			s.process,
		),
	)(ctx, input)
}

func (s *WaitState) Validate() error {
	s.SetType(to.Strp("Wait"))

	if err := ValidateNameAndType(s); err != nil {
		return fmt.Errorf("%v %v", errorPrefix(s), err)
	}

	// Next xor End
	if err := endValid(s.Next, s.End); err != nil {
		return fmt.Errorf("%v %v", errorPrefix(s), err)
	}

	exactly_one := []bool{
		s.Seconds != nil,
		s.Timestamp != nil,
		s.TimestampPath != nil,
	}

	count := 0
	for _, c := range exactly_one {
		if c {
			count += 1
		}
	}

	if count != 1 {
		return fmt.Errorf("%v Exactly One (Seconds,TimeStamp,TimeStampPath)", errorPrefix(s))
	}

	return nil
}

func (s *WaitState) SetType(t *string) {
	s.Type = t
}

func (s *WaitState) GetType() *string {
	return s.Type
}
