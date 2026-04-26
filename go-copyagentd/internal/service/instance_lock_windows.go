//go:build windows

package service

import "fmt"

type InstanceLock struct{}

func AcquireInstanceLock(string) (*InstanceLock, error) {
	return &InstanceLock{}, nil
}

func (lock *InstanceLock) Release() {}

func (lock *InstanceLock) Path() string {
	return ""
}

func KillExistingInstance(string) bool {
	return false
}

func readPIDFromLockFile(string) int {
	return 0
}

var _ = fmt.Sprintf
