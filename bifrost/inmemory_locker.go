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
	mu    sync.RWMutex
	locks map[string][]*Lock
}

func NewInMemoryLocker() *InMemoryLocker {
	return &InMemoryLocker{
		locks: make(map[string][]*Lock),
	}
}

func (l *InMemoryLocker) GrabLock(namespace string, lockPath string, uuid string, reason string) (bool, error) {
	existingLock := l.GetLockByPath(namespace, lockPath)
	if existingLock != nil {
		return existingLock.uuid == uuid, nil
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	l.locks[namespace] = append(l.locks[namespace], &Lock{
		lockPath: lockPath,
		uuid:     uuid,
		reason:   reason,
	})

	return true, nil
}

func (l *InMemoryLocker) ReleaseLock(namespace string, lockPath string, uuid string) error {
	existingLock := l.GetLockByPath(namespace, lockPath)
	if existingLock != nil && existingLock.uuid != uuid {
		return fmt.Errorf("failed to release lock: %s is currently held by UUID(%v)", lockPath, existingLock.uuid)
	}

	l.mu.Lock()
	defer l.mu.Unlock()

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
	l.mu.RLock()
	defer l.mu.RUnlock()

	locks, found := l.locks[namespace]
	if !found {
		return []*Lock{}
	}

	return locks
}

func (l *InMemoryLocker) GetLockByPath(namespace string, lockPath string) *Lock {
	l.mu.RLock()
	defer l.mu.RUnlock()

	for _, lock := range l.GetLockByNamespace(namespace) {
		if lock.lockPath == lockPath {
			return lock
		}
	}

	return nil
}
