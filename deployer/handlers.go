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
	"github.com/coinbase/step/aws/dynamodb"
	"github.com/coinbase/step/errors"
	"github.com/coinbase/step/utils/to"
)

////////
// ERRORS
///////

type DeploySFNError struct {
	err error
}

type DeployLambdaError struct {
	err error
}

func (e DeploySFNError) Error() string {
	return fmt.Sprintf("DeploySFNError: %v", e.err.Error())
}

func (e DeployLambdaError) Error() string {
	return fmt.Sprintf("DeployLambdaError: %v", e.err.Error())
}

////////////
// HANDLERS
////////////

var assumed_role = to.Strp("coinbase-step-deployer-assumed")

func ValidateHandler(awsc aws.AwsClients) interface{} {
	return func(ctx context.Context, release *Release) (*Release, error) {
		// Override any attributes set by the client
		release.ReleaseSHA256 = to.SHA256Struct(release)
		release.WipeControlledValues()

		region, account := to.AwsRegionAccountFromContext(ctx)
		release.SetDefaults(region, account, "coinbase-step-deployer-")

		// Validate the attributes for the release
		if err := release.Validate(awsc.S3Client(nil, nil, nil)); err != nil {
			return nil, errors.BadReleaseError{err.Error()}
		}

		return release, nil
	}
}

func LockHandler(awsc aws.AwsClients) interface{} {
	return func(ctx context.Context, release *Release) (*Release, error) {
		// returns LockExistsError, LockError
		locker := dynamodb.NewDynamoDBLocker(awsc.DynamoDBClient(nil, nil, nil))
		return release, release.GrabLocks(awsc.S3Client(nil, nil, nil), locker, getLockTableNameFromContext(ctx, "-locks"))
	}
}

func ValidateResourcesHandler(awsc aws.AwsClients) interface{} {
	return func(ctx context.Context, release *Release) (*Release, error) {
		// Validate the Resources for the release
		if err := release.ValidateResources(awsc.LambdaClient(release.AwsRegion, release.AwsAccountID, assumed_role), awsc.SFNClient(release.AwsRegion, release.AwsAccountID, assumed_role)); err != nil {
			return nil, errors.BadReleaseError{err.Error()}
		}

		return release, nil
	}
}

func DeployHandler(awsc aws.AwsClients) interface{} {
	return func(ctx context.Context, release *Release) (*Release, error) {

		// Update Step Function first because State Machine if it fails we can recover
		if err := release.DeployStepFunction(awsc.SFNClient(release.AwsRegion, release.AwsAccountID, assumed_role)); err != nil {
			return nil, DeploySFNError{err}
		}

		if err := release.DeployLambda(awsc.LambdaClient(release.AwsRegion, release.AwsAccountID, assumed_role), awsc.S3Client(nil, nil, nil)); err != nil {
			return nil, DeployLambdaError{err}
		}

		release.Success = to.Boolp(true)
		locker := dynamodb.NewDynamoDBLocker(awsc.DynamoDBClient(nil, nil, nil))
		release.UnlockRoot(locker, getLockTableNameFromContext(ctx, "-locks"))

		return release, nil
	}
}

func ReleaseLockFailureHandler(awsc aws.AwsClients) interface{} {
	return func(ctx context.Context, release *Release) (*Release, error) {
		locker := dynamodb.NewDynamoDBLocker(awsc.DynamoDBClient(nil, nil, nil))
		if err := release.UnlockRoot(locker, getLockTableNameFromContext(ctx, "-locks")); err != nil {
			return nil, errors.LockError{err.Error()}
		}

		return release, nil
	}
}

func getLockTableNameFromContext(ctx context.Context, postfix string) string {
	_, _, lambdaName := to.AwsRegionAccountLambdaNameFromContext(ctx)
	return fmt.Sprintf("%s%s", lambdaName, postfix)
}
