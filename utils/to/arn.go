package to

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/aws/aws-sdk-go/aws/arn"
)

// LambdaArn takes a name OR arn and returns Arn defaulting to AWS Environment variables
func LambdaArn(region *string, account_id *string, name_or_arn *string) *string {
	return createArn("arn:aws:lambda:%v:%v:function:%v", region, account_id, name_or_arn)
}

// StepArn takes a name OR arn and returns Arn defaulting to AWS Environment variables
func StepArn(region *string, account_id *string, name_or_arn *string) *string {
	return createArn("arn:aws:states:%v:%v:stateMachine:%v", region, account_id, name_or_arn)
}

func RoleArn(account_id *string, name_or_arn *string) *string {
	return createArn("arn:aws:iam::%v%v:role/%v", account_id, Strp(""), name_or_arn)
}

// InterpolateArnVariables replaces any resource parameter templates with the appropriate values
func InterpolateArnVariables(state_machine *string, region *string, account_id *string, name_or_arn *string) *string {
	variableTemplate := map[string]*string{
		"{{aws_account}}": account_id,
		"{{aws_region}}":  region,
		"{{lambda_name}}": name_or_arn,
	}
	for k, v := range variableTemplate {
		*state_machine = strings.Replace(*state_machine, k, *v, -1)
	}
	return state_machine
}

func ArnPath(arn string) string {
	_, _, res := ArnRegionAccountResource(arn)

	path := strings.Split(res, "/")

	switch len(path) {
	case 0:
		return "/"
	case 1:
		return "/"
	case 2:
		return "/"
	default:
		return fmt.Sprintf("/%v/", strings.Join(path[1:len(path)-1], "/"))
	}
}

func LambdaArnFromContext(ctx context.Context) (string, error) {
	lc, ok := lambdacontext.FromContext(ctx)
	if !ok || lc == nil {
		return "", fmt.Errorf("Incorrect Lambda Context")
	}

	return lc.InvokedFunctionArn, nil
}

func AwsRegionAccountFromContext(ctx context.Context) (*string, *string) {
	arn, err := LambdaArnFromContext(ctx)
	if err != nil {
		return nil, nil
	}

	region, account, _ := ArnRegionAccountResource(arn)
	return &region, &account
}

func AwsRegionAccountLambdaNameFromContext(ctx context.Context) (region, account, lambdaName string) {
	arn, err := LambdaArnFromContext(ctx)
	if err != nil {
		return "", "", ""
	}

	region, account, resource := ArnRegionAccountResource(arn)
	// function:<lambda name>
	resourceParts := strings.SplitN(strings.ToLower(resource), ":", 2)
	if len(resourceParts) < 2 {
		return region, account, ""
	}

	return region, account, resourceParts[1]
}

func ArnRegionAccountResource(arnstr string) (string, string, string) {
	a, err := arn.Parse(arnstr)
	if err != nil {
		return "", "", ""
	}
	return a.Region, a.AccountID, a.Resource
}

func createArn(arn_str string, region *string, account_id *string, name_or_arn *string) *string {
	if len(*name_or_arn) < 5 || (*name_or_arn)[:4] == "arn:" {
		return name_or_arn
	}

	if region == nil || account_id == nil || name_or_arn == nil {
		return name_or_arn
	}

	arn := fmt.Sprintf(arn_str, *region, *account_id, *name_or_arn)
	return &arn
}
