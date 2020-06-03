package machine

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/coinbase/step/jsonpath"
	"github.com/coinbase/step/utils/to"
)

type ChoiceState struct {
	stateStr // Include Defaults

	Type    *string
	Comment *string `json:",omitempty"`

	InputPath  *jsonpath.Path `json:",omitempty"`
	OutputPath *jsonpath.Path `json:",omitempty"`

	Default *string `json:",omitempty"` // Default State if no choices match

	Choices []*Choice `json:",omitempty"`
}

type Choice struct {
	ChoiceRule

	Next *string `json:",omitempty"`
}

type ChoiceRule struct {
	Variable *jsonpath.Path `json:",omitempty"`

	StringEquals            *string `json:",omitempty"`
	StringLessThan          *string `json:",omitempty"`
	StringGreaterThan       *string `json:",omitempty"`
	StringLessThanEquals    *string `json:",omitempty"`
	StringGreaterThanEquals *string `json:",omitempty"`

	NumericEquals            *float64 `json:",omitempty"`
	NumericLessThan          *float64 `json:",omitempty"`
	NumericGreaterThan       *float64 `json:",omitempty"`
	NumericLessThanEquals    *float64 `json:",omitempty"`
	NumericGreaterThanEquals *float64 `json:",omitempty"`

	BooleanEquals *bool `json:",omitempty"`

	TimestampEquals            *time.Time `json:",omitempty"`
	TimestampLessThan          *time.Time `json:",omitempty"`
	TimestampGreaterThan       *time.Time `json:",omitempty"`
	TimestampLessThanEquals    *time.Time `json:",omitempty"`
	TimestampGreaterThanEquals *time.Time `json:",omitempty"`

	And []*ChoiceRule `json:",omitempty"`
	Or  []*ChoiceRule `json:",omitempty"`
	Not *ChoiceRule   `json:",omitempty"`
}

func (cr *ChoiceRule) String() string {
	if cr.And != nil {
		var strs []string
		for _, and := range cr.And {
			strs = append(strs, and.String())
		}
		return strings.Join(strs, " && ")
	}

	if cr.Or != nil {
		var strs []string
		for _, or := range cr.Or {
			strs = append(strs, or.String())
		}
		return strings.Join(strs, " || ")
	}

	if cr.Not != nil {
		return fmt.Sprintf("!%v", *cr.Variable)
	}

	op := ""

	if cr.StringEquals != nil {
		op = fmt.Sprintf("=%v", *cr.StringEquals)
	} else if cr.StringLessThan != nil {
		op = fmt.Sprintf("<%v", *cr.StringLessThan)
	} else if cr.StringGreaterThan != nil {
		op = fmt.Sprintf(">%v", *cr.StringGreaterThan)
	} else if cr.StringLessThanEquals != nil {
		op = fmt.Sprintf("<=%v", *cr.StringLessThanEquals)
	} else if cr.StringGreaterThanEquals != nil {
		op = fmt.Sprintf(">=%v", *cr.StringGreaterThanEquals)
	} else if cr.NumericEquals != nil {
		op = fmt.Sprintf("=%v", *cr.NumericEquals)
	} else if cr.NumericLessThan != nil {
		op = fmt.Sprintf("<%v", *cr.NumericLessThan)
	} else if cr.NumericGreaterThan != nil {
		op = fmt.Sprintf(">%v", *cr.NumericGreaterThan)
	} else if cr.NumericLessThanEquals != nil {
		op = fmt.Sprintf("<=%v", *cr.NumericLessThanEquals)
	} else if cr.NumericGreaterThanEquals != nil {
		op = fmt.Sprintf(">=%v", *cr.NumericGreaterThanEquals)
	} else if cr.BooleanEquals != nil {
		op = fmt.Sprintf("=%v", *cr.BooleanEquals)
	} else if cr.TimestampEquals != nil {
		op = fmt.Sprintf("=%v", *cr.TimestampEquals)
	} else if cr.TimestampLessThan != nil {
		op = fmt.Sprintf("<%v", *cr.TimestampLessThan)
	} else if cr.TimestampGreaterThan != nil {
		op = fmt.Sprintf(">%v", *cr.TimestampGreaterThan)
	} else if cr.TimestampLessThanEquals != nil {
		op = fmt.Sprintf("<=%v", *cr.TimestampLessThanEquals)
	} else if cr.TimestampGreaterThanEquals != nil {
		op = fmt.Sprintf(">=%v", *cr.TimestampGreaterThanEquals)
	}

	return fmt.Sprintf("%v%v", cr.Variable.String(), op)
}

