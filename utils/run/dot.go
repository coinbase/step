package run

import (
	"fmt"
	"os"
	"strings"

	"github.com/coinbase/step/utils/to"

	"github.com/coinbase/step/machine/state"

	"github.com/coinbase/step/machine"
)

// Output Dot Format For State Machine

// JSON prints a state machine as JSON
func Dot(stateMachine *machine.StateMachine, err error) {
	if err != nil {
		fmt.Println("ERROR", err)
		os.Exit(1)
	}

	dotStr := toDot(stateMachine)
	fmt.Println(string(dotStr))
	os.Exit(0)
}

func toDot(stateMachine *machine.StateMachine) string {
	return fmt.Sprintf(`digraph StateMachine {
	node      [ style="rounded,filled,bold", shape=box, fixedsize=true, width=2, fontname="Arial" ];
	edge      [ style=bold, fontname="Arial", ]
  _Start    [ fillcolor=black, shape=circle, label="", width=0.25 ];
  _End      [ fillcolor=black, shape=doublecircle, label="", width=0.3 ];

  _Start -> %v [weight=1000]

  # States
  %v
}`, *stateMachine.StartAt, processStates(stateMachine.States))
}

func processStates(states map[string]state.State) string {
	str := ""
	for _, s := range states {
		str += processState(s)
	}
	return str
}

func processState(s state.State) string {
	strs := []string{}
	name := *s.Name()
	switch s.(type) {
	case *state.PassState:
		sstate := s.(*state.PassState)
		if *sstate.Next == machine.TaskFnName(name) {
			return "" // Skip if TaskFn Pass
		}
		strs = append(strs, fmt.Sprintf(`%q [fillcolor="#b0b0b0"]`, name))
		if sstate.Next != nil {
			strs = append(strs, fmt.Sprintf(`%q -> %q [weight=100]`, name, *sstate.Next))
		}
		if sstate.End != nil {
			strs = append(strs, fmt.Sprintf(`%q -> _End`, name))
		}

	case *state.TaskState:
		sstate := s.(*state.TaskState)
		name = machine.RemoveTaskFnName(name)
		strs = append(strs, fmt.Sprintf(`%q [fillcolor="#b0b0b0"]`, name))

		if sstate.Retry != nil {
			strs = append(strs, fmt.Sprintf(`%q -> %q [color="#FFa0a0"]`, name, name))
		}

		if sstate.Catch != nil {
			for _, c := range sstate.Catch {
				cname := fmt.Sprintf("%q", strings.Join(to.StrSlice(c.ErrorEquals), ","))
				if len(c.ErrorEquals) == 1 && *c.ErrorEquals[0] == "States.ALL" {
					cname = ""
				}
				strs = append(strs, fmt.Sprintf(`%q -> %q [color="#FFa0a0", label=%q]`, name, *c.Next, cname))
			}
		}

		if sstate.Next != nil {
			strs = append(strs, fmt.Sprintf(`%q -> %q [weight=100]`, name, *sstate.Next))
		}

		if sstate.End != nil {
			strs = append(strs, fmt.Sprintf(`%q -> _End`, name))
		}

	case *state.ChoiceState:
		sstate := s.(*state.ChoiceState)
		strs = append(strs, fmt.Sprintf(`%q [shape=diamond, fillcolor="#b0b0b0"]`, name))

		if sstate.Default != nil {
			strs = append(strs, fmt.Sprintf(`%q -> %q [label="default", weight=10]`, name, *sstate.Default))
		}

		if sstate.Choices != nil {
			for _, c := range sstate.Choices {
				strs = append(strs, fmt.Sprintf(`%q -> %q [label=%q, weight=100]`, name, *c.Next, choiceStr(c.ChoiceRule)))
			}
		}

	case *state.WaitState:
		sstate := s.(*state.WaitState)

		wait_til := ""
		if sstate.Seconds != nil {
			wait_til = fmt.Sprintf("%vs", *sstate.Seconds)
		}
		if sstate.Timestamp != nil {
			wait_til = fmt.Sprintf("%v", *sstate.Timestamp)
		}
		if sstate.TimestampPath != nil {
			wait_til = fmt.Sprintf("%v", *sstate.TimestampPath)
		}

		strs = append(strs, fmt.Sprintf(`%q [width=0.5, shape=doublecircle, fillcolor="#b0b0b0", label="Wait"]`, name))

		if sstate.Next != nil {
			strs = append(strs, fmt.Sprintf(`%q -> %q [weight=100, label=%q]`, name, *sstate.Next, wait_til))
		}

	case *state.FailState:
		strs = append(strs, fmt.Sprintf(`%q [fillcolor="#ffa0a0"]`, name))
		strs = append(strs, fmt.Sprintf(`%q -> _End [weight=1000]`, name))
	case *state.SucceedState:
		strs = append(strs, fmt.Sprintf(`%q [fillcolor="#a0ffa0"]`, name))
		strs = append(strs, fmt.Sprintf(`%q -> _End [weight=1000]`, name))

	}
	strs = append(strs, "\n  ")
	return strings.Join(strs, "\n  ")
}

