// to is a list of Functions use to convert things to things
package to

import (
	"fmt"
	"math/rand"
	"os"
	"reflect"
	"strings"
	"time"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandomString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func StrSlice(strps []*string) []string {
	strs := []string{}
	for _, sp := range strps {
		str := ""
		if sp != nil {
			str = *sp
		}
		strs = append(strs, str)
	}
	return strs
}

// Take from aws-lambda-go.Function#lambdaErrorResponse
func ErrorType(invokeError error) string {
	var errorName string
	if errorType := reflect.TypeOf(invokeError); errorType.Kind() == reflect.Ptr {
		errorName = errorType.Elem().Name()
	} else {
		errorName = errorType.Name()
	}
	return errorName
}

func RegionAccount() (*string, *string) {
	region := os.Getenv("AWS_REGION")
	account_id := os.Getenv("AWS_ACCOUNT_ID")

	if region == "" || account_id == "" {
		return nil, nil
	}

	return &region, &account_id
}

func RegionAccountOrExit() (*string, *string) {
	region, account_id := RegionAccount()

	if region == nil || account_id == nil {
		fmt.Println("AWS_REGION or AWS_ACCOUNT_ID not defined")
		os.Exit(1)
	}

	return region, account_id
}

// TimeUUID returns time base UUID with prefix
func TimeUUID(prefix string) *string {
	tf := strings.Replace(time.Now().UTC().Format(time.RFC3339), ":", "-", -1)
	rs := RandomString(7)
	rid := fmt.Sprintf("%v%v-%v", prefix, tf, rs)
	return &rid
}
