package client

import (
	"time"

	"github.com/coinbase/step/aws"
	"github.com/coinbase/step/aws/s3"
	"github.com/coinbase/step/deployer"
	"github.com/coinbase/step/utils/to"
)

// NewRelease returns a valid Release Structure
func NewRelease(states *string, lambda *string, step *string, bucket *string, region *string, account_id *string, lambda_sha *string) *deployer.Release {
	r := &deployer.Release{
		ReleaseId:        to.TimeUUID("release-"),
		CreatedAt:        to.Timep(time.Now()),
		LambdaName:       lambda,
		StepFnName:       step,
		Bucket:           bucket,
		StateMachineJSON: to.Strp(to.CompactJSONStr(states)),
		LambdaSHA256:     lambda_sha,
	}

	r.SetDefaults(region, account_id)
	return r
}

// PrepareRelease returns a release with additional information filled in by querying AWS
func PrepareRelease(awsc aws.AwsClients, states *string, lambda *string, step *string, bucket *string, zip_file_path *string) (*deployer.Release, error) {
	region, account_id := to.RegionAccount()

	lambda_sha, err := to.SHA256File(*zip_file_path)

	if err != nil {
		return nil, err
	}

	release := NewRelease(states, lambda, step, bucket, region, account_id, &lambda_sha)

	// We get the release values from the lambda
	project, config, _, err := release.LambdaProjectConfigDeployerTags(awsc.LambdaClient())

	if err != nil {
		return nil, err
	}

	release.ProjectName = project
	release.ConfigName = config

	return release, release.ValidateClientAttributes()
}

// PrepareReleaseDeploy builds and uploads necessary info for a deploy
func PrepareReleaseBundle(awsc aws.AwsClients, states *string, lambda *string, step *string, bucket *string, zip_file_path *string) (*deployer.Release, error) {
	release, err := PrepareRelease(awsc, states, lambda, step, bucket, zip_file_path)
	if err != nil {
		return nil, err
	}

	err = s3.PutFile(
		awsc.S3Client(),
		zip_file_path,
		release.Bucket,
		release.LambdaZipPath(),
	)

	// Uploading the SHA of the
	if err := s3.PutStruct(awsc.S3Client(), release.Bucket, release.ReleasePath(), release); err != nil {
		return nil, err
	}

	return release, nil
}
