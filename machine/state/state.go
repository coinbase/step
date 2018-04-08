package state

import (
	"context"
	"fmt"

	"github.com/coinbase/step/jsonpath"
	"github.com/coinbase/step/utils/is"
	"github.com/coinbase/step/utils/to"
)

// TYPES

type Execution func(context.Context, interface{}) (interface{}, *string, error)

type State interface {
	Execute(context.Context, interface{}) (interface{}, *string, error)
	Validate() error

	SetName(*string)
	SetType(*string)

	Name() *string
	GetType() *string
}

type stateStr struct {
	name *string `json:"-"`
}

type Catcher struct {
	ErrorEquals []*string      `json:",omitempty"`
	ResultPath  *jsonpath.Path `json:",omitempty"`
	Next        *string        `json:",omitempty"`
}

type Retrier struct {
	ErrorEquals     []*string `json:",omitempty"`
	IntervalSeconds *int      `json:",omitempty"`
	MaxAttempts     *int      `json:",omitempty"`
	BackoffRate     *float64  `json:",omitempty"`
	attempts        int       `json:"-"` // Used to remember attempts
}

func errorOutputFromError(err error) map[string]interface{} {
	return errorOutput(to.Strp(to.ErrorType(err)), to.Strp(err.Error()))
}

func errorOutput(err *string, cause *string) map[string]interface{} {
	errstr := ""
	causestr := ""
	if err != nil {
		errstr = *err
	}
	if cause != nil {
		causestr = *cause
	}
	return map[string]interface{}{
		"Error": errstr,
		"Cause": causestr,
	}
}

func errorIncluded(errorEquals []*string, err error) bool {
	error_type := to.ErrorType(err)

	for _, et := range errorEquals {
		if *et == error_type {
			return true
		}
	}
	return false
}

// Default State Methods

func (s *stateStr) Name() *string {
	return s.name
}

func (s *stateStr) SetName(name *string) {
	s.name = name
}

func nextState(next *string, end *bool) *string {
	if next != nil {
		return next
	}
	// If End is true then it should be nil
	// If End is false then Next must be defined so invalid
	// If End is nil then Next must be defined so invalid
	return nil
}

//////
// Shared Methods
//////

func processRetrier(retryName *string, retriers []*Retrier, exec Execution) Execution {
	return func(ctx context.Context, input interface{}) (interface{}, *string, error) {
		// Simulate Retry once, not actually waiting
		output, next, err := exec(ctx, input)
		if len(retriers) == 0 || err == nil {
			return output, next, err
		}

		// Is Error in a Retrier
		for _, retrier := range retriers {
			// If the error type is defined in the retrier AND we have not attempted the retry yet
			if retrier.MaxAttempts == nil {
				// Default retries is 3
				retrier.MaxAttempts = to.Intp(3)
			}

			if errorIncluded(retrier.ErrorEquals, err) && retrier.attempts < *retrier.MaxAttempts {
				retrier.attempts++
				// Returns the name of the state to the state-machine to re-execute
				return input, retryName, nil
			}
		}

		// Otherwise, just return
		return output, next, err
	}
}

func processCatcher(catchers []*Catcher, exec Execution) Execution {
	return func(ctx context.Context, input interface{}) (interface{}, *string, error) {
		output, next, err := exec(ctx, input)

		if len(catchers) == 0 || err == nil {
			return output, next, err
		}

		for _, catcher := range catchers {
			if errorIncluded(catcher.ErrorEquals, err) {

				eo := errorOutputFromError(err)
				output, err := catcher.ResultPath.Set(input, eo)

				return output, catcher.Next, err
			}
		}

		// Otherwise continue
		return output, next, err
	}
}

func processError(s State, exec Execution) Execution {
	return func(ctx context.Context, input interface{}) (interface{}, *string, error) {
		output, next, err := exec(ctx, input)

		if err != nil {
			return nil, nil, fmt.Errorf("%v %v", errorPrefix(s), err.Error())
		}
		return output, next, nil
	}
}
func inputOutput(inputPath *jsonpath.Path, outputPath *jsonpath.Path, exec Execution) Execution {
	return func(ctx context.Context, input interface{}) (interface{}, *string, error) {
		input, err := inputPath.Get(input)

		if err != nil {
			return nil, nil, fmt.Errorf("Input Error: %v", err)
		}

		output, next, err := exec(ctx, input)

		if err != nil {
			return nil, nil, err
		}

		output, err = outputPath.Get(output)

		if err != nil {
			return nil, nil, fmt.Errorf("Output Error: %v", err)
		}

		return output, next, nil
	}
}

func result(resultPath *jsonpath.Path, exec Execution) Execution {
	return func(ctx context.Context, input interface{}) (interface{}, *string, error) {
		result, next, err := exec(ctx, input)

		if err != nil {
			return nil, nil, err
		}

		if result != nil {
			input, err := resultPath.Set(input, result)

			if err != nil {
				return nil, nil, err
			}

			return input, next, nil
		}

		return input, next, nil
	}
}

//////
// Shared Validity Methods
//////

func endValid(next *string, end *bool) error {
	if end == nil && next == nil {
		return fmt.Errorf("End and Next both undefined")
	}

	if end != nil && next != nil {
		return fmt.Errorf("End and Next both defined")
	}

	if end != nil && *end == false {
		return fmt.Errorf("End can only be true or nil")
	}

	return nil
}

func errorPrefix(s State) string {
	if !is.EmptyStr(s.Name()) {
		return fmt.Sprintf("%vState(%v) Error:", *s.GetType(), *s.Name())
	} else {
		return fmt.Sprintf("%vState Error:", *s.GetType())
	}
}

func ValidateNameAndType(s State) error {
	if is.EmptyStr(s.Name()) {
		return fmt.Errorf("Must have Name")
	}

	if is.EmptyStr(s.GetType()) {
		return fmt.Errorf("Must have Type")
	}

	return nil
}

func retrierValid(r *Retrier) error {
	if r.ErrorEquals == nil || len(r.ErrorEquals) == 0 {
		return fmt.Errorf("Retrier requires ErrorEquals")
	}

	return nil
}

func catcherValid(c *Catcher) error {
	if c.ErrorEquals == nil || len(c.ErrorEquals) == 0 {
		return fmt.Errorf("Catcher requires ErrorEquals")
	}

	if is.EmptyStr(c.Next) {
		return fmt.Errorf("Catcher requires Next")
	}
	return nil
}
