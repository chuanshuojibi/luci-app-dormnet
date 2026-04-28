package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/openwrt-dormnet/dormnet/shared/errx"
)

type PidFile interface {
	Exist() bool
	Pid() (int, errx.Exception)
	Create() errx.Exception
	Remove() errx.Exception
	Lock() errx.Exception
	Unlock() errx.Exception
	IsSelf() (bool, errx.Exception)
	KillSelf() errx.Exception
	Close() errx.Exception
}

type pidFile struct {
	path string
	file *os.File
}

func NewPidFile(path string) PidFile {
	return &pidFile{
		path: path,
	}
}

func (p *pidFile) tryOpen(createWhenNotExist bool) errx.Exception {
	if p.file == nil {
		var err error
		flag := os.O_RDWR
		if createWhenNotExist {
			flag |= os.O_CREATE
		}
		p.file, err = os.OpenFile(p.path, flag, 0666)
		if err != nil {
			if os.IsNotExist(err) {
				return errx.NewExceptionWithError(err, "pid file is not exist")
			}
			return errx.NewExceptionWithError(err, "failed to open pid file")
		}
	}
	return nil
}

func (p *pidFile) Pid() (int, errx.Exception) {
	if err := p.tryOpen(false); err != nil {
		return 0, err
	}
	content, err := os.ReadFile(p.path)
	if err != nil {
		return 0, errx.NewExceptionWithError(err, "failed to read pid file")
	}

	pid, err := strconv.Atoi(string(content))
	if err != nil {
		return 0, errx.NewExceptionWithError(err, "failed to parse pid file")
	}
	return pid, nil
}

func (p *pidFile) Exist() bool {
	_, err := os.Stat(p.path)
	return err == nil
}

func (p *pidFile) Create() errx.Exception {
	pid := os.Getpid()
	err := os.WriteFile(p.path, []byte(strconv.Itoa(pid)), 0644)
	if err != nil {
		return errx.NewExceptionWithError(err, "failed to create pid file")
	}
	return nil
}

func (p *pidFile) Remove() errx.Exception {
	err := os.Remove(p.path)
	if err == nil || os.IsNotExist(err) {
		return nil
	}
	return errx.NewExceptionWithError(err, "failed to remove pid file")
}

func (p *pidFile) Lock() errx.Exception {
	if err := p.tryOpen(true); err != nil {
		return err
	}
	if err := syscall.Flock(int(p.file.Fd()), syscall.LOCK_EX); err != nil {
		return errx.NewExceptionWithError(err, "failed to lock pid file")
	}
	return nil
}

func (p *pidFile) Unlock() errx.Exception {
	if !p.Exist() {
		return nil
	}
	if err := p.tryOpen(true); err != nil {
		return err
	}
	if err := syscall.Flock(int(p.file.Fd()), syscall.LOCK_UN|syscall.LOCK_NB); err != nil {
		return errx.NewExceptionWithError(err, "failed to unlock pid file")
	}
	return nil
}

func (p *pidFile) IsSelf() (bool, errx.Exception) {
	if !p.Exist() {
		return false, nil
	}

	pid, errX := p.Pid()
	if errX != nil {
		return false, errx.NewExceptionWithCause(errX, "failed to read pid file")
	}

	if pid == os.Getpid() {
		return true, nil
	}

	if err := syscall.Kill(pid, 0); err != nil {
		return false, nil
	}

	exe, err := os.Readlink(fmt.Sprintf("/proc/%d/exe", pid))
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, errx.NewExceptionWithError(err, "failed to read pid state")
	}

	self, err := os.Executable()
	if err != nil {
		return false, errx.NewExceptionWithError(err, "failed to get self executable")
	}

	exe, _ = filepath.EvalSymlinks(exe)
	self, _ = filepath.EvalSymlinks(self)

	return exe == self, nil
}

func (p *pidFile) KillSelf() errx.Exception {
	pid, errX := p.Pid()
	if errX != nil {
		return errx.NewExceptionWithCause(errX, "failed to read pid file")
	}

	isSelf, err := p.IsSelf()
	if err != nil {
		return errx.NewExceptionWithCause(errX, "failed to check whether pid is self")
	}
	if !isSelf {
		return errx.NewException("not allow to kill other process")
	}

	if err := syscall.Kill(pid, 2); err != nil {
		return errx.NewExceptionWithError(err, "failed to call kill -2")
	}

	_ = p.Remove()

	return nil
}

func (p *pidFile) Close() errx.Exception {
	if p.file != nil {
		_ = p.Unlock()
		_ = p.file.Close()
	}
	return nil
}
