package bifrost

import (
	"fmt"
	"sync"
)

type Lock struct {
	lockPath string
	uuid     string
	reason   string
}

type InMemoryLocker struct {
	locks map[string][]*Lock

	mx sync.Mutex
}

func NewInMemoryLocker() *InMemoryLocker {
	return &InMemoryLocker{}
}

func (l *InMemoryLocker) init() {
	if l.locks == nil {
		l.locks = map[string][]*Lock{}
	}
}

func (l *InMemoryLocker) GrabLock(namespace string, lockPath string, uuid string, reason string) (bool, error) {
	l.mx.Lock()
	defer l.mx.Unlock()

	l.init()

	existingLock := l.GetLockByPath(namespace, lockPath)
	if existingLock != nil {
		if existingLock.uuid == uuid {
			return true, nil
		} else {
			return false, nil
		}
	}

	l.locks[namespace] = append(l.locks[namespace], &Lock{
		lockPath: lockPath,
		uuid:     uuid,
		reason:   reason,
	})

	return true, nil
}

func (l *InMemoryLocker) ReleaseLock(namespace string, lockPath string, uuid string) error {
	l.mx.Lock()
	defer l.mx.Unlock()

	l.init()

	existingLock := l.GetLockByPath(namespace, lockPath)
	if existingLock != nil && existingLock.uuid != uuid {
		return fmt.Errorf("failed to release lock: %s is currently held by UUID(%v)", lockPath, existingLock.uuid)
	}

	var updatedLocks []*Lock
	for _, lock := range l.locks[namespace] {
		if lock.uuid == uuid {
			continue
		}
		updatedLocks = append(updatedLocks, lock)
	}

	l.locks[namespace] = updatedLocks

	return nil
}

func (l *InMemoryLocker) GetLockByNamespace(namespace string) []*Lock {
	locks, found := l.locks[namespace]
	if !found {
		return []*Lock{}
	}

	return locks
}

func (l *InMemoryLocker) GetLockByPath(namespace string, lockPath string) *Lock {
	for _, lock := range l.GetLockByNamespace(namespace) {
		if lock.lockPath == lockPath {
			return lock
		}
	}

	return nil
}
