// errors has a list of common errors and error functions
package errors

import (
	"fmt"
)

//
// General Errors that represent levels of action to be taken
//

type AlertError struct {
	Cause string
}

func (e AlertError) Error() string {
	return fmt.Sprintf("AlertError: %v", e.Cause)
}

type NotifyError struct {
	Cause string
}

func (e NotifyError) Error() string {
	return fmt.Sprintf("NotifyError: %v", e.Cause)
}

type LogError struct {
	Cause string
}

func (e LogError) Error() string {
	return fmt.Sprintf("LogError: %v", e.Cause)
}

//
// Specific Deploy/Release errors
//

type BadReleaseError struct {
	Cause string
}

func (e BadReleaseError) Error() string {
	return fmt.Sprintf("BadReleaseError: %v", e.Cause)
}

type LockExistsError struct {
	Cause string
}

func (e LockExistsError) Error() string {
	return fmt.Sprintf("LockExistsError: %v", e.Cause)
}

type LockError struct {
	Cause string
}

func (e LockError) Error() string {
	return fmt.Sprintf("LockError: %v", e.Cause)
}