func choiceStr(cr state.ChoiceRule) string {
	if cr.And != nil {
		strs := []string{}
		for _, a := range cr.And {
			strs = append(strs, choiceStr(*a))
		}
		return strings.Join(strs, " && ")
	}

	if cr.Or != nil {
		strs := []string{}
		for _, a := range cr.And {
			strs = append(strs, choiceStr(*a))
		}
		return strings.Join(strs, " || ")
	}

	if cr.Not != nil {
		return fmt.Sprintf("!%v", *cr.Variable)
	}

	op := ""

	if cr.StringEquals != nil {
		op = fmt.Sprintf("=%v", *cr.StringEquals)
	}

	if cr.StringLessThan != nil {
		op = fmt.Sprintf("<%v", *cr.StringLessThan)
	}

	if cr.StringGreaterThan != nil {
		op = fmt.Sprintf(">%v", *cr.StringGreaterThan)
	}

	if cr.StringLessThanEquals != nil {
		op = fmt.Sprintf("<=%v", *cr.StringLessThanEquals)
	}

	if cr.StringGreaterThanEquals != nil {
		op = fmt.Sprintf(">=%v", *cr.StringGreaterThanEquals)
	}

	// NUMBERs
	if cr.NumericEquals != nil {
		op = fmt.Sprintf("=%v", *cr.NumericEquals)
	}

	if cr.NumericLessThan != nil {
		op = fmt.Sprintf("<%v", *cr.NumericLessThan)
	}

	if cr.NumericGreaterThan != nil {
		op = fmt.Sprintf(">%v", *cr.NumericGreaterThan)
	}

	if cr.NumericLessThanEquals != nil {
		op = fmt.Sprintf("<=%v", *cr.NumericLessThanEquals)
	}

	if cr.NumericGreaterThanEquals != nil {
		op = fmt.Sprintf(">=%v", *cr.NumericGreaterThanEquals)
	}

	if cr.BooleanEquals != nil {
		op = fmt.Sprintf("=%v", *cr.BooleanEquals)
	}

	if cr.TimestampEquals != nil {
		op = fmt.Sprintf("=%v", *cr.TimestampEquals)
	}

	if cr.TimestampLessThan != nil {
		op = fmt.Sprintf("<%v", *cr.TimestampLessThan)
	}

	if cr.TimestampGreaterThan != nil {
		op = fmt.Sprintf(">%v", *cr.TimestampGreaterThan)
	}

	if cr.TimestampLessThanEquals != nil {
		op = fmt.Sprintf("<=%v", *cr.TimestampLessThanEquals)
	}

	if cr.TimestampGreaterThanEquals != nil {
		op = fmt.Sprintf(">=%v", *cr.TimestampGreaterThanEquals)
	}

	return fmt.Sprintf("%v%v", cr.Variable.String(), op)
}
