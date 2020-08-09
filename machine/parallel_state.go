package machine

import (
	"context"
	"fmt"
	"github.com/coinbase/step/utils/to"
	"sync"
)

type ParallelState struct {
	stateStr // Include Defaults

	Type *string
	Comment *string `json:",omitempty"`
	Next *string
	Branches []StateMachine
}

type BranchExecution struct {
	Execution *Execution
	ExecutionError error
}

type ParallelStateExecution struct {
	BranchExecution []*BranchExecution
}

func (s *ParallelState) Execute(_ context.Context, input interface{}) (output interface{}, next *string, err error) {
	// parallel state according to asl spec does not assume Next therefore we could run each branch as a separate
    execution := ParallelStateExecution{}
    execution.BranchExecution = make([]*BranchExecution, len(s.Branches))
    awaiter := sync.WaitGroup{}
    awaiter.Add(len(s.Branches))
	for i, b := range s.Branches {
		go func(index int, branch *StateMachine){
			result, err := branch.Execute(input)
			execution.BranchExecution[index] = &BranchExecution{}
			execution.BranchExecution[index].Execution = result
			execution.BranchExecution[index].ExecutionError = err
			defer awaiter.Done()
		}(i, &b)
	}
	awaiter.Wait()
	// TODO: UMV: do think how to use execution (ParallelStateExecution)
	return nil, s.Next, nil
}

// TODO: umv: accumulate all errors before return
func (s *ParallelState) Validate() error {
	s.SetType(to.Strp("Parallel"))
	// TODO: umv: Default Validator (Do we need it?)
	if err := ValidateNameAndType(s); err != nil {
		return fmt.Errorf("%v %v", errorPrefix(s), err)
	}
	// 1. State Must contains Not Null Branches
	if s.Branches == nil || len(s.Branches) < 1 {
		return  fmt.Errorf("branches can't be nil or empty (len = 0)")
	}

	// 2. Every state (except Succeed) must have Next or End is Validating inside it own Validate method (i.e. in State)
	// 3. Reachable Next is testing in State machine validate method
	for _, b := range s.Branches {
		err := b.Validate()
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *ParallelState) SetType(t *string) {
	s.Type = t
}

func (s *ParallelState) GetType() *string {
	return s.Type
}

func (s *ParallelState) GetNext() *string {
	return s.Next
}