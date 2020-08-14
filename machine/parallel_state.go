package machine

import (
	"context"
	"fmt"
	"github.com/coinbase/step/jsonpath"
	"github.com/coinbase/step/utils/to"
	"sync"
)

type ParallelState struct {
	stateStr // Include Defaults
	Type *string
	Comment *string `json:",omitempty"`
	Next *string`json:",omitempty"`
	End *bool`json:",omitempty"`
	Branches []StateMachine

	InputPath  *jsonpath.Path `json:",omitempty"`
	OutputPath *jsonpath.Path `json:",omitempty"`
	ResultPath *jsonpath.Path `json:",omitempty"`
	Parameters interface{} `json:",omitempty"`
}

type BranchExecution struct {
	Execution *Execution
	ExecutionError error
}

type ParallelStateExecution struct {
	BranchExecution []*BranchExecution
}

func (s *ParallelState) Execute(_ context.Context, input interface{}) (output interface{}, next *string, err error) {
	// UMV: ParallelStateExecution could be used
    execution := ParallelStateExecution{}
    execution.BranchExecution = make([]*BranchExecution, len(s.Branches))
    outData := make([]interface{}, len(execution.BranchExecution))
    awaiter := sync.WaitGroup{}
    awaiter.Add(len(s.Branches))
	for i, b := range s.Branches {
		go func(index int, branch StateMachine){
			result, err := branch.Execute(input)
			execution.BranchExecution[index] = &BranchExecution{}
			execution.BranchExecution[index].Execution = result
			execution.BranchExecution[index].ExecutionError = err
			outData[index] = result.Output
			defer awaiter.Done()
		}(i, b)
	}
	awaiter.Wait()
	return outData, s.Next, nil
}

// TODO: umv: accumulate all errors before return
func (s *ParallelState) Validate() error {
	s.SetType(to.Strp("Parallel"))
	// 1. Validate Name & Type
	if err := ValidateNameAndType(s); err != nil {
		return fmt.Errorf("%v %v", errorPrefix(s), err)
	}
	// 2. State Must have either Next or End  field
	if s.Next == nil && s.End == nil {
		return fmt.Errorf("parallel state must have either \"Next\" or \"End\" property")
	}
	// 3. State Must contains Not Null Branches
	if s.Branches == nil || len(s.Branches) < 1 {
		return fmt.Errorf("branches can't be nil or empty (len = 0)")
	}

	// Following conditions checks in StateMachine Validate()
	// 4. Every state (except Succeed) must have Next or End is Validating inside it own Validate method (i.e. in State)
	// 5. Reachable Next is testing in State machine validate method
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