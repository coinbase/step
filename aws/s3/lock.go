package s3

import (
	"fmt"

	"github.com/coinbase/step/aws"
)

type Lock struct {
	UUID string `json:"uuid,omitempty"`
}

type UserLock struct {
	User       string `json:"user,omitempty"`
	LockReason string `json:"lock_reason", omitempty"`
}

func CheckUserLock(s3c aws.S3API, bucket *string, lock_path *string) error {
	var userLock UserLock
	err := GetStruct(s3c, bucket, lock_path, &userLock)
	if err != nil {
		switch err.(type) {
		case *NotFoundError:
			// good we want this
			return nil
		default:
			return err // All other errors return
		}
	}
	return fmt.Errorf("Deploys locked by %v for reason: %v", userLock.User, userLock.LockReason)
}

// GrabLock creates a lock file in S3 with a UUID
// it returns a grabbed bool, and error
// if the Lock already exists and UUID is equal to the existing lock it will returns true, otherwise false
// if the Lock doesn't exist it will create the file and return true
func GrabLock(s3c aws.S3API, bucket *string, lock_path *string, uuid string) (bool, error) {
	lock := &Lock{uuid}
	var s3_lock Lock

	err := GetStruct(s3c, bucket, lock_path, &s3_lock)
	if err != nil {
		switch err.(type) {
		case *NotFoundError:
			// good we want this
		default:
			return false, err // All other errors return
		}
	}

	// If s3_lock unmarshalled and the UUID
	if s3_lock.UUID != "" {
		// if UUID is the same
		if s3_lock.UUID == lock.UUID {
			// Already have the lock (caused by a retry ... maybe)
			return true, nil
		} else {
			return false, nil
		}
	}

	// After this point we might have created the lock so return true
	// Create the Lock
	err = PutStruct(s3c, bucket, lock_path, lock)

	if err != nil {
		return true, err
	}

	return true, nil
}

// ReleaseLock removes the lock file for UUID
// If the lock file exists and is not the same UUID it returns an error
func ReleaseLock(s3c aws.S3API, bucket *string, lock_path *string, uuid string) error {
	var s3_lock Lock

	err := GetStruct(s3c, bucket, lock_path, &s3_lock)
	if err != nil {
		switch err.(type) {
		case *NotFoundError:
			// No lock to release
			return nil
		default:
			return err // All other errors return
		}
	}

	// if s3_lock unmarshalled and the UUID is different then error
	if s3_lock.UUID != "" && s3_lock.UUID != uuid {
		return fmt.Errorf("Release with UUID(%v) is trying to unlock UUID(%v)", uuid, s3_lock.UUID)
	}

	return Delete(s3c, bucket, lock_path)
}
