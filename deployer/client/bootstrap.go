package client

import (
	"fmt"

	"github.com/coinbase/step/aws"
	"github.com/coinbase/step/utils/to"
)

// Bootstrap takes release information and uploads directly to Step Function and Lambda
func Bootstrap(states *string, lambda *string, step *string, bucket *string, zip_file_path *string) error {
	awsc := aws.CreateAwsClients()

	fmt.Println("Preparing Release Bundle")
	release, err := PrepareReleaseBundle(awsc, states, lambda, step, bucket, zip_file_path)
	if err != nil {
		return err
	}

	fmt.Println("Deploying Step Function")
	fmt.Println(to.PrettyJSONStr(release))
	err = release.DeployStepFunction(awsc.SFNClient())
	if err != nil {
		return err
	}

	fmt.Println("Deploying Lambda Function")
	err = release.DeployLambda(awsc.LambdaClient())
	if err != nil {
		return err
	}

	fmt.Println("Success")
	return nil
}
