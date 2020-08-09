package machine

import (
	"context"
	"fmt"
	"strings"

	"github.com/coinbase/step/jsonpath"
	"github.com/coinbase/step/utils/is"
	"github.com/coinbase/step/utils/to"
)

// TYPES

type ExecutionFn func(context.Context, interface{}) (interface{}, *string, error)

type State interface {
	Execute(context.Context, interface{}) (interface{}, *string, error)
	Validate() error

	SetName(*string)
	SetType(*string)

	Name() *string
	GetType() *string
	GetNext() *string
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
		if *et == "States.ALL" || *et == error_type {
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

func (s *stateStr) GetNext() *string {
	return nil
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

func processRetrier(retryName *string, retriers []*Retrier, exec ExecutionFn) ExecutionFn {
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

			// Match on first retrier
			if errorIncluded(retrier.ErrorEquals, err) {
				if retrier.attempts < *retrier.MaxAttempts {
					retrier.attempts++
					// Returns the name of the state to the state-machine to re-execute
					return input, retryName, nil
				} else {
					// Finished retrying so continue
					return output, next, err
				}
			}
		}

		// Otherwise, just return
		return output, next, err
	}
}

func processCatcher(catchers []*Catcher, exec ExecutionFn) ExecutionFn {
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

func processError(s State, exec ExecutionFn) ExecutionFn {
	return func(ctx context.Context, input interface{}) (interface{}, *string, error) {
		output, next, err := exec(ctx, input)

		if err != nil {
			return nil, nil, fmt.Errorf("%v %v", errorPrefix(s), err.Error())
		}
		return output, next, nil
	}
}
func inputOutput(inputPath *jsonpath.Path, outputPath *jsonpath.Path, exec ExecutionFn) ExecutionFn {
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

func withParams(params interface{}, exec ExecutionFn) ExecutionFn {
	return func(ctx context.Context, input interface{}) (interface{}, *string, error) {
		if params == nil {
			return exec(ctx, input)
		}
		// Loop through the input replace values with JSON paths
		input, err := replaceParamsJSONPath(params, input)
		if err != nil {
			return nil, nil, err
		}

		return exec(ctx, input)
	}
}

func replaceParamsJSONPath(params interface{}, input interface{}) (interface{}, error) {

	switch params.(type) {
	case map[string]interface{}:

		newParams := map[string]interface{}{}
		// Recurse over params find keys to replace
		for key, value := range params.(map[string]interface{}) {
			if strings.HasSuffix(key, ".$") {
				key = key[:len(key)-len(".$")]
				// value must be a JSON path string!
				switch value.(type) {
				case string:
				default:
					return nil, fmt.Errorf("value to key %q is not string", key)
				}
				valueStr := value.(string)
				path, err := jsonpath.NewPath(valueStr)
				if err != nil {
					return nil, err
				}
				newValue, err := path.Get(input)
				if err != nil {
					return nil, err
				}
				newParams[key] = newValue
			} else {
				newValue, err := replaceParamsJSONPath(value, input)
				if err != nil {
					return nil, err
				}
				newParams[key] = newValue
			}
		}
		return newParams, nil
	}
	return params, nil
}

func result(resultPath *jsonpath.Path, exec ExecutionFn) ExecutionFn {
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
	}

	return fmt.Sprintf("%vState Error:", *s.GetType())
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

func retryValid(retry []*Retrier) error {
	if retry == nil {
		return nil
	}

	for i, r := range retry {
		if err := errorEqualsValid(r.ErrorEquals, len(retry)-1 == i); err != nil {
			return err
		}
	}

	return nil
}

func catchValid(catch []*Catcher) error {
	if catch == nil {
		return nil
	}

	for i, c := range catch {
		if err := errorEqualsValid(c.ErrorEquals, len(catch)-1 == i); err != nil {
			return err
		}

		if is.EmptyStr(c.Next) {
			return fmt.Errorf("Catcher requires Next")
		}
	}
	return nil
}

func errorEqualsValid(errorEquals []*string, last bool) error {
	if errorEquals == nil || len(errorEquals) == 0 {
		return fmt.Errorf("Retrier requires ErrorEquals")
	}

	for _, e := range errorEquals {
		// If it is a States. Error, then must match one of the defined values
		if strings.HasPrefix(*e, "States.") {
			switch *e {
			case
				"States.ALL",
				"States.Timeout",
				"States.TaskFailed",
				"States.Permissions",
				"States.ResultPathMatchFailure",
				"States.BranchFailed",
				"States.NoChoiceMatched":
			default:
				return fmt.Errorf("Unknown States.* error found %q", *e)
			}
		}

		if *e == "States.ALL" {
			if len(errorEquals) != 1 {
				return fmt.Errorf(`"States.ALL" ErrorEquals must be only element in list`)
			}

			if !last {
				return fmt.Errorf(`"States.ALL" must be last Catcher/Retrier`)
			}
		}
	}

	return nil
}
