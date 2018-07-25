package deployer

import (
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/sfn"
	"github.com/coinbase/step/machine"
	"github.com/coinbase/step/utils/to"
	"github.com/stretchr/testify/assert"
)

///////////
// HAPPY PATH
///////////

func Test_DeployHandler_Execution_Works(t *testing.T) {
	release := MockRelease()
	awsc := MockAwsClients(release)
	state_machine := createTestStateMachine(t, awsc)

	output, err := state_machine.ExecuteToMap(release)

	assert.NoError(t, err)
	assert.Equal(t, output["success"], true)
	assert.NotRegexp(t, "error", state_machine.LastOutput())

	assertNoLock(t, awsc, release)
	assert.Equal(t, []string{
		"Validate",
		machine.TaskFnName("Validate"),
		"Lock",
		machine.TaskFnName("Lock"),
		"ValidateResources",
		machine.TaskFnName("ValidateResources"),
		"Deploy",
		machine.TaskFnName("Deploy"),
		"Success",
	}, state_machine.ExecutionPath())
}

func Test_DeployHandler_Execution_NoUUIDorSHA_Override(t *testing.T) {
	release := MockRelease()
	release.UUID = to.Strp("badString")
	release.ReleaseSHA256 = "badString"

	awsc := MockAwsClients(release)
	state_machine := createTestStateMachine(t, awsc)

	output, err := state_machine.ExecuteToMap(release)
	assert.NoError(t, err)
	assert.Equal(t, output["success"], true)
	assert.NotEqual(t, output["uuid"], "badString")
	assert.NotEqual(t, output["release_sha256"], "badString")
}

/////////
// UNHAPPY PATH :(
/////////

// Bad Release
func Test_DeployHandler_Execution_Errors_BadInput(t *testing.T) {
	release := MockRelease()
	awsc := MockAwsClients(release)
	state_machine := createTestStateMachine(t, awsc)

	_, err := state_machine.ExecuteToMap(struct{}{})

	assert.Error(t, err)
	assert.Regexp(t, "BadReleaseError", state_machine.LastOutput())
	assertNoLock(t, awsc, release)

	assert.Equal(t, []string{
		"Validate",
		machine.TaskFnName("Validate"),
		"FailureClean",
	}, state_machine.ExecutionPath())
}

func Test_DeployHandler_Execution_Errors_Release(t *testing.T) {
	release := MockRelease()

	awsc := MockAwsClients(release)
	release.ReleaseID = nil

	state_machine := createTestStateMachine(t, awsc)

	_, err := state_machine.ExecuteToMap(release)

	assert.Error(t, err)
	assert.Regexp(t, "BadReleaseError", state_machine.LastOutput())
	assert.Regexp(t, "ReleaseID must", state_machine.LastOutput())
	assertNoLock(t, awsc, release)

	assert.Equal(t, []string{
		"Validate",
		machine.TaskFnName("Validate"),
		"FailureClean",
	}, state_machine.ExecutionPath())
}

func Test_DeployHandler_Execution_Errors_CreatedAt(t *testing.T) {
	release := MockRelease()
	release.CreatedAt = to.Timep(time.Now().Add(10 * time.Minute))

	awsc := MockAwsClients(release)

	state_machine := createTestStateMachine(t, awsc)

	_, err := state_machine.ExecuteToMap(release)

	assert.Error(t, err)
	assert.Regexp(t, "BadReleaseError", state_machine.LastOutput())
	assert.Regexp(t, "older", state_machine.LastOutput())
	assertNoLock(t, awsc, release)

	assert.Equal(t, []string{
		"Validate",
		machine.TaskFnName("Validate"),
		"FailureClean",
	}, state_machine.ExecutionPath())
}

func Test_DeployHandler_Execution_Errors_LockError(t *testing.T) {
	release := MockRelease()
	awsc := MockAwsClients(release)

	state_machine := createTestStateMachine(t, awsc)

	awsc.S3.AddPutObject(*release.LockPath(), fmt.Errorf("PuttyError"))

	_, err := state_machine.ExecuteToMap(release)

	assert.Error(t, err)
	assert.Regexp(t, "LockError", state_machine.LastOutput())
	assert.Regexp(t, "PuttyError", state_machine.LastOutput())

	assertNoLock(t, awsc, release)

	assert.Equal(t, []string{
		"Validate",
		machine.TaskFnName("Validate"),
		"Lock",
		machine.TaskFnName("Lock"),
		"ReleaseLockFailure",
		machine.TaskFnName("ReleaseLockFailure"),
		"FailureClean",
	}, state_machine.ExecutionPath())
}

