package client

import (
	"fmt"
	"io/ioutil"

	"github.com/coinbase/step/aws"
	"github.com/coinbase/step/deployer"
	"github.com/coinbase/step/utils/to"
)

// Bootstrap takes release information and uploads directly to Step Function and Lambda
func Bootstrap(release *deployer.Release, zip_file_path *string) error {
	awsc := &aws.Clients{}

	fmt.Println("Preparing Release Bundle")
	err := PrepareRelease(release, zip_file_path)
	if err != nil {
		return err
	}

	bts, err := ioutil.ReadFile(*zip_file_path)
	if err != nil {
		return err
	}

	fmt.Println("Deploying Step Function")
	fmt.Println(to.PrettyJSONStr(release))

	err = release.DeployStepFunction(awsc.SFNClient(nil, nil, nil))
	if err != nil {
		return err
	}

	fmt.Println("Deploying Lambda Function")

	err = release.DeployLambdaCode(awsc.LambdaClient(nil, nil, nil), &bts)
	if err != nil {
		return err
	}

	fmt.Println("Success")
	return nil
}
