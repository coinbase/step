package client

import (
	"encoding/json"
	"fmt"

	"github.com/coinbase/step/aws"
	"github.com/coinbase/step/bifrost"
	"github.com/coinbase/step/deployer"
	"github.com/coinbase/step/execution"
	"github.com/coinbase/step/utils/to"
)

// Deploy takes release information and Calls the Step Deployer to deploy the release
func Deploy(release *deployer.Release, zip_file_path *string, deployer_arn *string) error {
	awsc := &aws.Clients{}

	fmt.Println("Preparing Release Bundle")
	err := PrepareReleaseBundle(awsc, release, zip_file_path)
	if err != nil {
		return err
	}

	fmt.Println("Preparing Deploy")
	fmt.Println(to.PrettyJSONStr(release))
	err = sendDeployToDeployer(awsc.SFNClient(nil, nil, nil), release.ReleaseID, release, deployer_arn)
	if err != nil {
		return err
	}

	return nil
}

// sendDeployToDeployer Calls the Step Deployer Step Function,
// This function will wait for the execution to finish but will timeout after 20 seconds
func sendDeployToDeployer(sfnc aws.SFNAPI, name *string, release *deployer.Release, deployer_arn *string) error {

	exec, err := execution.StartExecution(sfnc, deployer_arn, name, release)
	if err != nil {
		return err
	}

	fmt.Printf("\nStarting Deploy")

	exec.WaitForExecution(sfnc, 1, func(ed *execution.Execution, sd *execution.StateDetails, err error) error {
		if err != nil {
			return fmt.Errorf("Unexpected Error %v", err.Error())
		}

		var release_error struct {
			Error *bifrost.ReleaseError `json:"error,omitempty"`
		}

		fmt.Printf("\rExecution: %v", *ed.Status)

		if sd.LastOutput != nil {
			json.Unmarshal([]byte(*sd.LastOutput), &release_error)

			if release_error.Error != nil {
				fmt.Printf("\nError: %v\nCause: %v\n", to.Strs(release_error.Error.Error), to.Strs(release_error.Error.Cause))
			}
		}

		return nil
	})

	return nil
}
