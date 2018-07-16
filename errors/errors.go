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
// Low Level Step Errors
//

type UnmarshalError struct {
	Cause string
}

func (e UnmarshalError) Error() string {
	return fmt.Sprintf("UnmarshalError: %v", e.Cause)
}

type PanicError struct {
	Cause string
}

func (e PanicError) Error() string {
	return fmt.Sprintf("PanicError: %v", e.Cause)
}

//
// Specific Deploy/Release errors
//

// BadReleaseError error
type BadReleaseError struct {
	Cause string
}

func (e BadReleaseError) Error() string {
	return fmt.Sprintf("BadReleaseError: %v", e.Cause)
}

// LockExistsError error
type LockExistsError struct {
	Cause string
}

func (e LockExistsError) Error() string {
	return fmt.Sprintf("LockExistsError: %v", e.Cause)
}

// LockError error
type LockError struct {
	Cause string
}

func (e LockError) Error() string {
	return fmt.Sprintf("LockError: %v", e.Cause)
}

// DeployError error
type DeployError struct {
	Cause string
}

func (e DeployError) Error() string {
	return fmt.Sprintf("DeployError: %v", e.Cause)
}

// HealthError error
type HealthError struct {
	Cause string
}

func (e HealthError) Error() string {
	return fmt.Sprintf("HealthError: %v", e.Cause)
}

// HaltError error
type HaltError struct {
	Cause string
}

func (e HaltError) Error() string {
	return fmt.Sprintf("HaltError: %v", e.Cause)
}

// CleanUpError error
type CleanUpError struct {
	Cause string
}

func (e CleanUpError) Error() string {
	return fmt.Sprintf("CleanUpError: %v", e.Cause)
}

func throw(err error) error {
	fmt.Printf(err.Error())
	return err
}
