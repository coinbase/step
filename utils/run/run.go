// run takes arguments
package run

import (
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/coinbase/step/handler"
	"github.com/coinbase/step/machine"
	"github.com/coinbase/step/utils/is"
	"github.com/coinbase/step/utils/to"
)

// Exec returns a function that will execute the state machine
func Exec(state_machine *machine.StateMachine, err error) func(*string) {
	if err != nil {
		return func(input *string) {
			fmt.Println("ERROR", err)
			os.Exit(1)
		}
	}

	return func(input *string) {

		if is.EmptyStr(input) {
			input = to.Strp("{}")
		}

		exec, err := state_machine.Execute(input)
		output_json := exec.OutputJSON

		if err != nil {
			fmt.Println("ERROR", err)
			os.Exit(1)
		}

		fmt.Println(output_json)
		os.Exit(0)
	}
}

// JSON prints a state machine as JSON
func JSON(state_machine *machine.StateMachine, err error) {
	if err != nil {
		fmt.Println("ERROR", err)
		os.Exit(1)
	}

	json, err := to.PrettyJSON(state_machine)

	if err != nil {
		fmt.Println("ERROR", err)
		os.Exit(1)
	}

	fmt.Println(string(json))
	os.Exit(0)
}

// LambdaTasks takes task functions and and executes as a lambda
func LambdaTasks(task_functions *handler.TaskHandlers) {
	handler, err := handler.CreateHandler(task_functions)

	if err != nil {
		fmt.Println("ERROR", err)
		os.Exit(1)
	}

	lambda.Start(handler)

	fmt.Println("ERROR: lambda.Start returned, but should have blocked")
	os.Exit(1)
}
