package run

import (
	"fmt"
	"os"
	"strings"

	"github.com/coinbase/step/machine"
	"github.com/coinbase/step/machine/state"
	"github.com/coinbase/step/utils/to"
)

// Output Dot Format For State Machine

// JSON prints a state machine as JSON
func Dot(stateMachine *machine.StateMachine, err error) {
	if err != nil {
		fmt.Println("ERROR", err)
		os.Exit(1)
	}

	dotStr := toDot(stateMachine)
	fmt.Println(dotStr)
	os.Exit(0)
}

func toDot(stateMachine *machine.StateMachine) string {
	return fmt.Sprintf(`digraph StateMachine {
    node      [style="rounded,filled,bold", shape=box, width=2, fontname="Arial" fontcolor="#183153", color="#183153"];
    edge      [style=bold, fontname="Arial", fontcolor="#183153", color="#183153"];
    _Start    [fillcolor="#183153", shape=circle, label="", width=0.25];
    _End      [fillcolor="#183153", shape=doublecircle, label="", width=0.3];

    _Start -> "%v" [weight=1000];
    %v
}`, *stateMachine.StartAt, processStates(*stateMachine.StartAt, stateMachine.States))
}

func processStates(start string, states map[string]state.State) string {
	orderedStates := orderStates(start, states)

	var stateStrings []string
	for _, stateNode := range orderedStates {
		stateStrings = append(stateStrings, processState(stateNode))
	}
	return strings.Join(stateStrings, "\n\n    ")
}

// Order states from start to end consistently to generate deterministic graphs.
func orderStates(start string, states map[string]state.State) []state.State {
	var orderedStates []state.State
	startState := states[start]
	stateQueue := []state.State{startState}
	seenStates := make(map[string]struct{})

	for len(stateQueue) > 0 {
		var stateNode state.State
		stateNode, stateQueue = stateQueue[0], stateQueue[1:]

		orderedStates = append(orderedStates, stateNode)

		var connectedStates []state.State
		switch stateNode.(type) {
		case *state.PassState:
			stateNode := stateNode.(*state.PassState)
			if stateNode.Next != nil {
				connectedStates = append(connectedStates, states[*stateNode.Next])
			}
		case *state.TaskState:
			stateNode := stateNode.(*state.TaskState)

			if stateNode.Catch != nil {
				for _, catch := range stateNode.Catch {
					connectedStates = append(connectedStates, states[*catch.Next])
				}
			}

			if stateNode.Next != nil {
				connectedStates = append(connectedStates, states[*stateNode.Next])
			}
		case *state.ChoiceState:
			stateNode := stateNode.(*state.ChoiceState)

			if stateNode.Choices != nil {
				for _, choice := range stateNode.Choices {
					connectedStates = append(connectedStates, states[*choice.Next])
				}
			}
		case *state.WaitState:
			stateNode := stateNode.(*state.WaitState)

			if stateNode.Next != nil {
				connectedStates = append(connectedStates, states[*stateNode.Next])
			}
		}

		for _, connectedState := range connectedStates {
			stateName := *connectedState.Name()
			if _, seen := seenStates[stateName]; !seen {
				stateQueue = append(stateQueue, connectedState)
				seenStates[stateName] = struct{}{}
			}
		}
	}

	return orderedStates
}

func processState(stateNode state.State) string {
	var lines []string
	name := *stateNode.Name()
	switch stateNode.(type) {
	case *state.PassState:
		stateNode := stateNode.(*state.PassState)
		lines = append(lines, fmt.Sprintf(`%q [fillcolor="#FBFBFB"];`, name))
		if stateNode.Next != nil {
			lines = append(lines, fmt.Sprintf(`%q -> %q [weight=100];`, name, *stateNode.Next))
		}
		if stateNode.End != nil {
			lines = append(lines, fmt.Sprintf(`%q -> _End;`, name))
		}
	case *state.TaskState:
		stateNode := stateNode.(*state.TaskState)
		lines = append(lines, fmt.Sprintf(`%q [fillcolor="#FBFBFB"];`, name))

		if stateNode.Catch != nil {
			for _, catch := range stateNode.Catch {
				catchName := fmt.Sprintf("%q", strings.Join(to.StrSlice(catch.ErrorEquals), ","))
				if len(catch.ErrorEquals) == 1 && *catch.ErrorEquals[0] == "States.ALL" {
					catchName = ""
				}
				lines = append(lines, fmt.Sprintf(`%q -> %q [color="#949494", label=%q, style=solid];`, name, *catch.Next, catchName))
			}
		}

		if stateNode.Next != nil {
			lines = append(lines, fmt.Sprintf(`%q -> %q [weight=100];`, name, *stateNode.Next))
		}

		if stateNode.End != nil {
			lines = append(lines, fmt.Sprintf(`%q -> _End;`, name))
		}
	case *state.ChoiceState:
		stateNode := stateNode.(*state.ChoiceState)
		lines = append(lines, fmt.Sprintf(`%q [shape=egg, fillcolor="#FBFBFB"];`, name))

		if stateNode.Choices != nil {
			for _, choice := range stateNode.Choices {
				lines = append(lines, fmt.Sprintf(`%q -> %q [weight=100];`, name, *choice.Next))
			}
		}
	case *state.WaitState:
		stateNode := stateNode.(*state.WaitState)

		lines = append(lines, fmt.Sprintf(`%q [width=0.5, shape=doublecircle, fillcolor="#FBFBFB", label="Wait"];`, name))

		if stateNode.Next != nil {
			lines = append(lines, fmt.Sprintf(`%q -> %q [weight=100];`, name, *stateNode.Next))
		}
	case *state.FailState:
		lines = append(lines, fmt.Sprintf(`%q [fillcolor="#F9E4D1"];`, name))
		lines = append(lines, fmt.Sprintf(`%q -> _End [weight=1000];`, name))
	case *state.SucceedState:
		lines = append(lines, fmt.Sprintf(`%q [fillcolor="#e5eddb"];`, name))
		lines = append(lines, fmt.Sprintf(`%q -> _End [weight=1000];`, name))
	}

	return strings.Join(lines, "\n    ")
}