func Test_DeployHandler_Execution_Errors_LockExistsError(t *testing.T) {
	release := MockRelease()
	awsc := MockAwsClients(release)

	state_machine := createTestStateMachine(t, awsc)

	awsc.S3.AddGetObject(*release.LockPath(), `{"uuid":"notuuid"}`, nil)

	_, err := state_machine.ExecuteToMap(release)

	assert.Error(t, err)
	assert.Regexp(t, "LockExistsError", state_machine.LastOutput())
	assert.Regexp(t, "Lock Already Exists", state_machine.LastOutput())

	assert.Equal(t, []string{
		"Validate",
		machine.TaskFnName("Validate"),
		"Lock",
		machine.TaskFnName("Lock"),
		"FailureClean",
	}, state_machine.ExecutionPath())
}

// Bad Resource Errors
func Test_DeployHandler_Execution_Errors_WrongLambdaTags(t *testing.T) {
	release := MockRelease()
	awsc := MockAwsClients(release)
	awsc.Lambda.ListTagsResp = &lambda.ListTagsOutput{
		Tags: map[string]*string{"ProjectName": release.ProjectName, "ConfigName": release.ConfigName, "DeployWith": to.Strp("wrong_tag")},
	}

	state_machine := createTestStateMachine(t, awsc)

	_, err := state_machine.ExecuteToMap(release)
	assert.Error(t, err)
	assert.Regexp(t, "BadReleaseError", state_machine.LastOutput())
	assert.Regexp(t, "DeployWith", state_machine.LastOutput())
	assertNoLock(t, awsc, release)

	assert.Equal(t, []string{
		"Validate",
		machine.TaskFnName("Validate"),
		"Lock",
		machine.TaskFnName("Lock"),
		"ValidateResources",
		machine.TaskFnName("ValidateResources"),
		"ReleaseLockFailure",
		machine.TaskFnName("ReleaseLockFailure"),
		"FailureClean",
	}, state_machine.ExecutionPath())
}

func Test_DeployHandler_Execution_Errors_WrongSFNPath(t *testing.T) {
	release := MockRelease()
	awsc := MockAwsClients(release)
	awsc.SFN.DescribeStateMachineResp = &sfn.DescribeStateMachineOutput{
		RoleArn: to.Strp("arn:aws:iam::000000000000:role/step/wrongproject/config/role-name"),
	}

	state_machine := createTestStateMachine(t, awsc)

	_, err := state_machine.ExecuteToMap(release)
	assert.Error(t, err)
	assert.Regexp(t, "BadReleaseError", state_machine.LastOutput())
	assert.Regexp(t, "Role Path", state_machine.LastOutput())
	assertNoLock(t, awsc, release)

	assert.Equal(t, []string{
		"Validate",
		machine.TaskFnName("Validate"),
		"Lock",
		machine.TaskFnName("Lock"),
		"ValidateResources",
		machine.TaskFnName("ValidateResources"),
		"ReleaseLockFailure",
		machine.TaskFnName("ReleaseLockFailure"),
		"FailureClean",
	}, state_machine.ExecutionPath())
}

func Test_DeployHandler_Execution_Errors_BadLambdaSHA(t *testing.T) {
	release := MockRelease()
	release.LambdaSHA256 = to.Strp("wrongsha")

	awsc := MockAwsClients(release)

	state_machine := createTestStateMachine(t, awsc)

	_, err := state_machine.ExecuteToMap(release)
	assert.Error(t, err)
	assert.Regexp(t, "BadReleaseError", state_machine.LastOutput())
	assert.Regexp(t, "Lambda SHA", state_machine.LastOutput())
	assertNoLock(t, awsc, release)

	assert.Equal(t, []string{
		"Validate",
		machine.TaskFnName("Validate"),
		"FailureClean",
	}, state_machine.ExecutionPath())
}

