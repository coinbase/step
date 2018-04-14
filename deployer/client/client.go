package client

import (
	"github.com/coinbase/step/aws"
	"github.com/coinbase/step/aws/s3"
	"github.com/coinbase/step/deployer"
	"github.com/coinbase/step/machine"
	"github.com/coinbase/step/utils/to"
)

// PrepareRelease returns a release with additional information filled in
func PrepareRelease(release *deployer.Release, zip_file_path *string) error {
	region, account_id := to.RegionAccount()
	release.SetDefaults(region, account_id)

	lambda_sha, err := to.SHA256File(*zip_file_path)
	if err != nil {
		return err
	}
	release.LambdaSHA256 = &lambda_sha

	// Add the lambda resource to Tasks with nil resource
	state_machine, err := machine.FromJSON([]byte(*release.StateMachineJSON))
	if err != nil {
		return err
	}

	lambda_arn := to.LambdaArn(release.AwsRegion, release.AwsAccountID, release.LambdaName)
	state_machine.SetResource(lambda_arn)
	release.StateMachineJSON = to.Strp(to.CompactJSONStr(state_machine))

	return release.ValidateClientAttributes()
}

// PrepareReleaseDeploy builds and uploads necessary info for a deploy
func PrepareReleaseBundle(awsc aws.AwsClients, release *deployer.Release, zip_file_path *string) error {
	if err := PrepareRelease(release, zip_file_path); err != nil {
		return err
	}

	err := s3.PutFile(
		awsc.S3Client(nil, nil, nil),
		zip_file_path,
		release.Bucket,
		release.LambdaZipPath(),
	)

	if err != nil {
		return err
	}

	// Uploading the Release to S3 to match SHAs
	if err := s3.PutStruct(awsc.S3Client(nil, nil, nil), release.Bucket, release.ReleasePath(), release); err != nil {
		return err
	}

	return nil
}