func (s *ChoiceState) process(ctx context.Context, input interface{}) (interface{}, *string, error) {
	next := chooseNextState(input, s.Default, s.Choices)
	if next == nil {
		return nil, nil, fmt.Errorf("State Choice Error")
	}
	return input, next, nil
}

func (s *ChoiceState) Execute(ctx context.Context, input interface{}) (output interface{}, next *string, err error) {
	return processError(s,
		inputOutput(
			s.InputPath,
			s.OutputPath,
			s.process,
		),
	)(ctx, input)
}

func chooseNextState(input interface{}, default_choice *string, choices []*Choice) *string {
	for _, choice := range choices {
		if choiceRulePositive(input, &choice.ChoiceRule) {
			return choice.Next
		}
	}
	return default_choice
}

func choiceRulePositive(input interface{}, cr *ChoiceRule) bool {
	if cr.And != nil {
		for _, a := range cr.And {
			// if any choices have false then return false
			if !choiceRulePositive(input, a) {
				return false
			}
		}
		return true
	}

	if cr.Or != nil {
		for _, a := range cr.Or {
			// if any choices have true then return true
			if choiceRulePositive(input, a) {
				return true
			}
		}
		return false
	}

	if cr.Not != nil {
		return !choiceRulePositive(input, cr.Not)
	}

	if cr.StringEquals != nil {
		vstr, err := cr.Variable.GetString(input)
		if err != nil {
			return false // either not found or bad type
		}
		return *vstr == *cr.StringEquals
	}

	if cr.StringLessThan != nil {
		vstr, err := cr.Variable.GetString(input)
		if err != nil {
			return false // either not found or bad type
		}
		return *vstr < *cr.StringLessThan
	}

	if cr.StringGreaterThan != nil {
		vstr, err := cr.Variable.GetString(input)
		if err != nil {
			return false // either not found or bad type
		}
		return *vstr > *cr.StringGreaterThan
	}

	if cr.StringLessThanEquals != nil {
		vstr, err := cr.Variable.GetString(input)
		if err != nil {
			return false // either not found or bad type
		}
		return *vstr <= *cr.StringLessThanEquals
	}

	if cr.StringGreaterThanEquals != nil {
		vstr, err := cr.Variable.GetString(input)
		if err != nil {
			return false // either not found or bad type
		}
		return *vstr >= *cr.StringGreaterThanEquals
	}

	// NUMBERs
	if cr.NumericEquals != nil {
		vnum, err := cr.Variable.GetNumber(input)
		if err != nil {
			return false
		}
		return *vnum == *cr.NumericEquals
	}

	if cr.NumericLessThan != nil {
		vnum, err := cr.Variable.GetNumber(input)
		if err != nil {
			return false
		}
		return *vnum < *cr.NumericLessThan
	}

	if cr.NumericGreaterThan != nil {
		vnum, err := cr.Variable.GetNumber(input)
		if err != nil {
			return false
		}
		return *vnum > *cr.NumericGreaterThan
	}

	if cr.NumericLessThanEquals != nil {
		vnum, err := cr.Variable.GetNumber(input)
		if err != nil {
			return false
		}
		return *vnum <= *cr.NumericLessThanEquals
	}

	if cr.NumericGreaterThanEquals != nil {
		vnum, err := cr.Variable.GetNumber(input)
		if err != nil {
			return false
		}
		return *vnum >= *cr.NumericGreaterThanEquals
	}

	if cr.BooleanEquals != nil {
		vbool, err := cr.Variable.GetBool(input)
		if err != nil {
			return false
		}
		return *vbool == *cr.BooleanEquals
	}

	if cr.TimestampEquals != nil {
		vtime, err := cr.Variable.GetTime(input)
		if err != nil {
			return false
		}
		return *vtime == *cr.TimestampEquals
	}

	if cr.TimestampLessThan != nil {
		vtime, err := cr.Variable.GetTime(input)
		if err != nil {
			return false
		}
		return vtime.Before(*cr.TimestampLessThan)
	}

	if cr.TimestampGreaterThan != nil {
		vtime, err := cr.Variable.GetTime(input)
		if err != nil {
			return false
		}
		return vtime.After(*cr.TimestampGreaterThan)
	}

	if cr.TimestampLessThanEquals != nil {
		vtime, err := cr.Variable.GetTime(input)
		if err != nil {
			return false
		}
		return *vtime == *cr.TimestampLessThanEquals || vtime.Before(*cr.TimestampLessThanEquals)
	}

	if cr.TimestampGreaterThanEquals != nil {
		vtime, err := cr.Variable.GetTime(input)
		if err != nil {
			return false
		}
		return *vtime == *cr.TimestampGreaterThanEquals || vtime.After(*cr.TimestampGreaterThanEquals)
	}

	return false
}

