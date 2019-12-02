// +build !windows

package lock

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/flant/werf/pkg/util"
)

type fileLocker struct {
	baseLocker

	FileLock        *File
	openFileHandler *os.File
}

func (locker *fileLocker) lockFilePath() string {
	fileName := util.MurmurHash(locker.FileLock.GetName())
	return filepath.Join(locker.FileLock.LocksDir, fileName)
}

func (locker *fileLocker) Lock() error {
	f, err := os.OpenFile(locker.lockFilePath(), os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}

	locker.openFileHandler = f

	fd := int(locker.openFileHandler.Fd())
	var mode int
	if locker.ReadOnly {
		mode = syscall.LOCK_SH
	} else {
		mode = syscall.LOCK_EX
	}

	err = syscall.Flock(fd, mode|syscall.LOCK_NB)

	if err == syscall.EWOULDBLOCK {
		return locker.OnWait(func() error {
			return locker.pollFlock(fd, mode)
		})
	}

	return err
}

func (locker *fileLocker) pollFlock(fd int, mode int) error {
	flockRes := make(chan error)
	cancelPoll := make(chan bool)

	go func() {
		ticker := time.NewTicker(time.Millisecond * 500)

	PollFlock:
		for {
			select {
			case <-ticker.C:
				err := syscall.Flock(fd, mode|syscall.LOCK_NB)
				if err == nil || err != syscall.EWOULDBLOCK {
					flockRes <- err
				}
			case <-cancelPoll:
				break PollFlock
			}
		}
	}()

	select {
	case err := <-flockRes:
		return err
	case <-time.After(locker.Timeout):
		cancelPoll <- true
		return fmt.Errorf("lock `%s` timeout %s expired", locker.FileLock.GetName(), locker.Timeout)
	}
}

func (locker *fileLocker) Unlock() error {
	err := locker.openFileHandler.Close()
	if err != nil {
		return err
	}

	locker.openFileHandler = nil

	return nil
}