func Test_DeployHandler_Execution_Errors_BadReleasePath(t *testing.T) {
	release := MockRelease()
	release.AwsAccountID = to.Strp("0000000")
	awsc := MockAwsClients(release)

	awsc.S3.AddGetObject(*release.ReleasePath(), "bad_release", nil)
	state_machine := createTestStateMachine(t, awsc)

	_, err := state_machine.ExecuteToMap(release)
	assert.Error(t, err)
	assert.Regexp(t, "BadReleaseError", state_machine.LastOutput())
	assert.Regexp(t, "uploaded Release struct", state_machine.LastOutput())
	assertNoLock(t, awsc, release)

	assert.Equal(t, []string{
		"Validate",
		machine.TaskFnName("Validate"),
		"FailureClean",
	}, state_machine.ExecutionPath())
}

func Test_DeployHandler_Execution_Errors_WrongReleasePath(t *testing.T) {
	release := MockRelease()
	release.AwsAccountID = to.Strp("0000000")
	awsc := MockAwsClients(release)

	awsc.S3.AddGetObject(*release.ReleasePath(), "{}", nil)
	state_machine := createTestStateMachine(t, awsc)

	_, err := state_machine.ExecuteToMap(release)
	assert.Error(t, err)
	assert.Regexp(t, "BadReleaseError", state_machine.LastOutput())
	assert.Regexp(t, "Release SHA", state_machine.LastOutput())
	assertNoLock(t, awsc, release)

	assert.Equal(t, []string{
		"Validate",
		machine.TaskFnName("Validate"),
		"FailureClean",
	}, state_machine.ExecutionPath())
}

func Test_DeployHandler_Execution_Errors_DifferentReleaseSHA(t *testing.T) {
	release := MockRelease()
	awsc := MockAwsClients(release)

	// Change the release
	release.CreatedAt = to.Timep(time.Now())
	state_machine := createTestStateMachine(t, awsc)

	_, err := state_machine.ExecuteToMap(release)
	assert.Error(t, err)
	assert.Regexp(t, "BadReleaseError", state_machine.LastOutput())
	assert.Regexp(t, "Release SHA", state_machine.LastOutput())
	assertNoLock(t, awsc, release)

	assert.Equal(t, []string{
		"Validate",
		machine.TaskFnName("Validate"),
		"FailureClean",
	}, state_machine.ExecutionPath())
}

// Upload Errors
func Test_DeployHandler_Execution_Errors_DeploySFNError(t *testing.T) {
	release := MockRelease()
	awsc := MockAwsClients(release)

	awsc.SFN.UpdateStateMachineError = fmt.Errorf("AWSSFNError")

	state_machine := createTestStateMachine(t, awsc)

	_, err := state_machine.ExecuteToMap(release)

	assert.Error(t, err)
	assert.Regexp(t, "DeploySFNError", state_machine.LastOutput())
	assert.Regexp(t, "AWSSFNError", state_machine.LastOutput())

	assert.Equal(t, []string{
		"Validate",
		machine.TaskFnName("Validate"),
		"Lock",
		machine.TaskFnName("Lock"),
		"ValidateResources",
		machine.TaskFnName("ValidateResources"),
		"Deploy",
		machine.TaskFnName("Deploy"),
		"ReleaseLockFailure",
		machine.TaskFnName("ReleaseLockFailure"),
		"FailureClean",
	}, state_machine.ExecutionPath())
}

func Test_DeployHandler_Execution_Errors_DeployLambdaError(t *testing.T) {
	release := MockRelease()
	awsc := MockAwsClients(release)
	awsc.Lambda.UpdateFunctionCodeError = fmt.Errorf("AWSLambdaError")

	state_machine := createTestStateMachine(t, awsc)

	_, err := state_machine.ExecuteToMap(release)

	assert.Error(t, err)
	assert.Regexp(t, "DeployLambdaError", state_machine.LastOutput())
	assert.Regexp(t, "AWSLambdaError", state_machine.LastOutput())

	assert.Equal(t, []string{
		"Validate",
		machine.TaskFnName("Validate"),
		"Lock",
		machine.TaskFnName("Lock"),
		"ValidateResources",
		machine.TaskFnName("ValidateResources"),
		"Deploy",
		machine.TaskFnName("Deploy"),
		"FailureDirty",
	}, state_machine.ExecutionPath())
}