// VALIDATION LOGIC

func (s *ChoiceState) Validate() error {
	s.SetType(to.Strp("Choice"))

	if err := ValidateNameAndType(s); err != nil {
		return fmt.Errorf("%v %v", errorPrefix(s), err)
	}

	if len(s.Choices) == 0 {
		return fmt.Errorf("%v Must have Choices", errorPrefix(s))
	}

	for _, c := range s.Choices {
		err := validateChoice(c)
		if err != nil {
			return fmt.Errorf("%v %v", errorPrefix(s), err)
		}
	}

	return nil
}

func validateChoice(c *Choice) error {

	if c.Next == nil {
		return fmt.Errorf("Choice must have Next")
	}

	all_choice_rules := recursiveAllChoiceRule(&c.ChoiceRule)

	for _, cr := range all_choice_rules {
		if err := validateChoiceRule(cr); err != nil {
			return err
		}
	}

	return nil
}

func recursiveAllChoiceRule(c *ChoiceRule) []*ChoiceRule {
	if c == nil {
		return []*ChoiceRule{}
	}

	crs := []*ChoiceRule{c}

	if c.Not != nil {
		crs = append(crs, c.Not)
	}

	if c.And != nil {
		for _, cr := range c.And {
			crs = append(crs, recursiveAllChoiceRule(cr)...)
		}
	}

	if c.Or != nil {
		for _, cr := range c.Or {
			crs = append(crs, recursiveAllChoiceRule(cr)...)
		}
	}

	return crs
}

func validateChoiceRule(c *ChoiceRule) error {
	// Exactly One Comparison Operator
	all_comparison_operators := []bool{
		c.Not != nil,
		c.And != nil,
		c.Or != nil,
		c.StringEquals != nil,
		c.StringLessThan != nil,
		c.StringGreaterThan != nil,
		c.StringLessThanEquals != nil,
		c.StringGreaterThanEquals != nil,
		c.NumericEquals != nil,
		c.NumericLessThan != nil,
		c.NumericGreaterThan != nil,
		c.NumericLessThanEquals != nil,
		c.NumericGreaterThanEquals != nil,
		c.BooleanEquals != nil,
		c.TimestampEquals != nil,
		c.TimestampLessThan != nil,
		c.TimestampGreaterThan != nil,
		c.TimestampLessThanEquals != nil,
		c.TimestampGreaterThanEquals != nil,
	}

	count := 0
	for _, co := range all_comparison_operators {
		if co {
			count += 1
		}
	}

	if count != 1 {
		return fmt.Errorf("Not Exactly One comparison Operator")
	}

	// Variable must be defined, UNLESS AND NOT OR, in which case error if defined
	not_and_or := c.Not != nil || c.And != nil || c.Or != nil

	if not_and_or {
		if c.Variable != nil {
			return fmt.Errorf("Variable defined with Not And Or defined")
		}
	} else {
		if c.Variable == nil {
			return fmt.Errorf("Variable Not defined")
		}
	}

	if c.And != nil && len(c.And) == 0 {
		return fmt.Errorf("And Must have elements")
	}

	if c.Or != nil && len(c.Or) == 0 {
		return fmt.Errorf("Or Must have elements")
	}

	return nil
}

func (s *ChoiceState) SetType(t *string) {
	s.Type = t
}

func (s *ChoiceState) GetType() *string {
	return s.Type
}
