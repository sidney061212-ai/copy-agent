//go:build !windows

package service

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

type InstanceLock struct {
	file     *os.File
	path     string
	acquired bool
}

func AcquireInstanceLock(configPath string) (*InstanceLock, error) {
	configDir := filepath.Dir(configPath)
	configBase := filepath.Base(configPath)
	lockName := fmt.Sprintf(".%s.lock", configBase)
	lockPath := filepath.Join(configDir, lockName)

	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return nil, fmt.Errorf("cannot create config directory: %w", err)
	}

	file, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, fmt.Errorf("cannot open lock file: %w", err)
	}

	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		_ = file.Close()
		pid := readPIDFromLockFile(lockPath)
		if pid > 0 {
			return nil, fmt.Errorf("another copyagentd instance is already running (PID %d) with config %s", pid, configPath)
		}
		return nil, fmt.Errorf("another copyagentd instance is already running with config %s", configPath)
	}

	pid := os.Getpid()
	_ = file.Truncate(0)
	_, _ = file.Seek(0, 0)
	_, _ = fmt.Fprintf(file, "%d\n", pid)

	return &InstanceLock{
		file:     file,
		path:     lockPath,
		acquired: true,
	}, nil
}

func (lock *InstanceLock) Release() {
	if lock == nil || !lock.acquired {
		return
	}
	if lock.file != nil {
		_ = lock.file.Truncate(0)
		_ = syscall.Flock(int(lock.file.Fd()), syscall.LOCK_UN)
		_ = lock.file.Close()
		lock.file = nil
	}
	lock.acquired = false
}

func (lock *InstanceLock) Path() string {
	if lock == nil {
		return ""
	}
	return lock.path
}

func readPIDFromLockFile(path string) int {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	var pid int
	if _, err := fmt.Sscanf(string(data), "%d", &pid); err != nil {
		return 0
	}
	return pid
}

func KillExistingInstance(configPath string) bool {
	configDir := filepath.Dir(configPath)
	configBase := filepath.Base(configPath)
	lockName := fmt.Sprintf(".%s.lock", configBase)
	lockPath := filepath.Join(configDir, lockName)

	pid := readPIDFromLockFile(lockPath)
	if pid <= 0 {
		return false
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	if err := process.Signal(syscall.Signal(0)); err != nil {
		return false
	}
	return process.Kill() == nil
}
