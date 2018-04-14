/*
The deployer package contains the Step Deployer service
that is a Step Function that Deploys Step Functions.

It also contains a client for messaging and bootstrapping the Step Deployer.
*/
package deployer

import (
	"context"
	"fmt"

	"github.com/coinbase/step/aws"
	"github.com/coinbase/step/utils/to"
)

////////
// ERRORS
///////

type ErrorWrapper struct {
	err error
}

func (e *ErrorWrapper) Error() string {
	return fmt.Sprintf("ERROR: %v", e.err)
}

type BadReleaseError struct {
	*ErrorWrapper
}

type LockExistsError struct {
	*ErrorWrapper
}

type LockError struct {
	*ErrorWrapper
}

type DeploySFNError struct {
	*ErrorWrapper
}

type DeployLambdaError struct {
	*ErrorWrapper
}

////////////
// HANDLERS
////////////

var assumed_role = to.Strp("coinbase-step-deployer-assumed")

func ValidateHandler(awsc aws.AwsClients) interface{} {
	return func(ctx context.Context, release *Release) (*Release, error) {
		// Override any attributes set by the client
		release.ReleaseSHA256 = to.SHA256Struct(release)
		release.UUID = nil // Will be set later
		release.Success = to.Boolp(false)

		region, account_id := to.AwsRegionAccountFromContext(ctx)
		release.SetDefaults(&region, &account_id) // Fill in all the blank Attributes

		// Validate the attributes for the release
		if err := release.ValidateAttributes(); err != nil {
			return nil, &BadReleaseError{&ErrorWrapper{err}}
		}

		return release, nil
	}
}

func LockHandler(awsc aws.AwsClients) interface{} {
	return func(ctx context.Context, release *Release) (*Release, error) {
		// First Thing is to grab the Lock
		grabbed, err := release.GrabLock(awsc.S3Client(nil, nil, nil))

		// Check grabbed first because there are errors that can be thrown before anything is created
		if !grabbed {
			return nil, &LockExistsError{&ErrorWrapper{fmt.Errorf("Lock Already Exists")}}
		}

		if err != nil {
			return nil, &LockError{&ErrorWrapper{err}}
		}

		return release, nil
	}
}

func ValidateResourcesHandler(awsc aws.AwsClients) interface{} {
	return func(ctx context.Context, release *Release) (*Release, error) {
		// Validate the Resources for the release
		if err := release.ValidateResources(awsc.LambdaClient(release.AwsRegion, release.AwsAccountID, assumed_role), awsc.SFNClient(release.AwsRegion, release.AwsAccountID, assumed_role), awsc.S3Client(nil, nil, nil)); err != nil {
			return nil, &BadReleaseError{&ErrorWrapper{err}}
		}

		return release, nil
	}
}

func DeployHandler(awsc aws.AwsClients) interface{} {
	return func(ctx context.Context, release *Release) (*Release, error) {

		// Update Step Function first because State Machine if it fails we can recover
		if err := release.DeployStepFunction(awsc.SFNClient(release.AwsRegion, release.AwsAccountID, assumed_role)); err != nil {
			return nil, &DeploySFNError{&ErrorWrapper{err}}
		}

		if err := release.DeployLambda(awsc.LambdaClient(release.AwsRegion, release.AwsAccountID, assumed_role), awsc.S3Client(nil, nil, nil)); err != nil {
			return nil, &DeployLambdaError{&ErrorWrapper{err}}
		}

		release.Success = to.Boolp(true)
		release.ReleaseLock(awsc.S3Client(nil, nil, nil))

		return release, nil
	}
}

func ReleaseLockFailureHandler(awsc aws.AwsClients) interface{} {
	return func(ctx context.Context, release *Release) (*Release, error) {

		if err := release.ReleaseLock(awsc.S3Client(nil, nil, nil)); err != nil {
			return nil, &LockError{&ErrorWrapper{err}}
		}

		return release, nil
	}
}
