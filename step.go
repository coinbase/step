package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/coinbase/step/deployer"
	"github.com/coinbase/step/deployer/client"
	"github.com/coinbase/step/utils/run"
	"github.com/coinbase/step/utils/to"
)

func main() {
	defaultName := "coinbase-step-deployer"
	region, accountID := to.RegionAccount()
	defLambdaARN := to.Strp("")
	defStepARN := to.Strp("")

	if region != nil && accountID != nil {
		defLambdaARN = to.LambdaArn(region, accountID, &defaultName)
		defStepARN = to.StepArn(region, accountID, &defaultName)
	}

	// Step Subcommands
	jsonCommand := flag.NewFlagSet("json", flag.ExitOnError)
	execCommand := flag.NewFlagSet("exec", flag.ExitOnError)

	// Other Subcommands
	bootstrapCommand := flag.NewFlagSet("bootstrap", flag.ExitOnError)
	deployCommand := flag.NewFlagSet("deploy", flag.ExitOnError)

	// json args
	jsonLambda := jsonCommand.String("lambda", *defLambdaARN, "lambda name or arn to replace tasks Task.Resource")

	// exec args
	execInput := execCommand.String("input", "{}", "Input JSON to execute")

	// bootstrap args
	bootstrapStates := bootstrapCommand.String("states", "{}", "State Machine JSON")
	bootstrapLambda := bootstrapCommand.String("lambda", "", "lambda name or arn")
	bootstrapStep := bootstrapCommand.String("step", "", "step function name or arn")
	bootstrapBucket := bootstrapCommand.String("bucket", "", "s3 bucket to upload release to")
	bootstrapZip := bootstrapCommand.String("zip", "lambda.zip", "zip of lambda")
	bootstrapProject := bootstrapCommand.String("project", "", "project name")
	bootstrapConfig := bootstrapCommand.String("config", "", "config name")
	bootstrapRegion := bootstrapCommand.String("region", "", "AWS region")
	bootstrapAccount := bootstrapCommand.String("account", "", "AWS account id")

	// deploy args
	deployStates := deployCommand.String("states", "{}", "State Machine JSON")
	deployLambda := deployCommand.String("lambda", "", "lambda name or arn")
	deployStep := deployCommand.String("step", "", "step function name or arn")
	deployBucket := deployCommand.String("bucket", "", "s3 bucket to upload release to")
	deployDeployer := deployCommand.String("deployer", *defStepARN, "step function deployer name or arn")
	deployZip := deployCommand.String("zip", "lambda.zip", "zip of lambda")
	deployProject := deployCommand.String("project", "", "project name")
	deployConfig := deployCommand.String("config", "", "config name")
	deployRegion := deployCommand.String("region", "", "AWS region")
	deployAccount := deployCommand.String("account", "", "AWS account id")

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
		region, accountID := to.RegionAccount()
		jsonRun(to.LambdaArn(region, accountID, jsonLambda))
	} else if execCommand.Parsed() {
		execRun(execInput)
	} else if bootstrapCommand.Parsed() {
		r := newRelease(
			bootstrapProject,
			bootstrapConfig,
			bootstrapLambda,
			bootstrapStep,
			bootstrapBucket,
			bootstrapStates,
			bootstrapRegion,
			bootstrapAccount,
		)
		bootstrapRun(r, bootstrapZip)

	} else if deployCommand.Parsed() {
		region, accountID := to.RegionAccountOrExit()
		r := newRelease(
			deployProject,
			deployConfig,
			deployLambda,
			deployStep,
			deployBucket,
			deployStates,
			deployRegion,
			deployAccount,
		)
		arn := to.StepArn(region, accountID, deployDeployer)
		deployRun(r, deployZip, arn)
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

func bootstrapRun(release *deployer.Release, zip *string) {
	err := client.Bootstrap(release, zip)
	if err != nil {
		fmt.Println("ERROR", err)
		os.Exit(1)
	}
}

func deployRun(release *deployer.Release, zip *string, deployerARN *string) {
	err := client.Deploy(release, zip, deployerARN)
	if err != nil {
		fmt.Println("ERROR", err)
		os.Exit(1)
	}
}

func newRelease(project *string, config *string, lambda *string, step *string, bucket *string, states *string, region *string, accountID *string) *deployer.Release {
	return &deployer.Release{
		ReleaseId:        to.TimeUUID("release-"),
		CreatedAt:        to.Timep(time.Now()),
		ProjectName:      project,
		ConfigName:       config,
		LambdaName:       lambda,
		StepFnName:       step,
		Bucket:           bucket,
		StateMachineJSON: states,
		AwsRegion:        region,
		AwsAccountID:     accountID,
	}
}
