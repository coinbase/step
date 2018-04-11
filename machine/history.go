package machine

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/service/sfn"
	"github.com/coinbase/step/machine/state"
	"github.com/coinbase/step/utils/to"
)

type HistoryEvent struct {
	sfn.HistoryEvent
}

func (sm *StateMachine) ExecutionHistory() []HistoryEvent {
	return sm.executionHistory
}

func (sm *StateMachine) ExecutionPath() []string {
	path := []string{}
	for _, er := range sm.executionHistory {
		if er.StateEnteredEventDetails != nil {
			path = append(path, *er.StateEnteredEventDetails.Name)
		}
	}
	return path
}

func (sm *StateMachine) LastOutput() string {
	op := ""
	for _, er := range sm.executionHistory {
		if er.StateExitedEventDetails != nil {
			if er.StateExitedEventDetails.Output != nil {
				op = *er.StateExitedEventDetails.Output
			}
		}
	}
	return op
}

func startExecution() []HistoryEvent {
	return []HistoryEvent{createEvent("ExecutionStarted")}
}

func createEvent(name string) HistoryEvent {
	t := time.Now()
	return HistoryEvent{
		sfn.HistoryEvent{
			Type:      to.Strp(name),
			Timestamp: &t,
		},
	}
}

func createEnteredEvent(state state.State, input interface{}) HistoryEvent {
	event := createEvent(fmt.Sprintf("%vStateEntered", *state.GetType()))
	json_raw, err := json.Marshal(input)

	if err != nil {
		json_raw = []byte{}
	}

	event.StateEnteredEventDetails = &sfn.StateEnteredEventDetails{
		Name:  state.Name(),
		Input: to.Strp(string(json_raw)),
	}

	return event
}

func createExitedEvent(state state.State, output interface{}) HistoryEvent {
	event := createEvent(fmt.Sprintf("%vStateExited", *state.GetType()))
	json_raw, err := json.Marshal(output)

	if err != nil {
		json_raw = []byte{}
	}

	event.StateExitedEventDetails = &sfn.StateExitedEventDetails{
		Name:   state.Name(),
		Output: to.Strp(string(json_raw)),
	}

	return event
}
