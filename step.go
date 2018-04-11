package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/coinbase/step/deployer"
	"github.com/coinbase/step/deployer/client"
	"github.com/coinbase/step/utils/run"
	"github.com/coinbase/step/utils/to"
)

func main() {
	default_name := "coinbase-step-deployer"
	region, account_id := to.RegionAccount()
	def_lambda_arn := to.Strp("")
	def_step_arn := to.Strp("")
	if region != nil && account_id != nil {
		def_lambda_arn = to.LambdaArn(region, account_id, &default_name)
		def_step_arn = to.StepArn(region, account_id, &default_name)
	}

	// Step Subcommands
	jsonCommand := flag.NewFlagSet("json", flag.ExitOnError)
	execCommand := flag.NewFlagSet("exec", flag.ExitOnError)

	// Other Subcommands
	bootstrapCommand := flag.NewFlagSet("bootstrap", flag.ExitOnError)
	deployCommand := flag.NewFlagSet("deploy", flag.ExitOnError)

	// json args
	jsonLambda := jsonCommand.String("lambda", *def_lambda_arn, "lambda name or arn to replace tasks Task.Resource")

	// exec args
	execInput := execCommand.String("input", "{}", "Input JSON to execute")

	// bootstrap args
	bootstrapStates := bootstrapCommand.String("states", "{}", "State Machine JSON")
	bootstrapLambda := bootstrapCommand.String("lambda", "", "lambda name or arn")
	bootstrapStep := bootstrapCommand.String("step", "", "step function name or arn")
	bootstrapBucket := bootstrapCommand.String("bucket", "", "s3 bucket to upload release to")
	bootstrapZip := bootstrapCommand.String("zip", "lambda.zip", "zip of lambda")

	// deploy args
	deployStates := deployCommand.String("states", "{}", "State Machine JSON")
	deployLambda := deployCommand.String("lambda", "", "lambda name or arn")
	deployStep := deployCommand.String("step", "", "step function name or arn")
	deployBucket := deployCommand.String("bucket", "", "s3 bucket to upload release to")
	deployDeployer := deployCommand.String("deployer", *def_step_arn, "step function deployer name or arn")
	deployZip := deployCommand.String("zip", "lambda.zip", "zip of lambda")

	// By Default Run Lambda Function
	if len(os.Args) == 1 {
		fmt.Println("Starting Lambda")
		run.Lambda(deployer.StateMachineWithTaskHandlers())
	}

	switch os.Args[1] {
	case "json":
		jsonCommand.Parse(os.Args[2:])
	case "exec":
		execCommand.Parse(os.Args[2:])
	case "bootstrap":
		bootstrapCommand.Parse(os.Args[2:])
	case "deploy":
		deployCommand.Parse(os.Args[2:])
	default:
		fmt.Println("Usage of step: step <json|exec|bootstrap|deploy> <args> (No args starts Lambda)")
		fmt.Println("json")
		jsonCommand.PrintDefaults()
		fmt.Println("exec")
		execCommand.PrintDefaults()
		fmt.Println("bootstrap")
		bootstrapCommand.PrintDefaults()
		fmt.Println("deploy")
		deployCommand.PrintDefaults()
		os.Exit(1)
	}

	// Create the State machine
	if jsonCommand.Parsed() {
		region, account_id := to.RegionAccount()
		jsonRun(to.LambdaArn(region, account_id, jsonLambda))
	} else if execCommand.Parsed() {
		execRun(execInput)
	} else if bootstrapCommand.Parsed() {
		bootstrapRun(
			bootstrapStates,
			bootstrapLambda,
			bootstrapStep,
			bootstrapBucket,
			bootstrapZip,
		)
	} else if deployCommand.Parsed() {
		region, account_id := to.RegionAccountOrExit()
		deployRun(
			deployStates,
			deployLambda,
			deployStep,
			deployBucket,
			deployZip,
			to.StepArn(region, account_id, deployDeployer),
		)
	} else {
		fmt.Println("ERROR: Command Line Not Parsed")
		os.Exit(1)
	}
}

// Print the state JSON for the step function
func jsonRun(jsonLambda *string) {
	run.JSON(deployer.StateMachineWithLambdaArn(jsonLambda))
}

func execRun(input *string) {
	run.Exec(deployer.StateMachineWithTaskHandlers())(input)
}

func bootstrapRun(states *string, lambda *string, step *string, bucket *string, zip *string) {
	err := client.Bootstrap(states, lambda, step, bucket, zip)
	if err != nil {
		fmt.Println("ERROR", err)
		os.Exit(1)
	}
}

func deployRun(states *string, lambda *string, step *string, bucket *string, zip *string, deployer_arn *string) {
	err := client.Deploy(states, lambda, step, bucket, zip, deployer_arn)
	if err != nil {
		fmt.Println("ERROR", err)
		os.Exit(1)
	}
}
