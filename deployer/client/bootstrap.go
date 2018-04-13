package client

import (
	"fmt"
	"os"

	"github.com/coinbase/step/aws"
	"github.com/coinbase/step/deployer"
	"github.com/coinbase/step/utils/to"
)

// Bootstrap takes release information and uploads directly to Step Function and Lambda
func Bootstrap(release *deployer.Release, zip_file_path *string) error {
	awsc := &aws.AwsClientsStr{}

	fmt.Println("Preparing Release Bundle")
	err := PrepareRelease(release, zip_file_path)
	if err != nil {
		return err
	}

	bts, err := fileBytes(zip_file_path)
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

	err = release.DeployLambdaCode(awsc.LambdaClient(nil, nil, nil), bts)
	if err != nil {
		return err
	}

	fmt.Println("Success")
	return nil
}

func fileBytes(file_path *string) (*[]byte, error) {
	file, err := os.Open(*file_path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fileInfo, _ := file.Stat()
	var size int64 = fileInfo.Size()
	buffer := make([]byte, size)
	file.Read(buffer)
	return &buffer, nil
}
