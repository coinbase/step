package client

import (
	"time"

	"github.com/coinbase/step/aws"
	"github.com/coinbase/step/aws/s3"
	"github.com/coinbase/step/deployer"
	"github.com/coinbase/step/utils/to"
)

// PrepareRelease returns a release with additional information filled in
func PrepareRelease(release *deployer.Release, zip_file_path *string) error {
	region, account_id := to.RegionAccount()
	release.SetDefaults(region, account_id, "coinbase-step-deployer-")

	lambda_sha, err := to.SHA256File(*zip_file_path)
	if err != nil {
		return err
	}
	release.LambdaSHA256 = &lambda_sha

	// Interpolate variables for resource strings
	release.StateMachineJSON = to.InterpolateArnVariables(
		release.StateMachineJSON,
		release.AwsRegion,
		release.AwsAccountID,
		release.LambdaName,
	)

	return nil
}

// PrepareReleaseBundle builds and uploads necessary info for a deploy
func PrepareReleaseBundle(awsc aws.AwsClients, release *deployer.Release, zip_file_path *string) error {
	if err := PrepareRelease(release, zip_file_path); err != nil {
		return err
	}

	err := s3.PutFile(
		awsc.S3Client(release.AwsRegion, nil, nil),
		zip_file_path,
		release.Bucket,
		release.LambdaZipPath(),
	)

	if err != nil {
		return err
	}

	// reset CreateAt because it can take a while to upload the lambda
	release.CreatedAt = to.Timep(time.Now())

	// Uploading the Release to S3 to match SHAs
	if err := s3.PutStruct(awsc.S3Client(release.AwsRegion, nil, nil), release.Bucket, release.ReleasePath(), release); err != nil {
		return err
	}

	return nil
}
