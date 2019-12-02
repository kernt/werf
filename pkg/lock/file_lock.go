package lock

import "time"

func NewFileLock(name string, locksDir string) LockObject {
	return &File{Base: Base{Name: name}, LocksDir: locksDir}
}

type File struct {
	Base
	LocksDir string
	locker   *fileLocker
}

func (lock *File) newLocker(timeout time.Duration, readOnly bool, onWait func(doWait func() error) error) *fileLocker {
	return &fileLocker{
		baseLocker: baseLocker{
			Timeout:  timeout,
			ReadOnly: readOnly,
			OnWait:   onWait,
		},
		FileLock: lock,
	}
}

func (lock *File) Lock(timeout time.Duration, readOnly bool, onWait func(doWait func() error) error) error {
	lock.locker = lock.newLocker(timeout, readOnly, onWait)
	return lock.Base.Lock(lock.locker)
}

func (lock *File) Unlock() error {
	if lock.locker == nil {
		return nil
	}

	err := lock.Base.Unlock(lock.locker)
	if err != nil {
		return err
	}

	lock.locker = nil

	return nil
}

func (lock *File) WithLock(timeout time.Duration, readOnly bool, onWait func(doWait func() error) error, f func() error) error {
	lock.locker = lock.newLocker(timeout, readOnly, onWait)

	err := lock.Base.WithLock(lock.locker, f)
	if err != nil {
		return err
	}

	lock.locker = nil

	return nil
}
