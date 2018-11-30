package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/coinbase/step/machine"

	"github.com/coinbase/step/bifrost"
	"github.com/coinbase/step/client"
	"github.com/coinbase/step/deployer"
	"github.com/coinbase/step/utils/run"
	"github.com/coinbase/step/utils/to"
)

func main() {
	default_name := "coinbase-step-deployer"
	region, account_id := to.RegionAccount()
	def_step_arn := to.Strp("")
	if region != nil && account_id != nil {
		def_step_arn = to.StepArn(region, account_id, &default_name)
	}

	// Step Subcommands
	jsonCommand := flag.NewFlagSet("json", flag.ExitOnError)

	dotCommand := flag.NewFlagSet("dot", flag.ExitOnError)
	dotStates := dotCommand.String("states", "{}", "State Machine JSON")

	// Other Subcommands
	bootstrapCommand := flag.NewFlagSet("bootstrap", flag.ExitOnError)
	deployCommand := flag.NewFlagSet("deploy", flag.ExitOnError)

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
	deployDeployer := deployCommand.String("deployer", *def_step_arn, "step function deployer name or arn")
	deployZip := deployCommand.String("zip", "lambda.zip", "zip of lambda")
	deployProject := deployCommand.String("project", "", "project name")
	deployConfig := deployCommand.String("config", "", "config name")
	deployRegion := deployCommand.String("region", "", "AWS region")
	deployAccount := deployCommand.String("account", "", "AWS account id")

	// By Default Run Lambda Function
	if len(os.Args) == 1 {
		fmt.Println("Starting Lambda")
		run.LambdaTasks(deployer.TaskHandlers())
	}

	switch os.Args[1] {
	case "json":
		jsonCommand.Parse(os.Args[2:])
	case "dot":
		dotCommand.Parse(os.Args[2:])
	case "bootstrap":
		bootstrapCommand.Parse(os.Args[2:])
	case "deploy":
		deployCommand.Parse(os.Args[2:])
	default:
		fmt.Println("Usage of step: step <json|bootstrap|deploy|dot> <args> (No args starts Lambda)")
		fmt.Println("json")
		jsonCommand.PrintDefaults()
		fmt.Println("dot")
		dotCommand.PrintDefaults()
		fmt.Println("bootstrap")
		bootstrapCommand.PrintDefaults()
		fmt.Println("deploy")
		deployCommand.PrintDefaults()
		os.Exit(1)
	}

	// Create the State machine
	if jsonCommand.Parsed() {
		run.JSON(deployer.StateMachine())
	} else if dotCommand.Parsed() {
		run.Dot(machine.FromJSON([]byte(*dotStates)))
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
		region, account_id := to.RegionAccountOrExit()
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
		arn := to.StepArn(region, account_id, deployDeployer)
		deployRun(r, deployZip, arn)
	} else {
		fmt.Println("ERROR: Command Line Not Parsed")
		os.Exit(1)
	}
}

func check(err error) {
	if err == nil {
		return
	}
	fmt.Println("ERROR", err)
	os.Exit(1)
}

func bootstrapRun(release *deployer.Release, zip *string) {
	err := client.Bootstrap(release, zip)
	check(err)
}

func deployRun(release *deployer.Release, zip *string, deployer_arn *string) {
	err := client.Deploy(release, zip, deployer_arn)
	check(err)
}

func newRelease(project *string, config *string, lambda *string, step *string, bucket *string, states *string, region *string, account_id *string) *deployer.Release {
	return &deployer.Release{
		Release: bifrost.Release{
			AwsRegion:    region,
			AwsAccountID: account_id,
			ReleaseID:    to.TimeUUID("release-"),
			CreatedAt:    to.Timep(time.Now()),
			ProjectName:  project,
			ConfigName:   config,
			Bucket:       bucket,
		},
		StateMachineJSON: states,
		LambdaName:       lambda,
		StepFnName:       step,
	}
}
